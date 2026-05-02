// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"

	"github.com/5vnetwork/vx-core/common"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// ConcurrentDnsServers fans out a query to all underlying [DnsServer]s in parallel
// and returns the first definitive response. A definitive response is NOERROR with
// at least one answer record. NXDOMAIN and NODATA (NOERROR with empty answer) are
// kept as fallbacks and only returned if no server produces a definitive answer.
// SERVFAIL/REFUSED/transport errors are ignored entirely.
type ConcurrentDnsServers struct {
	DnsServers []DnsServer
}

func NewConcurrentDnsServers(servers ...DnsServer) *ConcurrentDnsServers {
	return &ConcurrentDnsServers{DnsServers: servers}
}

func (c *ConcurrentDnsServers) Start() error {
	return common.StartAll(c.DnsServers)
}

func (c *ConcurrentDnsServers) Close() error {
	return common.CloseAll(c.DnsServers)
}

func (c *ConcurrentDnsServers) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	if len(c.DnsServers) == 0 {
		return nil, ErrAllServersFailed
	}
	if len(c.DnsServers) == 1 {
		return c.DnsServers[0].HandleQuery(ctx, msg, tcp)
	}

	queryCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan *dns.Msg, len(c.DnsServers))
	for i, s := range c.DnsServers {
		go func(server DnsServer) {
			newMsg := msg
			if i != len(c.DnsServers)-1 {
				newMsg = msg.Copy()
			}
			resp, err := server.HandleQuery(queryCtx, newMsg, tcp)
			if err != nil || resp == nil {
				log.Ctx(queryCtx).Debug().Err(err).Msg("concurrent dns server failed")
				results <- nil
				return
			}
			results <- resp
		}(s)
	}

	var fallback *dns.Msg
	pending := len(c.DnsServers)
	for pending > 0 {
		select {
		case resp := <-results:
			pending--
			if resp == nil {
				continue
			}
			switch resp.Rcode {
			case dns.RcodeSuccess:
				if len(resp.Answer) > 0 {
					return resp, nil
				}
				if fallback == nil {
					fallback = resp
				}
			case dns.RcodeNameError:
				if fallback == nil {
					fallback = resp
				}
			default:
				log.Ctx(queryCtx).Debug().
					Str("rcode", dns.RcodeToString[int(resp.Rcode)]).
					Msg("dns server returned non-success rcode")
			}
		case <-ctx.Done():
			if fallback != nil {
				return fallback, nil
			}
			return nil, ctx.Err()
		}
	}

	if fallback != nil {
		return fallback, nil
	}
	return nil, ErrAllServersFailed
}
