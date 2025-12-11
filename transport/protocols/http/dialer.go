package http

import (
	"context"
	gotls "crypto/tls"
	"fmt"
	gonet "net"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/http2"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"
)

var (
	globalDialerMap    map[net.Destination]*http.Client
	globalDialerAccess sync.Mutex
)

type dialerCanceller func()

func getHTTPClient(ctx context.Context, dest net.Destination, securityEngine security.Engine, sc i.Dialer) (*http.Client, dialerCanceller) {
	globalDialerAccess.Lock()
	defer globalDialerAccess.Unlock()

	canceller := func() {
		globalDialerAccess.Lock()
		defer globalDialerAccess.Unlock()
		delete(globalDialerMap, dest)
	}

	if globalDialerMap == nil {
		globalDialerMap = make(map[net.Destination]*http.Client)
	}

	if client, found := globalDialerMap[dest]; found {
		return client, canceller
	}

	transport := &http2.Transport{
		DialTLSContext: func(_ context.Context, network, addr string, tlsConfig *gotls.Config) (gonet.Conn, error) {
			rawHost, rawPort, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			if len(rawPort) == 0 {
				rawPort = "443"
			}
			port, err := net.PortFromString(rawPort)
			if err != nil {
				return nil, err
			}
			address := net.ParseAddress(rawHost)

			pconn, err := sc.Dial(ctx, net.TCPDestination(address, port))
			if err != nil {
				return nil, err
			}

			cn, err := (securityEngine).GetClientConn(pconn,
				security.OptionWithDestination{Dest: dest})
			if err != nil {
				return nil, err
			}

			protocol := ""
			if connAPLNGetter, ok := cn.(security.ConnectionApplicationProtocol); ok {
				connectionALPN, err := connAPLNGetter.GetConnectionApplicationProtocol()
				if err != nil {
					return nil, fmt.Errorf("failed to get ALPN from connection: %w", err)
				}
				protocol = connectionALPN
			}

			if protocol != http2.NextProtoTLS {
				return nil, fmt.Errorf("http2: unexpected ALPN protocol %s; want %s", protocol, http2.NextProtoTLS)
			}
			return cn, nil
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	globalDialerMap[dest] = client
	return client, canceller
}

type Dialer struct {
	Config   *HttpConfig
	Security security.Engine
	Socket   i.Dialer
}

func NewDialer(config *HttpConfig, securityEngine security.Engine, socket i.Dialer) *Dialer {
	return &Dialer{
		Config:   config,
		Security: securityEngine,
		Socket:   socket,
	}
}

func (d *Dialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	return Dial(ctx, dest, d.Config, d.Security, d.Socket)
}

// Dial dials a new TCP connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, c *HttpConfig, se security.Engine, sc i.Dialer) (net.Conn, error) {
	if se == nil {
		return nil, fmt.Errorf("http: security engine is not set")
	}
	client, canceller := getHTTPClient(ctx, dest, se, sc)

	p := pipe.NewPipe(buf.BufferSize, false)
	breader := &buf.BufferedReader{Reader: p}

	httpMethod := "PUT"
	if c.Method != "" {
		httpMethod = c.Method
	}

	httpHeaders := make(http.Header)

	for _, httpHeader := range c.Header {
		for _, httpHeaderValue := range httpHeader.Value {
			httpHeaders.Set(httpHeader.Name, httpHeaderValue)
		}
	}

	request := &http.Request{
		Method: httpMethod,
		Host:   c.getRandomHost(),
		Body:   breader,
		URL: &url.URL{
			Scheme: "https",
			Host:   dest.NetAddr(),
			Path:   c.getNormalizedPath(),
		},
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		ProtoMinor: 0,
		Header:     httpHeaders,
	}
	// Disable any compression method from server.
	request.Header.Set("Accept-Encoding", "identity")

	response, err := client.Do(request) // nolint: bodyclose
	if err != nil {
		canceller()
		return nil, fmt.Errorf("http: failed to dial to %s: %w", dest.NetAddr(), err)
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	bwriter := buf.NewBufferedWriter(p)
	common.Must(bwriter.SetBuffered(false))
	return net.NewConnection(
		net.ConnectionOutput(response.Body),
		net.ConnectionInput(bwriter),
		net.ConnectionOnClose(common.ChainedClosable{breader, bwriter, response.Body}),
	), nil
}
