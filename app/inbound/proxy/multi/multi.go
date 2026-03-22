// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

//go:build server

package multi

import (
	"context"
	gotls "crypto/tls"
	"errors"
	"slices"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/rs/zerolog/log"
	"github.com/xtls/reality"

	"github.com/5vnetwork/vx-core/common/buf"
	protocolHttp "github.com/5vnetwork/vx-core/common/protocol/http"
	"github.com/5vnetwork/vx-core/common/strmatcher"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	transportHttp "github.com/5vnetwork/vx-core/transport/protocols/http"
	securityreality "github.com/5vnetwork/vx-core/transport/security/reality"
	securitytls "github.com/5vnetwork/vx-core/transport/security/tls"
)

func NewMultiInboundServer(config *configs.MultiProxyInboundConfig, ha i.Handler,
	tp i.TimeoutSetting, inStats *monitor.Stats, onUnauth i.UnauthorizedReport) (proxy.Inbound, error) {
	ports := make([]uint16, 0, 10)
	for _, port := range config.Ports {
		ports = append(ports, uint16(port))
	}

	servers, hysteriaConfig, err := create.GetServers(config.Users, config.Protocols,
		ha, tp, onUnauth)
	if err != nil {
		return nil, err
	}

	if config.GetAddress() == "" {
		config.Address = net.AnyIP.String()
	}

	// proxy inbound
	h := proxy.NewProxyInbound(config.Tag)
	for _, server := range servers {
		if i, ok := server.(proxy.UserManage); ok {
			h.AddUserManage(i)
		}
	}

	if hysteriaConfig != nil {
		in, err := hysteria2.NewInbound(&hysteria2.InboundConfig{
			Ports:                 ports,
			Hysteria2ServerConfig: hysteriaConfig,
			InStats:               inStats,
			Tag:                   config.Tag,
			OnUnauthorizedRequest: onUnauth,
			Handler:               ha,
		})
		if err != nil {
			return nil, err
		}
		for _, u := range append(hysteriaConfig.Users, config.Users...) {
			user, err := create.UserConfigToUser(u)
			if err != nil {
				return nil, err
			}
			in.AddUser(user)
		}
		h.AddWorker(in)
		h.AddUserManage(in)
	}

	address := net.ParseAddress(config.Address)

	for _, port := range ports {
		multi := &Multi{
			addr:     &net.TCPAddr{IP: address.IP(), Port: int(port)},
			connChan: make(chan net.Conn, 128),
			done:     done.New(),
			sniffer: sniff.NewSniffer(
				sniff.SniffSetting{
					Interval: 100 * time.Millisecond,
					Sniffers: []sniff.ProtocolSnifferWithNetwork{
						sniff.TlsSniff,
					},
				}),
		}
		for _, security := range config.SecurityConfigs {
			switch s := security.Security.(type) {
			case *configs.MultiProxyInboundConfig_Security_Tls:
				indexMatcher, err := toIndexMatcher(security.Domains, security.RegularExpression)
				if err != nil {
					return nil, err
				}
				c, err := securitytls.GetTLSConfig(s.Tls, securitytls.WithNextProtocol([]string{"h2", "http/1.1"}))
				if err != nil {
					return nil, err
				}
				multi.securitys = append(multi.securitys, indexMatcherSecurity{
					indexMatcher: indexMatcher,
					always:       security.Always,
					security: &tlsConfig{
						tlsConfig: c,
					},
				})
			case *configs.MultiProxyInboundConfig_Security_Reality:
				indexMatcher, err := toIndexMatcher(security.Domains, security.RegularExpression)
				if err != nil {
					return nil, err
				}
				multi.securitys = append(multi.securitys, indexMatcherSecurity{
					indexMatcher: indexMatcher,
					always:       security.Always,
					security: &realityConfig{
						realityConfig: securityreality.GetREALITYConfig(s.Reality),
					},
				})
			}
		}
		for _, protocol := range config.TransportProtocols {
			listenerAdapter := &listenerAdapter{
				addr:    multi.addr,
				channel: make(chan net.Conn, 128),
				done:    done.New(),
				condition: condition{
					alpn:   protocol.GetAlpn(),
					sni:    protocol.GetSni(),
					path:   protocol.GetPath(),
					h2:     protocol.GetH2(),
					always: protocol.GetAlways(),
				},
			}
			switch p := protocol.GetProtocol().(type) {
			case *configs.MultiProxyInboundConfig_Protocol_Websocket:
				l, err := websocket.Listen(context.Background(), net.DestinationFromAddr(multi.addr),
					p.Websocket, listenerAdapter, multi.handleConn)
				if err != nil {
					return nil, err
				}
				listenerAdapter.listener = l
			case *configs.MultiProxyInboundConfig_Protocol_Httpupgrade:
				l, err := httpupgrade.Listen(context.Background(), net.DestinationFromAddr(multi.addr),
					p.Httpupgrade, listenerAdapter, multi.handleConn)
				if err != nil {
					return nil, err
				}
				listenerAdapter.listener = l
			case *configs.MultiProxyInboundConfig_Protocol_Splithttp:
				l, err := splithttp.ListenXH(context.Background(), net.DestinationFromAddr(multi.addr),
					p.Splithttp, listenerAdapter, multi.handleConn)
				if err != nil {
					return nil, err
				}
				listenerAdapter.listener = l
			case *configs.MultiProxyInboundConfig_Protocol_Http:
				l, err := transportHttp.Listen(context.Background(), net.DestinationFromAddr(multi.addr),
					p.Http, listenerAdapter, multi.handleConn)
				if err != nil {
					return nil, err
				}
				listenerAdapter.listener = l
			case *configs.MultiProxyInboundConfig_Protocol_Grpc:
				l, err := grpc.Listen(context.Background(), net.DestinationFromAddr(multi.addr), p.Grpc,
					listenerAdapter, multi.handleConn)
				if err != nil {
					return nil, err
				}
				listenerAdapter.listener = l
			}
			multi.protocols = append(multi.protocols, listenerAdapter)
		}

		// proxy protocols
		hasHys := hysteriaConfig != nil
		var tcpServers []proxy.ProxyServer
		var udpServers []proxy.ProxyServer
		for _, server := range servers {
			if slices.Contains(server.Network(), net.Network_TCP) {
				tcpServers = append(tcpServers, server)
			}
			if slices.Contains(server.Network(), net.Network_UDP) {
				udpServers = append(udpServers, server)
			}
		}
		if len(tcpServers) > 0 {
			var connHandler proxy.ConnHandler
			if len(tcpServers) == 1 {
				connHandler = tcpServers[0]
			} else {
				proxyServers := &proxy.ProxyServers{}
				for _, server := range tcpServers {
					if fp, ok := server.(proxy.FallbackProxyServer); ok {
						proxyServers.FallbackProxyServers = append(proxyServers.FallbackProxyServers, fp)
					} else {
						if proxyServers.ProxyServer != nil {
							log.Warn().Msg("there are two non-fallback proxy servers for the same port")
						}
						proxyServers.ProxyServer = server
					}
				}
				// if there is no non-fallback proxy server, make the last fallback server as it
				if proxyServers.ProxyServer == nil && len(proxyServers.FallbackProxyServers) > 0 {
					proxyServers.ProxyServer = proxyServers.FallbackProxyServers[len(proxyServers.FallbackProxyServers)-1]
					proxyServers.FallbackProxyServers[len(proxyServers.FallbackProxyServers)-1] = nil
					proxyServers.FallbackProxyServers = proxyServers.FallbackProxyServers[:len(proxyServers.FallbackProxyServers)-1]
				}
				connHandler = proxyServers
			}
			tcpWorker := proxy.NewTcpWorker(proxy.TcpWorkerConfig{
				Addr:        &net.TCPAddr{IP: address.IP(), Port: int(port)},
				Listener:    multi,
				Tag:         h.Tag(),
				ConnHandler: connHandler,
			})
			h.AddWorker(tcpWorker)
		}
		if !hasHys && len(udpServers) > 0 {
			udpWorker := proxy.NewUdpWorker(proxy.UdpWorkerConfig{
				Addr:        &net.UDPAddr{IP: address.IP(), Port: int(port)},
				Tag:         h.Tag(),
				ConnHandler: udpServers[0],
				Listener:    create.SocketConfigToMemoryConfig(config.GetSocket(), nil, nil),
			})
			h.AddWorker(udpWorker)
		}
	}
	return h, nil
}

func toIndexMatcher(domains []string, regularExpression string) (strmatcher.IndexMatcher, error) {
	indexMatcher := strmatcher.NewMphIndexMatcher()
	for _, domain := range domains {
		matcher, err := strmatcher.Full.New(domain)
		if err != nil {
			return nil, err
		}
		indexMatcher.Add(matcher)
	}
	if regularExpression != "" {
		matcher, err := strmatcher.Regex.New(regularExpression)
		if err != nil {
			return nil, err
		}
		indexMatcher.Add(matcher)
	}
	if err := indexMatcher.Build(); err != nil {
		return nil, err
	}
	return indexMatcher, nil
}

type Multi struct {
	addr net.Addr
	// value is either security or protocol
	securitys    []indexMatcherSecurity
	protocols    []*listenerAdapter
	connChan     chan net.Conn
	done         *done.Instance
	socketConfig *dlhelper.SocketSetting
	sniffer      *sniff.Sniffer
	listener     net.Listener
}

type condition struct {
	alpn   string
	sni    string
	path   string
	h2     bool
	always bool
}

func (m *Multi) Addr() net.Addr {
	return m.addr
}

func (m *Multi) Listen(ctx context.Context, addr net.Addr) (net.Listener, error) {
	m.addr = addr
	listener, err := m.socketConfig.Listen(ctx, addr)
	if err != nil {
		return nil, err
	}
	m.listener = listener
	go m.keepAccepting()
	return m, nil
}

func (m *Multi) keepAccepting() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			return
		}
		go m.process(conn)
	}
}

func (m *Multi) Close() error {
	m.done.Close()
	m.listener.Close()
	for _, protocol := range m.protocols {
		protocol.close()
	}
	return nil
}

func (m *Multi) Accept() (net.Conn, error) {
	select {
	case conn := <-m.connChan:
		return conn, nil
	case <-m.done.Wait():
		return nil, errors.New("listener closed")
	}
}

func (m *Multi) process(conn net.Conn) {

	sni := ""
	// sniff tls
	if len(m.securitys) > 0 && !m.securitys[0].always {
		// read at least once
		b := buf.New()
		_, err := b.ReadOnce(conn)
		if err != nil {
			log.Error().Err(err).Msg("failed to read from conn")
			conn.Close()
			b.Release()
			return
		}
		conn = net.NewMbConn(conn, buf.MultiBuffer{b})

		var result sniff.SniffResult
		conn, result, err = m.sniffer.SniffConn(context.Background(), conn)
		if err == nil {
			if tlsResult, ok := result.(*tls.SniffHeader); ok {
				sni = tlsResult.Domain()
			} else {
				log.Debug().Err(err).Msg("not tls")
			}
		} else {
			log.Debug().Err(err).Msg("failed to sniff")
		}
	}

	alpn := ""
	handshakeAddress := ""
	// security layer: apply security to the conn if matched
	if len(m.securitys) > 0 {
		var securityToApply security
		if m.securitys[0].always {
			securityToApply = m.securitys[0].security
		} else {
			for _, security := range m.securitys {
				if security.indexMatcher.MatchAny(sni) {
					securityToApply = security.security
					break
				}
			}
		}
		if securityToApply != nil {
			secureConn, err := securityToApply.Server(conn)
			if err != nil {
				log.Error().Err(err).Msg("failed to server tls")
				conn.Close()
				return
			}
			alpn, err = secureConn.GetConnectionApplicationProtocol()
			if err != nil {
				log.Error().Err(err).Msg("failed to get connection application protocol")
				conn.Close()
				return
			}
			if serverAddress := secureConn.HandshakeAddress(); serverAddress != nil {
				handshakeAddress = serverAddress.String()
			}
			conn = secureConn
		}
	}

	h2 := false
	h1Path := ""
	// sniff path and h2
	b := buf.New()
	_, err := b.ReadOnce(conn)
	if err != nil {
		log.Error().Err(err).Msg("failed to read from conn")
		conn.Close()
		b.Release()
		return
	}
	conn = net.NewMbConn(conn, buf.MultiBuffer{b})
	sniffResult, err := protocolHttp.SniffHttp1(b.Bytes())
	if err == nil {
		h1Path = sniffResult.Path()
	} else {
		log.Debug().Err(err).Msg("not http1")
		h2 = protocolHttp.IsHttp2(b.Bytes())
	}

	// transport protocol layer: apply transport protocol to the conn
	if len(m.protocols) > 0 {
		var protocolToApply *listenerAdapter
		if m.protocols[0].condition.always {
			protocolToApply = m.protocols[0]
		} else {
			for _, protocol := range m.protocols {
				if protocol.condition.alpn != "" && protocol.condition.alpn != alpn {
					continue
				}
				if protocol.condition.sni != "" && protocol.condition.sni != handshakeAddress {
					continue
				}
				if protocol.condition.path != "" && protocol.condition.path != h1Path {
					continue
				}
				if protocol.condition.h2 && !h2 {
					continue
				}
				protocolToApply = protocol
				break
			}
		}
		if protocolToApply != nil {
			protocolToApply.process(conn)
			return
		}
	}
	// hand the conn to the upper layer
	m.connChan <- conn
}

func (m *Multi) handleConn(conn net.Conn) {
	m.connChan <- conn
}

type secureConn interface {
	net.Conn
	GetConnectionApplicationProtocol() (string, error)
	HandshakeAddress() net.Address
}

type security interface {
	Server(conn net.Conn) (secureConn, error)
}

type indexMatcherSecurity struct {
	indexMatcher strmatcher.IndexMatcher
	always       bool
	security     security
}

type tlsConfig struct {
	tlsConfig *gotls.Config
}

func (t *tlsConfig) Server(conn net.Conn) (secureConn, error) {
	c := gotls.Server(conn, t.tlsConfig)
	return &securitytls.Conn{Conn: c}, nil
}

type realityConfig struct {
	realityConfig *reality.Config
}

func (r *realityConfig) Server(conn net.Conn) (secureConn, error) {
	return securityreality.Server(conn, r.realityConfig)
}

type listenerAdapter struct {
	addr net.Addr
	// input channel
	channel   chan net.Conn
	done      *done.Instance
	listener  transport.Listener
	condition condition
}

func (n *listenerAdapter) process(conn net.Conn) {
	select {
	case n.channel <- conn:
	default:
		conn.Close()
	}
}

func (n *listenerAdapter) Listen(ctx context.Context, addr net.Addr) (net.Listener, error) {
	return n, nil
}

func (n *listenerAdapter) Addr() net.Addr {
	return n.addr
}

func (n *listenerAdapter) Accept() (net.Conn, error) {
	select {
	case conn := <-n.channel:
		return conn, nil
	case <-n.done.Wait():
		return nil, errors.New("listener closed")
	}
}

func (n *listenerAdapter) Close() error {
	n.done.Close()
	return nil
}

func (n *listenerAdapter) close() {
	n.listener.Close()
}
