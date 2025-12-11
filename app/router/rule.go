package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
)

type Condition interface {
	Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool)
}

type rule struct {
	conditions  []Condition
	outboundTag string
	selectorTag string
	name        string
}

func NewRule(name string, outboundTag, selectorTag string, conditions ...Condition) *rule {
	r := &rule{
		name:        name,
		conditions:  conditions,
		outboundTag: outboundTag,
		selectorTag: selectorTag,
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
