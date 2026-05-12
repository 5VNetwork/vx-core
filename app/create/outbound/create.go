// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package outbound

import (
	"crypto/x509"
	"fmt"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/user"
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
	"github.com/5vnetwork/vx-core/proxy/shadowsocks_2022"
	"github.com/5vnetwork/vx-core/proxy/socks"
	"github.com/5vnetwork/vx-core/proxy/trojan"
	"github.com/5vnetwork/vx-core/proxy/wireguard"

	"github.com/5vnetwork/vx-core/proxy/vless"
	vless_client "github.com/5vnetwork/vx-core/proxy/vless/outbound"
	"github.com/5vnetwork/vx-core/proxy/vmess"
	vmess_client "github.com/5vnetwork/vx-core/proxy/vmess/client"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/apernet/hysteria/core/v2/client"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type HandlerConfig struct {
	*configs.HandlerConfig
	DialerFactory transport.DialerFactory
	Policy        *policy.Policy
	// for node domain
	IPResolver                  i.IPResolver
	EchResolver                 i.ECHResolver
	IPResolverForRequestAddress i.IPResolver
	RejectQuic                  bool
}

func NewHandler(config *HandlerConfig) (i.Outbound, error) {
	var out i.Outbound
	var err error
	if config.GetOutbound() != nil {
		out, err = NewOutHandler(&Config{
			OutboundHandlerConfig:       config.GetOutbound(),
			DialerFactory:               config.DialerFactory,
			Policy:                      config.Policy,
			IPResolver:                  config.IPResolver,
			ECHResolver:                 config.EchResolver,
			IPResolverForRequestAddress: config.IPResolverForRequestAddress,
			RejectQuic:                  config.RejectQuic,
		})
	} else {
		out, err = NewChainHandler(&ChainHandlerConfig{
			ChainHandlerConfig:          config.GetChain(),
			Policy:                      config.Policy,
			IPResolver:                  config.IPResolver,
			EchResolver:                 config.EchResolver,
			DF:                          config.DialerFactory,
			IPResolverForRequestAddress: config.IPResolverForRequestAddress,
			RejectQuic:                  config.RejectQuic,
		})
	}
	if err != nil {
		return nil, err
	}
	if config.SupportIpv6 != nil {
		return outbound.NewHandlerWithSupport6Info(out, *config.SupportIpv6), nil
	}
	return out, nil
}

type Config struct {
	*configs.OutboundHandlerConfig
	DialerFactory transport.DialerFactory
	Policy        i.TimeoutSetting
	// some outbound require it to lookup server addresses
	IPResolver  i.IPResolver
	ECHResolver i.ECHResolver
	// some outbounds need it to lookup ips of request addresses
	IPResolverForRequestAddress i.IPResolver
	RejectQuic                  bool
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

	if _, ok := m.(*configs.FreedomConfig); ok {
		transportConfig := create.TransportConfigToMemoryConfig(config.Transport, nil, nil, config.ECHResolver)
		transportConfig.DomainStrategy = domain.DomainStrategy(config.DomainStrategy)
		dialer, err := df.GetDialer(transportConfig)
		if err != nil {
			return nil, err
		}
		pl, err := df.GetPacketListener(
			transportConfig)
		if err != nil {
			return nil, err
		}
		if ipr == nil {
			ipr = &dns.GoDnsResolver{}
		}
		freedomHandler := freedom.New(dialer, pl, config.Tag, ipr)
		return freedomHandler, nil
	}

	// dialer
	transportConfig := create.TransportConfigToMemoryConfig(config.Transport,
		readCounter, writeCounter, config.ECHResolver)
	transportConfig.DomainStrategy = domain.DomainStrategy(config.DomainStrategy)
	dialer, err := df.GetDialer(transportConfig)
	if err != nil {
		return nil, err
	}

	if conf, ok := m.(*configs.WireguardClientConfig); ok {
		handler, err := wireguard.New(wireguard.HandlerSettings{
			Name:                 config.Tag,
			Conf:                 conf,
			Dialer:               dialer,
			Strategy:             domain.DomainStrategy(config.DomainStrategy),
			DnsForRequestAddress: config.IPResolverForRequestAddress,
			DnsForEndpoint:       config.IPResolver,
		})
		if err != nil {
			return nil, err
		}
		return handler, nil
	}

	sp, err := getPortSelector(config.OutboundHandlerConfig)
	if err != nil {
		return nil, err
	}
	var pc i.Handler
	switch m := m.(type) {
	case *configs.HttpClientConfig:
		pc = http.NewClient(http.ClientSettings{
			Address:            address,
			PortPicker:         sp,
			Account:            m.Account,
			H1SkipWaitForReply: m.H1SkipWaitForReply,
			Dialer:             dialer,
		})
	case *configs.ShadowsocksClientConfig:
		account, err := shadowsocks.NewMemoryAccount(
			user.NewUser("", m.Password),
			shadowsocks.CipherType(m.CipherType),
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
	case *configs.SocksClientConfig:
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
	case *configs.TrojanClientConfig:
		account := trojan.NewMemoryAccount(user.NewUser("", m.Password))
		pc = trojan.NewClient(
			trojan.ClientSettings{
				Address:    address,
				PortPicker: sp,
				Account:    account,
				Dialer:     dialer,
				Vision:     m.Vision,
			})
	case *configs.VmessClientConfig:
		account := vmess.NewMemoryAccount(user.NewUser("", uuid.StringToUUID(m.Id).String()),
			uint16(m.AlterId), protocol.SecurityType(m.Security), false, false)
		sp, err := getServerPicker(account, config.Address, config.Port, config.Ports)
		if err != nil {
			return nil, err
		}
		pc = vmess_client.NewClient(vmess_client.ClientSettings{
			ServerPicker: sp,
			Dialer:       dialer,
		})
	case *configs.VlessClientConfig:
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
	case *configs.Hysteria2ClientConfig:
		var rootCAs *x509.CertPool
		if len(m.GetTlsConfig().RootCas) > 0 {
			rootCAs, err = tls.CertsToCertPool(m.GetTlsConfig().RootCas)
			if err != nil {
				return nil, err
			}
		}
		lis, _ := df.GetPacketListener(create.TransportConfigToMemoryConfig(config.Transport,
			readCounter, writeCounter, config.ECHResolver))
		if ipr == nil {
			ipr = &dns.GoDnsResolver{}
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
		}
		maxConnectionReceiveWindow := uint64(m.Quic.GetMaxConnectionReceiveWindow() * 1024 * 1024)
		if maxConnectionReceiveWindow == 0 {
			maxConnectionReceiveWindow = m.Quic.GetMaxConnectionReceiveWindowBytes()
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
			Tag:                      config.Tag,
			Address:                  address,
			PortSelector:             sp,
			IpResolverForNodeAddress: ipr,
			DomainStrategy:           domain.DomainStrategy(config.DomainStrategy),
			PacketListener:           lis,
			RejectQuic:               config.RejectQuic,
			HysteriaClientConfig: &client.Config{
				Auth: m.Auth,
				TLSConfig: client.TLSConfig{
					ServerName:                     serverName,
					InsecureSkipVerify:             m.GetTlsConfig().GetAllowInsecure(),
					VerifyPeerCertificate:          tls.VerifyPeerCert(m.GetTlsConfig()),
					RootCAs:                        rootCAs,
					EncryptedClientHelloConfigList: m.GetTlsConfig().GetEchConfig(),
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
	case *configs.AnytlsClientConfig:
		pc = anytls.NewClient(
			&anytls.ClientConfig{
				Address:                  address,
				PortPicker:               sp,
				Password:                 m.Password,
				Dialer:                   dialer,
				IdleSessionCheckInterval: time.Duration(m.IdleSessionCheckInterval) * time.Second,
				IdleSessionTimeout:       time.Duration(m.IdleSessionTimeout) * time.Second,
				MinIdleSession:           int(m.MinIdleSession),
			})
	case *configs.Shadowsocks2022ClientConfig:
		pc, err = shadowsocks_2022.NewClient(
			&shadowsocks_2022.ClientSettings{
				Address:    address,
				PortPicker: sp,
				Method:     m.Method,
				Key:        m.GetKey(),
				Dialer:     dialer,
			})
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown proxy client config: %v", reflect.TypeOf(m))
	}

	settings := outbound.ProxyHandlerSettings{
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
	h := outbound.NewProxyHandler(
		settings,
	)
	return h, nil
}

type HandlerCreator func(config interface{}, address string, portSelector i.PortSelector,
) (i.Outbound, error)

func getSinglePort(config *configs.OutboundHandlerConfig) uint16 {
	if len(config.Ports) > 0 {
		return uint16(config.Ports[0].GetFrom())
	} else {
		return uint16(config.Port)
	}
}

func getPortSelector(config *configs.OutboundHandlerConfig) (i.PortSelector, error) {
	var ranges []*net.PortRange
	if len(config.Ports) > 0 {
		ranges = config.Ports
	} else if config.Port != 0 {
		ranges = []*net.PortRange{
			{From: config.Port, To: config.Port},
		}
	} else {
		return nil, fmt.Errorf("no port and ports")
	}

	interval, minInterval, maxInterval, useOne := getOnePortStrategySeconds(config)
	if useOne {
		return outbound.NewOnePortSelector(
			ranges,
			time.Second*time.Duration(interval),
			time.Second*time.Duration(minInterval),
			time.Second*time.Duration(maxInterval),
		), nil
	}
	return outbound.NewRandomPortSelector(ranges), nil
}

func getOnePortStrategySeconds(config *configs.OutboundHandlerConfig) (uint32, uint32, uint32, bool) {
	msg := config.ProtoReflect()
	if !msg.IsValid() {
		return 0, 0, 0, false
	}
	oneField := msg.Descriptor().Fields().ByName(protoreflect.Name("one"))
	if oneField == nil || !msg.Has(oneField) {
		return 0, 0, 0, false
	}
	oneMsg := msg.Get(oneField).Message()
	if !oneMsg.IsValid() {
		return 0, 0, 0, false
	}

	var interval uint32
	var minInterval uint32
	var maxInterval uint32

	if intervalField := oneMsg.Descriptor().Fields().ByName(protoreflect.Name("interval")); intervalField != nil && oneMsg.Has(intervalField) {
		interval = uint32(oneMsg.Get(intervalField).Uint())
	}
	if minField := oneMsg.Descriptor().Fields().ByName(protoreflect.Name("min_interval")); minField != nil && oneMsg.Has(minField) {
		minInterval = uint32(oneMsg.Get(minField).Uint())
	}
	if maxField := oneMsg.Descriptor().Fields().ByName(protoreflect.Name("max_interval")); maxField != nil && oneMsg.Has(maxField) {
		maxInterval = uint32(oneMsg.Get(maxField).Uint())
	}
	return interval, minInterval, maxInterval, true
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
