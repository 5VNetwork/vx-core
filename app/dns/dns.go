// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/miekg/dns"
)

type AllDnsServers struct {
	dnsServers []DnsServer
}

func NewAllDnsServers(dnsServers []DnsServer) *AllDnsServers {
	d := &AllDnsServers{
		dnsServers: dnsServers,
	}
	return d
}

func (dsp *AllDnsServers) Start() error {
	for _, client := range dsp.dnsServers {
		if err := client.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (dsp *AllDnsServers) Close() error {
	for _, dnsServer := range dsp.dnsServers {
		if err := dnsServer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (dsp *AllDnsServers) IsIPInIPPool(ip net.Address) bool {
	for _, dnsServer := range dsp.dnsServers {
		if isFakeDns(dnsServer) {
			if fakeDns, ok := dnsServer.(*FakeDns); ok {
				if fakeDns.IsIPInIPPool(ip) {
					return true
				}
			}
		}
	}
	return false
}

func (dsp *AllDnsServers) GetDomainFromFakeDNS(ip net.Address) string {
	for _, dnsServer := range dsp.dnsServers {
		if isFakeDns(dnsServer) {
			if fakeDns, ok := dnsServer.(*FakeDns); ok {
				if d := fakeDns.GetDomainFromFakeDNS(ip); d != "" {
					return d
				}
			}
		}
	}
	return ""
}

func isFakeDns(dnsServer DnsServer) bool {
	_, ok := dnsServer.(*FakeDns)
	return ok
}

func addClientIP(msg *dns.Msg, clientIp net.IP) {
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT

	subnet := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        1, // IPv4
		SourceNetmask: 24,
		SourceScope:   0,
		Address:       clientIp.To4(),
	}
	if clientIp.To4() == nil && clientIp.To16() != nil {
		subnet.Family = 2 // IPv6
		subnet.SourceNetmask = 64
		subnet.Address = clientIp.To16()
	}
	o.Option = append(o.Option, subnet)
	msg.Extra = append(msg.Extra, o)
}
