//go:build android

package x_android

import (
	"sync"

	"github.com/5vnetwork/vx-core/app/api"

	"google.golang.org/protobuf/proto"
)

var lock sync.Mutex
var apiServer *api.Api

func StartApiServer(configBytes []byte) error {
	lock.Lock()
	defer lock.Unlock()
	if apiServer != nil {
		apiServer.Stop()
		apiServer = nil
	}
	var config api.ApiServerConfig
	err := proto.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}
	as, err := api.StartApiServer(&config)
	if err != nil {
		return err
	}
	apiServer = as
	return nil
}

func StopApiServer() {
	lock.Lock()
	defer lock.Unlock()
	if apiServer != nil {
		apiServer.Stop()
		apiServer = nil
	}
}

func SetSecret(id uint32, secret string) {
	apiServer.SetSecret(id, secret)
}
