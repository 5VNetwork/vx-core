//go:build !android

package api

import (
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/nic"
)

func getNicMonitor(tunName string) (i.DefaultInterfaceInfo, error) {
	mon, err := nic.NewInterfaceMonitor(tunName)
	if err != nil {
		return nil, err
	}
	return mon, nil
}
