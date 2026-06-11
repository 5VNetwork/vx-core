// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package router

import (
	"context"

	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

type DomainMatcher struct {
	DomainSet i.DomainSet
	SkipSniff bool
	Sniffer   *sniff.Sniffer
}

func (m *DomainMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	if info.Target.Address == nil {
		return rw, false
	}
	if targetDomain := info.GetTargetDomain(); targetDomain != "" {
		return rw, m.DomainSet.Match(targetDomain)
	}
	if m.SkipSniff {
		return rw, false
	}
	if !info.Sniffed && rw != nil && m.Sniffer != nil {
		if readerWriter, ok := rw.(buf.ReaderWriter); ok {
			rw, _ = m.Sniffer.Sniff(c, info, readerWriter)
		}
	}
	if info.SniffedDomain != "" {
		return rw, m.DomainSet.Match(info.SniffedDomain)
	}
	return rw, false
}

type HasNoDomain struct {
}

func (h *HasNoDomain) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	if targetDomain := info.GetTargetDomain(); targetDomain == "" {
		return rw, true
	}
	return rw, false
}
