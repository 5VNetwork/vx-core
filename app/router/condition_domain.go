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
	if info.Target.Address.Family().IsDomain() {
		return rw, m.DomainSet.Match(info.Target.Address.Domain())
	}
	if m.SkipSniff {
		return rw, false
	}
	if !info.Sniffed && rw != nil {
		if readerWriter, ok := rw.(buf.ReaderWriter); ok {
			rw, _ = m.Sniffer.Sniff(c, info, readerWriter)
		}
	}
	if info.SniffedDomain != "" {
		return rw, m.DomainSet.Match(info.SniffedDomain)
	}
	// only consider ipToDomain when sniffed but no domain sniffed out to avoid cdn issues:
	// when the ip is cdn ip, the ipToDomain might be wrong. But if sniff failed to get domain,
	// the ip is not likely to be cdn ip.
	// if info.Sniffed && info.IpToDomain != "" {
	// 	return rw, m.DomainSet.Match(info.IpToDomain)
	// }
	return rw, false
}
