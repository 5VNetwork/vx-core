package router

import (
	"context"
	"errors"
	"net/netip"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/sniff"
	cgeo "github.com/5vnetwork/vx-core/common/geo"
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

func (r *RouterWrapper) PickHandlerWithData(ctx context.Context, si *session.Info, rw interface{}) (interface{}, i.Outbound, error) {
	router := r.Value.Load()
	if router == nil {
		return rw, nil, ErrNoHandlerPick
	}
	return router.(*Router).PickHandlerWithData(ctx, si, rw)
}

var ErrNoHandlerPick = errors.New("no handler picked")

// determine a outbound handler for a session
type Router struct {
	om        i.OutboundManager
	rules     []*rule
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
		conditions := make([]Condition, 0, 20)
		if routerRuleConfig.InboundTags != nil {
			conditions = append(conditions, NewInboundTagMatcher(routerRuleConfig.InboundTags))
		}
		if routerRuleConfig.Ipv6 {
			conditions = append(conditions, &Ipv6Matcher{})
		}
		if len(routerRuleConfig.SrcCidrs) > 0 || len(routerRuleConfig.SrcIpTags) > 0 {
			var cidrs []*cgeo.CIDR
			for _, cidr := range routerRuleConfig.SrcCidrs {
				prefix, err := netip.ParsePrefix(cidr)
				if err != nil {
					return nil, err
				}
				cidrs = append(cidrs, &cgeo.CIDR{
					Ip:     prefix.Addr().AsSlice(),
					Prefix: uint32(prefix.Bits()),
				})
			}
			srcIPSet, err := geo.NewIPSet(routerRuleConfig.SrcIpTags, config.GeoHelper, cidrs...)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, &IpMatcher{
				MatchSourceIp: true,
				IpSet:         srcIPSet,
			})
		}
		if len(routerRuleConfig.DstCidrs) > 0 || len(routerRuleConfig.DstIpTags) > 0 {
			var cidrs []*cgeo.CIDR
			for _, cidr := range routerRuleConfig.DstCidrs {
				prefix, err := netip.ParsePrefix(cidr)
				if err != nil {
					return nil, err
				}
				cidrs = append(cidrs, &cgeo.CIDR{
					Ip:     prefix.Addr().AsSlice(),
					Prefix: uint32(prefix.Bits()),
				})
			}
			dstIPSet, err := geo.NewIPSet(routerRuleConfig.DstIpTags, config.GeoHelper, cidrs...)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, &IpMatcher{
				IpSet:      dstIPSet,
				IpResolver: config.IpResolver,
				Resolve:    routerRuleConfig.ResolveDomain,
			})
		}
		if len(routerRuleConfig.GeoDomains) > 0 || len(routerRuleConfig.DomainTags) > 0 {
			domainSet, err := geo.NewDomainSet(routerRuleConfig.DomainTags, config.GeoHelper, routerRuleConfig.GeoDomains...)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, &DomainMatcher{
				DomainSet: domainSet,
				SkipSniff: routerRuleConfig.SkipSniff,
				Sniffer:   sniff.NewSniffer(),
			})
		}
		if len(routerRuleConfig.Networks) > 0 {
			conditions = append(conditions, NewNetworkMatcher(routerRuleConfig.Networks))
		}
		if len(routerRuleConfig.SrcPortRanges) > 0 {
			conditions = append(conditions, NewPortMatcher(routerRuleConfig.SrcPortRanges, true))
		}
		if len(routerRuleConfig.DstPortRanges) > 0 {
			conditions = append(conditions, NewPortMatcher(routerRuleConfig.DstPortRanges, false))
		}
		if len(routerRuleConfig.Usernames) > 0 {
			conditions = append(conditions, NewUserMatcher(routerRuleConfig.Usernames))
		}
		if len(routerRuleConfig.AppIds) > 0 || len(routerRuleConfig.AppTags) > 0 {
			appSet, err := geo.NewAppSet(routerRuleConfig.AppTags, config.GeoHelper, routerRuleConfig.AppIds...)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, &AppIdMatcher{
				AppSet: appSet,
			})
		}
		if routerRuleConfig.FakeIp {
			conditions = append(conditions, &ConditionFakeIp{})
		}
		if len(routerRuleConfig.AllTags) > 0 {
			domainSet, err := geo.NewDomainSet(routerRuleConfig.AllTags, config.GeoHelper)
			if err != nil {
				return nil, err
			}
			ipSet, err := geo.NewIPSet(routerRuleConfig.AllTags, config.GeoHelper)
			if err != nil {
				return nil, err
			}
			appSet, err := geo.NewAppSet(routerRuleConfig.AllTags, config.GeoHelper)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, &AllMatcher{
				domainMatcher: &DomainMatcher{
					DomainSet: domainSet,
					SkipSniff: routerRuleConfig.SkipSniff,
					Sniffer:   sniff.NewSniffer(),
				},
				ipMatcher: &IpMatcher{
					IpSet:      ipSet,
					IpResolver: config.IpResolver,
					Resolve:    routerRuleConfig.ResolveDomain,
				},
				appIdMatcher: &AppIdMatcher{
					AppSet: appSet,
				},
			})
		}
		if routerRuleConfig.MatchAll {
			conditions = []Condition{
				&ConditionTrue{},
			}
		}
		rule := NewRule(routerRuleConfig.RuleName, routerRuleConfig.OutboundTag, routerRuleConfig.SelectorTag, conditions...)
		r.AddRule(rule)
	}
	return r, nil
}

func (r *Router) AddRule(rule *rule) {
	r.rules = append(r.rules, rule)
}

var ErrNoHandler = errors.New("no handler")
var ErrSelectorNotFound = errors.New("selector not found")
var ErrBlocked = errors.New("block")
var ErrNoRule = errors.New("no rule matched")

func (r *Router) PickHandler(ctx context.Context, si *session.Info) (i.Outbound, error) {
	_, handler, err := r.PickHandlerWithData(ctx, si, nil)
	return handler, err
}

func (r *Router) PickHandlerWithData(ctx context.Context, si *session.Info, rw interface{}) (interface{}, i.Outbound, error) {
	// for tests
	if len(r.rules) == 0 {
		if h := r.om.GetHandler(""); h != nil {
			return rw, h, nil
		} else {
			return rw, r.om.GetHandler("direct"), nil
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
					return rw, h, nil
				}
				return rw, nil, ErrNoHandler
			} else if rule.selectorTag != "" {
				if se := r.selectors.GetSelector(rule.selectorTag); se != nil {
					si.UsedSelector = rule.selectorTag
					if h := se.GetHandler(si); h != nil {
						return rw, h, nil
					}
					return rw, nil, ErrNoHandler
				}
				log.Ctx(ctx).Warn().Str("selector_tag", rule.selectorTag).Msg("selector not found")
				return rw, nil, ErrSelectorNotFound
			} else {
				return rw, nil, ErrBlocked
			}
		}
	}

	return rw, nil, ErrNoRule
}

// func ipv6Process(si *session.Info, handler i.Outbound) i.Outbound {
// 	// Only for global unicast IPv6?
// 	if si.Target.IsValid() && si.Target.Address.Family().IsIP() &&
// 		si.Target.Address.Family().IsIPv6() && si.Target.Address.IP().IsGlobalUnicast() {
// 		if h, ok := handler.(*outbound.Handler); ok && h.NotSupport6() {
// 			return nil
// 		}
// 	}
// 	return handler
// }

// func (r *Router) AddRule(v *rule) {
// 	r.Lock()
// 	defer r.Unlock()
// 	newRules := make([]*rule, 0, len(r.rules)+1)
// 	newRules = append(newRules, v)
// 	newRules = append(newRules, r.rules...)
// 	r.rules = newRules
// }
// func (r *Router) RemoveRule(v *rule) {
// 	r.Lock()
// 	defer r.Unlock()
// 	newRules := make([]*rule, 0, len(r.rules))
// 	for _, rule := range r.rules {
// 		if rule != v {
// 			newRules = append(newRules, rule)
// 		}
// 	}
// 	r.rules = newRules
// }
