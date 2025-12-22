package hysteria2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/domain"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/helper"

	hysErrors "github.com/apernet/hysteria/core/v2/errors"

	"github.com/apernet/hysteria/core/v2/client"
	"github.com/apernet/hysteria/extras/v2/obfs"
	"github.com/apernet/quic-go"
	"github.com/rs/zerolog/log"
)

// TODO: solve no recent network activity. Use activity checker at the raw packet conn level.
// Close dead connections immediately.

type HysClient struct {
	tag                        string
	address                    net.Address
	serverPicker               i.PortSelector
	IpResolverForNodeAddress   i.IPResolver
	IpResolverForTargetAddress i.IPResolver
	DomainStrategy             domain.DomainStrategy
	config                     *client.Config
	id                         atomic.Int32
	RejectQuic                 bool
	sync.Mutex
	clients                   []*wrappedClient
	concurrentCreateNewClient *concurrentCreateNewClient
}

type Config struct {
	Tag                        string
	PacketListener             i.PacketListener
	HysteriaClientConfig       *client.Config
	SalamanderPassword         string
	Address                    net.Address
	PortSelector               i.PortSelector
	IpResolverForNodeAddress   i.IPResolver
	DomainStrategy             domain.DomainStrategy
	IpResolverForTargetAddress i.IPResolver
	RejectQuic                 bool
}

func NewClient(config *Config) (*HysClient, error) {
	config.HysteriaClientConfig.ConnFactory = &hysConnFactory{
		packetListener: config.PacketListener,
	}
	if config.SalamanderPassword != "" {
		obfuscator, err := obfs.NewSalamanderObfuscator([]byte(config.SalamanderPassword))
		if err != nil {
			return nil, fmt.Errorf("failed to create obfuscator: %w", err)
		}
		config.HysteriaClientConfig.ConnFactory = &hysConnFactory{
			packetListener: config.PacketListener,
			obfuscator:     obfuscator,
		}
	}
	log.Debug().Msgf("keepalive period: %v", config.HysteriaClientConfig.QUICConfig.KeepAlivePeriod.Seconds())
	log.Debug().Msgf("max idle timeout: %v", config.HysteriaClientConfig.QUICConfig.MaxIdleTimeout.Seconds())

	d := &HysClient{
		tag:                        config.Tag,
		address:                    config.Address,
		config:                     config.HysteriaClientConfig,
		serverPicker:               config.PortSelector,
		IpResolverForNodeAddress:   config.IpResolverForNodeAddress,
		DomainStrategy:             config.DomainStrategy,
		IpResolverForTargetAddress: config.IpResolverForTargetAddress,
		RejectQuic:                 config.RejectQuic,
	}
	return d, nil
}

type concurrentCreateNewClient struct {
	sync.Mutex
	client *wrappedClient
	err    error
}

type wrappedClient struct {
	// lock sync.Mutex
	client.Client
	id          int32
	idle        int64 //seconds
	usedSession atomic.Int32

	timerLock sync.Mutex
	timer     *time.Timer

	dialing atomic.Int32

	lastActiveTime *atomic.Int64
}

func (c *wrappedClient) isActive() bool {
	// if runtime.GOOS == "ios" {
	// 	if time.Now().Unix()-c.lastActiveTime.Load() < 5 {
	// 		log.Debug().Int32("id", c.id).Msg("hys client active")
	// 		return true
	// 	}
	// } else {
	if time.Now().Unix()-c.lastActiveTime.Load() < c.idle {
		log.Debug().Int32("id", c.id).Msg("hys client active")
		return true
	}
	// }

	return false
}

func (w *wrappedClient) addTimer(hc *HysClient) {
	w.timerLock.Lock()
	defer w.timerLock.Unlock()
	if w.timer == nil {
		w.timer = time.AfterFunc(time.Duration(w.idle)*time.Second, func() {
			hc.removeClient(w)
		})
	}
}

func (w *wrappedClient) removeTimer() {
	w.timerLock.Lock()
	defer w.timerLock.Unlock()
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
}

// if not idle, ok. if idle, check if used session is 0, if so, remove it
func (hys *HysClient) okayToUse(c *wrappedClient) bool {
	if c.isActive() {
		return true
	}

	// if used session is 0, close it
	if c.usedSession.Load() == 0 {
		hys.removeClient(c)
	}
	return false
}

func (hys *HysClient) increaseUsedSession(c *wrappedClient) {
	if c.usedSession.Add(1) == 1 {
		c.removeTimer()
	}
}

func (hys *HysClient) decreaseUsedSession(c *wrappedClient) {
	if c.usedSession.Add(-1) == 0 {
		// idle, remove it
		if !c.isActive() && c.dialing.Load() == 0 {
			hys.removeClient(c)
		} else {
			c.addTimer(hys)
		}
	}
}

var streamLimitReachedError = quic.StreamLimitReachedError{}

func (d *HysClient) Tag() string {
	return d.tag
}

func (d *HysClient) removeClient(clientToRemove *wrappedClient) {
	log.Debug().Int32("id", clientToRemove.id).Msg("remove hys client")

	d.Lock()
	defer d.Unlock()

	if slices.Contains(d.clients, clientToRemove) {
		newClients := make([]*wrappedClient, 0, len(d.clients))
		for _, cl := range d.clients {
			if cl == clientToRemove {
				cl.Close()
				continue
			}
			newClients = append(newClients, cl)
		}
		d.clients = newClients
	}

}

func (d *HysClient) addNewClientConcurrent() (*wrappedClient, error) {
	d.Lock()
	if ccc := d.concurrentCreateNewClient; ccc != nil {
		d.Unlock()
		ccc.Lock()
		defer ccc.Unlock()
		if ccc.client != nil {
			return ccc.client, nil
		} else {
			return nil, ccc.err
		}
	} else {
		ccc = &concurrentCreateNewClient{}
		d.concurrentCreateNewClient = ccc
		ccc.Lock()
		d.Unlock()
		defer func() {
			ccc.Unlock()
			d.Lock()
			d.concurrentCreateNewClient = nil
			d.Unlock()
		}()
		ccc.client, ccc.err = d.addNewClientCommon()
		return ccc.client, ccc.err
	}
}

func (d *HysClient) addNewClient() (*wrappedClient, error) {
	// restrict the num of clients on ios for memory
	if runtime.GOOS == "ios" {
		d.Lock()
		numOfClient := len(d.clients)
		d.Unlock()
		if numOfClient > 1 {
			return d.addNewClientConcurrent()
		}
	}

	return d.addNewClientCommon()
}

func (d *HysClient) addNewClientCommon() (*wrappedClient, error) {
	id := d.id.Add(1)

	port := d.serverPicker.SelectPort()
	udpAddr := &net.UDPAddr{
		Port: int(port),
	}
	config := *d.config
	var cl client.Client
	var err error
	factory := d.config.ConnFactory

	var idleTimer *atomic.Int64
	if d.address.Family().IsDomain() {
		ctx := log.With().Int32("id", id).Str("domain", d.address.Domain()).
			Logger().WithContext(context.Background())
		ips := domain.GetIPs(ctx, d.address.Domain(), d.DomainStrategy,
			d.IpResolverForNodeAddress)
		if len(ips) == 0 {
			return nil, errors.New("failed to lookup server address")
		}
		for _, ip := range ips {
			udpAddr.IP = ip
			config.ServerAddr = udpAddr
			conn, err := factory.New(udpAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to create connection: %w", err)
			}
			c := &ddlPacketConn{
				PacketConn: conn,
				id:         id,
				idle:       int64(config.QUICConfig.MaxIdleTimeout.Seconds())}
			c.lastReadTime.Store(time.Now().Unix())
			idleTimer = &c.lastReadTime
			config.ConnFactory = &connFactory{
				PacketConn: c}
			cl, _, err = client.NewClient(&config)
			if err == nil {
				break
			}
		}
	} else {
		udpAddr.IP = d.address.IP()
		config.ServerAddr = udpAddr
		conn, err1 := factory.New(udpAddr)
		if err1 != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}
		c := &ddlPacketConn{
			PacketConn: conn,
			id:         id,
			idle:       int64(config.QUICConfig.MaxIdleTimeout.Seconds())}
		c.lastReadTime.Store(time.Now().Unix())
		idleTimer = &c.lastReadTime
		config.ConnFactory = &connFactory{
			PacketConn: c}
		cl, _, err = client.NewClient(&config)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	wrappedClient := &wrappedClient{
		Client: cl,
		id:     id,
		idle:   int64(min(15, config.QUICConfig.MaxIdleTimeout.Seconds())),
	}
	wrappedClient.lastActiveTime = idleTimer

	// ccc.client = wrappedClient
	d.Lock()
	d.clients = append(d.clients, wrappedClient)
	log.Debug().Interface("client", cl).Int32("id", wrappedClient.id).Msg("new hys client")
	d.Unlock()
	return wrappedClient, nil
}

type connFactory struct {
	net.PacketConn
}

func (c *connFactory) New(addr net.Addr) (net.PacketConn, error) {
	return c.PacketConn, nil
}

type ddlPacketConn struct {
	net.PacketConn
	id           int32
	idle         int64
	lastReadTime atomic.Int64
}

func (c *ddlPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(p)
	if err != nil {
		return n, addr, err
	}
	c.lastReadTime.Store(time.Now().Unix())
	return n, addr, err
}

func (c *ddlPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if time.Now().Unix()-c.lastReadTime.Load() > c.idle {
		log.Debug().Int32("id", c.id).Msg("hys client no read activity but still sending data")
	}
	return c.PacketConn.WriteTo(p, addr)
}

type hysConnFactory struct {
	packetListener i.PacketListener
	obfuscator     obfs.Obfuscator
}

// to prevent "connection already exists" panic
var correntConns = make(map[string]*packetConn)
var correntConnsLock sync.Mutex

type packetConn struct {
	key string
	net.PacketConn
}

func key(c net.PacketConn) string {
	return c.LocalAddr().Network() + " " + c.LocalAddr().String()
}

func (c *packetConn) Close() error {
	defer func() {
		time.Sleep(time.Millisecond * 100)
		correntConnsLock.Lock()
		delete(correntConns, c.key)
		log.Debug().Str("local_addr", c.LocalAddr().String()).Int("remaining", len(correntConns)).Msg("hys packetConn close")
		correntConnsLock.Unlock()
	}()
	return c.PacketConn.Close()
}

// addr is the remote address
func (c *hysConnFactory) New(addr net.Addr) (net.PacketConn, error) {
	log.Debug().Str("addr", addr.String()).Str("network", addr.Network()).Msg("hysteria2 client listen udp")
	network := addr.Network()
	if network == "udp" {
		ip, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return nil, fmt.Errorf("failed to split host port: %w", err)
		}
		if net.ParseAddress(ip).Family().IsIPv6() {
			network = "udp6"
		} else {
			network = "udp4"
		}
	}

	ctx := log.Logger.WithContext(context.Background())

	var conn net.PacketConn
	for range 5 {
		tmpConn, err := c.packetListener.ListenPacket(ctx, network, "")
		if err != nil {
			return nil, fmt.Errorf("failed to listen system packet: %w", err)
		}

		k := key(tmpConn)

		correntConnsLock.Lock()
		_, existed := correntConns[k]
		if existed {
			correntConnsLock.Unlock()
			defer tmpConn.Close()
			log.Debug().Str("local_addr", tmpConn.LocalAddr().String()).Msg("connection already exists, try to create new one")
			continue
		}
		packetConn := &packetConn{key: k, PacketConn: tmpConn}
		conn = packetConn
		correntConns[k] = packetConn
		correntConnsLock.Unlock()
		break
	}

	if conn == nil {
		return nil, errors.New("failed to listen system packet after 5 times")
	}

	log.Debug().Str("local_addr", conn.LocalAddr().String()).Msg("hysteria2 client listen udp succ")
	if c.obfuscator != nil {
		conn = obfs.WrapPacketConn(conn, c.obfuscator)
	}
	return conn, nil
}

func (d *HysClient) dialCommon(ctx context.Context, dst net.Destination) (net.Conn, *wrappedClient, error) {
	var conn net.Conn
	var err error
	var wrappedClient *wrappedClient
	if dst.Network == net.Network_UDP {
		if dst.Port == 443 && d.RejectQuic {
			return nil, nil, errors.New("reject quic over hysteria2")
		}
		var udpConn client.HyUDPConn
		udpConn, wrappedClient, err = d.udp(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to dial UDP: %w", err)
		}
		target := dst
		// change target to ip if it is domain. Because if target is domain,
		//  this connection will likely fail
		if d.IpResolverForTargetAddress != nil && dst.Address.Family().IsDomain() {
			ips, _ := d.IpResolverForTargetAddress.LookupIP(
				ctx, dst.Address.Domain())
			if len(ips) > 0 {
				target.Address = net.IPAddress(ips[rand.Intn(len(ips))])
			}
		}
		conn = &HyUdpConnToNetConn{addr: target.NetAddr(), hyUdpConn: udpConn}
	} else {
		conn, wrappedClient, err = d.tcp(ctx, dst)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to dial: %w", err)
		}
	}
	return conn, wrappedClient, nil
}

func (d *HysClient) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	conn, wrappedClient, err := d.dialCommon(ctx, dst)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer d.decreaseUsedSession(wrappedClient)
	defer conn.Close()
	return helper.Relay(ctx, rw, rw, buf.NewReader(conn), buf.NewWriter(conn))
}

func (d *HysClient) ProxyDial(ctx context.Context, dst net.Destination,
	initialData buf.MultiBuffer) (i.FlowConn, error) {
	conn, wrappedClient, err := d.dialCommon(ctx, dst)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	c := proxy.NewFlowConn(
		proxy.FlowConnOption{
			Reader:      buf.NewReader(conn),
			Writer:      buf.NewWriter(conn),
			SetDeadline: conn,
			Close: func() error {
				d.decreaseUsedSession(wrappedClient)
				return nil
			},
		})
	if initialData.Len() > 0 {
		err = c.WriteMultiBuffer(initialData)
		if err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, nil
}

func (d *HysClient) tcp(ctx context.Context, dest net.Destination) (net.Conn, *wrappedClient, error) {
	d.Lock()
	clients := d.clients
	d.Unlock()

	var conn net.Conn
	var err error

	target := dest
	start := 0
	if len(clients) != 0 {
		start = rand.Intn(len(clients))
	}
	// find a client and use it to dial
	for i := 0; i < len(clients); i++ {
		cl := clients[(start+i)%len(clients)]

		if !d.okayToUse(cl) {
			log.Ctx(ctx).Debug().Int32("id", cl.id).Int32("used_session", cl.usedSession.Load()).Msg("hys client not okay to use")
			continue
		}

		// succ := cl.lock.TryLock()
		// if !succ {
		// 	// it is being used, so skip it
		// 	continue
		// }

		cl.dialing.Add(1)
		conn, err = cl.TCP(target.String())
		cl.dialing.Add(-1)
		// cl.lock.Unlock()
		if err != nil {
			if !errors.As(err, &streamLimitReachedError) {
				log.Ctx(ctx).Debug().Int32("id", cl.id).Err(err).Msg("hys client failed to TCP")
				if errors.As(err, &hysErrors.ClosedError{}) {
					d.removeClient(cl)
					continue
				} else {
					return nil, nil, err
				}
			} else {
				log.Ctx(ctx).Debug().Int32("id", cl.id).Msg("hys client stream limit reached")
				// err is stream limit reached
				continue
			}
		}
		// if runtime.GOOS == "ios" {
		// conn = &ActiveTimeConn{Conn: conn, lastActiveTime: &cl.lastActiveTime}
		// }
		d.increaseUsedSession(cl)
		log.Ctx(ctx).Debug().Int32("id", cl.id).Int32("used_session", cl.usedSession.Load()).Msg("using hys client")
		return conn, cl, nil
	}

	newClient, err := d.addNewClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add new client: %w", err)
	}
	log.Ctx(ctx).Debug().Int32("id", newClient.id).Int32("used_session", newClient.usedSession.Load()).Msg("using hys client")
	conn, err = newClient.TCP(target.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial: %w", err)
	}
	d.increaseUsedSession(newClient)
	return conn, newClient, nil
}

type ActiveTimeConn struct {
	net.Conn
	lastActiveTime *atomic.Int64
}

func (c *ActiveTimeConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	c.lastActiveTime.Store(time.Now().Unix())
	return n, err
}

func (d *HysClient) HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error {
	udpConn, wrappedClient, err := d.udp(ctx)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}
	conn := &hyUdpConnToUDPPacketConn{hyUdpConn: udpConn, hysClient: d, wrappedClient: wrappedClient}
	defer conn.Close()
	return helper.RelayUDPPacketConn(ctx, p, conn)
}

func (d *HysClient) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	udpConn, wrappedClient, err := d.udp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP: %w", err)
	}
	conn := &hyUdpConnToUDPPacketConn{hyUdpConn: udpConn, hysClient: d, wrappedClient: wrappedClient}
	return conn, nil
}

func (d *HysClient) udp(ctx context.Context) (client.HyUDPConn, *wrappedClient, error) {
	d.Lock()
	clients := d.clients
	d.Unlock()

	var udpConn client.HyUDPConn
	var err error

	start := 0
	if len(clients) != 0 {
		start = rand.Intn(len(clients))
	}

	for i := 0; i < len(clients); i++ {
		cl := clients[(start+i)%len(clients)]

		if !d.okayToUse(cl) {
			continue
		}

		cl.dialing.Add(1)
		udpConn, err = cl.UDP()
		cl.dialing.Add(-1)

		if err != nil {
			if !errors.As(err, &streamLimitReachedError) {
				log.Ctx(ctx).Error().Int32("id", cl.id).Err(err).Msg("hys client failed to UDP")
				if errors.As(err, &hysErrors.ClosedError{}) {
					d.removeClient(cl)
				} else {
					return nil, nil, err
				}
			}
			continue
		}
		// if runtime.GOOS == "ios" {
		// udpConn = &ActiveTimeHyUDPConn{HyUDPConn: udpConn, lastActiveTime: &cl.lastActiveTime}
		// }

		d.increaseUsedSession(cl)
		log.Ctx(ctx).Debug().Int32("id", cl.id).Int32("used_session", cl.usedSession.Load()).Msg("using hys client")
		return udpConn, cl, nil
	}
	newClient, err := d.addNewClient()
	if err != nil {
		return nil, nil, err
	}
	log.Ctx(ctx).Debug().Int32("id", newClient.id).Int32("used_session", newClient.usedSession.Load()).Msg("using hys client")
	udpConn, err = newClient.UDP()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial UDP: %w", err)
	}
	d.increaseUsedSession(newClient)
	return udpConn, newClient, nil
}

type ActiveTimeHyUDPConn struct {
	client.HyUDPConn
	lastActiveTime *atomic.Int64
}

func (c *ActiveTimeHyUDPConn) Receive() ([]byte, string, error) {
	b, src, err := c.HyUDPConn.Receive()
	if err != nil {
		return b, src, err
	}
	c.lastActiveTime.Store(time.Now().Unix())
	return b, src, nil
}

type hyUdpConnToUDPPacketConn struct {
	hyUdpConn     client.HyUDPConn
	hysClient     *HysClient
	wrappedClient *wrappedClient
}

func (c *hyUdpConnToUDPPacketConn) ReadPacket() (*udp.Packet, error) {
	b, src, err := c.hyUdpConn.Receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive: %w", err)
	}
	srcDest, err := net.ParseDestination(src)
	if err != nil {
		return nil, err
	}
	srcDest.Network = net.Network_UDP
	return &udp.Packet{
		Payload: buf.FromBytes(b),
		Source:  srcDest,
	}, nil
}

func (c *hyUdpConnToUDPPacketConn) WritePacket(p *udp.Packet) error {
	return c.hyUdpConn.Send(p.Payload.Bytes(), p.Target.String())
}

func (c *hyUdpConnToUDPPacketConn) Close() error {
	err := c.hyUdpConn.Close()
	c.hysClient.decreaseUsedSession(c.wrappedClient)
	return err
}

// Adapter for client.HyUDPConn to implement net.PacketConn
type HyPacketConn struct {
	hyConn client.HyUDPConn
}

// ReadFrom implements net.PacketConn
func (c *HyPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	data, remoteAddr, err := c.hyConn.Receive()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to receive: %w", err)
	}

	n = copy(p, data)
	if n < len(data) {
		return n, &net.UDPAddr{IP: net.ParseIP(remoteAddr)}, io.ErrShortBuffer
	}

	return n, &net.UDPAddr{IP: net.ParseIP(remoteAddr)}, nil
}

// WriteTo implements net.PacketConn
func (c *HyPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	err = c.hyConn.Send(p, addr.String())
	if err != nil {
		return 0, fmt.Errorf("failed to send: %w", err)
	}
	return len(p), nil
}

// Close implements net.PacketConn
func (c *HyPacketConn) Close() error {
	return c.hyConn.Close()
}

// LocalAddr implements net.PacketConn
func (c *HyPacketConn) LocalAddr() net.Addr {
	// Since HyUDPConn doesn't provide local address info,
	// return a placeholder address
	return &net.UDPAddr{IP: net.AnyIP.IP()}
}

// SetDeadline implements net.PacketConn
func (c *HyPacketConn) SetDeadline(t time.Time) error {
	// HyUDPConn doesn't support deadlines
	return nil
}

// SetReadDeadline implements net.PacketConn
func (c *HyPacketConn) SetReadDeadline(t time.Time) error {
	// HyUDPConn doesn't support deadlines
	return nil
}

// SetWriteDeadline implements net.PacketConn
func (c *HyPacketConn) SetWriteDeadline(t time.Time) error {
	// HyUDPConn doesn't support deadlines
	return nil
}

type HyUdpConnToNetConn struct {
	addr      string
	hyUdpConn client.HyUDPConn
}

func (c *HyUdpConnToNetConn) Read(p []byte) (n int, err error) {
	for {
		data, src, err := c.hyUdpConn.Receive()
		if err != nil {
			return 0, fmt.Errorf("failed to receive: %w", err)
		}
		if src == c.addr {
			n = copy(p, data)
			return n, nil
		}
	}
}

func (c *HyUdpConnToNetConn) Write(p []byte) (int, error) {
	return len(p), c.hyUdpConn.Send(p, c.addr)
}

func (c *HyUdpConnToNetConn) Close() error {
	return c.hyUdpConn.Close()
}

func (c *HyUdpConnToNetConn) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.AnyIP.IP(), Port: 0}
}

func (c *HyUdpConnToNetConn) RemoteAddr() net.Addr {
	return &net.UDPAddr{IP: net.AnyIP.IP(), Port: 0}
}

func (c *HyUdpConnToNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *HyUdpConnToNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *HyUdpConnToNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}
