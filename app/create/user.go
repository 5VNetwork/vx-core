// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package create

import (
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/proxy/shadowsocks"
)

func UserConfigToUser(config *configs.UserConfig) (*user.User, error) {
	return user.NewUser(config.Id, config.Secret), nil
}

func ShadowsocksAccountToMemoryAccount(account *configs.ShadowsocksAccount) (*shadowsocks.MemoryAccount, error) {
	return shadowsocks.NewMemoryAccount(
		user.NewUser(account.User.Id, account.User.Secret),
		shadowsocks.CipherType(account.CipherType),
		account.ExperimentReducedIvHeadEntropy,
		account.IvCheck,
	)
}
