package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/domain"
	net1 "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	"github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/kcp"
	"github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/rs/zerolog/log"
)

type DialerFactory interface {
	GetDialer(*Config) (i.Dialer, error)
	GetPacketListener(*Config) (i.PacketListener, error)
}

var DefaultDialer = tcp.NewTcpDialer(nil, nil, &dlhelper.SocketSetting{})
var DefaultPacketListener = &dlhelper.SocketSetting{}

type Config struct {
	Security interface{}
	Protocol interface{}
	Socket   *dlhelper.SocketSetting
	// used to lookup ech config
	DnsServer i.ECHResolver
	// dialer
	DomainStrategy domain.DomainStrategy
	Address        net1.Address
	PortSelector   i.PortSelector
}

func NewDialer(protocolConfig, securityConfig interface{}, dl i.DialerListener, dnsServer i.ECHResolver) (i.Dialer, error) {
	if protocolConfig == nil {
		protocolConfig = &tcp.TcpConfig{}
	}

	var securityEngine security.Engine
	var err error
	switch sc := securityConfig.(type) {
	case *tls.TlsConfig:
		securityEngine, err = tls.NewEngine(tls.EngineConfig{Config: sc, DnsServer: dnsServer})
		if err != nil {
			return nil, fmt.Errorf("failed to create tls engine: %w", err)
		}
	case *reality.RealityConfig:
		securityEngine, err = reality.NewEngine(sc)
		if err != nil {
			return nil, fmt.Errorf("failed to create reality engine: %w", err)
		}
	}
	var d i.Dialer
	switch transportConfig := protocolConfig.(type) {
	//TODO: currently, udp dial also uses this
	case *tcp.TcpConfig:
		d = tcp.NewTcpDialer(transportConfig, securityEngine, dl)
	case *kcp.KcpConfig:
		d = kcp.NewKcpDialer(transportConfig, securityEngine, dl)
	case *http.HttpConfig:
		d = http.NewDialer(transportConfig, securityEngine, dl)
	case *websocket.WebsocketConfig:
		if securityConfig != nil && transportConfig.Host == "" {
			if tls, ok := securityConfig.(*tls.TlsConfig); ok {
				transportConfig.Host = tls.ServerName
			} else if reality, ok := securityConfig.(*reality.RealityConfig); ok {
				transportConfig.Host = reality.ServerName
			}
		}
		d = websocket.NewWebsocketDialer(transportConfig, securityEngine, dl)
	// case *quic.QuicConfig:
	// 	return quic.NewQuicDialer(transportConfig, securityEngine, dl), nil
	case *grpc.GrpcConfig:
		d = grpc.NewGrpcDialer(transportConfig, securityEngine, dl)
	case *splithttp.SplitHttpConfig:
		d, err = splithttp.NewXhttpDialer(transportConfig, securityEngine, dl, dnsServer)
		if err != nil {
			return nil, err
		}
	case *httpupgrade.HttpUpgradeConfig:
		d = httpupgrade.NewHttpUpgradeDialer(transportConfig, securityEngine, dl)
	default:
		return nil, errors.New("invalid transport config")
	}
	return &UdpDialer{
		tcpDialer: d,
		udpDialer: dl,
	}, nil
}

type UdpDialer struct {
	tcpDialer i.Dialer
	udpDialer i.Dialer
}

func (d *UdpDialer) Dial(ctx context.Context, dest net1.Destination) (net.Conn, error) {
	if dest.Network == net1.Network_UDP {
		return d.udpDialer.Dial(ctx, dest)
	}
	return d.tcpDialer.Dial(ctx, dest)
}

type RetryDialer struct {
	i.Dialer
}

func (d *RetryDialer) Dial(ctx context.Context, target net1.Destination) (net.Conn, error) {
	var conn net.Conn
	err := retry.ExponentialBackoff(5, 100).On(func() error {
		rawConn, err := d.Dialer.Dial(ctx, target)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("dial failed")
			return err
		}
		conn = rawConn
		return nil
	})
	return conn, err
}

type BindToDefaultNICDialer struct {
	defaultNICMonitor i.DefaultInterfaceInfo
	socketSetting     atomic.Value // *dlhelper.SocketSetting
}

func NewBindToDefaultNICDialer(defaultNICMonitor i.DefaultInterfaceInfo, socketSetting *dlhelper.SocketSetting) *BindToDefaultNICDialer {
	socketSettingValue := atomic.Value{}
	socketSettingValue.Store(socketSetting)
	return &BindToDefaultNICDialer{
		defaultNICMonitor: defaultNICMonitor,
		socketSetting:     socketSettingValue,
	}
}

func (d *BindToDefaultNICDialer) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	socket := d.socketSetting.Load().(*dlhelper.SocketSetting)
	if socket.BindToDevice4 != d.defaultNICMonitor.DefaultInterface4() ||
		socket.BindToDevice6 != d.defaultNICMonitor.DefaultInterface6() ||
		socket.BindToDeviceName != d.defaultNICMonitor.DefaultInterfaceName4() {
		socket = socket.Dulplicate()
		socket.BindToDevice4 = d.defaultNICMonitor.DefaultInterface4()
		socket.BindToDevice6 = d.defaultNICMonitor.DefaultInterface6()
		socket.BindToDeviceName = d.defaultNICMonitor.DefaultInterfaceName4()
		d.socketSetting.Store(socket)
	}
	return socket.ListenPacket(ctx, network, address)
}

// when DefaultInterfaceInfo is nil, LookupIP might return ipv6 address while the default nic does not support ipv6.
func (d *BindToDefaultNICDialer) Dial(ctx context.Context, dst net1.Destination) (net.Conn, error) {
	socket := d.socketSetting.Load().(*dlhelper.SocketSetting)
	if socket.BindToDevice4 != d.defaultNICMonitor.DefaultInterface4() ||
		socket.BindToDevice6 != d.defaultNICMonitor.DefaultInterface6() ||
		socket.BindToDeviceName != d.defaultNICMonitor.DefaultInterfaceName4() {
		socket = socket.Dulplicate()
		socket.BindToDevice4 = d.defaultNICMonitor.DefaultInterface4()
		socket.BindToDevice6 = d.defaultNICMonitor.DefaultInterface6()
		socket.BindToDeviceName = d.defaultNICMonitor.DefaultInterfaceName4()
		d.socketSetting.Store(socket)
	}
	if dst.Address == net1.LocalHostIP || dst.Address == net1.LocalHostIPv6 {
		socket = socket.Dulplicate()
		socket.BindToDevice4 = 0
		socket.BindToDevice6 = 0
		socket.BindToDeviceName = ""
		return socket.Dial(ctx, dst)
	}

	if socket.BindToDevice4 == 0 && socket.BindToDevice6 == 0 && socket.BindToDeviceName == "" {
		return nil, errors.New("failed to get default nic")
	}
	return socket.Dial(ctx, dst)
}

func DefaultDialerFactory() *DialerFactoryImp {
	return defaultDf
}

var defaultDf = &DialerFactoryImp{}

// To prevent route loop
type DialerFactoryImp struct {
	DialerFactoryOption
}

type DialerFactoryOption struct {
	Retry                   bool
	BindToDefaultNIC        bool
	IpResolver              i.IPResolver
	DefaultInterfaceMonitor i.DefaultInterfaceInfo
	// used to bind to default nic
	FdFunc FdFunc
}

type FdFunc func(fd uintptr) error

func NewDialerFactoryImp(options DialerFactoryOption) *DialerFactoryImp {
	return &DialerFactoryImp{
		DialerFactoryOption: options,
	}
}

func (d *DialerFactoryImp) GetDialer(config *Config) (i.Dialer, error) {
	if config == nil {
		config = &Config{}
	}
	if config.Socket == nil {
		config.Socket = &dlhelper.SocketSetting{}
	}

	var dialer i.DialerListener
	dialer = config.Socket

	if d.FdFunc != nil {
		config.Socket.FdFunc = d.FdFunc
	}

	if d.BindToDefaultNIC && config.Socket.BindToDevice4 == 0 &&
		config.Socket.BindToDevice6 == 0 &&
		config.Socket.BindToDeviceName == "" {
		socketSetting := atomic.Value{}
		socketSetting.Store(config.Socket)
		dialer = &BindToDefaultNICDialer{
			defaultNICMonitor: d.DefaultInterfaceMonitor,
			socketSetting:     socketSetting,
		}
	}

	if d.IpResolver != nil {
		dialer = &ResolveDomainDialer{
			Dns:                  d.IpResolver,
			DialerListener:       dialer,
			DefaultInterfaceInfo: d.DefaultInterfaceMonitor,
			Strategy:             config.DomainStrategy,
		}
	}

	// config.Socket.Resolver = resolver
	da, err := NewDialer(config.Protocol, config.Security,
		dialer, config.DnsServer)
	if err != nil {
		return nil, err
	}

	if d.Retry {
		da = &RetryDialer{Dialer: dialer}
	}

	return da, nil
}

func (d *DialerFactoryImp) GetPacketListener(config *Config) (i.PacketListener, error) {
	if config == nil {
		config = &Config{}
	}
	if config.Socket == nil {
		config.Socket = &dlhelper.SocketSetting{}
	}

	if d.FdFunc != nil {
		config.Socket.FdFunc = d.FdFunc
	}

	if d.BindToDefaultNIC && config.Socket.BindToDevice4 == 0 &&
		config.Socket.BindToDevice6 == 0 && config.Socket.BindToDeviceName == "" {
		socketSetting := atomic.Value{}
		socketSetting.Store(config.Socket)
		return &BindToDefaultNICDialer{
			defaultNICMonitor: d.DefaultInterfaceMonitor,
			socketSetting:     socketSetting,
		}, nil
	}
	return config.Socket, nil
}

// This dialer will resolve domain to ip address before dial.
// If the default interface does not support ipv6, it will only lookup ipv4.
type ResolveDomainDialer struct {
	i.DialerListener
	Dns i.IPResolver
	// if nil, use lookupIP to lookup domain
	// when not-nil, Dialer tyically use default nic to dial.
	DefaultInterfaceInfo i.DefaultInterfaceInfo
	Strategy             domain.DomainStrategy
}

// beaware that when DefaultInterfaceInfo is nil, LookupIP might return ipv6 address while the default nic does not support ipv6.
func (d *ResolveDomainDialer) Dial(ctx context.Context, dst net1.Destination) (net.Conn, error) {
	if dst.Address.Family().IsDomain() {
		strategy := d.Strategy
		if d.DefaultInterfaceInfo != nil &&
			d.DefaultInterfaceInfo.SupportIPv6() < 0 &&
			(strategy == domain.DomainStrategy_PreferIPv6 ||
				strategy == domain.DomainStrategy_PreferIPv4) {
			strategy = domain.DomainStrategy_IPv4Only
		}
		ips := domain.GetIPs(ctx, dst.Address.Domain(), strategy, d.Dns)
		for _, ip := range ips {
			dst.Address = net1.IPAddress(ip)
			conn, err := d.DialerListener.Dial(ctx, dst)
			if err == nil {
				return conn, nil
			}
		}
	}
	return d.DialerListener.Dial(ctx, dst)
}

// use a handler to dial and listen
type HandlerDialerFactory struct {
	Handler   i.Handler
	DnsServer i.ECHResolver
}

func (d *HandlerDialerFactory) GetDialer(config *Config) (i.Dialer, error) {
	return NewDialer(
		config.Protocol,
		config.Security,
		&util.HandlerToDialerListener{
			FlowHandlerToDialer:     util.FlowHandlerToDialer{FlowHandler: d.Handler},
			PacketHandlerToListener: util.PacketHandlerToListener{PacketHandler: d.Handler},
		}, d.DnsServer)
}

func (d *HandlerDialerFactory) GetPacketListener(config *Config) (i.PacketListener, error) {
	return &util.PacketHandlerToListener{PacketHandler: d.Handler}, nil
}

type Prefer4Dialer struct {
	i.Dialer
	IpResolver i.IPResolver
}

func (p *Prefer4Dialer) Dial(ctx context.Context, dst net1.Destination) (net.Conn, error) {
	if dst.Address.Family().IsDomain() {
		ips, _ := p.IpResolver.LookupIPv4(ctx, dst.Address.Domain())
		if len(ips) > 0 {
			dst.Address = net1.IPAddress(ips[0])
			return p.Dialer.Dial(ctx, dst)
		}
	}
	return p.Dialer.Dial(ctx, dst)
}
