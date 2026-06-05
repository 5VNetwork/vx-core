package nic

import (
	"log"
	"testing"

	"golang.org/x/sys/windows"
)

func TestNICMonitorName(t *testing.T) {
	m, err := NewInterfaceMonitor("")
	if err != nil {
		t.Error(err)
	}
	t.Logf("name4: %s, name6: %s", m.name4, m.name6)
}

func TestNICMonitorDns(t *testing.T) {
	m, err := NewInterfaceMonitor("singbox_tun")
	if err != nil {
		t.Error(err)
	}
	log.Println(m.dnsAddrs4)
	if len(m.dnsAddrs4) == 0 {
		t.Error("dnsAddrs is empty")
	}
}

func TestNICMonitorStartClose(t *testing.T) {
	m, err := NewInterfaceMonitor("")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNICMonitorFindDefaultLUID1(t *testing.T) {
	name, luid, index, err := FindDefaultLUID1(windows.AF_INET)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("name: %s, luid: %d, index: %d", name, luid, index)
}
