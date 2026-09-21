// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"fmt"
	"runtime"

	"github.com/5vnetwork/vx-core/app/client"
	cdns "github.com/5vnetwork/vx-core/app/create/dns"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/transport"

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
	client.StaticDnsServer = staticDnsServer

	// dns servers
	if len(dnsConfig.DnsServers) > 0 {
		err := fc.requireFeature(func(h *dispatcher.Dispatcher, gh i.GeoHelper,
			om *outbound.Manager, dii i.DefaultInterfaceInfo, df transport.DialerFactory) error {
			internalDns := &idns.InternalDns{
				StaticDns: staticDnsServer,
			}
			client.IPResolver.UpdateIPResolver(internalDns)
			client.EchResolver.UpdateECHResolver(internalDns)

			dailer, err := df.GetDialer(&transport.Config{})
			if err != nil {
				return err
			}
			// dns
			dnsServers, dnsServerMap, err := cdns.GetDnsServers(dnsConfig, internalDns, gh, dii, dailer, ipToDomain, h)
			if err != nil {
				return err
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
				dnsServerMap["Hijack"] = hijackDnsToDnsServer
				client.Dns = dns
				common.Must(fc.addComponent(dns))
				om.AddHandlers(idns.NewHandlerV().WithTag("dns").WithDns(dns))
			}

			// resolver used in dialing
			err = cdns.InternalDns(internalDns, config.GetDns().GetInternalResolver(), dnsServerMap)
			if err != nil {
				return err
			}

			// resolver used to lookup request domains in router and dispatcher
			err = cdns.PopulateResolverForRequestAddress(client,
				config.GetDns().GetRequestDomainResolver(), dnsServerMap)
			if err != nil {
				return err
			}
			// all dns servers
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
		client.IPResolverForRequestAddress.UpdateIPResolver(&dns.GoDnsResolver{})
		client.IPResolver.UpdateIPResolver(&dns.GoDnsResolver{})
		client.EchResolver.UpdateECHResolver(dns.DefaultCfResolver())
		client.Dns = idns.NewHijackDns(staticDnsServer, nil, false)
		allDnsServers := idns.NewAllDnsServers(nil)
		client.AllDnsServers = allDnsServers
		common.Must(fc.addComponent(&dns.GoDnsResolver{}))
		common.Must(fc.addFeature(client.Dns))
		common.Must(fc.addComponent(allDnsServers))
	}

	return nil
}
