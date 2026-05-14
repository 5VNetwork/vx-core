// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package router

import (
	"context"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

type Condition interface {
	Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool)
}

type rule struct {
	conditions  []Condition
	outboundTag string
	selectorTag string
	name        string
	retries     []i.Fallback
}

func NewRule(name string, outboundTag, selectorTag string,
	fallbackers []i.Fallback, conditions ...Condition) *rule {
	r := &rule{
		name:        name,
		conditions:  conditions,
		outboundTag: outboundTag,
		selectorTag: selectorTag,
		retries:     fallbackers,
	}

	return r
}

func (r *rule) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	if len(r.conditions) == 0 {
		return rw, false
	}
	for _, cond := range r.conditions {
		rw0, match := cond.Apply(c, info, rw)
		rw = rw0
		if !match {
			return rw, false
		}
	}
	return rw, true
}

func (r *rule) Name() string {
	return r.name
}

type Fallback struct {
	// one of selectorTag and outboundTag must be set
	selectorTag string
	outboundTag string
	action      *configs.RuleConfig_Fallback_Action
	// must have at least one condition
	conditions []Condition
	om         i.OutboundManager
	selectors  *selector.Selectors
	final      bool
}

func NewFallbacker(selectorTag, outboundTag string,
	action *configs.RuleConfig_Fallback_Action, om i.OutboundManager,
	selectors *selector.Selectors, final bool, conditions ...Condition) *Fallback {
	return &Fallback{
		selectorTag: selectorTag,
		outboundTag: outboundTag,
		action:      action,
		conditions:  conditions,
		om:          om,
		selectors:   selectors,
		final:       final,
	}
}

func (f *Fallback) GetHandler(ctx context.Context, info *session.Info, err error) (i.Outbound, bool) {
	if len(f.conditions) == 0 {
		return nil, false
	}
	for _, cond := range f.conditions {
		_, match := cond.Apply(ctx, info, nil)
		if !match {
			return nil, false
		}
	}
	if f.action != nil {
		if f.action.IpToDomain && info.GetTargetIP() != nil && info.GetTargetDomain() != "" {
			info.Target.Address = net.ParseAddress(info.GetTargetDomain())
		}
	}
	if f.selectorTag != "" {
		if se := f.selectors.GetSelector(f.selectorTag); se != nil {
			return se.GetHandler(info), f.final
		} else {
			return nil, false
		}
	} else {
		return f.om.GetHandler(f.outboundTag), f.final
	}
}
