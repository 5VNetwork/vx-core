//go:build darwin || linux

package util_test

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/tun"
)

func TestIpv6(t *testing.T) {
	handler := freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "freedom", nil)
	response, err := util.TestIpv6(context.Background(), handler, util.AliDNS6)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
}

func TestNICSupportIPv6(t *testing.T) {
	device, err := tun.GetPrimaryPhysicalInterface()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("default nic %s", device.Name)
	support6 := util.NICSupportIPv6Index(uint32(device.Index))
	t.Logf("support6: %v", support6)
}

func TestNICHasGlobalIPv6Address(t *testing.T) {
	device, err := tun.GetPrimaryPhysicalInterface()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("default nic %s", device.Name)
	support6, err := util.NICHasGlobalIPv6Address(uint32(device.Index))
	t.Logf("support6: %v", support6)
}
