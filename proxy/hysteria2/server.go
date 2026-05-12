// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

//go:build server

package hysteria2

import (
	"context"
	"fmt"
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

	"github.com/apernet/hysteria/extras/v2/obfs"
	"github.com/rs/zerolog/log"
)

type Inbound struct {
	config *InboundConfig
	server []server.Server

	usersLock sync.RWMutex
	users     map[string]i.User // secret to User

	cLock      sync.RWMutex
	srcAddrMap map[netip.AddrPort]*srcAddrInfo

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

	var obfuscator obfs.Obfuscator
	if in.config.GetObfs().GetSalamander() != nil {
		obfuscator, err = obfs.NewSalamanderObfuscator(
			[]byte(config.GetObfs().GetSalamander().Password))
		if err != nil {
			return fmt.Errorf("failed to create obfuscator: %w", err)
		}
	}

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
			if obfuscator != nil {
				pc = obfs.WrapPacketConn(pc, obfuscator)
			}
			hysConfig := &server.Config{
				Tag: in.config.Tag,
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
				Conn:     pc,
				Outbound: in.handler,
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
	}
	return nil
}

func (in *Inbound) Close() error {
	return common.CloseAll(in.server)
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
func (in *Inbound) Connect(addr net.Addr, id string, tx uint64) {
	log.Debug().Any("src_addr", addr).Str("user_id", id).Uint64("tx", tx).Msgf("hysteria2 connect")
	in.usersLock.RLock()
	defer in.usersLock.RUnlock()
	user, ok := in.users[id]
	if !ok {
		return
	}

	in.cLock.Lock()
	in.srcAddrMap[addr.(*net.UDPAddr).AddrPort()] = &srcAddrInfo{
		counter: user.Counter(),
	}
	in.cLock.Unlock()
}

func (in *Inbound) Disconnect(addr net.Addr, id string, err error) {
	in.cLock.Lock()
	delete(in.srcAddrMap, addr.(*net.UDPAddr).AddrPort())
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
	info, ok := c.inbound.srcAddrMap[addr.(*net.UDPAddr).AddrPort()]
	if !ok {
		// log.Debug().Str("addr", addr.String()).Msgf("hysteria2 src addr not found")
		return n, nil
	}
	info.counter.Add(uint64(n))

	// log.Debug().Str("src_addr", addr.String()).Uint64("traffic", info.counter.Load()).Msgf("hysteria2 traffic")
	return n, nil
}
