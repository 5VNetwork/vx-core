// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package hysteria2

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/hysteria2/internal/server"
	tlssec "github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/apernet/hysteria/extras/v2/correctnet"
	"github.com/apernet/hysteria/extras/v2/obfs"
	"github.com/apernet/hysteria/extras/v2/realm"
	"github.com/rs/zerolog/log"
)

type Inbound struct {
	config *InboundConfig
	server []server.Server

	usersLock sync.RWMutex
	users     map[string]i.User // secret to User

	cLock      sync.RWMutex
	srcAddrMap map[netip.AddrPort]*srcAddrInfo

	cleanup []io.Closer

	realmRuntime   *realmServerRuntime
	realmCtxCancel context.CancelFunc

	onUnauthorizedRequest i.UnauthorizedReport
	handler               i.Handler
}

type srcAddrInfo struct {
	counter *atomic.Uint64
}

func NewInbound(config *InboundConfig) (*Inbound, error) {
	in := &Inbound{
		config:                config,
		users:                 make(map[string]i.User),
		srcAddrMap:            make(map[netip.AddrPort]*srcAddrInfo),
		onUnauthorizedRequest: config.OnUnauthorizedRequest,
		handler:               config.Handler,
	}
	return in, nil
}

type InboundConfig struct {
	*configs.Hysteria2ServerConfig
	Addresses             []string
	Ports                 []uint16
	Tag                   string
	OnUnauthorizedRequest i.UnauthorizedReport
	Handler               i.Handler
	Listener              i.PacketListener
	Realm                 RealmConfig
}

func (in *Inbound) Tag() string {
	return in.config.Tag
}

func (in *Inbound) AddUser(user i.User) {
	in.usersLock.Lock()
	defer in.usersLock.Unlock()
	in.users[user.Secret()] = user
}

func (in *Inbound) RemoveUser(user i.User) {
	in.usersLock.Lock()
	defer in.usersLock.Unlock()
	delete(in.users, user.Secret())
}

func (in *Inbound) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	in.onUnauthorizedRequest = f
}

func (in *Inbound) Start() error {
	config := in.config

	tlsConfig, err := tlssec.GetTLSConfig(config.Hysteria2ServerConfig.TlsConfig)
	if err != nil {
		return err
	}

	var salamanderPSK []byte
	if in.config.GetObfs().GetSalamander() != nil {
		psk := []byte(config.GetObfs().GetSalamander().Password)
		if len(psk) < 4 {
			return fmt.Errorf("failed to create obfuscator: password too short")
		}
		salamanderPSK = psk
	}

	if in.config.Realm.RealmAddr != "" {
		if realmAddr, ok, err := parseServerRealmAddr(in.config.Realm.RealmAddr); ok || err != nil {
			if err != nil {
				return err
			}
			ctx, cancel := context.WithCancel(context.Background())
			in.realmCtxCancel = cancel
			go func() {
				for ctx.Err() == nil {
					if err := in.startRealmServer(ctx, realmAddr,
						int(in.config.Realm.LocalPort), salamanderPSK,
						tlsConfig, in.config.Realm); err != nil {
						log.Error().Err(err).Msg("failed to start realm server")
					}
				}
			}()
		}
	} else {
		if len(config.Addresses) == 0 {
			config.Addresses = []string{""}
		}
		for _, addr := range config.Addresses {

			network := "udp"
			if strings.HasPrefix(addr, "udp6:") {
				network = "udp6"
				addr = strings.TrimPrefix(addr, "udp6:")
			} else if strings.HasPrefix(addr, "udp4:") {
				network = "udp4"
				addr = strings.TrimPrefix(addr, "udp4:")
			}
			log.Info().Msgf("hysteria2 listen on %s network %s", addr, network)
			for _, p := range config.Ports {
				var pc net.PacketConn
				var err error
				if in.config.Listener != nil {
					pc, err = in.config.Listener.ListenPacket(context.Background(),
						network, net.JoinHostPort(addr, fmt.Sprintf("%d", p)))
				} else {
					pc, err = net.ListenPacket(network, net.JoinHostPort(addr, fmt.Sprintf("%d", p)))
				}
				if err != nil {
					return err
				}
				pc = &statsPacketConn{
					PacketConn: pc,
					inbound:    in,
				}
				if len(salamanderPSK) > 0 {
					pc, err = obfs.WrapPacketConnSalamander(pc, salamanderPSK)
					if err != nil {
						_ = pc.Close()
						return err
					}
				}
				s, err := server.NewServer(in.newHysServerConfig(tlsConfig, pc))
				if err != nil {
					return err
				}
				in.startHysServer(s)
			}
		}
	}

	return nil
}

func (in *Inbound) Close() error {
	items := make([]interface{}, 0, len(in.server)+len(in.cleanup))
	for _, s := range in.server {
		items = append(items, s)
	}
	for _, c := range in.cleanup {
		items = append(items, c)
	}
	if in.realmCtxCancel != nil {
		in.realmCtxCancel()
	}
	return common.CloseAll(items...)
}

func (in *Inbound) Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, user i.User) {
	// segments := strings.Split(auth, ":")
	// if len(segments) != 2 {
	// 	return false, ""
	// }
	// uid := segments[0]
	// secret := segments[1]
	in.usersLock.RLock()
	defer in.usersLock.RUnlock()
	user, ok = in.users[auth]
	if !ok {
		if in.onUnauthorizedRequest != nil {
			in.onUnauthorizedRequest.ReportUnauthorized(addr.String(), auth)
		}
		return false, nil
	}

	return true, user
}

func (in *Inbound) LogOnlineState(user i.User, online bool) {}

// rx is what received from final dest, and the data will be sent back to client.
// tx is the data sent to the final dest, in other words, the data received from the client
// data sent back to client is rx, data received from client is tx
// send(to client) traffic meter happens at the transport layer, so we don't need to count it here
func (in *Inbound) LogTraffic(user i.User, tx, rx uint64) (ok bool) {
	if tx != 0 {
		in.usersLock.RLock()
		defer in.usersLock.RUnlock()
		user, ok := in.users[user.Secret()]
		if !ok {
			return false
		}
		user.Counter().Add(tx + rx)
	}
	return true
}

func (in *Inbound) TraceStream(stream server.HyStream, stats *server.StreamStats) {
}
func (in *Inbound) UntraceStream(stream server.HyStream) {
}

// Implements server.EventLogger
func (in *Inbound) Connect(addr net.Addr, secret string, tx uint64) {
	in.usersLock.RLock()
	defer in.usersLock.RUnlock()
	user, ok := in.users[secret]
	if !ok {
		return
	}
	log.Debug().Any("src_addr", addr).Str("user_id", user.Uid()).
		Uint64("tx", tx).Msg("hysteria2 connect")

	in.cLock.Lock()
	in.srcAddrMap[peerAddrMapKeyFrom(addr)] = &srcAddrInfo{
		counter: user.Counter(),
	}
	in.cLock.Unlock()
}

func (in *Inbound) Disconnect(addr net.Addr, id string, err error) {
	in.cLock.Lock()
	delete(in.srcAddrMap, peerAddrMapKeyFrom(addr))
	in.cLock.Unlock()

	log.Debug().Err(err).Any("src_addr", addr).Str("user_id", id).Msgf("hysteria2 disconnect")
}

func (in *Inbound) TCPRequest(addr net.Addr, id, reqAddr string) {
	log.Debug().Any("src_addr", addr).Str("user_id", id).Str("req_addr", reqAddr).Msgf("hysteria2 tcp request")
}
func (in *Inbound) TCPError(addr net.Addr, id, reqAddr string, err error) {
	log.Debug().Err(err).Any("src_addr", addr).Str("user_id", id).Str("req_addr", reqAddr).Msgf("hysteria2 tcp error")
}
func (in *Inbound) UDPRequest(addr net.Addr, id string, sessionID uint32, reqAddr string) {
	log.Debug().Any("src_addr", addr).Str("user_id", id).Uint32("session_id", sessionID).
		Str("req_addr", reqAddr).Msgf("hysteria2 udp request")
}
func (in *Inbound) UDPError(addr net.Addr, id string, sessionID uint32, err error) {
	log.Debug().Err(err).Any("src_addr", addr).Str("user_id", id).Uint32("session_id", sessionID).Msgf("hysteria2 udp error")
}

type statsPacketConn struct {
	net.PacketConn

	inbound *Inbound
}

// TODO: more efficient
func (c *statsPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := c.PacketConn.WriteTo(b, addr)
	if err != nil {
		return 0, err
	}

	c.inbound.cLock.RLock()
	defer c.inbound.cLock.RUnlock()
	info, ok := c.inbound.srcAddrMap[peerAddrMapKeyFrom(addr)]
	if !ok {
		// log.Debug().Str("addr", addr.String()).Msgf("hysteria2 src addr not found")
		return n, nil
	}
	info.counter.Add(uint64(n))

	// log.Debug().Str("src_addr", addr.String()).Uint64("traffic", info.counter.Load()).Msgf("hysteria2 traffic")
	return n, nil
}

func parseServerRealmAddr(listen string) (*realm.Addr, bool, error) {
	addr, err := realm.ParseAddr(listen)
	if err == nil {
		return addr, true, nil
	}
	if strings.HasPrefix(listen, realm.SchemeHTTPS+":") || strings.HasPrefix(listen, realm.SchemeHTTP+":") {
		return nil, true, err
	}
	return nil, false, nil
}

func (in *Inbound) startRealmServer(ctx context.Context, addr *realm.Addr, port int,
	salamanderPSK []byte, tlsConfig *tls.Config, realmCfg RealmConfig) error {
	log.Debug().Str("realm", addr.RealmID).Str("realmServer", addr.HostPort).
		Str("scheme", addr.RendezvousScheme).Msg("realm server mode detected")
	family, network, err := realmIPMode(realmCfg.IPMode)
	if err != nil {
		return err
	}
	listenAddr := &net.UDPAddr{Port: port}
	if addr.LocalPort != 0 {
		listenAddr.Port = addr.LocalPort
	}
	var baseConn net.PacketConn
	if in.config.Listener != nil {
		baseConn, err = in.config.Listener.ListenPacket(ctx, network, listenAddr.String())
	} else {
		baseConn, err = correctnet.ListenUDP(network, listenAddr)
	}
	if err != nil {
		return err
	}
	log.Debug().Str("realm", addr.RealmID).Str("local", baseConn.LocalAddr().String()).Msg("realm server UDP socket opened")
	punchConn, err := realm.NewPunchPacketConn(baseConn, 0)
	if err != nil {
		_ = baseConn.Close()
		return fmt.Errorf("failed to create punch packet conn: %w", err)
	}
	var pc net.PacketConn = punchConn
	if len(salamanderPSK) > 0 {
		pc, err = obfs.WrapPacketConnSalamander(pc, salamanderPSK)
		if err != nil {
			_ = pc.Close()
			return fmt.Errorf("failed to wrap packet conn with salamander: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	rt, err := startRealmServerRuntime(ctx, cancel, addr, punchConn, family, realmCfg)
	if err != nil {
		cancel()
		_ = pc.Close()
		return fmt.Errorf("failed to start realm server runtime: %w", err)
	}
	in.cleanup = append(in.cleanup, rt)
	in.realmRuntime = rt

	s, err := server.NewServer(in.newHysServerConfig(tlsConfig, pc))
	if err != nil {
		cancel()
		_ = pc.Close()
		return fmt.Errorf("failed to create hysteria2 server: %w", err)
	}
	in.startHysServer(s)
	return nil
}

func (in *Inbound) newHysServerConfig(tlsConfig *tls.Config, pc net.PacketConn) *server.Config {
	config := in.config
	quic := config.GetQuic()
	bandwidth := config.GetBandwidth()
	return &server.Config{
		Tag: in.config.Tag,
		TLSConfig: server.TLSConfig{
			Certificates:             tlsConfig.Certificates,
			EncryptedClientHelloKeys: tlsConfig.EncryptedClientHelloKeys,
		},
		QUICConfig: server.QUICConfig{
			InitialStreamReceiveWindow:     uint64(quic.GetInitialStreamReceiveWindow()) * 1024 * 1024,
			MaxStreamReceiveWindow:         uint64(quic.GetMaxStreamReceiveWindow()) * 1024 * 1024,
			InitialConnectionReceiveWindow: uint64(quic.GetInitialConnectionReceiveWindow()) * 1024 * 1024,
			MaxConnectionReceiveWindow:     uint64(quic.GetMaxConnectionReceiveWindow()) * 1024 * 1024,
			MaxIdleTimeout:                 time.Duration(quic.GetMaxIdleTimeout()) * time.Second,
			DisablePathMTUDiscovery:        quic.GetDisablePathMtuDiscovery(),
			MaxIncomingStreams:             int64(quic.GetMaxIncomingStreams()),
		},
		Conn:     pc,
		Outbound: in.handler,
		BandwidthConfig: server.BandwidthConfig{
			MaxTx: uint64(bandwidth.GetMaxTx() * 1024 * 1024),
			MaxRx: uint64(bandwidth.GetMaxRx() * 1024 * 1024),
		},
		IgnoreClientBandwidth: config.GetIgnoreClientBandwidth(),
		Authenticator:         in,
		TrafficLogger:         in,
		EventLogger:           in,
	}
}

func (in *Inbound) startHysServer(s server.Server) {
	in.server = append(in.server, s)
	go func() {
		if err := s.Serve(); err != nil {
			log.Error().Err(err).Msg("hysteria2 server serve error")
		}
	}()
}

// RealmStatus returns a snapshot of the realm server state.
// active is false when no realm server is running (realmRuntime is nil or session is empty).
// peers counts authenticated QUIC clients that connected via a successful realm STUN punch.
func (in *Inbound) RealmStatus() (active bool, realmID string, publicAddrs []string, peers int) {
	log.Debug().Msg("getting realm status")
	if in.realmRuntime == nil {
		log.Debug().Msg("realm runtime is nil")
		return false, "", nil, 0
	}
	sess := in.realmRuntime.currentSession()
	raw := in.realmRuntime.currentAddrs()
	in.cLock.RLock()
	log.Debug().Int("peers", len(in.srcAddrMap)).Msg("src addr map length")
	peers = in.realmRuntime.activePunchedPeers(in.srcAddrMap)
	in.cLock.RUnlock()
	return sess.id != "", in.realmRuntime.realmID, addrPortStrings(raw), peers
}
