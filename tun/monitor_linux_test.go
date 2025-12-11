package tun

import (
	"net"
	"testing"
)

func TestDnsServers(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	servers, err := DnsServers(interfaces[0].Index)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("servers: %v", servers)
}

func TestGetDnsFromSystemdResolved(t *testing.T) {
	servers, err := getDnsFromSystemdResolved("wlp1s0")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(" servers: %v", servers)
}

func TestGetDnsFromNetworkManager(t *testing.T) {
	servers, err := getDnsFromNetworkManager("wlp1s0")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("servers: %v", servers)
}

func TestGetDnsFromResolvConf(t *testing.T) {
	servers, err := getDnsFromResolvConf()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("servers: %v", servers)
}
