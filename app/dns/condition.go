package dns

import (
	"slices"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
)

type HasSrcCondition struct{}

func (h *HasSrcCondition) Match(msg *DnsMsgMeta) bool {
	return msg.Src != nil
}

type FakeDnsCondition struct {
	FakeDnsEnabled *atomic.Bool
}

func (f *FakeDnsCondition) Match(msg *DnsMsgMeta) bool {
	return f.FakeDnsEnabled.Load()
}

type ExcludeDomainCondition struct {
	DomainSet i.DomainSet
}

func (e *ExcludeDomainCondition) Match(msg *dns.Msg) bool {
	return !e.DomainSet.Match(UnFqdn(msg.Question[0].Name))
}

type PreferDomainCondition struct {
	lastMatched   string
	lastUnmathced string
	DomainSet     i.DomainSet
}

func (p *PreferDomainCondition) Match(msg *DnsMsgMeta) bool {
	domain := UnFqdn(msg.Question[0].Name)
	if p.lastMatched == domain {
		return true
	}
	if p.lastUnmathced == domain {
		return false
	}
	matched := p.DomainSet.Match(domain)
	if matched {
		p.lastMatched = domain
	} else {
		p.lastUnmathced = domain
	}
	return matched
}

type IncludedTypesCondition struct {
	Types []uint16
}

func (i *IncludedTypesCondition) Match(msg *DnsMsgMeta) bool {
	return slices.Contains(i.Types, msg.Question[0].Qtype)
}
