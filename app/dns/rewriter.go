package dns

import (
	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type MsgRewriter interface {
	Rewrite(msg *dns.Msg) *dns.Msg
}

type msgRewriter struct {
	ipset i.IPSet
}

type MsgRewriterOption struct {
	IPSet i.IPSet
}

func NewMsgRewriter(opts MsgRewriterOption) MsgRewriter {
	return &msgRewriter{
		ipset: opts.IPSet,
	}
}

func (r *msgRewriter) Rewrite(msg *dns.Msg) *dns.Msg {
	if r.ipset == nil || msg == nil {
		return msg
	}
	msg.Answer = filterIPRecords(msg.Answer, r.ipset)
	msg.Ns = filterIPRecords(msg.Ns, r.ipset)
	msg.Extra = filterIPRecords(msg.Extra, r.ipset)
	return msg
}

// filterIPRecords keeps only A/AAAA RRs whose IP is in the set; other RR types are kept unchanged.
// It filters in-place and returns a slice over the same backing array, so it does not allocate.
func filterIPRecords(rrs []dns.RR, set i.IPSet) []dns.RR {
	if set == nil {
		return rrs
	}
	n := 0
	for _, rr := range rrs {
		keep := true
		switch r := rr.(type) {
		case *dns.A:
			keep = set.Match(r.A)
			// log.Debug().IPAddr("ip", r.A).Bool("keep", keep).Msg("filtered A record")
		case *dns.AAAA:
			keep = set.Match(r.AAAA)
			// log.Debug().IPAddr("ip", r.AAAA).Bool("keep", keep).Msg("filtered AAAA record")
		default:
			keep = true
		}
		if keep {
			rrs[n] = rr
			n++
		}
	}
	log.Debug().Int("n", n).Int("len", len(rrs)).Msg("filtered IP records")
	return rrs[:n]
}
