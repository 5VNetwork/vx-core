// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package router

import (
	"context"
	"errors"
	"net/netip"
	"sync/atomic"
	"time"

	cgeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"

	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type RouterWrapper struct {
	atomic.Value //*Router
}

func (r *RouterWrapper) GetRouter() *Router {
	router := r.Value.Load()
	if router == nil {
		return nil
	}
	return router.(*Router)
}

func (r *RouterWrapper) UpdateRouter(router *Router) {
	r.Value.Store(router)
}

func (r *RouterWrapper) PickHandler(ctx context.Context, si *session.Info) (i.Outbound, error) {
	router := r.Value.Load()
	if router == nil {
		return nil, ErrNoHandlerPick
	}
	return router.(*Router).PickHandler(ctx, si)
}

func (r *RouterWrapper) PickHandlerWithData(ctx context.Context, si *session.Info,
	rw interface{}) (interface{}, i.Outbound, []i.Fallback, error) {
	router := r.Value.Load()
	if router == nil {
		return rw, nil, nil, ErrNoHandlerPick
	}
	return router.(*Router).PickHandlerWithData(ctx, si, rw)
}

var ErrNoHandlerPick = errors.New("no handler picked")

// determine a outbound handler for a session
type Router struct {
	rules     []*rule
	om        i.OutboundManager
	selectors *selector.Selectors
}

type RouterConfig struct {
	*configs.RouterConfig
	GeoHelper       i.GeoHelper
	Selectors       *selector.Selectors
	OutboundManager i.OutboundManager
	IpResolver      i.IPResolver
}

func NewRouter(config *RouterConfig) (*Router, error) {
	if config == nil {
		config = &RouterConfig{
			RouterConfig: &configs.RouterConfig{},
		}
	}
	if config.RouterConfig == nil {
		config.RouterConfig = &configs.RouterConfig{}
	}

	r := &Router{
		om:        config.OutboundManager,
		selectors: config.Selectors,
	}
	for _, routerRuleConfig := range config.Rules {
		var conditions []Condition
		if routerRuleConfig.MatchAll {
			conditions = []Condition{&ConditionTrue{}}
		} else {
			var err error
			conditions, err = getConditions(toConditionConfigs(routerRuleConfig),
				config.GeoHelper, config.IpResolver)
			if err != nil {
				return nil, err
			}
		}
		fallbackers := make([]i.Fallback, 0, len(routerRuleConfig.Fallbacks))
		for _, fallbackerConfig := range routerRuleConfig.Fallbacks {
			var conditions []Condition
			if fallbackerConfig.MatchAll {
				conditions = []Condition{&ConditionTrue{}}
			} else {
				var err error
				conditions, err = getConditions(fallbackConditions(fallbackerConfig),
					config.GeoHelper, config.IpResolver)
				if err != nil {
					return nil, err
				}
			}
			fallbackers = append(fallbackers, NewFallbacker(
				fallbackerConfig.SelectorTag, fallbackerConfig.OutboundTag,
				fallbackerConfig.Action, r.om, r.selectors, fallbackerConfig.Last, conditions))
		}
		rule := NewRule(routerRuleConfig.RuleName, routerRuleConfig.OutboundTag,
			routerRuleConfig.SelectorTag, fallbackers,
			conditions)
		r.AddRule(rule)
	}
	return r, nil
}

func fallbackConditions(fallbackerConfig *configs.RuleConfig_Fallback) []*configs.Condition {
	if len(fallbackerConfig.Conditions) > 0 {
		return fallbackerConfig.Conditions
	}
	if fallbackerConfig.Condition != nil {
		return []*configs.Condition{fallbackerConfig.Condition}
	}
	return []*configs.Condition{{
		DomainTags: fallbackerConfig.DomainTags,
		DstIpTags:  fallbackerConfig.DstIpTags,
	}}
}

func toConditionConfigs(ruleConfig *configs.RuleConfig) []*configs.Condition {
	if len(ruleConfig.Conditions) > 0 {
		return ruleConfig.Conditions
	}
	if ruleConfig.Condition != nil {
		return []*configs.Condition{ruleConfig.Condition}
	}
	return []*configs.Condition{{
		SrcCidrs:             ruleConfig.SrcCidrs,
		SrcIpTags:            ruleConfig.SrcIpTags,
		DstCidrs:             ruleConfig.DstCidrs,
		DstIpTags:            ruleConfig.DstIpTags,
		ResolveDomain:        ruleConfig.ResolveDomain,
		ResolveSoftRewrite:   ruleConfig.ResolveSoftRewrite,
		ResolveSoftNoRewrite: ruleConfig.ResolveSoftNoRewrite,
		GeoDomains:           ruleConfig.GeoDomains,
		DomainTags:           ruleConfig.DomainTags,
		SkipSniff:            ruleConfig.SkipSniff,
		Usernames:            ruleConfig.Usernames,
		InboundTags:          ruleConfig.InboundTags,
		Networks:             ruleConfig.Networks,
		SrcPortRanges:        ruleConfig.SrcPortRanges,
		DstPortRanges:        ruleConfig.DstPortRanges,
		AppIds:               ruleConfig.AppIds,
		Ipv6:                 ruleConfig.Ipv6,
		FakeIp:               ruleConfig.FakeIp,
		AppTags:              ruleConfig.AppTags,
		AllTags:              ruleConfig.AllTags,
		Protocols:            ruleConfig.Protocols,
	}}
}

func getConditions(conditionConfigs []*configs.Condition, gh i.GeoHelper, ipResolver i.IPResolver) ([]Condition, error) {
	conditions := make([]Condition, 0, len(conditionConfigs))
	for _, conditionConfig := range conditionConfigs {
		conds, err := getCondition(conditionConfig, gh, ipResolver)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, conds)
	}
	return conditions, nil
}

func getCondition(conditionConfig *configs.Condition, gh i.GeoHelper, ipResolver i.IPResolver) (*Conditions, error) {
	conditions := make([]Condition, 0, 20)
	if conditionConfig.InboundTags != nil {
		conditions = append(conditions, NewInboundTagMatcher(conditionConfig.InboundTags))
	}
	if conditionConfig.Ipv6 {
		conditions = append(conditions, &Ipv6Matcher{})
	}
	if len(conditionConfig.SrcCidrs) > 0 || len(conditionConfig.SrcIpTags) > 0 {
		var cidrs []*cgeo.CIDR
		for _, cidr := range conditionConfig.SrcCidrs {
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				return nil, err
			}
			cidrs = append(cidrs, &cgeo.CIDR{
				Ip:     prefix.Addr().AsSlice(),
				Prefix: uint32(prefix.Bits()),
			})
		}
		srcIPSet, err := geo.NewIPSet(conditionConfig.SrcIpTags, gh, cidrs...)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, &IpMatcher{
			MatchSourceIp: true,
			IpSet:         srcIPSet,
		})
	}
	if len(conditionConfig.DstCidrs) > 0 || len(conditionConfig.DstIpTags) > 0 {
		var cidrs []*cgeo.CIDR
		for _, cidr := range conditionConfig.DstCidrs {
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				return nil, err
			}
			cidrs = append(cidrs, &cgeo.CIDR{
				Ip:     prefix.Addr().AsSlice(),
				Prefix: uint32(prefix.Bits()),
			})
		}
		dstIPSet, err := geo.NewIPSet(conditionConfig.DstIpTags, gh, cidrs...)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, &IpMatcher{
			IpSet:                 dstIPSet,
			IpResolver:            ipResolver,
			ResolveHard:           conditionConfig.ResolveDomain,
			ResolveSoftAndRewrite: conditionConfig.ResolveSoftRewrite,
			ResolveSoftNoRewrite:  conditionConfig.ResolveSoftNoRewrite,
		})
	}
	if len(conditionConfig.Protocols) > 0 {
		sniffers := make([]sniff.ProtocolSnifferWithNetwork, 0, len(conditionConfig.Protocols))
		for _, protocol := range conditionConfig.Protocols {
			switch protocol {
			case "tls":
				sniffers = append(sniffers, sniff.TlsSniff)
			case "http1":
				sniffers = append(sniffers, sniff.HTTP1Sniff)
			case "quic":
				sniffers = append(sniffers, sniff.QUICSniff)
			case "bittorrent":
				sniffers = append(sniffers, sniff.BTScniff)
				sniffers = append(sniffers, sniff.UTPSniff)
			default:
				log.Warn().Str("protocol", protocol).Msg("unknown protocol")
				continue
			}
		}
		conditions = append(conditions, &ConditionProtocol{
			protocols: conditionConfig.Protocols,
			Sniffer: sniff.NewSniffer(sniff.SniffSetting{
				Interval: 10 * time.Millisecond,
				Sniffers: sniffers,
			}),
		})
	}
	if len(conditionConfig.GeoDomains) > 0 || len(conditionConfig.DomainTags) > 0 {
		domainSet, err := geo.NewDomainSet(conditionConfig.DomainTags, gh,
			conditionConfig.GeoDomains...)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, &DomainMatcher{
			DomainSet: domainSet,
			SkipSniff: conditionConfig.SkipSniff,
			Sniffer: sniff.NewSniffer(sniff.SniffSetting{
				Interval: 10 * time.Millisecond,
				Sniffers: []sniff.ProtocolSnifferWithNetwork{
					sniff.TlsSniff,
					sniff.HTTP1Sniff,
					sniff.QUICSniff,
					sniff.BTScniff,
					sniff.UTPSniff,
				},
			}),
		})
	}
	if len(conditionConfig.Networks) > 0 {
		conditions = append(conditions, NewNetworkMatcher(conditionConfig.Networks))
	}
	if len(conditionConfig.SrcPortRanges) > 0 {
		conditions = append(conditions, NewPortMatcher(conditionConfig.SrcPortRanges, true))
	}
	if len(conditionConfig.DstPortRanges) > 0 {
		conditions = append(conditions, NewPortMatcher(conditionConfig.DstPortRanges, false))
	}
	if len(conditionConfig.Usernames) > 0 {
		conditions = append(conditions, NewUserMatcher(conditionConfig.Usernames))
	}
	if len(conditionConfig.AppIds) > 0 || len(conditionConfig.AppTags) > 0 {
		appSet, err := geo.NewAppSet(conditionConfig.AppTags, gh, conditionConfig.AppIds...)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, &AppIdMatcher{
			AppSet: appSet,
		})
	}
	if conditionConfig.FakeIp {
		conditions = append(conditions, &ConditionFakeIp{})
	}
	if conditionConfig.HasNoDomain {
		conditions = append(conditions, &HasNoDomain{})
	}
	if len(conditionConfig.AllTags) > 0 {
		domainSet, err := geo.NewDomainSet(conditionConfig.AllTags, gh)
		if err != nil {
			return nil, err
		}
		ipSet, err := geo.NewIPSet(conditionConfig.AllTags, gh)
		if err != nil {
			return nil, err
		}
		appSet, err := geo.NewAppSet(conditionConfig.AllTags, gh)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, &AllMatcher{
			domainMatcher: &DomainMatcher{
				DomainSet: domainSet,
				SkipSniff: conditionConfig.SkipSniff,
				Sniffer: sniff.NewSniffer(
					sniff.SniffSetting{
						Interval: 10 * time.Millisecond,
						Sniffers: []sniff.ProtocolSnifferWithNetwork{
							sniff.TlsSniff,
							sniff.HTTP1Sniff,
							sniff.QUICSniff,
							sniff.BTScniff,
							sniff.UTPSniff,
						}}),
			},
			ipMatcher: &IpMatcher{
				IpSet:                 ipSet,
				IpResolver:            ipResolver,
				ResolveHard:           conditionConfig.ResolveDomain,
				ResolveSoftAndRewrite: conditionConfig.ResolveSoftRewrite,
				ResolveSoftNoRewrite:  conditionConfig.ResolveSoftNoRewrite,
			},
			appIdMatcher: &AppIdMatcher{
				AppSet: appSet,
			},
		})
	}
	return &Conditions{
		conditions: conditions,
	}, nil
}

func (r *Router) AddRule(rule *rule) {
	r.rules = append(r.rules, rule)
}

var ErrNoHandler = errors.New("no handler")
var ErrSelectorNotFound = errors.New("selector not found")
var ErrBlocked = errors.New("block")
var ErrNoRule = errors.New("no rule matched")

func (r *Router) PickHandler(ctx context.Context, si *session.Info) (i.Outbound, error) {
	_, handler, _, err := r.PickHandlerWithData(ctx, si, nil)
	return handler, err
}

func (r *Router) PickHandlerWithData(ctx context.Context, si *session.Info,
	rw interface{}) (interface{}, i.Outbound, []i.Fallback, error) {
	// for tests
	if len(r.rules) == 0 {
		if h := r.om.GetHandler(""); h != nil {
			return rw, h, nil, nil
		} else {
			return rw, r.om.GetHandler("direct"), nil, nil
		}
	}

	info := si

	for _, rule := range r.rules {
		rw0, t := rule.Apply(ctx, info, rw)
		rw = rw0
		if t {
			si.MatchedRule = rule.Name()
			log.Ctx(ctx).Debug().Str("matched_rule", si.MatchedRule).Msg("matched rule")
			if rule.outboundTag != "" {
				if h := r.om.GetHandler(rule.outboundTag); h != nil {
					return rw, h, rule.retries, nil
				}
				return rw, nil, nil, ErrNoHandler
			} else if rule.selectorTag != "" {
				if se := r.selectors.GetSelector(rule.selectorTag); se != nil {
					si.UsedSelector = rule.selectorTag
					if h := se.GetHandler(si); h != nil {
						return rw, h, rule.retries, nil
					}
					return rw, nil, nil, ErrNoHandler
				}
				log.Ctx(ctx).Warn().Str("selector_tag", rule.selectorTag).Msg("selector not found")
				return rw, nil, nil, ErrSelectorNotFound
			} else {
				return rw, nil, nil, ErrBlocked
			}
		}
	}

	return rw, nil, nil, ErrNoRule
}
