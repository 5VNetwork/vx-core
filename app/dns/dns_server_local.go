package dns

import (
	"errors"
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type StaticDnsServer struct {
	cache map[dns.Question]*dns.Msg
}

func NewStaticDnsServer(records []*configs.Record) *StaticDnsServer {
	s := &StaticDnsServer{
		cache: make(map[dns.Question]*dns.Msg),
	}
	for _, record := range records {
		var ip4List, ip6List []net.Address
		for _, ip := range record.Ip {
			ipAddr := net.ParseAddress(ip)
			if ipAddr == nil {
				log.Warn().Str("ip", ip).Msg("invalid ip")
				continue
			}
			if ipAddr.Family().IsIPv4() {
				ip4List = append(ip4List, ipAddr)
			} else {
				ip6List = append(ip6List, ipAddr)
			}
		}
		if len(ip4List) > 0 {
			msg := new(dns.Msg)
			msg.Answer = make([]dns.RR, 0, len(ip4List))
			for _, ip := range ip4List {
				rr, err := dns.NewRR(fmt.Sprintf("%s 3600 IN A %s", record.Domain, ip.IP()))
				if err != nil {
					log.Warn().Err(err).Str("domain", record.Domain).Msg("failed to create A record")
					continue
				}
				msg.Answer = append(msg.Answer, rr)
			}
			question := dns.Question{
				Name:   dns.Fqdn(record.Domain), // Ensures domain ends with a dot
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			}
			msg.Question = []dns.Question{question}
			s.cache[question] = msg
		}
		if len(ip6List) > 0 {
			msg := new(dns.Msg)
			msg.Answer = make([]dns.RR, 0, len(ip6List))
			for _, ip := range ip6List {
				rr, err := dns.NewRR(fmt.Sprintf("%s 3600 IN AAAA %s", record.Domain, ip.IP()))
				if err != nil {
					log.Warn().Err(err).Str("domain", record.Domain).Msg("failed to create AAAA record")
					continue
				}
				msg.Answer = append(msg.Answer, rr)
			}
			question := dns.Question{
				Name:   dns.Fqdn(record.Domain), // Ensures domain ends with a dot
				Qtype:  dns.TypeAAAA,
				Qclass: dns.ClassINET,
			}
			msg.Question = []dns.Question{question}
			s.cache[question] = msg
		}
		if record.ProxiedDomain != "" {
			msg := new(dns.Msg)
			question := dns.Question{
				Name:   dns.Fqdn(record.Domain),
				Qtype:  dns.TypeCNAME,
				Qclass: dns.ClassINET,
			}
			msg.Question = []dns.Question{question}
			rr, err := dns.NewRR(fmt.Sprintf("%s 3600 IN CNAME %s", record.Domain, record.ProxiedDomain))
			if err != nil {
				log.Warn().Err(err).Str("domain", record.Domain).Msg("failed to create CNAME record")
				continue
			}
			msg.Answer = []dns.RR{rr}
			s.cache[question] = msg
		}
	}
	return s
}

var ErrNotFound = errors.New("not found")

func (s *StaticDnsServer) ReplyFor(msg *dns.Msg) (*dns.Msg, bool) {
	entry, ok := s.cache[msg.Question[0]]
	if !ok {
		return nil, false
	}
	return entry, true
}
