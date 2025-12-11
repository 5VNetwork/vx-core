package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
)

type ConditionTrue struct{}

func (c *ConditionTrue) Apply(ctx context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	return rw, true
}
