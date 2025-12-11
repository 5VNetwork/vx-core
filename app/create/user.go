package create

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfigs "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/proxy/shadowsocks"
)

func UserConfigToUser(config *configs.UserConfig) (*user.User, error) {
	return user.NewUser(config.Id, config.UserLevel, config.Secret), nil
}

func ShadowsocksAccountToMemoryAccount(account *proxyconfigs.ShadowsocksAccount) (*shadowsocks.MemoryAccount, error) {
	return shadowsocks.NewMemoryAccount(
		account.User.Id,
		shadowsocks.CipherType(account.CipherType),
		account.User.Secret,
		account.ExperimentReducedIvHeadEntropy,
		account.IvCheck,
	)
}
