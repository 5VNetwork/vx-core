// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
)

func DialerFactory(config *configs.TmConfig, fc *Builder, client *client.Client) error {
	// dialer factory
	if config.GetDialerFactory() != nil {
		err := fc.requireFeature(func(bdl i.DefaultInterfaceInfo, ipResolver i.IPResolver) error {
			opt := transport.DialerFactoryOption{
				BindToDefaultNIC:        config.GetDialerFactory().GetShouldBindDevice(),
				IpResolver:              ipResolver,
				DefaultInterfaceMonitor: bdl,
				DialTimeout:             time.Duration(config.GetDialerFactory().GetDialTimeout()) * time.Second,
			}
			if config.GetDialerFactory().GetShouldBindDevice() && runtime.GOOS == "android" {
				fdFunc := fc.getFeature(reflect.TypeOf((*transport.FdFunc)(nil)).Elem())
				opt.FdFunc = fdFunc.(transport.FdFunc)
				opt.BindToDefaultNIC = false
			}

			df := transport.NewDialerFactoryImp(opt)
			client.DialerFactory = df
			return fc.addComponent(df)
		})
		if err != nil {
			return fmt.Errorf("failed to require features: %w", err)
		}
	} else {
		client.DialerFactory = transport.DefaultDialerFactory()
		common.Must(fc.addComponent(client.DialerFactory))
	}

	return nil
}
