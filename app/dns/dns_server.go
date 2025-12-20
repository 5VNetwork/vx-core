package dns

import (
	"context"
	"errors"
	"strings"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"

	"github.com/miekg/dns"
)

type DnsServer interface {
	common.Runnable
	i.DnsServer
}

type IPOption struct {
	IPv4Enable bool
	IPv6Enable bool
}

type ResolverGetter interface {
	// return true and the nameserver if [ResolverGetter] actually handles the dns query for [domain]
	// and get result of [ip]
	GetResolver(domain string, ip net.Address) (string, bool)
}

type ConditionDnsServer struct {
	DnsServer
	conditions []Condition
}

type Condition interface {
	Match(msg *DnsMsgMeta) bool
}

func NewConditionDnsServer(dnsServer DnsServer, conditions ...Condition) *ConditionDnsServer {
	return &ConditionDnsServer{
		DnsServer:  dnsServer,
		conditions: conditions,
	}
}

func (d *ConditionDnsServer) MatchConditions(msg *DnsMsgMeta) bool {
	for _, condition := range d.conditions {
		if !condition.Match(msg) {
			return false
		}
	}
	return true
}

// UnFqdn removes the trailing dot from the domain
func UnFqdn(domain string) string {
	if len(domain) > 1 && strings.HasSuffix(domain, ".") {
		return domain[:len(domain)-1]
	}
	return domain
}

var ErrConditionNotMatch = errors.New("condition not match")
var ErrAllServersFailed = errors.New("all dns servers failed")
var ErrNoQuestion = errors.New("no question in dns query")

// if ip option conflicts, return a response with no answer
type IpOptionDnsServer struct {
	IpOption *IPOption
	DnsServer
}

// msg should contain a domain that will go proxy
func (p *IpOptionDnsServer) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	if !p.IpOption.IPv4Enable && msg.Question[0].Qtype == dns.TypeA ||
		!p.IpOption.IPv6Enable && msg.Question[0].Qtype == dns.TypeAAAA {
		log.Ctx(ctx).Debug().Str("domain", msg.Question[0].Name).Str("type", dns.Type(msg.Question[0].Qtype).String()).
			Uint16("id", msg.Id).Msg("ip option dns server return empty answer")
		return emptyReply(msg), nil
	}
	return p.DnsServer.HandleQuery(ctx, msg, tcp)
}

func emptyReply(msg *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	resp.SetReply(msg)
	resp.RecursionAvailable = true
	resp.Rcode = dns.RcodeSuccess
	return resp
}
