// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package hysteria2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	stdHTTP "net/http"
	"net/netip"
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
	"github.com/apernet/hysteria/extras/v2/realm"
	"github.com/apernet/quic-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HysClient struct {
	tag                        string
	address                    net.Address
	dialer                     i.Dialer
	serverPicker               i.PortSelector
	IpResolverForNodeAddress   i.IPResolver
	IpResolverForTargetAddress i.IPResolver
	DomainStrategy             domain.DomainStrategy
	config                     *client.Config
	id                         atomic.Int32
	RejectQuic                 bool
	realmConfig                RealmConfig
	connFactory                *hysConnFactory
	realmConnFactory           *hysConnFactory
	salamanderPSK              []byte
	sync.Mutex
	clients                   []*wrappedClient
	concurrentCreateNewClient *concurrentCreateNewClient
}

type RealmConfig struct {
	RealmAddr         string
	LocalPort         uint16
	STUNServers       []string
	STUNTimeout       time.Duration
	PunchTimeout      time.Duration
	HeartbeatInterval time.Duration
	Insecure          bool
	IPMode            string
	PortMapping       RealmPortMappingConfig
	Resolver          i.IPResolver
}

type RealmPortMappingConfig struct {
	Enabled  bool
	Timeout  time.Duration
	Lifetime time.Duration
}

type Config struct {
	Tag                      string
	PacketListener           i.PacketListener
	Dialer                   i.Dialer
	HysteriaClientConfig     *client.Config
	SalamanderPassword       string
	Address                  net.Address
	PortSelector             i.PortSelector
	IpResolverForNodeAddress i.IPResolver
	DomainStrategy           domain.DomainStrategy
	RejectQuic               bool
	Realm                    RealmConfig
}

func NewClient(config *Config) (*HysClient, error) {
	connFactory := &hysConnFactory{
		packetListener: config.PacketListener,
	}
	realmConnFactory := &hysConnFactory{
		packetListener: config.PacketListener,
	}
	var salamanderPSK []byte
	if config.SalamanderPassword != "" {
		if len(config.SalamanderPassword) < 4 {
			return nil, fmt.Errorf("failed to create obfuscator: password too short")
		}
		salamanderPSK = []byte(config.SalamanderPassword)
		connFactory.salamanderPSK = salamanderPSK
	}
	config.HysteriaClientConfig.ConnFactory = connFactory
	log.Debug().Msgf("keepalive period: %v", config.HysteriaClientConfig.QUICConfig.KeepAlivePeriod.Seconds())
	log.Debug().Msgf("max idle timeout: %v", config.HysteriaClientConfig.QUICConfig.MaxIdleTimeout.Seconds())

	d := &HysClient{
		tag:                      config.Tag,
		address:                  config.Address,
		dialer:                   config.Dialer,
		config:                   config.HysteriaClientConfig,
		serverPicker:             config.PortSelector,
		IpResolverForNodeAddress: config.IpResolverForNodeAddress,
		DomainStrategy:           config.DomainStrategy,
		RejectQuic:               config.RejectQuic,
		realmConfig:              config.Realm,
		connFactory:              connFactory,
		realmConnFactory:         realmConnFactory,
		salamanderPSK:            salamanderPSK,
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
	if time.Now().Unix()-c.lastActiveTime.Load() <= c.idle {
		log.Debug().Int32("id", c.id).Msg("hys client active")
		return true
	}
	// if runtime.GOOS == "ios" {
	// 	if time.Now().Unix()-c.lastActiveTime.Load() < 5 {
	// 		log.Debug().Int32("id", c.id).Msg("hys client active")
	// 		return true
	// 	}
	// } else {
	// 	if time.Now().Unix()-c.lastActiveTime.Load() < c.idle {
	// 		log.Debug().Int32("id", c.id).Msg("hys client active")
	// 		return true
	// 	}
	// }
	return false
}

func (w *wrappedClient) addTimer(hc *HysClient) {
	w.timerLock.Lock()
	defer w.timerLock.Unlock()
	if w.timer == nil {
		w.timer = time.AfterFunc(10*time.Second, func() {
			if w.dialing.Load() == 0 && w.usedSession.Load() == 0 {
				hc.removeClient(w)
			}
		})
	}
}

func (w *wrappedClient) removeTimer() {
	w.timerLock.Lock()
	defer w.timerLock.Unlock()
	if w.timer != nil {
		if !w.timer.Stop() {
			log.Warn().Int32("id", w.id).Msg("hys client timer not stopped")
		}
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
var defaultRealmSTUNServers = []string{
	"stun.nextcloud.com:3478",
	"stun.sip.us:3478",
	"global.stun.twilio.com:3478",
}

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
	factory := d.connFactory
	var idleTimer *atomic.Int64
	if d.realmConfig.RealmAddr != "" {
		if realmAddr, parseErr := d.parseRealmAddr(); parseErr == nil {
			return d.addNewRealmClient(id, &config, realmAddr)
		} else {
			return nil, parseErr
		}
	}

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
			conn, err1 := factory.New(udpAddr)
			if err1 != nil {
				return nil, fmt.Errorf("failed to create connection: %w", err1)
			}
			c := &ddlPacketConn{
				PacketConn: conn,
				id:         id,
				debug:      zerolog.GlobalLevel() == zerolog.DebugLevel,
				idle:       int64(config.QUICConfig.MaxIdleTimeout.Seconds())}
			c.lastReadTime.Store(time.Now().Unix())
			idleTimer = &c.lastReadTime
			config.ConnFactory = &connFactory{
				PacketConn: c}
			cl, _, err = client.NewClient(&config)
			if err == nil {
				log.Debug().Int32("id", id).Any("local_addr", conn.LocalAddr()).Msg("hys client created")
				break
			}
		}
	} else {
		udpAddr.IP = d.address.IP()
		config.ServerAddr = udpAddr
		conn, err1 := factory.New(udpAddr)
		if err1 != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err1)
		}
		c := &ddlPacketConn{
			PacketConn: conn,
			id:         id,
			debug:      zerolog.GlobalLevel() == zerolog.DebugLevel,
			idle:       int64(config.QUICConfig.MaxIdleTimeout.Seconds())}
		c.lastReadTime.Store(time.Now().Unix())
		idleTimer = &c.lastReadTime
		config.ConnFactory = &connFactory{
			PacketConn: c}
		cl, _, err = client.NewClient(&config)
		if err == nil {
			log.Debug().Int32("id", id).Any("local_addr", conn.LocalAddr()).Msg("hys client created")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	wrappedClient := &wrappedClient{
		Client: cl,
		id:     id,
		idle:   5,
	}
	wrappedClient.lastActiveTime = idleTimer

	// ccc.client = wrappedClient
	d.Lock()
	d.clients = append(d.clients, wrappedClient)
	log.Debug().Interface("client", cl).Int32("id", wrappedClient.id).Msg("new hys client")
	d.Unlock()
	return wrappedClient, nil
}

func (d *HysClient) parseRealmAddr() (*realm.Addr, error) {
	raw := d.realmConfig.RealmAddr
	addr, err := realm.ParseAddr(raw)
	if err == nil {
		return addr, nil
	}
	return nil, errors.New("invalid realm address")
}

func (d *HysClient) addNewRealmClient(id int32, config *client.Config, addr *realm.Addr) (*wrappedClient, error) {
	ctx := log.With().Uint32("sid", rand.Uint32()).Logger().WithContext(context.Background())

	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("realmServer", addr.HostPort).
		Str("scheme", addr.RendezvousScheme).Msg("realm client mode detected")
	family, network, err := realmIPMode(d.realmConfig.IPMode)
	if err != nil {
		return nil, err
	}
	baseConn, err := d.newRealmBaseConn(network)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("local", baseConn.LocalAddr().String()).
		Msg("realm client UDP socket opened")
	success := false
	defer func() {
		if !success {
			_ = baseConn.Close()
		}
	}()
	var mapper *realm.PortMapper
	if d.realmConfig.PortMapping.Enabled {
		if udpAddr, ok := baseConn.LocalAddr().(*net.UDPAddr); ok {
			mapper = newRealmPortMapper(ctx, addr.RealmID, udpAddr.Port, d.realmConfig.PortMapping)
			if mapper != nil {
				defer func() {
					if !success {
						_ = mapper.Close()
					}
				}()
			}
		}
	}
	localAddrs, err := d.realmDiscover(ctx, baseConn, addr, family, mapper)
	if err != nil {
		return nil, err
	}
	meta, err := realm.NewPunchMetadata()
	if err != nil {
		return nil, err
	}
	attempt := shortAttempt(meta.Nonce)
	rClient, err := realm.NewClientFromAddr(addr, d.realmHTTPClient())
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("attempt", attempt).
		Strs("addresses", addrPortStrings(localAddrs)).Msg("realm client connect request started")
	connectStart := time.Now()
	connectResp, err := rClient.Connect(ctx, addr.RealmID, realm.ConnectRequest{
		Addresses:     addrPortStrings(localAddrs),
		PunchMetadata: meta,
	})
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("attempt", attempt).
		Strs("serverAddresses", connectResp.Addresses).
		Str("duration", formatLogDuration(time.Since(connectStart))).Msg("realm client connect response received")
	peerAddrs, err := parseRealmAddrPorts(connectResp.Addresses)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("attempt", attempt).
		Strs("candidates", connectResp.Addresses).Msg("realm client punch started")
	punchStart := time.Now()
	result, err := realm.Punch(ctx, baseConn, localAddrs, peerAddrs, connectResp.PunchMetadata, realm.PunchConfig{
		Timeout: d.realmConfig.PunchTimeout,
		Family:  family,
	})
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("attempt", attempt).Str("peer", result.PeerAddr.String()).
		Str("packet", punchPacketTypeString(result.Packet.Type)).
		Str("duration", formatLogDuration(time.Since(punchStart))).Msg("realm client punch completed")
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Str("peer", result.PeerAddr.String()).
		Msg("realm client handing socket to QUIC")
	finalConn := baseConn
	if len(d.salamanderPSK) > 0 {
		finalConn, err = obfs.WrapPacketConnSalamander(finalConn, d.salamanderPSK)
		if err != nil {
			return nil, err
		}
	}
	if mapper != nil {
		mapCtx, mapCancel := context.WithCancel(context.Background())
		go realmPortMapLoop(mapCtx, addr.RealmID, mapper)
		finalConn = &cleanupPacketConn{PacketConn: finalConn, cleanup: mapCancel}
	}
	ddlConn := &ddlPacketConn{
		PacketConn: finalConn,
		id:         id,
		debug:      zerolog.GlobalLevel() == zerolog.DebugLevel,
		idle:       int64(config.QUICConfig.MaxIdleTimeout.Seconds()),
	}
	ddlConn.lastReadTime.Store(time.Now().Unix())
	config.ServerAddr = udpAddrFromAddrPort(result.PeerAddr)
	config.ConnFactory = &connFactory{
		PacketConn: ddlConn,
	}
	c, _, err := client.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create realm client: %w", err)
	}
	wrapped := &wrappedClient{
		Client: c,
		id:     id,
		idle:   5,
	}
	wrapped.lastActiveTime = &ddlConn.lastReadTime
	d.Lock()
	d.clients = append(d.clients, wrapped)
	d.Unlock()
	success = true
	return wrapped, nil
}

func (d *HysClient) newRealmBaseConn(network string) (net.PacketConn, error) {
	switch network {
	case "udp4":
		return d.realmConnFactory.New(&net.UDPAddr{IP: net.AnyIP.IP(),
			Port: int(d.realmConfig.LocalPort)})
	case "udp6":
		return d.realmConnFactory.New(&net.UDPAddr{IP: net.AnyIPv6.IP(),
			Port: int(d.realmConfig.LocalPort)})
	default:
		return d.realmConnFactory.New(&net.UDPAddr{Port: int(d.realmConfig.LocalPort)})
	}
}

func (d *HysClient) realmDiscover(ctx context.Context, baseConn net.PacketConn, addr *realm.Addr, family realm.AddrFamily, mapper *realm.PortMapper) ([]netip.AddrPort, error) {
	stunServers := d.realmSTUNServers(addr)
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Strs("stunServers", stunServers).Msg("realm client STUN discovery started")
	stunStart := time.Now()
	localAddrs, err := realm.Discover(ctx, baseConn, realm.STUNConfig{
		Servers:  stunServers,
		Timeout:  d.realmConfig.STUNTimeout,
		Family:   family,
		Resolver: &ipResolverAdapter{resolver: d.realmConfig.Resolver},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover STUN servers: %w", err)
	}
	log.Ctx(ctx).Debug().Str("realm", addr.RealmID).Strs("addresses", addrPortStrings(localAddrs)).
		Str("duration", formatLogDuration(time.Since(stunStart))).Msg("realm client STUN discovery completed")
	return withMappedAddr(localAddrs, mapper), nil
}

func (d *HysClient) realmSTUNServers(addr *realm.Addr) []string {
	if stunServers := addr.Params["stun"]; len(stunServers) > 0 {
		return append([]string(nil), stunServers...)
	}
	if len(d.realmConfig.STUNServers) > 0 {
		return append([]string(nil), d.realmConfig.STUNServers...)
	}
	return append([]string(nil), defaultRealmSTUNServers...)
}

func (d *HysClient) realmHTTPClient() *stdHTTP.Client {
	tr := stdHTTP.DefaultTransport.(*stdHTTP.Transport).Clone()
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		log.Ctx(ctx).Debug().Str("addr", addr).Msg("dialing")
		dest, err := net.ParseDestination(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse destination: %w", err)
		}
		dest.Network = net.Network_TCP
		return d.dialer.Dial(ctx, dest)
	}
	tr.TLSClientConfig.InsecureSkipVerify = d.realmConfig.Insecure
	return &stdHTTP.Client{Transport: tr}
}

func parseRealmAddrPorts(values []string) ([]netip.AddrPort, error) {
	addrs := make([]netip.AddrPort, 0, len(values))
	for _, value := range values {
		addr, err := netip.ParseAddrPort(value)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func addrPortStrings(addrs []netip.AddrPort) []string {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		out = append(out, addr.String())
	}
	return out
}

func udpAddrFromAddrPort(addr netip.AddrPort) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IP(addr.Addr().AsSlice()),
		Port: int(addr.Port()),
	}
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
	debug        bool
}

func (c *ddlPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(p)
	if err != nil {
		return n, addr, err
	}
	c.lastReadTime.Store(time.Now().Unix())
	if c.debug {
		// log.Debug().Int32("id", c.id).Msg("hys client read from")
	}
	return n, addr, err
}

func (c *ddlPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if c.debug {
		if time.Now().Unix()-c.lastReadTime.Load() > c.idle {
			log.Debug().Int32("id", c.id).Msg("hys client no read activity but still sending data")
		}
		// log.Debug().Int32("id", c.id).Msg("hys client write to")
	}
	return c.PacketConn.WriteTo(p, addr)
}

type hysConnFactory struct {
	packetListener i.PacketListener
	salamanderPSK  []byte
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
		if ip == "" {
			network = "udp"
		} else if net.ParseAddress(ip).Family().IsIPv6() {
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
	if len(c.salamanderPSK) > 0 {
		var err error
		conn, err = obfs.WrapPacketConnSalamander(conn, c.salamanderPSK)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

var ErrRejectQuic = errors.New("reject quic over hysteria2")

func (d *HysClient) dialCommon(ctx context.Context, dst net.Destination) (net.Conn, *wrappedClient, error) {
	var conn net.Conn
	var err error
	var wrappedClient *wrappedClient
	if dst.Network == net.Network_UDP {
		if dst.Port == 443 && d.RejectQuic {
			return nil, nil, ErrRejectQuic
		}
		var udpConn client.HyUDPConn
		udpConn, wrappedClient, err = d.udp(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to dial UDP: %w", err)
		}
		// if d.IpResolverForTargetAddress != nil && dst.Address.Family().IsDomain() {
		// 	// TODO: should consider whether server support ipv6
		// 	ips, _ := d.IpResolverForTargetAddress.LookupIP(
		// 		ctx, dst.Address.Domain())
		// 	if len(ips) > 0 {
		// 		target.Address = net.IPAddress(ips[rand.Intn(len(ips))])
		// 	}
		// }
		conn = &HyUdpConnToNetConn{addr: dst.NetAddr(), hyUdpConn: udpConn}
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
	data, src, err := c.hyUdpConn.Receive()
	if err != nil {
		return 0, fmt.Errorf("failed to receive: %w", err)
	}
	if src != c.addr {
		log.Warn().Str("src", src).Str("addr", c.addr).Msg("hys udp conn to net conn src not equal to addr")
	}
	n = copy(p, data)
	return n, nil
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
