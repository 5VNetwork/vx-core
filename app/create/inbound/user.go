// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package inbound

import (
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/user"
)

func UserConfigToUser(config *configs.UserConfig) (*user.User, error) {
	return user.NewUser(config.Id, config.Secret), nil
}
