package buildserver

import (
	"context"
	"fmt"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	cdns "github.com/5vnetwork/vx-core/app/create/dns"
	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

type DNSParams struct {
	fx.In
	Config    *configs.DnsConfig
	Handler   i.Handler `name:"dispatcher"`
	GeoHelper i.GeoHelper
	Monitor   i.DefaultInterfaceInfo
}

type DNSResult struct {
	fx.Out
	// Resolvers             []i.DnsResolver `group:"resolvers"`
	InternalResolver      i.IPResolver `name:"internal_resolver"`
	RequestDomainResolver i.IPResolver `name:"request_domain_resolver"`
	HijackDnsHandler      i.Handler    `name:"hijack_dns_handler"`
}

func NewDNS(lc fx.Lifecycle, params DNSParams) (DNSResult, error) {
	result := DNSResult{}

	dnsConfig := params.Config
	if dnsConfig == nil {
		result.InternalResolver = &idns.GoDnsResolver{}
		result.RequestDomainResolver = &idns.GoDnsResolver{}
		return result, nil
	}

	// static
	staticDnsServer := idns.NewStaticDnsServer(dnsConfig.GetRecords(),
		dnsConfig.GetRecordStrings()...)

	internalDns := &idns.InternalDns{
		StaticDns: staticDnsServer,
	}
	result.InternalResolver = internalDns

	// dns
	var dnsServers []idns.DnsServer
	dnsServerMap := make(map[string]idns.DnsServer)
	for _, dsConfig := range dnsConfig.DnsServers {
		ds, err := cdns.NewDnsServer(dsConfig, params.Handler,
			nil, params.Monitor, nil,
			internalDns, params.GeoHelper)
		if err != nil {
			return DNSResult{}, fmt.Errorf("failed to create dns server: %w", err)
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
			concurrentDns := idns.NewConcurrentDnsServers(concurrentDnsServer.Name, servers...)
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
				serialDnsServer.Name,
				time.Duration(serialDnsServer.Interval)*time.Second, servers...)
			dnsServers = append(dnsServers, serialDns)
			dnsServerMap[serialDnsServer.Name] = serialDns
			createdSerial[serialDnsServer.Name] = true
			remaining--
			progressed = true
		}

		if !progressed {
			return DNSResult{}, fmt.Errorf("unable to resolve dns server dependencies for concurrent/serial dns servers")
		}
	}

	// dns hijack
	{
		var dnsRules []*idns.DnsRule
		for _, dnsRule := range dnsConfig.GetDnsHijack().GetDnsRules() {
			if ds, ok := dnsServerMap[dnsRule.DnsServerName]; ok {
				dr, err := cdns.NewDnsRule(dnsRule, ds, params.GeoHelper)
				if err != nil {
					return DNSResult{}, fmt.Errorf("failed to create dns rule: %w", err)
				}
				dnsRules = append(dnsRules, dr)
			} else {
				return DNSResult{}, fmt.Errorf("dns server %s not found", dnsRule.DnsServerName)
			}
		}
		dns := idns.NewHijackDns(staticDnsServer, dnsRules,
			dnsConfig.GetDnsHijack().GetEnableFakeDns())
		hijackDnsToDnsServer := &idns.HijackDnsToDnsServer{
			HijackDns: dns,
		}
		dnsServers = append(dnsServers, hijackDnsToDnsServer)
		dnsServerMap["Hijack"] = hijackDnsToDnsServer
		result.HijackDnsHandler = idns.NewHandlerV().WithTag("dns").WithDns(dns)
	}

	// resolver used in dialing
	{
		var servers []idns.DnsServer
		for _, name := range dnsConfig.GetInternalResolver().GetDnsServers() {
			if ds, ok := dnsServerMap[name]; ok {
				servers = append(servers, ds)
			} else {
				return DNSResult{}, fmt.Errorf("dns server %s not found", name)
			}
		}
		resolver := idns.NewDnsServerToResolver(
			idns.DnsServerToResolverOption{DnsServers: servers,
				Interval: time.Duration(dnsConfig.GetInternalResolver().GetInterval()) * time.Second})
		internalDns.Resolver = resolver
	}

	// resolver used to lookup request domains in router and dispatcher
	{
		var servers []idns.DnsServer
		for _, name := range dnsConfig.GetRequestDomainResolver().GetDnsServers() {
			if ds, ok := dnsServerMap[name]; ok {
				servers = append(servers, ds)
			} else {
				return DNSResult{}, fmt.Errorf("dns server %s not found", name)
			}
		}
		resolver := idns.NewDnsServerToResolver(
			idns.DnsServerToResolverOption{DnsServers: servers,
				Interval: time.Duration(dnsConfig.GetRequestDomainResolver().GetInterval()) * time.Second})
		result.RequestDomainResolver = resolver
	}
	//
	allDnsServers := idns.NewAllDnsServers(dnsServers)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return allDnsServers.Start()
		},
		OnStop: func(ctx context.Context) error {
			return allDnsServers.Close()
		},
	})
	return result, nil
}
