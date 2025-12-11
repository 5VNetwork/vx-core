package outbound

import (
	"crypto/x509"
	"fmt"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfigs "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common/domain"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/anytls"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/proxy/http"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	"github.com/5vnetwork/vx-core/proxy/shadowsocks"
	"github.com/5vnetwork/vx-core/proxy/socks"
	"github.com/5vnetwork/vx-core/proxy/trojan"

	"github.com/5vnetwork/vx-core/proxy/vless"
	vless_client "github.com/5vnetwork/vx-core/proxy/vless/outbound"
	"github.com/5vnetwork/vx-core/proxy/vmess"
	vmess_client "github.com/5vnetwork/vx-core/proxy/vmess/client"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/apernet/hysteria/core/v2/client"
)

type HandlerConfig struct {
	*configs.HandlerConfig
	DialerFactory transport.DialerFactory
	Policy        *policy.Policy
	// for node domain
	IPResolver                  i.IPResolver
	IPResolverForRequestAddress i.IPResolver
	DnsServer                   i.ECHResolver
	RejectQuic                  bool
}

func NewHandler(config *HandlerConfig) (i.Outbound, error) {
	if config.GetOutbound() != nil {
		return NewOutHandler(&Config{
			OutboundHandlerConfig:       config.GetOutbound(),
			DialerFactory:               config.DialerFactory,
			Policy:                      config.Policy,
			IPResolver:                  config.IPResolver,
			IPResolverForRequestAddress: config.IPResolverForRequestAddress,
			RejectQuic:                  config.RejectQuic,
		})
	} else {
		return NewChainHandler(&ChainHandlerConfig{
			ChainHandlerConfig:          config.GetChain(),
			Policy:                      config.Policy,
			IPResolver:                  config.IPResolver,
			DF:                          config.DialerFactory,
			IPResolverForRequestAddress: config.IPResolverForRequestAddress,
			RejectQuic:                  config.RejectQuic,
			DnsServer:                   config.DnsServer,
		})
	}
}

type Config struct {
	*configs.OutboundHandlerConfig
	DialerFactory transport.DialerFactory
	Policy        i.TimeoutSetting
	// some outbound require it to lookup server addresses
	IPResolver i.IPResolver
	// some outbounds need it to lookup ips of request addresses
	IPResolverForRequestAddress i.IPResolver
	//
	RejectQuic bool
}

// TODO: Validate config
func NewOutHandler(config *Config) (i.Outbound, error) {
	if config == nil {
		return nil, fmt.Errorf("outbound handler config is nil")
	}

	df := config.DialerFactory
	ipr := config.IPResolver
	policy := config.Policy
	address := net.ParseAddress(config.Address)

	var readCounter, writeCounter *atomic.Uint64

	m, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy client config: %w", err)
	}

	if _, ok := m.(*proxyconfigs.FreedomConfig); ok {
		dialer, err := df.GetDialer(create.TransportConfigToMemoryConfig(config.Transport, nil, nil))
		if err != nil {
			return nil, err
		}
		pl, err := df.GetPacketListener(create.TransportConfigToMemoryConfig(config.Transport, nil, nil))
		if err != nil {
			return nil, err
		}
		if ipr == nil {
			ipr = &dns.DnsResolver{}
		}
		freedomHandler := freedom.New(dialer, pl, config.Tag, ipr)
		return freedomHandler, nil
	}

	sp, err := getPortSelector(config.Address, config.Port, config.Ports)
	if err != nil {
		return nil, err
	}

	// dialer
	transportConfig := create.TransportConfigToMemoryConfig(config.Transport,
		readCounter, writeCounter)
	transportConfig.Address = address
	transportConfig.PortSelector = sp
	transportConfig.DomainStrategy = domain.DomainStrategy(config.DomainStrategy)
	dialer, err := df.GetDialer(transportConfig)
	if err != nil {
		return nil, err
	}

	var pc i.Handler
	switch m := m.(type) {
	case *proxyconfigs.HttpClientConfig:
		pc = http.NewClient(http.ClientSettings{
			Address:            address,
			PortPicker:         sp,
			Account:            m.Account,
			H1SkipWaitForReply: m.H1SkipWaitForReply,
			Dialer:             dialer,
		})
	case *proxyconfigs.ShadowsocksClientConfig:
		account, err := shadowsocks.NewMemoryAccount(
			"",
			shadowsocks.CipherType(m.CipherType),
			m.Password,
			false,
			false,
		)
		if err != nil {
			return nil, err
		}
		pc = shadowsocks.NewClient(&shadowsocks.ClientSettings{
			Address:    address,
			PortPicker: sp,
			Account:    account,
			Dialer:     dialer,
		})
	case *proxyconfigs.SocksClientConfig:
		pc = socks.NewClient(&socks.ClientSettings{
			ServerDest: net.TCPDestination(address,
				net.Port(getSinglePort(config.OutboundHandlerConfig))),
			User:           m.Name,
			Secret:         m.Password,
			DelayAuthWrite: m.DelayAuthWrite,
			DNS:            ipr,
			Policy:         policy,
			Dialer:         dialer,
		})
	case *proxyconfigs.TrojanClientConfig:
		account := trojan.NewMemoryAccount("", m.Password)
		pc = trojan.NewClient(
			trojan.ClientSettings{
				Address:    address,
				PortPicker: sp,
				Account:    account,
				Dialer:     dialer,
				Vision:     m.Vision,
			})
	case *proxyconfigs.VmessClientConfig:
		account := vmess.NewMemoryAccount("", uuid.StringToUUID(m.Id).String(),
			uint16(m.AlterId), protocol.SecurityType(m.Security), false, false)
		sp, err := getServerPicker(account, config.Address, config.Port, config.Ports)
		if err != nil {
			return nil, err
		}
		pc = vmess_client.NewClient(vmess_client.ClientSettings{
			ServerPicker: sp,
			Dialer:       dialer,
		})
	case *proxyconfigs.VlessClientConfig:
		uid, err := uuid.ParseString(m.Id)
		if err != nil {
			return nil, err
		}
		account := &vless.MemoryAccount{
			ID:         protocol.NewID(uid),
			Flow:       m.Flow,
			Encryption: m.Encryption,
		}
		pc = vless_client.New(
			vless_client.ClientSettings{
				ServerPicker:   sp,
				Address:        address,
				Account:        account,
				TimeoutSetting: policy,
				Dialer:         dialer,
			},
		)
	case *proxyconfigs.Hysteria2ClientConfig:
		var rootCAs *x509.CertPool
		if len(m.GetTlsConfig().RootCas) > 0 {
			rootCAs, err = tls.CertsToCertPool(m.GetTlsConfig().RootCas)
			if err != nil {
				return nil, err
			}
		}
		lis, _ := df.GetPacketListener(create.TransportConfigToMemoryConfig(config.Transport,
			readCounter, writeCounter))
		if ipr == nil {
			ipr = &dns.DnsResolver{}
		}
		serverName := m.GetTlsConfig().ServerName
		if serverName == "" {
			serverName = config.Address
		}
		initialStreamReceiveWindow := uint64(m.Quic.GetInitialStreamReceiveWindow() * 1024 * 1024)
		if initialStreamReceiveWindow == 0 {
			initialStreamReceiveWindow = m.Quic.GetInitialStreamReceiveWindowBytes()
			if initialStreamReceiveWindow == 0 && runtime.GOOS == "ios" {
				initialStreamReceiveWindow = 80 * 1024
			}
		}
		initialConnectionReceiveWindow := uint64(m.Quic.GetInitialConnectionReceiveWindow() * 1024 * 1024)
		if initialConnectionReceiveWindow == 0 {
			initialConnectionReceiveWindow = m.Quic.GetInitialConnectionReceiveWindowBytes()
			if initialConnectionReceiveWindow == 0 && runtime.GOOS == "ios" {
				initialConnectionReceiveWindow = 200 * 1024
			}
		}
		maxStreamReceiveWindow := uint64(m.Quic.GetMaxStreamReceiveWindow() * 1024 * 1024)
		if maxStreamReceiveWindow == 0 {
			maxStreamReceiveWindow = m.Quic.GetMaxStreamReceiveWindowBytes()
			if maxStreamReceiveWindow == 0 && runtime.GOOS == "ios" {
				maxStreamReceiveWindow = 800 * 1024
			}
		}
		maxConnectionReceiveWindow := uint64(m.Quic.GetMaxConnectionReceiveWindow() * 1024 * 1024)
		if maxConnectionReceiveWindow == 0 {
			maxConnectionReceiveWindow = m.Quic.GetMaxConnectionReceiveWindowBytes()
			if maxConnectionReceiveWindow == 0 && runtime.GOOS == "ios" {
				maxConnectionReceiveWindow = 2000 * 1024
			}
		}
		keepAlive := m.Quic.GetKeepAlivePeriod()
		if keepAlive == 0 {
			keepAlive = 10
		}
		maxIdleTimeout := m.Quic.GetMaxIdleTimeout()
		if maxIdleTimeout == 0 {
			maxIdleTimeout = 30
		}
		hys, err := hysteria2.NewClient(&hysteria2.Config{
			Tag:                        config.Tag,
			Address:                    address,
			PortSelector:               sp,
			IpResolverForNodeAddress:   ipr,
			DomainStrategy:             domain.DomainStrategy(config.DomainStrategy),
			IpResolverForTargetAddress: config.IPResolverForRequestAddress,
			PacketListener:             lis,
			RejectQuic:                 config.RejectQuic,
			HysteriaClientConfig: &client.Config{
				Auth: m.Auth,
				TLSConfig: client.TLSConfig{
					ServerName:                     serverName,
					InsecureSkipVerify:             m.GetTlsConfig().AllowInsecure,
					VerifyPeerCertificate:          m.GetTlsConfig().VerifyPeerCert,
					RootCAs:                        rootCAs,
					EncryptedClientHelloConfigList: m.GetTlsConfig().EchConfig,
				},
				QUICConfig: client.QUICConfig{
					DisablePathMTUDiscovery:        m.Quic.GetDisablePathMtuDiscovery(),
					InitialStreamReceiveWindow:     initialStreamReceiveWindow,
					InitialConnectionReceiveWindow: initialConnectionReceiveWindow,
					MaxIdleTimeout:                 time.Duration(maxIdleTimeout) * time.Second,
					KeepAlivePeriod:                time.Duration(keepAlive) * time.Second,
					MaxConnectionReceiveWindow:     maxConnectionReceiveWindow,
					MaxStreamReceiveWindow:         maxStreamReceiveWindow,
				},
				BandwidthConfig: client.BandwidthConfig{
					MaxTx: uint64(m.Bandwidth.GetMaxTx() * 1024 * 1024),
					MaxRx: uint64(m.Bandwidth.GetMaxRx() * 1024 * 1024),
				},
				FastOpen: m.FastOpen,
			},
			SalamanderPassword: m.Obfs.GetSalamander().GetPassword(),
		})
		if err != nil {
			return nil, err
		}
		return hys, nil
	case *proxyconfigs.AnytlsClientConfig:
		pc = anytls.NewClient(
			&anytls.ClientConfig{
				Address:    address,
				PortPicker: sp,
				Password:   m.Password,
				Dialer:     dialer,
			})
	default:
		return nil, fmt.Errorf("unknown proxy client config: %v", reflect.TypeOf(m))
	}

	settings := ProxyHandlerSettings{
		Tag:       config.Tag,
		Handler:   pc,
		Uot:       config.Uot,
		EnableMux: config.EnableMux,
		MuxConfig: mux.DefaultClientStrategy,
	}
	if config.EnableMux && config.MuxConfig != nil {
		settings.MuxConfig = mux.ClientStrategy{
			MaxConnection:  config.MuxConfig.MaxConnection,
			MaxConcurrency: config.MuxConfig.MaxConcurrency,
		}
	}
	h := NewProxyHandler(
		settings,
	)
	return h, nil
}

func getSinglePort(config *configs.OutboundHandlerConfig) uint16 {
	if len(config.Ports) > 0 {
		return uint16(config.Ports[0].GetFrom())
	} else {
		return uint16(config.Port)
	}
}

// TODO: support udp
func getPortSelector(address string, port uint32, ports []*net.PortRange) (i.PortSelector, error) {
	if len(ports) > 0 {
		return NewRandomPortSelector(ports), nil
	} else if port != 0 {
		return NewRandomPortSelector([]*net.PortRange{
			{From: port, To: port},
		}), nil
	} else {
		return nil, fmt.Errorf("no port and ports")
	}
}

func getServerPicker(account interface{}, address string, port uint32, ports []*net.PortRange) (protocol.ServerPicker, error) {
	serverList := protocol.NewServerList()
	if len(ports) > 0 {
		for _, pr := range ports {
			for i := pr.GetFrom(); i <= pr.GetTo(); i++ {
				if i == 0 {
					continue
				}
				serverList.AddServer(
					protocol.NewServerSpec(net.TCPDestination(net.ParseAddress(address),
						net.Port(i)), protocol.AlwaysValid(), account))
			}
		}
	} else if port != 0 {
		serverList.AddServer(
			protocol.NewServerSpec(net.TCPDestination(net.ParseAddress(address),
				net.Port(port)), protocol.AlwaysValid(), account))

	} else {
		return nil, fmt.Errorf("no port and ports")
	}
	if serverList.Size() == 0 {
		return nil, fmt.Errorf("no servers")
	}
	return protocol.NewRoundRobinServerPicker(serverList), nil
}
