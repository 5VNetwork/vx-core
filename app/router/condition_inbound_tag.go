package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
)

type InboundTagMatcher struct {
	tags []string
}

func NewInboundTagMatcher(tags []string) *InboundTagMatcher {
	tagsCopy := make([]string, 0, len(tags))
	for _, tag := range tags {
		if len(tag) > 0 {
			tagsCopy = append(tagsCopy, tag)
		}
	}
	return &InboundTagMatcher{
		tags: tagsCopy,
	}
}

// Apply implements Condition.
func (v *InboundTagMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	tag := info.GetInboundTag()
	if len(tag) == 0 {
		return rw, false
	}
	for _, t := range v.tags {
		if t == tag {
			return rw, true
		}
	}
	return rw, false
}
