package buildclient

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"runtime"
	"sort"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/common"
	pd "github.com/5vnetwork/vx-core/common/dispatcher"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/rs/zerolog/log"

	mdns "github.com/miekg/dns"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/i"
)

func NewDNS(config *configs.TmConfig, fc *Builder, client *client.Client) error {
	dnsConfig := config.GetDns()
	if dnsConfig == nil {
		dnsConfig = &configs.DnsConfig{}
	}

	// ip to domain
	size := 500
	if runtime.GOOS == "ios" {
		size = 100
	}
	ipToDomain := idns.NewIPToDomain(size)
	client.IPToDomain = ipToDomain
	common.Must(fc.addComponent(ipToDomain))

	// static
	staticDnsServer := idns.NewStaticDnsServer(dnsConfig.GetRecords())
	if len(dnsConfig.DnsServers) > 0 {
		err := fc.requireFeature(func(h *dispatcher.Dispatcher, gh i.GeoHelper,
			om *outbound.Manager, dii i.DefaultInterfaceInfo) error {
			var dailer i.Dialer
			if config.GetTun().GetShouldBindDevice() {
				if runtime.GOOS == "android" {
					fdFunc := fc.getFeature(reflect.TypeOf((*transport.FdFunc)(nil)).Elem())
					dailer = &dlhelper.SocketSetting{
						FdFunc: fdFunc.(transport.FdFunc),
					}
				} else {
					dailer = transport.NewBindToDefaultNICDialer(dii, &dlhelper.SocketSetting{})
				}
			} else {
				dailer = transport.DefaultDialer
			}
			// dns server for direct
			ctx := inbound.ContextWithInboundTag(
				log.With().Str("tag", "internal-dns-direct").Logger().WithContext(
					context.Background()), "internal-dns-direct")
			dis := pd.NewPacketDispatcher(ctx, h)
			internalDnsDirect := idns.NewDnsServerConcurrent(idns.DnsServerConcurrentOption{
				Name:       "internal-dns-direct",
				RrCache:    idns.NewRrCache(idns.RrCacheSetting{}),
				Handler:    h,
				IPToDomain: ipToDomain,
				Dispatcher: dis,
			})
			setDefaultNICNameservers(internalDnsDirect, dii, []mynet.AddressPort{
				{
					Address: mynet.CfDns4,
					Port:    53,
				},
				{
					Address: mynet.AliyunDns4,
					Port:    53,
				},
			}, dailer)
			ctx = inbound.ContextWithInboundTag(
				log.With().Str("tag", "internal-dns-proxy").Logger().WithContext(
					context.Background()), "internal-dns-proxy")
			dis = pd.NewPacketDispatcher(ctx, h,
				// pd.WithRequestTimeout(time.Second*4),
				// pd.WithResponseTimeout(time.Second*4),
				pd.WithLinkLifetime(time.Minute*5),
			)
			internalDndProxy := idns.NewDnsServerConcurrent(idns.DnsServerConcurrentOption{
				Name:    "internal-dns-proxy",
				RrCache: idns.NewRrCache(idns.RrCacheSetting{Duration: 3600}),
				NameserverAddrs: []mynet.AddressPort{
					{
						Address: mynet.CfDns4,
						Port:    53,
					},
				},
				Handler:    h,
				Dispatcher: dis,
			})
			internalDns := idns.NewInternalDns(staticDnsServer, internalDnsDirect, internalDndProxy)
			client.IPResolver = internalDns
			// dns
			var dnsServers []idns.DnsServer
			var dnsRules []*idns.DnsRule
			for _, dsConfig := range config.Dns.DnsServers {
				ds, err := newDnsServer(dsConfig, h, ipToDomain, fc, client, dailer, internalDns, config.Dns.CacheDuration)
				if err != nil {
					return err
				}
				dnsServers = append(dnsServers, ds)
				for _, dnsRule := range dnsConfig.DnsRules {
					if dnsRule.DnsServerName == dsConfig.Name {
						dr, err := newDnsRule(dnsRule, ds, gh, client)
						if err != nil {
							return err
						}
						dnsRules = append(dnsRules, dr)
					}
				}
			}
			dns := idns.NewDns(staticDnsServer, dnsRules, dnsServers)
			client.IPResolverForRequestAddress = &idns.Prefer4IPResolver{
				IPResolver: &idns.DnsServerToResolver{
					DnsServers: []idns.DnsServer{&idns.DnsToDnsServer{Dns: dns}},
				},
			}

			if err := fc.addComponent(internalDns); err != nil {
				return err
			}
			if err := fc.addComponent(dns); err != nil {
				return err
			}
			om.AddHandlers(idns.NewHandlerV().WithTag("dns").WithDns(dns))
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		ipResolver := &idns.DnsResolver{
			Resolver: net.DefaultResolver,
		}
		client.Dns = idns.NewDns(staticDnsServer, nil, nil)
		client.IPResolver = ipResolver
		client.IPResolverForRequestAddress = ipResolver
		common.Must(fc.addComponent(ipResolver))
		common.Must(fc.addComponent(client.Dns))
	}

	return nil
}
func newDnsServer(config *configs.DnsServerConfig, handler i.Handler, ipToDomain *idns.IPToDomain,
	fc *Builder, client *client.Client, dailer i.Dialer, ipResolver i.IPResolver, globalDuration uint32) (idns.DnsServer, error) {
	duration := config.CacheDuration
	if duration == 0 {
		duration = globalDuration
	}
	rrCache := idns.NewRrCache(idns.RrCacheSetting{Duration: duration})

	switch c := config.Type.(type) {
	case *configs.DnsServerConfig_DohDnsServer:
		return idns.NewDoHNameServer(idns.DoHNameServerOption{
			Handler:    handler,
			Name:       config.Name,
			Url:        c.DohDnsServer.Url,
			IpToDomain: ipToDomain,
			RrCache:    rrCache,
			ClientIP:   net.ParseIP(config.ClientIp),
		})
	case *configs.DnsServerConfig_FakeDnsServer:
		pools, err := idns.NewPools(c.FakeDnsServer.GetPoolConfigs())
		if err != nil {
			return nil, fmt.Errorf("failed to create fake dns pool: %w", err)
		}
		fakeDns := dns.NewFakeDns(pools)
		client.AllFakeDns.AddFakeDns(fakeDns)
		return fakeDns, nil
	case *configs.DnsServerConfig_PlainDnsServer:
		var addressPorts []mynet.AddressPort
		for _, addr := range c.PlainDnsServer.Addresses {
			address, port, _ := net.SplitHostPort(addr)
			addressPorts = append(addressPorts, mynet.AddressPort{
				Address: mynet.ParseAddress(address),
				Port:    common.Must2(mynet.PortFromString(port)).(mynet.Port),
			})
		}
		ctx := inbound.ContextWithInboundTag(log.With().Str("tag", config.Name).Logger().WithContext(
			context.Background()), config.Name)
		dis := pd.NewPacketDispatcher(ctx,
			handler,
			// pd.WithRequestTimeout(time.Second*4),
			// pd.WithResponseTimeout(time.Second*4),
			pd.WithLinkLifetime(time.Minute*5),
		)

		ns := idns.NewDnsServerConcurrent(idns.DnsServerConcurrentOption{
			Name:            config.Name,
			Handler:         handler,
			IPToDomain:      ipToDomain,
			NameserverAddrs: addressPorts,
			ClientIp:        net.ParseIP(config.ClientIp),
			Dispatcher:      dis,
			RrCache:         rrCache,
		})
		if c.PlainDnsServer.UseDefaultDns {
			fc.requireFeature(func(info i.DefaultInterfaceInfo) {
				setDefaultNICNameservers(ns, info, addressPorts, dailer)
			})
		}
		return ns, nil
	case *configs.DnsServerConfig_TlsDnsServer:
		var addressPorts []mynet.AddressPort
		for _, addr := range c.TlsDnsServer.Addresses {
			address, port, _ := net.SplitHostPort(addr)
			addressPorts = append(addressPorts, mynet.AddressPort{
				Address: mynet.ParseAddress(address),
				Port:    common.Must2(mynet.PortFromString(port)).(mynet.Port),
			})
		}
		ctx := inbound.ContextWithInboundTag(
			log.With().Str("tag", config.Name).Logger().WithContext(
				context.Background()), config.Name)
		dis := pd.NewPacketDispatcher(ctx,
			handler,
			// pd.WithRequestTimeout(time.Second*4),
			// pd.WithResponseTimeout(time.Second*4),
			pd.WithLinkLifetime(time.Minute*5),
		)
		return idns.NewDnsServerConcurrent(idns.DnsServerConcurrentOption{
			Name:            config.Name,
			Handler:         handler,
			Dispatcher:      dis,
			IPToDomain:      ipToDomain,
			Tls:             true,
			NameserverAddrs: addressPorts,
			RrCache:         rrCache,
			ClientIp:        net.ParseIP(config.ClientIp),
		}), nil
	case *configs.DnsServerConfig_QuicDnsServer:
		dst, _ := mynet.ParseDestination(c.QuicDnsServer.Address)
		dst.Network = mynet.Network_UDP
		return idns.NewQUICNameServer(idns.QuicNameServerOption{
			Name:        config.Name,
			Destination: dst,
			Handler:     handler,
			IpToDomain:  ipToDomain,
			IPResolver:  ipResolver,
			ClientIp:    net.ParseIP(config.ClientIp),
			RrCache:     rrCache,
		})
	default:
		return nil, fmt.Errorf("unsupported DNS server type: %s", config.Type)
	}
}

type setDests interface {
	SetDests(dests []mynet.AddressPort)
	RemoveDest(remove mynet.AddressPort, fallback []mynet.AddressPort)
}

func setDefaultNICNameservers(ns setDests, info i.DefaultInterfaceInfo,
	fallbackNameservers []mynet.AddressPort, dailer i.Dialer) {
	var onInterfaceChange i.OnDefaultInterfaceChanged = func() {
		var newDests []mynet.AddressPort
		for _, dns := range info.DefaultDns4() {
			newDests = append(newDests, mynet.AddressPort{
				Address: mynet.AddressFromNetIpAddr(dns),
				Port:    mynet.Port(53),
			})
		}
		if info.DefaultInterfaceName6() != info.DefaultInterfaceName4() {
			for _, dns := range info.DefaultDns6() {
				newDests = append(newDests, mynet.AddressPort{
					Address: mynet.AddressFromNetIpAddr(dns),
					Port:    mynet.Port(53),
				})
			}
		}
		log.Info().Any("servers", newDests).Msg("default dns servers")
		if len(newDests) > 0 {
			// make ipv4 dests first
			// since some nic might contain ipv6 dns server but acutally does not support ipv6
			sort.Slice(newDests, func(i, j int) bool {
				return newDests[i].Address.Family() == mynet.AddressFamilyIPv4 && newDests[j].Address.Family() != mynet.AddressFamilyIPv4
			})
			ns.SetDests(newDests)

			msg := &mdns.Msg{}
			msg.SetQuestion("www.baidu.com.", mdns.TypeA)
			for _, dnsServer := range newDests {
				go func(ap mynet.AddressPort) {
					conn, err := dailer.Dial(context.Background(), mynet.UDPDestination(ap.Address, ap.Port))
					if err != nil {
						log.Debug().Err(err).Str("server", ap.String()).Msg("test dns server of default nic failed to dial")
						ns.RemoveDest(ap, fallbackNameservers)
						log.Info().Str("server", ap.String()).Msg("dns server removed")
						return
					}
					defer conn.Close()
					client := &mdns.Client{
						Net: "udp",
					}
					dnsConn := &mdns.Conn{
						Conn:    conn,
						UDPSize: client.UDPSize,
					}
					_, _, err = client.ExchangeWithConn(msg, dnsConn)
					if err != nil {
						log.Error().Err(err).Str("server", ap.String()).Msg("test dns server failed to exchange dns")
						ns.RemoveDest(ap, fallbackNameservers)
						log.Info().Str("server", ap.String()).Msg("dns server removed")
					}
				}(dnsServer)
			}
		} else {
			ns.SetDests(fallbackNameservers)
			log.Info().Interface("nameservers", fallbackNameservers).Msg("use config nameservers for direct")
		}
	}
	onInterfaceChange()
	info.Register(onInterfaceChange)
}

func newDnsRule(config *configs.DnsRuleConfig, dnsServer idns.DnsServer, gh i.GeoHelper, client *client.Client) (*dns.DnsRule, error) {
	var conditions []dns.Condition
	if len(config.DomainTags) > 0 || len(config.Domains) > 0 {
		domainSet, err := geo.NewDomainSet(config.DomainTags, gh, config.Domains...)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, &idns.PreferDomainCondition{
			DomainSet: domainSet,
		})
	}
	if _, ok := dnsServer.(*dns.FakeDns); ok {
		conditions = append(conditions, &idns.FakeDnsCondition{
			FakeDnsEnabled: &client.FakeDnsEnabled,
		})
		conditions = append(conditions, &idns.HasSrcCondition{})
	}
	if len(config.IncludedTypes) > 0 {
		c := &idns.IncludedTypesCondition{}
		for _, t := range config.IncludedTypes {
			c.Types = append(c.Types, typeToNumber(t))
		}
		conditions = append(conditions, c)
	}

	return idns.NewDnsRule(dnsServer, conditions...), nil
}

func typeToNumber(t configs.DnsType) uint16 {
	switch t {
	case configs.DnsType_DnsType_A:
		return mdns.TypeA
	case configs.DnsType_DnsType_AAAA:
		return mdns.TypeAAAA
	default:
		panic("not supported dns type")
	}
}
