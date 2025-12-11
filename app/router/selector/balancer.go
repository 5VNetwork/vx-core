package selector

import (
	"path/filepath"
	"runtime"

	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

type Balancer interface {
	GetHandler(*session.Info) i.Outbound
	UpdateHandlers(handlers []i.HandlerWith6Info)
	Support6() bool
}

func getApp(app string) string {
	if runtime.GOOS == "android" {
		return app
	}
	dir := filepath.Dir(app)
	if dir == "." {
		return app
	}
	return dir
}
