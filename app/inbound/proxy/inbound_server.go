//go:build server

package proxy

import (
	"math/rand"
	"slices"

	"fmt"

	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/anypb"

	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfigs "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/anytls"
	"github.com/5vnetwork/vx-core/proxy/dokodemo"
	"github.com/5vnetwork/vx-core/proxy/http"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	"github.com/5vnetwork/vx-core/proxy/shadowsocks"
	"github.com/5vnetwork/vx-core/proxy/socks"
	"github.com/5vnetwork/vx-core/proxy/trojan"
	vless_server "github.com/5vnetwork/vx-core/proxy/vless/inbound"
	vmess_server "github.com/5vnetwork/vx-core/proxy/vmess/server"
)

func NewInboundServer(config *configs.ProxyInboundConfig, ha i.Handler, router i.Router,
	tp i.TimeoutSetting, inStats *monitor.Stats, onUnauth i.UnauthorizedReport) (Inbound, error) {

	ports := make([]uint16, 0, 10)
	if config.Port != 0 {
		ports = append(ports, uint16(config.Port))
	} else if len(config.Ports) > 0 {
		for _, port := range config.Ports {
			ports = append(ports, uint16(port))
		}
	} else {
		for i := 0; i < 5; i++ {
			ports = append(ports, uint16(rand.Intn(40000-1024)+1024))
		}
	}

	if len(config.Protocols) == 0 {
		config.Protocols = append(config.Protocols, config.Protocol)
	}

	servers, hysteriaConfig, err := getServers(config.Users, config.Protocols,
		ha, tp, onUnauth)
	if err != nil {
		return nil, err
	}

	if config.GetAddress() == "" {
		config.Address = net.AnyIP.String()
	}

	// proxy inbound
	h := &ProxyInbound{
		tag: config.Tag,
	}
	for _, server := range servers {
		if i, ok := server.(UserManage); ok {
			h.userManages = append(h.userManages, i)
		}
	}
	address := net.ParseAddress(config.Address)
	transport := create.TransportConfigToMemoryConfig(config.GetTransport(), nil, nil)

	for _, port := range ports {
		// hysteria
		hasHys := false
		if hysteriaConfig != nil {
			hasHys = true
			in, err := hysteria2.NewInbound(&hysteria2.InboundConfig{
				Ports:                 ports,
				Hysteria2ServerConfig: hysteriaConfig,
				InStats:               inStats,
				Tag:                   config.Tag,
				Router:                router,
				OnUnauthorizedRequest: onUnauth,
				Dialer:                &net.NetDialer{Dialer: net.Dialer{}},
				Listener:              &net.NetPacketListener{ListenConfig: net.ListenConfig{}},
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
			h.workers = append(h.workers, in)
			h.userManages = append(h.userManages, in)
		}

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
		if !hasHys && len(udpServers) > 0 {
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

func getServers(users []*configs.UserConfig, protocols []*anypb.Any, ha i.Handler, tp i.TimeoutSetting, onUnauth i.UnauthorizedReport) ([]ProxyServer, *proxyconfigs.Hysteria2ServerConfig, error) {

	var servers []ProxyServer
	var hysteriaConfig *proxyconfigs.Hysteria2ServerConfig
	for _, protocol := range protocols {
		var server ProxyServer
		serverConfig, err := serial.GetInstanceOf(protocol)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get instance of ProxyServerConfig: %w", err)
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
		case *proxyconfigs.SocksServerConfig:
			server = socks.NewServer(&socks.SocksServerConfig{
				Address:    net.ParseAddress(c.Address),
				UdpEnabled: c.UdpEnabled,
				AuthType:   proxyconfigs.AuthType(c.AuthType),
				Policy:     tp,
				Handler:    ha,
			})
			for _, u := range c.Accounts {
				user, err := create.UserConfigToUser(u)
				if err != nil {
					return nil, nil, err
				}
				server.(*socks.Server).AddUser(user)
			}
		case *proxyconfigs.HttpServerConfig:
			server = http.NewServer(http.ServerSettings{
				PolicyManager: tp,
				Handler:       ha,
			})
		case *proxyconfigs.ShadowsocksServerConfig:
			server = shadowsocks.NewServer(
				shadowsocks.ServerSettings{
					Cipher:           shadowsocks.CipherType(c.CipherType),
					ReducedIVEntropy: c.ExperimentReducedIvHeadEntropy,
					IvCheck:          c.IvCheck,
					PolicyManager:    tp,
					Handler:          ha,
				})
			if c.User != nil {
				user, err := create.UserConfigToUser(c.User)
				if err != nil {
					return nil, nil, err
				}
				server.(*shadowsocks.Server).AddUser(user)
			}
		case *proxyconfigs.VlessServerConfig:
			server, err = vless_server.New(vless_server.HandlerSettings{
				PolicyManager: tp,
				Handler:       ha,
			})
			if err != nil {
				return nil, nil, err
			}
			for _, u := range c.Users {
				user, err := create.UserConfigToUser(u)
				if err != nil {
					return nil, nil, err
				}
				server.(*vless_server.Handler).AddUser(user)
			}
		case *proxyconfigs.VmessServerConfig:
			server = vmess_server.New(
				vmess_server.ServerSettings{
					Secure:                c.SecureEncryptionOnly,
					Handler:               ha,
					OnUnauthorizedRequest: onUnauth,
					PolicyManager:         tp,
				},
			)
			for _, u := range c.Accounts {
				user, err := create.UserConfigToUser(u)
				if err != nil {
					return nil, nil, err
				}
				server.(*vmess_server.Server).AddUser(user)
			}
		case *proxyconfigs.TrojanServerConfig:
			server = trojan.NewServer(
				trojan.ServerSettings{
					Handler:               ha,
					OnUnauthorizedRequest: onUnauth,
					PolicyManager:         tp,
					Vision:                c.Vision,
				},
			)
			for _, u := range c.Users {
				user, err := create.UserConfigToUser(u)
				if err != nil {
					return nil, nil, err
				}
				server.(*trojan.Server).AddUser(user)
			}
		case *proxyconfigs.Hysteria2ServerConfig:
			hysteriaConfig = c
			continue
		case *proxyconfigs.AnytlsServerConfig:
			server = anytls.NewServer(
				anytls.ServerSettings{
					Handler:               ha,
					OnUnauthorizedRequest: onUnauth,
				},
			)
			for _, u := range c.Users {
				user, err := create.UserConfigToUser(u)
				if err != nil {
					return nil, nil, err
				}
				server.(*anytls.Server).AddUser(user)
			}
		default:
			return nil, nil, fmt.Errorf("unknown proxy server config: %T", c)
		}
		for _, u := range users {
			user, err := create.UserConfigToUser(u)
			if err != nil {
				return nil, nil, err
			}
			if i, ok := server.(UserManage); ok {
				i.AddUser(user)
			}
		}
		servers = append(servers, server)
	}

	return servers, hysteriaConfig, nil
}
