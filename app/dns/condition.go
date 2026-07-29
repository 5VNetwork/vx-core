// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"slices"

	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
)

type HasSrcCondition struct{}

func (h *HasSrcCondition) Match(msg *DnsMsgMeta) bool {
	return msg.Src != nil
}

type ExcludeDomainCondition struct {
	DomainSet i.DomainSet
}

func (e *ExcludeDomainCondition) Match(msg *dns.Msg) bool {
	return !e.DomainSet.Match(UnFqdn(msg.Question[0].Name))
}

type PreferDomainCondition struct {
	DomainSet i.DomainSet
}

func (p *PreferDomainCondition) Match(msg *DnsMsgMeta) bool {
	return p.DomainSet.Match(UnFqdn(msg.Question[0].Name))
}

type IncludedTypesCondition struct {
	Types []uint16
}

func (i *IncludedTypesCondition) Match(msg *DnsMsgMeta) bool {
	return slices.Contains(i.Types, msg.Question[0].Qtype)
}
