// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package grpcservice

import (
	"context"
	"fmt"
	"runtime"

	vxdns "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/dns"
	vxrouter "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/router"
	cdns "github.com/5vnetwork/vx-core/app/create/dns"
	"github.com/5vnetwork/vx-core/app/dns"
	idns "github.com/5vnetwork/vx-core/app/dns"

	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

func (s *GrpcService) UpdateRouter(ctx context.Context, in *UpdateRouterRequest) (*UpdateRouterResponse, error) {
	log.Info().Msg("update router")

	err := s.updateRouter(in.RouterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update router: %w", err)
	}

	return &UpdateRouterResponse{}, nil
}

func (s *GrpcService) ChangeRoutingMode(ctx context.Context, in *ChangeRoutingModeRequest) (*ChangeRoutingModeResponse, error) {
	log.Debug().Msg("ChangeRoutingMode")
	err := s.Client.Geo.UpdateGeo(in.GeoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create geo: %w", err)
	}
	log.Debug().Msg("geo updated")

	err = s.updateDns(in.DnsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update dns: %w", err)
	}
	log.Debug().Msg("dns updated")

	if err := s.updateRouter(in.RouterConfig); err != nil {
		return nil, fmt.Errorf("failed to updateRouter: %w", err)
	}
	log.Debug().Msg("routing mode changed")
	return &ChangeRoutingModeResponse{}, nil
}

func (s *GrpcService) updateDns(dnsConfig *vxdns.DnsConfig) error {
	client := s.Client
	if len(dnsConfig.DnsServers) > 0 {
		internalDns := &idns.InternalDns{
			StaticDns: client.StaticDnsServer,
		}

		dailer, err := client.DialerFactory.GetDialer(nil)
		if err != nil {
			return err
		}
		dnsServers, dnsServerMap, err := cdns.GetDnsServers(dnsConfig,
			internalDns, client.Geo, client.NetMon,
			dailer, client.IPToDomain, client.Dispatcher)
		if err != nil {
			return err
		}

		client.AllDnsServers.UpdateDnsServers(dnsServers)

		// dns hijack
		{
			var dnsRules []*idns.DnsRule
			for _, dnsRule := range dnsConfig.GetDnsHijack().GetDnsRules() {
				if ds, ok := dnsServerMap[dnsRule.DnsServerName]; ok {
					dr, err := cdns.NewDnsRule(dnsRule, ds, client.Geo)
					if err != nil {
						return err
					}
					dnsRules = append(dnsRules, dr)
				} else {
					return fmt.Errorf("dns server %s not found", dnsRule.DnsServerName)
				}
			}
			client.Dns.UpdateDnsRules(dnsRules)
			hijackDnsToDnsServer := &idns.HijackDnsToDnsServer{
				HijackDns: client.Dns,
			}
			dnsServerMap["Hijack"] = hijackDnsToDnsServer
		}

		err = cdns.InternalDns(internalDns, dnsConfig.GetInternalResolver(),
			dnsServerMap)
		if err != nil {
			return err
		}
		client.IPResolver.UpdateIPResolver(internalDns)
		client.EchResolver.UpdateECHResolver(internalDns)

		err = cdns.PopulateResolverForRequestAddress(client,
			dnsConfig.GetRequestDomainResolver(), dnsServerMap)
		if err != nil {
			return err
		}
	} else {
		client.IPResolverForRequestAddress.UpdateIPResolver(&dns.GoDnsResolver{})
		client.IPResolver.UpdateIPResolver(&dns.GoDnsResolver{})
		client.EchResolver.UpdateECHResolver(dns.DefaultCfResolver())
		client.Dns.UpdateDnsRules(nil)
		client.AllDnsServers.UpdateDnsServers(nil)
	}

	return nil
}

func (s *GrpcService) updateRouter(config *vxrouter.RouterConfig) error {
	newRouter, err := router.NewRouter(&router.RouterConfig{
		RouterConfig:    config,
		OutboundManager: s.Client.OutboundManager,
		GeoHelper:       s.Client.Geo,
		Selectors:       s.Client.Selectors,
		IpResolver:      s.Client.IPResolverForRequestAddress,
	})
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}
	s.Client.Router.UpdateRouter(newRouter)
	return nil
}

func (s *GrpcService) ChangeOutbound(ctx context.Context, in *ChangeOutboundRequest) (*ChangeOutboundResponse, error) {
	log.Debug().Msg("ChangeOutbound")
	// s.AutoOutbound = in.GetAutoOutbound()
	// s.Policy.SetOutboundStats(s.AutoOutbound)
	om := s.Client.OutboundManager
	handlers := make([]i.Outbound, 0, len(in.GetHandlers()))
	for _, handler := range in.GetHandlers() {
		h, err := s.Client.HandlerFactory.CreateHandler(handler)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbound handler: %w", err)
		}
		handlers = append(handlers, h)
	}
	if in.GetDeleteAll() {
		om.ReplaceHandlers(handlers)
	} else {
		om.RemoveHandlers(in.GetTags())
		om.AddHandlers(handlers...)
	}
	return &ChangeOutboundResponse{}, nil
}

func (s *GrpcService) ChangeHandlerStore(ctx context.Context, in *ChangeHandlerStoreRequest) (*ChangeHandlerStoreResponse, error) {
	log.Debug().Msg("ChangeHandlerStore")
	if in.GetDeleteAll() {
		if s.Client.HandlerStore == nil {
			s.Client.HandlerStore = selector.NewHandlerStore()
		}
		s.Client.HandlerStore.ReplaceFromProtos(in.GetOutboundHandlers())
		return &ChangeHandlerStoreResponse{}, nil
	}
	if s.Client.HandlerStore == nil {
		s.Client.HandlerStore = selector.NewHandlerStore()
	}
	s.Client.HandlerStore.RemoveTags(in.GetTags())
	s.Client.HandlerStore.SetFromProtos(in.GetOutboundHandlers())
	return &ChangeHandlerStoreResponse{}, nil
}

func (s *GrpcService) CurrentOutbound(ctx context.Context, in *CurrentOutboundRequest) (*CurrentOutboundResponse, error) {
	om := s.Client.OutboundManager
	handlers := GetAllProxyhandlers(om)
	tags := make([]string, 0, len(handlers))
	for _, h := range handlers {
		tags = append(tags, h.Tag())
	}
	return &CurrentOutboundResponse{
		OutboundTags: tags,
	}, nil
}

func (s *GrpcService) SelectedHandlers(ctx context.Context, in *SelectedHandlersRequest) (*SelectedHandlersResponse, error) {
	selectors := s.Client.Selectors.GetAllSelectors()
	selectedHandlers := make(map[string]*SelectedHandlersResponse_Strings)
	for _, selector := range selectors {
		selectedHandlers[selector.Tag()] = &SelectedHandlersResponse_Strings{
			Strings: selector.GetHandlersBeingUsed(),
		}
	}
	return &SelectedHandlersResponse{
		SelectedHandlers: selectedHandlers,
	}, nil
}

func (s *GrpcService) NotifyHandlerChange(context.Context, *HandlerChangeNotify) (*HandlerChangeNotifyResponse, error) {
	log.Info().Msg("NotifyHandlerChange")
	s.Client.Selectors.OnHandlerChanged()
	return &HandlerChangeNotifyResponse{}, nil
}

func (s *GrpcService) ChangeSelector(ctx context.Context, in *ChangeSelectorRequest) (*ChangeSelectorResponse, error) {
	log.Info().Msg("ChangeSelector")
	if in.SelectorToRemove != "" {
		s.Client.Selectors.RemoveSelector(in.SelectorToRemove)
	}
	if in.DeleteAll {
		s.Client.Selectors.RemoveAllSelectors()
	}
	for _, selectorConfig := range in.GetSelectorsToAdd() {
		landHandlers, err := selector.ResolveLandHandlers(s.Client.DB, selectorConfig.LandHandlers, s.Client.HandlerStore)
		if err != nil {
			return nil, err
		}
		var filter selector.Filter
		if selectorConfig.SelectFromOm {
			filter = selector.NewOmFilter(selectorConfig.GetFilter(), s.Client.OutboundManager)
		} else if _, ok := s.Client.DB.(*xsqlite.Db); ok && runtime.GOOS == "android" {
			filter = selector.NewDbOrOmFilter(s.Client.DB, selectorConfig.GetFilter(),
				landHandlers, s.Client.HandlerFactory, s.Client.HandlerStore)
		} else {
			filter = selector.NewDbFilter(s.Client.DB, selectorConfig.GetFilter(),
				landHandlers, s.Client.HandlerFactory)
		}
		s.Client.Selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
			SelectorConfig:            selectorConfig,
			CreateHandler:             s.Client.HandlerFactory,
			HandlerErrorChangeSubject: s.Client.Dispatcher,
			Tester:                    s.Client.Tester,
			Filter:                    filter,
			HandlerInfo:               s.Client.OutStats,
		}))
	}
	return &ChangeSelectorResponse{}, nil
}

func (s *GrpcService) UpdateSelectorBalancer(ctx context.Context, in *UpdateSelectorBalancerRequest) (*Receipt, error) {
	log.Info().Msg("UpdateSelectorBalancer")

	var balancer selector.Balancer
	switch in.BalanceStrategy {
	case vxrouter.SelectorConfig_RANDOM:
		balancer = selector.NewRandomBanlancer()
	case vxrouter.SelectorConfig_MEMORY:
		balancer = selector.NewMemoryBalancer()
	}
	se := s.Client.Selectors.GetSelector(in.Tag)
	if se == nil {
		return nil, fmt.Errorf("selector not found: %s", in.Tag)
	}
	se.UpdateBalancer(balancer)
	return &Receipt{}, nil
}

func (s *GrpcService) UpdateSelectorFilter(ctx context.Context, in *UpdateSelectorFilterRequest) (*Receipt, error) {
	log.Info().Msg("UpdateSelectorFilter")
	se := s.Client.Selectors.GetSelector(in.Tag)
	if se == nil {
		return nil, fmt.Errorf("selector not found: %s", in.Tag)
	}

	filter := se.GetFilter()
	filter.(selector.FilterUpdate).UpdateFilterConfig(in.Filter)
	se.UpdateFilter(filter)

	return &Receipt{}, nil
}

func (s *GrpcService) SetOutboundHandlerSpeed(ctx context.Context, in *SetOutboundHandlerSpeedRequest) (*SetOutboundHandlerSpeedResponse, error) {
	log.Debug().Str("tag", in.GetTag()).Int32("speed", in.GetSpeed()).Msg("SetOutboundHandlerSpeed")
	s.Client.Selectors.OnHandlerSpeedChanged(in.GetTag(), in.GetSpeed())
	return &SetOutboundHandlerSpeedResponse{}, nil
}

const DirectHandlerTag = "direct"
const DnsHandlerTag = "dns"

// replace all proxy handlers with new ones
func ReplaceHandlers(om i.OutboundManager, handlers ...i.Outbound) {
	directHandler := om.GetHandler(DirectHandlerTag)
	dnsHandler := om.GetHandler(DnsHandlerTag)
	om.ReplaceHandlers(append([]i.Outbound{directHandler, dnsHandler}, handlers...))
}

// return all handlers except direct and dns
func GetAllProxyhandlers(om i.OutboundManager) []i.Outbound {
	all := om.GetAllHandlers()
	proxyHandlers := make([]i.Outbound, 0, len(all))
	for _, handler := range all {
		if handler.Tag() == DirectHandlerTag || handler.Tag() == DnsHandlerTag {
			continue
		}
		proxyHandlers = append(proxyHandlers, handler)
	}
	return proxyHandlers
}
