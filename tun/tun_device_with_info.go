package tun

import "net/netip"

type tunDeviceWithInfo struct {
	TunDevice
	ip4        netip.Addr
	ip6        netip.Addr
	dnsServers []netip.Addr
}

func NewTunDeviceWithInfo(tunDevice TunDevice, ip4 netip.Addr,
	ip6 netip.Addr, dnsServers []netip.Addr) TunDeviceWithInfo {
	return &tunDeviceWithInfo{
		TunDevice:  tunDevice,
		ip4:        ip4,
		ip6:        ip6,
		dnsServers: dnsServers,
	}
}

func (t *tunDeviceWithInfo) IP4() netip.Addr {
	return t.ip4
}

func (t *tunDeviceWithInfo) IP6() netip.Addr {
	return t.ip6
}

func (t *tunDeviceWithInfo) DnsServers() []netip.Addr {
	return t.dnsServers
}
