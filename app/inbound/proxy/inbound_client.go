package proxy

import (
	"math/rand"
	"slices"

	"fmt"

	"github.com/5vnetwork/vx-core/app/create"
	"github.com/rs/zerolog/log"

	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfigs "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/dokodemo"
	"github.com/5vnetwork/vx-core/proxy/http"
	"github.com/5vnetwork/vx-core/proxy/socks"
)

func NewInbound(config *configs.ProxyInboundConfig, ha i.Handler, tp i.TimeoutSetting) (Inbound, error) {
	if len(config.Protocols) == 0 {
		config.Protocols = append(config.Protocols, config.Protocol)
	}

	var servers []ProxyServer
	for _, protocol := range config.Protocols {
		var server ProxyServer
		serverConfig, err := serial.GetInstanceOf(protocol)
		if err != nil {
			return nil, fmt.Errorf("failed to get instance of ProxyServerConfig: %w", err)
		}
		switch c := serverConfig.(type) {
		case *proxyconfigs.DokodemoConfig:
			server = dokodemo.New(
				dokodemo.DoorSettings{
					Address:  net.ParseAddress(c.Address),
					Port:     net.Port(c.Port),
					Networks: c.Networks,
					Handler:  ha,
				},
			)
			servers = append(servers, server)
		case *proxyconfigs.SocksServerConfig:
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
				user, err := create.UserConfigToUser(u)
				if err != nil {
					return nil, err
				}
				server.(*socks.Server).AddUser(user)
			}
			servers = append(servers, server)
		case *proxyconfigs.HttpServerConfig:
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
	h := &ProxyInbound{
		tag: config.Tag,
	}
	address := net.ParseAddress(config.Address)
	transport := create.TransportConfigToMemoryConfig(config.GetTransport(), nil, nil)
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
		var tcpServers []ProxyServer
		var udpServers []ProxyServer
		for _, server := range servers {
			if slices.Contains(server.Network(), net.Network_TCP) {
				tcpServers = append(tcpServers, server)
			}
			if slices.Contains(server.Network(), net.Network_UDP) {
				udpServers = append(udpServers, server)
			}
		}
		if len(tcpServers) > 0 {
			tcpWorker := &tcpWorker{
				addr:     &net.TCPAddr{IP: address.IP(), Port: int(port)},
				listener: transport,
				tag:      h.tag,
			}
			if len(tcpServers) == 1 {
				tcpWorker.connHandler = tcpServers[0]
			} else {
				proxyServers := &proxyServers{}
				for _, server := range tcpServers {
					if fp, ok := server.(FallbackProxyServer); ok {
						proxyServers.fallbackProxyServers = append(proxyServers.fallbackProxyServers, fp)
					} else {
						if proxyServers.proxyServer != nil {
							log.Warn().Msg("there are two non-fallback proxy servers for the same port")
						}
						proxyServers.proxyServer = server
					}
				}
				// if there is no non-fallback proxy server, make the last fallback server as it
				if proxyServers.proxyServer == nil && len(proxyServers.fallbackProxyServers) > 0 {
					proxyServers.proxyServer = proxyServers.fallbackProxyServers[len(proxyServers.fallbackProxyServers)-1]
					proxyServers.fallbackProxyServers[len(proxyServers.fallbackProxyServers)-1] = nil
					proxyServers.fallbackProxyServers = proxyServers.fallbackProxyServers[:len(proxyServers.fallbackProxyServers)-1]
				}
				tcpWorker.connHandler = proxyServers
			}
			h.workers = append(h.workers, tcpWorker)
		}
		if len(udpServers) > 0 {
			udpWorker := &udpWorker{
				tag:         h.tag,
				addr:        &net.UDPAddr{IP: address.IP(), Port: int(port)},
				address:     address.IP(),
				port:        port,
				connHandler: udpServers[0],
				listener:    transport.Socket,
			}
			h.workers = append(h.workers, udpWorker)
		}
	}
	return h, nil
}
