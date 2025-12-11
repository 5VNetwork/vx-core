package tun

import (
	"net"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/tun/netmon"
)

func TestDnsServers(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("failed to get interfaces: %v", err)
	}
	// Test with first available interface
	servers, err := DnsServers(interfaces[0].Index)
	if err != nil {
		t.Logf("Note: DNS servers might not be configured for interface")
		// Don't fail the test as the interface might not have DNS configured
	}

	// Validate returned servers if any were found
	for _, server := range servers {
		if !server.IsValid() {
			t.Errorf("Invalid IP address returned: %v", server)
		}
	}

	// Test with invalid interface index
	invalidIndex := -1
	_, err = DnsServers(invalidIndex)
	if err == nil {
		t.Error("Expected error for invalid interface index, got nil")
	}

	// Test with non-existent interface index
	largeIndex := 99999
	_, err = DnsServers(largeIndex)
	if err == nil {
		t.Error("Expected error for non-existent interface index, got nil")
	}
}

func TestDefaultInterfaceInfo(t *testing.T) {
	mon, err := NewInterfaceMonitor("")
	if err != nil {
		t.Fatalf("failed to create interface monitor: %v", err)
	}
	mon.Start()
	defer mon.Close()
	time.Sleep(time.Second * 1)
	t.Logf("default interface: %v", mon.DefaultInterface4())
	for _, d := range mon.DefaultDns4() {
		t.Logf("default dns: %v", d)
	}
	state := mon.monitor.InterfaceState()
	t.Logf("state: %v", state)

	has, err := mon.HasGlobalIPv6()
	if err != nil {
		t.Fatalf("failed to get global ipv6: %v", err)
	}
	t.Logf("support6: %v", has)
}

func TestNetmonState(t *testing.T) {
	state, err := netmon.GetState()
	if err != nil {
		t.Fatalf("failed to get netmon state: %v", err)
	}
	t.Logf("state: %v", state)
}

func TestGetPrimaryPhysicalInterface0(t *testing.T) {
	device, err := GetPrimaryPhysicalInterface()
	if err != nil {
		t.Fatalf("failed to get primary physical interface: %v", err)
	}
	t.Logf("primary physical interface: %v", device)
	iface, err := net.InterfaceByIndex(device.Index)
	if err != nil {
		t.Fatalf("failed to get interface by index: %v", err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		t.Fatalf("failed to get addresses: %v", err)
	}
	for _, addr := range addrs {
		t.Logf("address: %v", addr)
	}
}
