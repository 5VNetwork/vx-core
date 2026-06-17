// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	cdns "github.com/5vnetwork/vx-core/app/create/dns"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/i"
)

func NewDNS(config *configs.TmConfig, fc *Builder, client *client.Client) error {
	dnsConfig := config.GetDns()
	if dnsConfig == nil {
		dnsConfig = &configs.DnsConfig{}
	}

	// ip to domain
	size := 500
	maxDomainAndResolversPerIp := 4
	if runtime.GOOS == "ios" {
		size = 100
		maxDomainAndResolversPerIp = 2
	}
	ipToDomain := idns.NewIPToDomain(size, maxDomainAndResolversPerIp)
	client.IPToDomain = ipToDomain
	common.Must(fc.addComponent(ipToDomain))

	// static
	staticDnsServer := idns.NewStaticDnsServer(dnsConfig.GetRecords(),
		dnsConfig.GetRecordStrings()...)

	// dns servers
	if len(dnsConfig.DnsServers) > 0 {
		err := fc.requireFeature(func(h *dispatcher.Dispatcher, gh i.GeoHelper,
			om *outbound.Manager, dii i.DefaultInterfaceInfo) error {
			internalDns := &idns.InternalDns{
				StaticDns: staticDnsServer,
			}
			client.IPResolver = internalDns
			client.EchResolver = internalDns
			if err := fc.addComponent(internalDns); err != nil {
				return err
			}

			var dailer i.Dialer
			if config.GetDialerFactory().GetShouldBindDevice() {
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
			// dns
			var dnsServers []idns.DnsServer
			dnsServerMap := make(map[string]idns.DnsServer)
			for _, dsConfig := range config.Dns.DnsServers {
				ds, err := cdns.NewDnsServer(dsConfig, h, ipToDomain, dii, dailer,
					internalDns, gh)
				if err != nil {
					return err
				}
				log.Info().Str("name", dsConfig.Name).Msg("new dns server")
				dnsServers = append(dnsServers, ds)
				dnsServerMap[dsConfig.Name] = ds
			}
			createdConcurrent := make(map[string]bool, len(dnsConfig.ConcurrentDnsServers))
			createdSerial := make(map[string]bool, len(dnsConfig.SerialDnsServers))
			remaining := len(dnsConfig.ConcurrentDnsServers) + len(dnsConfig.SerialDnsServers)
			for remaining > 0 {
				progressed := false

				for _, concurrentDnsServer := range dnsConfig.ConcurrentDnsServers {
					if createdConcurrent[concurrentDnsServer.Name] {
						continue
					}
					servers := make([]idns.DnsServer, 0, len(concurrentDnsServer.DnsServers))
					ready := true
					for _, name := range concurrentDnsServer.DnsServers {
						ds, ok := dnsServerMap[name]
						if !ok {
							ready = false
							break
						}
						servers = append(servers, ds)
					}
					if !ready {
						continue
					}
					concurrentDns := idns.NewConcurrentDnsServers(servers...)
					dnsServers = append(dnsServers, concurrentDns)
					dnsServerMap[concurrentDnsServer.Name] = concurrentDns
					createdConcurrent[concurrentDnsServer.Name] = true
					remaining--
					progressed = true
				}

				for _, serialDnsServer := range dnsConfig.SerialDnsServers {
					if createdSerial[serialDnsServer.Name] {
						continue
					}
					servers := make([]idns.DnsServer, 0, len(serialDnsServer.DnsServers))
					ready := true
					for _, name := range serialDnsServer.DnsServers {
						ds, ok := dnsServerMap[name]
						if !ok {
							ready = false
							break
						}
						servers = append(servers, ds)
					}
					if !ready {
						continue
					}
					serialDns := idns.NewSerialDnsServers(
						time.Duration(serialDnsServer.Interval)*time.Second, servers...)
					dnsServers = append(dnsServers, serialDns)
					dnsServerMap[serialDnsServer.Name] = serialDns
					createdSerial[serialDnsServer.Name] = true
					remaining--
					progressed = true
				}

				if !progressed {
					return fmt.Errorf("unable to resolve dns server dependencies for concurrent/serial dns servers")
				}
			}

			// dns hijack
			{
				var dnsRules []*idns.DnsRule
				for _, dnsRule := range dnsConfig.GetDnsHijack().GetDnsRules() {
					if ds, ok := dnsServerMap[dnsRule.DnsServerName]; ok {
						dr, err := cdns.NewDnsRule(dnsRule, ds, gh)
						if err != nil {
							return err
						}
						dnsRules = append(dnsRules, dr)
					} else {
						return fmt.Errorf("dns server %s not found", dnsRule.DnsServerName)
					}
				}
				dns := idns.NewHijackDns(staticDnsServer, dnsRules,
					config.GetDns().GetDnsHijack().GetEnableFakeDns())
				hijackDnsToDnsServer := &idns.HijackDnsToDnsServer{
					HijackDns: dns,
				}
				dnsServers = append(dnsServers, hijackDnsToDnsServer)
				dnsServerMap["hijack"] = hijackDnsToDnsServer
				client.Dns = dns
				common.Must(fc.addFeature(dns))
				om.AddHandlers(idns.NewHandlerV().WithTag("dns").WithDns(dns))
			}

			// resolver used in dialing
			{
				var servers []idns.DnsServer
				for _, name := range config.GetDns().GetInternalResolver().GetDnsServers() {
					if ds, ok := dnsServerMap[name]; ok {
						servers = append(servers, ds)
					} else {
						return fmt.Errorf("dns server %s not found", name)
					}
				}
				if len(servers) == 0 {
					resolver := idns.DefaultCfResolver()
					internalDns.Resolver = resolver
				} else {
					resolver := idns.NewDnsServerToResolver(
						idns.DnsServerToResolverOption{DnsServers: servers,
							Interval: time.Duration(config.Dns.InternalResolver.Interval) * time.Second})
					internalDns.Resolver = resolver
				}
			}

			// resolver used to lookup request domains in router and dispatcher
			{
				var servers []idns.DnsServer
				for _, name := range config.GetDns().GetRequestDomainResolver().GetDnsServers() {
					if ds, ok := dnsServerMap[name]; ok {
						servers = append(servers, ds)
					} else {
						return fmt.Errorf("dns server %s not found", name)
					}
				}
				if len(servers) == 0 {
					resolver := idns.DefaultCfResolver()
					client.IPResolverForRequestAddress = resolver
				} else {
					resolver := idns.NewDnsServerToResolver(
						idns.DnsServerToResolverOption{DnsServers: servers,
							Interval: time.Duration(config.Dns.RequestDomainResolver.Interval) * time.Second})
					client.IPResolverForRequestAddress = resolver
				}
			}
			//
			allDnsServers := idns.NewAllDnsServers(dnsServers)
			client.AllDnsServers = allDnsServers
			if err := fc.addComponent(allDnsServers); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	} else {
		client.IPResolverForRequestAddress = &dns.GoDnsResolver{}
		client.IPResolver = &dns.GoDnsResolver{}
		client.EchResolver = dns.DefaultCfResolver()
		client.Dns = idns.NewHijackDns(staticDnsServer, nil, false)
		allDnsServers := idns.NewAllDnsServers(nil)
		client.AllDnsServers = allDnsServers
		common.Must(fc.addComponent(&dns.GoDnsResolver{}))
		common.Must(fc.addFeature(client.Dns))
		common.Must(fc.addComponent(allDnsServers))
	}

	return nil
}
