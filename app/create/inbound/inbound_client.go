// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package inbound

import (
	"slices"

	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/dokodemo"
	"github.com/5vnetwork/vx-core/proxy/http"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	"github.com/5vnetwork/vx-core/proxy/socks"
	"github.com/5vnetwork/vx-core/transport"
)

func NewInbound(config *configs.ProxyInboundConfig, ha i.Handler,
	tp i.TimeoutSetting, resolver i.IPResolver, df transport.DialerFactory) (proxy.Inbound, error) {
	ports := make([]uint16, 0, 10)
	if config.Port != 0 {
		ports = append(ports, uint16(config.Port))
	}
	if len(config.Ports) > 0 {
		for _, port := range config.Ports {
			ports = append(ports, uint16(port))
		}
	}

	if config.Protocol != nil {
		config.Protocols = append(config.Protocols, config.Protocol)
	}

	servers, hysteriaConfig, err := getServers(config.Users, config.Protocols,
		ha, tp)
	if err != nil {
		return nil, err
	}

	// proxy inbound
	h := proxy.NewProxyInbound(config.Tag)
	for _, server := range servers {
		if i, ok := server.(proxy.UserManage); ok {
			h.AddUserManage(i)
		}
	}

	// hysteria
	hasHys := hysteriaConfig != nil
	if hysteriaConfig != nil {
		var addresses []string
		if config.Address != "" {
			addresses = append(addresses, config.Address)
		}
		addresses = append(addresses, hysteriaConfig.Addresses...)
		listener, err := df.GetPacketListener(nil)
		if err != nil {
			return nil, err
		}
		in, err := hysteria2.NewInbound(&hysteria2.InboundConfig{
			Ports:                 ports,
			Addresses:             addresses,
			Hysteria2ServerConfig: hysteriaConfig,
			Tag:                   config.Tag,
			Handler:               ha,
			Realm:                 create.HysteriaRealmConfig(hysteriaConfig.GetRealm(), resolver),
			Listener:              listener,
		})
		if err != nil {
			return nil, err
		}
		for _, u := range append(hysteriaConfig.Users, config.Users...) {
			user, err := UserConfigToUser(u)
			if err != nil {
				return nil, err
			}
			in.AddUser(user)
		}
		h.AddWorker(in)
		h.AddUserManage(in)
	}

	if config.GetAddress() == "" {
		config.Address = net.AnyIP.String()
	}
	address := net.ParseAddress(config.Address)
	transport := create.TransportConfigToMemoryConfig(config.GetTransport(), nil,
		nil, nil)
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
		if !hasHys && len(udpServers) > 0 {
			udpWorker := proxy.NewUdpWorker(
				proxy.UdpWorkerConfig{
					Addr:        &net.UDPAddr{IP: address.IP(), Port: int(port)},
					Listener:    transport.Socket,
					Tag:         h.Tag(),
					ConnHandler: udpServers[0],
				},
			)
			h.AddWorker(udpWorker)
		}
	}
	return h, nil
}

func getServers(users []*configs.UserConfig, protocols []*anypb.Any, ha i.Handler,
	tp i.TimeoutSetting) ([]proxy.ProxyServer,
	*configs.Hysteria2ServerConfig, error) {

	var servers []proxy.ProxyServer
	var hysteriaConfig *configs.Hysteria2ServerConfig
	for _, protocol := range protocols {
		var server proxy.ProxyServer
		serverConfig, err := serial.GetInstanceOf(protocol)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get instance of ProxyServerConfig: %w", err)
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
		case *configs.SocksServerConfig:
			server = socks.NewServer(&socks.SocksServerConfig{
				Address:    net.ParseAddress(c.Address),
				UdpEnabled: c.UdpEnabled,
				AuthType:   configs.AuthType(c.AuthType),
				Policy:     tp,
				Handler:    ha,
			})
			for _, u := range c.Accounts {
				user, err := UserConfigToUser(u)
				if err != nil {
					return nil, nil, err
				}
				server.(*socks.Server).AddUser(user)
			}
		case *configs.HttpServerConfig:
			server = http.NewServer(http.ServerSettings{
				PolicyManager: tp,
				Handler:       ha,
			})
		case *configs.Hysteria2ServerConfig:
			hysteriaConfig = c
			continue
		default:
			return nil, nil, fmt.Errorf("unknown proxy server config: %T", c)
		}
		for _, u := range users {
			user, err := UserConfigToUser(u)
			if err != nil {
				return nil, nil, err
			}
			if i, ok := server.(proxy.UserManage); ok {
				i.AddUser(user)
			}
		}
		servers = append(servers, server)
	}

	return servers, hysteriaConfig, nil
}
