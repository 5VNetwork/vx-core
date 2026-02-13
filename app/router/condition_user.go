// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
)

type userMatcher struct {
	user []string
}

func NewUserMatcher(users []string) *userMatcher {
	usersCopy := make([]string, 0, len(users))
	for _, user := range users {
		if len(user) > 0 {
			usersCopy = append(usersCopy, user)
		}
	}
	return &userMatcher{
		user: usersCopy,
	}
}

// Apply implements Condition.
func (v *userMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	user := info.GetUser()
	for _, u := range v.user {
		if u == user.Uid() {
			return rw, true
		}
	}
	return rw, false
}
