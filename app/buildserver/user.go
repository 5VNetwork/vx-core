//go:build server || test

package buildserver

import (
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/user"
	"go.uber.org/fx"
)

type UserManagerParams struct {
	fx.In
	Configs []*configs.UserConfig
}
type UserManagerResult struct {
	fx.Out
	UserManager *user.Manager
}

func NewUserManager(lc fx.Lifecycle, params UserManagerParams) (UserManagerResult, error) {
	um := user.NewManager()
	for _, userConfig := range params.Configs {
		u := user.NewUser(userConfig.Id, userConfig.Secret)
		um.AddUser(u)
	}
	return UserManagerResult{UserManager: um}, nil
}
