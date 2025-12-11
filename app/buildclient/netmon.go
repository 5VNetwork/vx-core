package buildclient

import (
	"fmt"
	"reflect"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
)

func Netmon(config *configs.TmConfig, builder *Builder, client *client.Client) error {
	if builder.getFeature(reflect.TypeOf((*i.DefaultInterfaceInfo)(nil)).Elem()) != nil {
		return nil
	}

	// monitor
	monitor, err := NewInterfaceMonotor(config.GetTun().GetDevice().GetName(), builder)
	if err != nil {
		return fmt.Errorf("failed to create tun interface monitor: %w", err)
	}
	common.Must(builder.addComponent(monitor))
	return nil
}
