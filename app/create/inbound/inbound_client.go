// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package inbound

import (
	"math/rand"
	"slices"

	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/dokodemo"
	"github.com/5vnetwork/vx-core/proxy/http"
	"github.com/5vnetwork/vx-core/proxy/socks"
)

func NewInbound(config *configs.ProxyInboundConfig, ha i.Handler, tp i.TimeoutSetting) (proxy.Inbound, error) {
	if len(config.Protocols) == 0 {
		config.Protocols = append(config.Protocols, config.Protocol)
	}

	var servers []proxy.ProxyServer
	for _, protocol := range config.Protocols {
		var server proxy.ProxyServer
		serverConfig, err := serial.GetInstanceOf(protocol)
		if err != nil {
			return nil, fmt.Errorf("failed to get instance of ProxyServerConfig: %w", err)
		}
		switch c := serverConfig.(type) {
		case *configs.DokodemoConfig:
			server = dokodemo.New(
				dokodemo.DoorSettings{
					Address:  net.ParseAddress(c.Address),
					Port:     net.Port(c.Port),
					Networks: c.Networks,
					Handler:  ha,
				},
			)
			servers = append(servers, server)
		case *configs.SocksServerConfig:
			config := &socks.SocksServerConfig{
				UdpEnabled: c.UdpEnabled,
				AuthType:   c.AuthType,
				Policy:     tp,
				Handler:    ha,
			}
			if c.Address != "" {
				config.Address = net.ParseAddress(c.Address)
			}
			server = socks.NewServer(config)
			for _, u := range c.Accounts {
				user, err := UserConfigToUser(u)
				if err != nil {
					return nil, err
				}
				server.(*socks.Server).AddUser(user)
			}
			servers = append(servers, server)
		case *configs.HttpServerConfig:
			server = http.NewServer(http.ServerSettings{
				PolicyManager: tp,
				Handler:       ha,
			})
			servers = append(servers, server)
		default:
			return nil, fmt.Errorf("unknown z proxy server config: %T", c)
		}
	}

	if config.GetAddress() == "" {
		config.Address = net.AnyIP.String()
	}

	// proxy inbound
	h := proxy.NewProxyInbound(config.Tag)
	address := net.ParseAddress(config.Address)
	transport := create.TransportConfigToMemoryConfig(config.GetTransport(), nil, nil, nil)
	ports := make([]uint16, 0, 10)
	if config.Port != 0 {
		ports = append(ports, uint16(config.Port))
	} else if config.Ports != nil {
		for _, port := range config.Ports {
			ports = append(ports, uint16(port))
		}
	} else {
		for i := 0; i < 5; i++ {
			ports = append(ports, uint16(rand.Intn(40000-1024)+1024))
		}
	}

	for _, port := range ports {
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
				Listener:    transport,
				Tag:         h.Tag(),
				ConnHandler: connHandler,
			})
			h.AddWorker(tcpWorker)
		}
		if len(udpServers) > 0 {
			udpWorker := proxy.NewUdpWorker(proxy.UdpWorkerConfig{
				Addr:        &net.UDPAddr{IP: address.IP(), Port: int(port)},
				Listener:    transport.Socket,
				Tag:         h.Tag(),
				ConnHandler: udpServers[0],
			})
			h.AddWorker(udpWorker)

		}
	}
	return h, nil
}
