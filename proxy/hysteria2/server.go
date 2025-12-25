//go:build server

package hysteria2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/common"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"

	"github.com/apernet/hysteria/core/v2/server"
	"github.com/apernet/hysteria/extras/v2/obfs"
	"github.com/rs/zerolog/log"
)

type Inbound struct {
	config *InboundConfig
	server []server.Server

	usersLock sync.RWMutex
	users     map[string]*User // secret to User

	cLock      sync.RWMutex
	srcAddrMap map[netip.AddrPort]*srcAddrInfo

	onUnauthorizedRequest i.UnauthorizedReport
	dialer                i.Dialer
	listener              i.PacketListener
}

type srcAddrInfo struct {
	counter *atomic.Uint64
}

type User struct {
	Uid     string
	Secret  string
	Counter *atomic.Uint64
}

func NewInbound(config *InboundConfig) (*Inbound, error) {
	in := &Inbound{
		config:                config,
		users:                 make(map[string]*User),
		srcAddrMap:            make(map[netip.AddrPort]*srcAddrInfo),
		onUnauthorizedRequest: config.OnUnauthorizedRequest,
		dialer:                config.Dialer,
		listener:              config.Listener,
	}
	return in, nil
}

type InboundConfig struct {
	*proxy.Hysteria2ServerConfig
	Ports                 []uint16
	Tag                   string
	InStats               *monitor.Stats
	Router                i.Router
	OnUnauthorizedRequest i.UnauthorizedReport
	Dialer                i.Dialer
	Listener              i.PacketListener
}

func (in *Inbound) Tag() string {
	return in.config.Tag
}

func (in *Inbound) AddUser(user i.User) {
	in.usersLock.Lock()
	defer in.usersLock.Unlock()
	in.users[user.Secret()] = &User{
		Uid:     user.Uid(),
		Secret:  user.Secret(),
		Counter: user.Counter(),
	}
}

func (in *Inbound) RemoveUser(uid, secret string) {
	in.usersLock.Lock()
	defer in.usersLock.Unlock()
	if secret == "" {
		for _, user := range in.users {
			if user.Uid == uid {
				delete(in.users, user.Secret)
				return
			}
		}
	} else {
		delete(in.users, secret)
	}
}

func (in *Inbound) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	in.onUnauthorizedRequest = f
}

func (in *Inbound) Start() error {
	config := in.config

	tlsConfig, err := config.Hysteria2ServerConfig.TlsConfig.GetTLSConfig()
	if err != nil {
		return err
	}

	var obfuscator obfs.Obfuscator
	if in.config.GetObfs().GetSalamander() != nil {
		obfuscator, err = obfs.NewSalamanderObfuscator(
			[]byte(config.GetObfs().GetSalamander().Password))
		if err != nil {
			return fmt.Errorf("failed to create obfuscator: %w", err)
		}
	}
	for _, p := range config.Ports {
		pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", p))
		if err != nil {
			return err
		}
		pc = &statsPacketConn{
			PacketConn: pc,
			inbound:    in,
		}
		if obfuscator != nil {
			pc = obfs.WrapPacketConn(pc, obfuscator)
		}
		log.Info().Msgf("hysteria2 listen on %d", p)
		hysConfig := &server.Config{
			RequestHook: &RouterToRequestHook{
				Router: in.config.Router,
			},
			TLSConfig: server.TLSConfig{
				Certificates:             tlsConfig.Certificates,
				EncryptedClientHelloKeys: tlsConfig.EncryptedClientHelloKeys,
			},
			QUICConfig: server.QUICConfig{
				InitialStreamReceiveWindow:     uint64(config.GetQuic().GetInitialStreamReceiveWindow()) * 1024 * 1024,
				MaxStreamReceiveWindow:         uint64(config.GetQuic().GetMaxStreamReceiveWindow()) * 1024 * 1024,
				InitialConnectionReceiveWindow: uint64(config.GetQuic().GetInitialConnectionReceiveWindow()) * 1024 * 1024,
				MaxConnectionReceiveWindow:     uint64(config.GetQuic().GetMaxConnectionReceiveWindow()) * 1024 * 1024,
				MaxIdleTimeout:                 time.Duration(config.GetQuic().GetMaxIdleTimeout()) * time.Second,
				DisablePathMTUDiscovery:        config.GetQuic().GetDisablePathMtuDiscovery(),
				MaxIncomingStreams:             int64(config.GetQuic().GetMaxIncomingStreams()),
			},
			Conn: pc,
			Outbound: &outboundAdapter{
				Dialer:         in.dialer,
				PacketListener: in.listener,
			},
			BandwidthConfig: server.BandwidthConfig{
				MaxTx: uint64(config.GetBandwidth().GetMaxTx()),
				MaxRx: uint64(config.GetBandwidth().GetMaxRx()),
			},
			IgnoreClientBandwidth: config.GetIgnoreClientBandwidth(),
			Authenticator:         in,
			TrafficLogger:         in,
			EventLogger:           in,
		}
		s, err := server.NewServer(hysConfig)
		if err != nil {
			return err
		}
		in.server = append(in.server, s)
		go func(s server.Server) {
			err := s.Serve()
			if err != nil {
				log.Error().Msgf("hysteria2 server serve error: %v", err)
			}
		}(s)
	}
	return nil
}

func (in *Inbound) Close() error {
	return common.CloseAll(in.server)
}

func (in *Inbound) Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, id string) {
	// segments := strings.Split(auth, ":")
	// if len(segments) != 2 {
	// 	return false, ""
	// }
	// uid := segments[0]
	// secret := segments[1]
	in.usersLock.RLock()
	defer in.usersLock.RUnlock()
	_, ok = in.users[auth]
	if !ok {
		if in.onUnauthorizedRequest != nil {
			in.onUnauthorizedRequest.ReportUnauthorized(addr.String(), auth)
		}
		return false, ""
	}

	return true, auth
}

func (in *Inbound) LogOnlineState(id string, online bool) {}

// rx is what received from final dest, and the data will be sent back to client.
// tx is the data sent to the final dest, in other words, the data received from the client
// data sent back to client is rx, data received from client is tx
// send(to client) traffic meter happens at the transport layer, so we don't need to count it here
func (in *Inbound) LogTraffic(id string, tx, rx uint64) (ok bool) {
	if tx != 0 {
		in.usersLock.RLock()
		defer in.usersLock.RUnlock()
		user, ok := in.users[id]
		if !ok {
			return false
		}
		user.Counter.Add(tx + rx)

		if in.config.InStats != nil {
			in.config.InStats.Traffic.Add(tx + rx)
		}
	}
	return true
}

type RouterToRequestHook struct {
	i.Router
}

func (r *RouterToRequestHook) Check(isUDP bool, reqAddr string) bool {
	return true
}
func (r *RouterToRequestHook) TCP(stream server.HyStream, reqAddr *string) ([]byte, error) {
	dest, err := mynet.ParseDestination(*reqAddr)
	if err != nil {
		return nil, err
	}
	dest.Network = mynet.Network_TCP
	info := session.Info{
		Target: dest,
	}
	if h, _ := r.Router.PickHandler(context.Background(), &info); h == nil {
		return nil, errors.New("destination not allowed")
	}
	return nil, nil
}
func (r *RouterToRequestHook) UDP(data []byte, reqAddr *string) error {
	dest, err := mynet.ParseDestination(*reqAddr)
	if err != nil {
		return err
	}
	dest.Network = mynet.Network_UDP
	info := session.Info{
		Target: dest,
	}
	if h, _ := r.Router.PickHandler(context.Background(), &info); h == nil {
		return errors.New("destination not allowed")
	}
	return nil
}

func (in *Inbound) TraceStream(stream server.HyStream, stats *server.StreamStats) {}
func (in *Inbound) UntraceStream(stream server.HyStream)                          {}

// Implements server.EventLogger
func (in *Inbound) Connect(addr net.Addr, id string, tx uint64) {
	log.Debug().Str("src_addr", addr.String()).Str("user_id", id).Uint64("tx", tx).Msgf("hysteria2 connect")
	in.usersLock.RLock()
	defer in.usersLock.RUnlock()
	user, ok := in.users[id]
	if !ok {
		return
	}

	in.cLock.Lock()
	defer in.cLock.Unlock()
	in.srcAddrMap[addr.(*net.UDPAddr).AddrPort()] = &srcAddrInfo{
		counter: user.Counter,
	}
}

func (in *Inbound) Disconnect(addr net.Addr, id string, err error) {
	in.cLock.Lock()
	defer in.cLock.Unlock()
	delete(in.srcAddrMap, addr.(*net.UDPAddr).AddrPort())
}
func (in *Inbound) TCPRequest(addr net.Addr, id, reqAddr string)                          {}
func (in *Inbound) TCPError(addr net.Addr, id, reqAddr string, err error)                 {}
func (in *Inbound) UDPRequest(addr net.Addr, id string, sessionID uint32, reqAddr string) {}
func (in *Inbound) UDPError(addr net.Addr, id string, sessionID uint32, err error)        {}

// Implements server.Outbound
type outboundAdapter struct {
	i.Dialer
	i.PacketListener
}

func (o *outboundAdapter) TCP(reqAddr string) (net.Conn, error) {
	ctx := session.GetCtx(context.Background())
	log.Ctx(ctx).Debug().Str("target", reqAddr).Msgf("hysteria2 tcp session")

	dest, err := mynet.ParseDestination(reqAddr)
	if err != nil {
		return nil, err
	}
	dest.Network = mynet.Network_TCP

	conn, err := o.Dialer.Dial(ctx, dest)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (o *outboundAdapter) UDP(reqAddr string) (server.UDPConn, error) {
	ctx := session.GetCtx(context.Background())
	log.Ctx(ctx).Debug().Str("target", reqAddr).Msgf("hysteria2 udp session")

	pc, err := o.PacketListener.ListenPacket(ctx, "udp", "")
	if err != nil {
		return nil, err
	}
	return &netUdpConnToServerUDPConn{pc}, nil
}

type netUdpConnToServerUDPConn struct {
	net.PacketConn
}

func (c *netUdpConnToServerUDPConn) ReadFrom(b []byte) (int, string, error) {
	n, addr, err := c.PacketConn.ReadFrom(b)
	if err != nil {
		return 0, "", err
	}
	return n, addr.String(), nil
}

func (c *netUdpConnToServerUDPConn) WriteTo(b []byte, addr string) (int, error) {
	netAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return 0, err
	}
	return c.PacketConn.WriteTo(b, netAddr)
}

type statsPacketConn struct {
	net.PacketConn

	inbound *Inbound
}

func (c *statsPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := c.PacketConn.WriteTo(b, addr)
	if err != nil {
		return 0, err
	}

	c.inbound.cLock.RLock()
	defer c.inbound.cLock.RUnlock()
	info, ok := c.inbound.srcAddrMap[addr.(*net.UDPAddr).AddrPort()]
	if !ok {
		// log.Debug().Str("addr", addr.String()).Msgf("hysteria2 src addr not found")
		return n, nil
	}
	info.counter.Add(uint64(n))

	// log.Debug().Str("src_addr", addr.String()).Uint64("traffic", info.counter.Load()).Msgf("hysteria2 traffic")

	if c.inbound.config.InStats != nil {
		c.inbound.config.InStats.Traffic.Add(uint64(n))
	}

	return n, nil
}
