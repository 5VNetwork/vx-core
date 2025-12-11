package dns

import (
	"context"
	"errors"

	"github.com/5vnetwork/vx-core/common/net"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type AllFakeDns struct {
	fakeDns []*FakeDns
}

func (a *AllFakeDns) AddFakeDns(fakeDns *FakeDns) {
	a.fakeDns = append(a.fakeDns, fakeDns)
}

func (a *AllFakeDns) IsIPInIPPool(ip net.Address) bool {
	for _, fakeDns := range a.fakeDns {
		if fakeDns.IsIPInIPPool(ip) {
			return true
		}
	}
	return false
}

func (a *AllFakeDns) GetDomainFromFakeDNS(ip net.Address) string {
	for _, fakeDns := range a.fakeDns {
		if domain := fakeDns.GetDomainFromFakeDNS(ip); domain != "" {
			return domain
		}
	}
	return ""
}

type FakeDns struct {
	pools Pools
}

func (*FakeDns) Start() error {
	return nil
}

func (*FakeDns) Close() error {
	return nil
}

func (*FakeDns) Name() string {
	return "fakedns"
}

func NewFakeDns(pools Pools) *FakeDns {
	f := &FakeDns{
		pools: pools,
	}
	return f
}

func (f *FakeDns) IsIPInIPPool(ip net.Address) bool {
	pools := f.pools
	if pools == nil {
		return false
	}
	return pools.IsIPInIPPool(ip)
}

func (f *FakeDns) GetDomainFromFakeDNS(ip net.Address) string {
	pools := f.pools
	if pools == nil {
		return ""
	}
	return pools.GetDomainFromFakeDNS(ip)
}

func (f *FakeDns) GetFakeIP(domain string) ([]net.IP, error) {
	pools := f.pools
	if pools == nil {
		return nil, errors.New("fake dns pool is not initialized")
	}
	return pools.GetFakeIP(domain)
}

func (f *FakeDns) GetResolver(domain string, ip net.Address) (string, bool) {
	if f.IsIPInIPPool(ip) {
		if f.GetDomainFromFakeDNS(ip) == domain {
			return "fakedns", true
		}
	}
	return "", false
}

func (f *FakeDns) HandleQuery(ctx context.Context, msg *dns.Msg, _ bool) (*dns.Msg, error) {
	resp := new(dns.Msg)
	resp.SetReply(msg)
	resp.RecursionAvailable = true

	// only handle A and AAAA queries
	if msg.Question[0].Qtype != dns.TypeA && msg.Question[0].Qtype != dns.TypeAAAA {
		return nil, errors.New("only A and AAAA queries are supported")
	}
	pools := f.pools
	if pools == nil {
		return nil, errors.New("fake dns pool is not initialized")
	}
	domain := UnFqdn(msg.Question[0].Name)
	if msg.Question[0].Qtype == dns.TypeA {
		ip, err := pools.GetFakeIPv4(domain)
		if err != nil {
			return nil, err
		}
		log.Ctx(ctx).Debug().Str("domain", domain).IPAddr("ip", ip).Msg("fake dns")
		resp.Answer = append(resp.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    5,
			},
			A: ip,
		})
	} else {
		ip, err := pools.GetFakeIPv6(domain)
		if err != nil {
			return nil, err
		}
		log.Ctx(ctx).Debug().Str("domain", domain).IPAddr("ip", ip).Msg("fake dns")
		resp.Answer = append(resp.Answer, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    5,
			},
			AAAA: ip,
		})
	}
	return resp, nil
}
