package dns

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// DoHNameServer implemented DNS over HTTPS (RFC8484) Wire Format,
// which is compatible with traditional dns over udp(RFC1035),
// thus most of the DOH implementation is copied from udpns.go
type DoHNameServer struct {
	sync.RWMutex
	cache      *rrCache
	httpClient *http.Client
	dohURL     string
	name       string
	clientIp   net.IP

	ipToDomain *IPToDomain
}

type DoHNameServerOption struct {
	ClientIP   net.IP
	Handler    i.FlowHandler
	Name       string
	Url        string
	IpToDomain *IPToDomain
	Tls        *tls.Config
}

// NewDoHNameServer creates DOH server object for remote resolving.
func NewDoHNameServer(option DoHNameServerOption) (*DoHNameServer, error) {
	s := &DoHNameServer{
		name:     option.Name,
		dohURL:   option.Url,
		cache:    NewRrCache(),
		clientIp: option.ClientIP,
	}

	// Dispatched connection will be closed (interrupted) after each request
	// This makes DOH inefficient without a keep-alived connection
	// See: core/app/proxyman/outbound/handler.go:113
	// Using mux (https request wrapped in a stream layer) improves the situation.
	// Recommend to use NewDoHLocalNameServer (DOHL:) if v2ray instance is running on
	//  a normal network eg. the server side of v2ray
	tr := &http.Transport{
		MaxIdleConns:        30,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 30 * time.Second,
		ForceAttemptHTTP2:   true,
		TLSClientConfig:     option.Tls,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			logger := log.With().Uint32("sid", uint32(session.NewID())).Logger()
			ctx = logger.WithContext(ctx)

			logger.Debug().Str("addr", addr).Msg("doh dial")

			d, err := net.ParseDestination(addr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse destination: %w", err)
			}
			d.Network = net.Network_TCP
			ctx = inbound.ContextWithInboundTag(ctx, s.name)
			hd := &util.FlowHandlerToDialer{
				FlowHandler: option.Handler}
			return hd.Dial(ctx, d)
		},
	}

	dispatchedClient := &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}

	s.httpClient = dispatchedClient
	return s, nil
}

func (d *DoHNameServer) Name() string {
	return d.name
}

func (d *DoHNameServer) Start() error {
	d.cache.Start()
	return nil
}

func (d *DoHNameServer) Close() error {
	if d.cache != nil {
		d.cache.Close()
	}
	return nil
}

func (s *DoHNameServer) dohHTTPSContext(ctx context.Context, b []byte) ([]byte, error) {
	body := bytes.NewBuffer(b)
	req, err := http.NewRequest("POST", s.dohURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/dns-message")
	req.Header.Add("Content-Type", "application/dns-message")

	resp, err := s.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body) // flush resp.Body so that the conn is reusable
		return nil, fmt.Errorf("DOH server returned code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (d *DoHNameServer) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, errors.New("invalid dns query")
	}
	question := msg.Question[0]
	cachedMsg, ok := d.cache.Get(&question)
	if ok {
		log.Ctx(ctx).Debug().Str("domain", question.Name).
			Str("type", dns.TypeToString[question.Qtype]).
			Any("reply", cachedMsg).Msg("doh cache hit")
		return makeReply(msg, cachedMsg), nil
	}

	// if there is clientIp, set it in EDNS0 Client Subnet option
	if len(d.clientIp) > 0 {
		addClientIP(msg, d.clientIp)
	}

	b := buf.New()
	defer b.Release()
	dnsMsgBytes, err := msg.PackBuffer(b.BytesTo(b.Cap()))
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	resp, err := d.dohHTTPSContext(ctx, dnsMsgBytes)
	if err != nil {
		return nil, err
	}

	rply := new(dns.Msg)
	if err := rply.Unpack(resp); err != nil {
		return nil, fmt.Errorf("failed to unpack DOH response: %w", err)
	}

	log.Ctx(ctx).Debug().Str("domain", msg.Question[0].Name).
		Dur("time", time.Since(startTime)).
		Str("type", dns.TypeToString[msg.Question[0].Qtype]).
		Str("reply", rply.String()).Msg("doh reply")

	if d.ipToDomain != nil {
		d.ipToDomain.SetDomain(msg, net.ParseAddress(d.dohURL))
	}
	d.cache.Set(rply)

	return rply, nil
}
