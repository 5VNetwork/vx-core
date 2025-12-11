package i

import (
	"context"
	"io"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/miekg/dns"
)

type Handler interface {
	FlowHandler
	PacketHandler
}

type ConnHandler interface {
	HandleConn(ctx context.Context, dst net.Destination, rw io.ReadWriter) error
}

type FlowHandler interface {
	// Read request data from rw; meanwhile write response data to rw.
	// Returns when there is no more data from rw (io.EOF) and no more response data; or when an
	// error occurs; or when ctx is canceled.
	// rw.CloseWrite() will be called when no more respone data;
	// When an error occurs and it returns, rw might still be used by some goroutines
	// for reading or writing.
	HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error
}

type DeadlineRW interface {
	buf.ReaderWriter
	ReadDeadline
}

type PacketHandler interface {
	// Read request data from rw; meanwhile write response data to rw.
	// Returns when there is no request for a while; or when an
	// error occurs; or when ctx is canceled.
	HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error
}

type ReadDeadline interface {
	SetReadDeadline(t time.Time) error
}

type ProxyDialer interface {
	ProxyDial(ctx context.Context, dst net.Destination, initialData buf.MultiBuffer) (FlowConn, error)
}
type ProxyPacketListener interface {
	// dst is the destination of the initial udp packet
	ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error)
}
type FlowConn interface {
	buf.ReaderWriter
	Close() error
}

type Outbound interface {
	Tag() string
	Handler
}

type HandlerWith6Info interface {
	Outbound
	Support6() bool
}

type GeoHelper interface {
	MatchDomain(domain string, tag string) bool
	MatchIP(ip net.IP, tag string) bool
	MatchAppId(appId string, tag string) bool
}

type TimeoutSetting interface {
	HandshakeTimeout() time.Duration
	TcpIdleTimeout() time.Duration
	UdpIdleTimeout() time.Duration
	SshIdleTimeout() time.Duration
	DnsIdleTimeout() time.Duration
	UpLinkOnlyTimeout() time.Duration
	DownLinkOnlyTimeout() time.Duration
}

type StatsSetting interface {
	CalculateUserStats() bool
	// whether sample throughput and ping
	CalculateOutboundLinkStats() bool
	CalculateInboundLinkStats() bool
	CalculateInboundStats() bool
	CalculateSessionStats() bool
}

type BufferPolicy interface {
	UserBufferSize(level uint32) int32
}

type IPSet interface {
	Match(ip net.IP) bool
}

type DomainSet interface {
	Match(domain string) bool
}

type AppSet interface {
	Match(appId string) bool
}

type Dialer interface {
	Dial(ctx context.Context, dst net.Destination) (net.Conn, error)
}

type PacketListener interface {
	ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error)
}

type Listener interface {
	Listen(ctx context.Context, addr net.Addr) (net.Listener, error)
}

type DialerListener interface {
	Dialer
	PacketListener
}

type UserValidator interface {
	Validate(secret []byte) bool
}

type User interface {
	Uid() string
	Level() uint32
	Secret() string
	Counter() *atomic.Uint64
}

type DefaultInterfaceInfo interface {
	// Name of the default ipv4 interface
	DefaultInterface4() uint32
	DefaultInterface6() uint32
	// Name of the default ipv4 interface
	DefaultInterfaceName4() string
	DefaultInterfaceName6() string
	// Dns servers of the default ipv4 interface
	DefaultDns4() []netip.Addr
	DefaultDns6() []netip.Addr
	// whether the default interface actually support ipv6
	// -1: no; 0: unknown for now, is checking; 1: yes
	SupportIPv6() int
	// whether the default interface has global ipv6 address
	// as determined by netip.Addr.IsGlobalUnicast()
	HasGlobalIPv6() (bool, error)
	DefaultInterfaceChangeSubject
}

type DefaultInterfaceChangeSubject interface {
	Register(observer DefaultInterfaceChangeObserver)
	Unregister(observer DefaultInterfaceChangeObserver)
	Notify()
}

type DefaultInterfaceChangeObserver interface {
	OnDefaultInterfaceChanged()
}

type OnDefaultInterfaceChanged func()

func (f OnDefaultInterfaceChanged) OnDefaultInterfaceChanged() {
	f()
}

type IPv6SupportChangeSubject interface {
	Register(observer IPv6SupportChangeObserver)
	Unregister(observer IPv6SupportChangeObserver)
	Notify()
}
type IPv6SupportChangeObserver interface {
	OnIPv6SupportChanged()
}

type IPResolver interface {
	LookupIP(ctx context.Context, domain string) ([]net.IP, error)
	LookupIPv4(ctx context.Context, domain string) ([]net.IP, error)
	LookupIPv6(ctx context.Context, domain string) ([]net.IP, error)
}

type ECHResolver interface {
	LookupECH(ctx context.Context, domain string) ([]byte, error)
}

type DnsResolver interface {
	IPResolver
	ECHResolver
}

type DnsServer interface {
	HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error)
}

type FakeDnsPool interface {
	GetDomainFromFakeDNS(ip net.Address) string
	IsIPInIPPool(ip net.Address) bool
}

type OutboundManager interface {
	GetHandler(tag string) Outbound
	GetAllHandlers() []Outbound
}

type Router interface {
	PickHandler(ctx context.Context, si *session.Info) (Outbound, error)
	// rw is either a buf.ReaderWriter or a udp.PacketConn
	PickHandlerWithData(ctx context.Context, si *session.Info, rw interface{}) (interface{}, Outbound, error)
}

type IpToDomain interface {
	// if there are multiple domains with same ip, return nothing
	GetDomain(ip net.IP) []string
}

type HandlerErrorObserver interface {
	OnHandlerError(tag string, err error)
}

type PortSelector interface {
	SelectPort() uint16
}

type UnauthorizedReport interface {
	ReportUnauthorized(ip string, credential string)
}
