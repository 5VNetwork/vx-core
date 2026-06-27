// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// SerialDnsServers queries each [DnsServer] in order, kicking off the next one
// after Interval if no definitive answer has arrived. Earlier queries keep running
// in the background and can still win if it returns a result before the current
// DnsServer returns a result, so a slow-but-correct upstream is not wasted.
//
// Definitive answer = NOERROR with at least one answer record (returned immediately).
// NXDOMAIN and NODATA (NOERROR with empty answer) are held as fallbacks and only
// returned if no server produces a definitive answer.
// SERVFAIL/REFUSED/transport errors are ignored.
type SerialDnsServers struct {
	name       string
	DnsServers []DnsServer
	Interval   time.Duration
}

func NewSerialDnsServers(name string, interval time.Duration, servers ...DnsServer) *SerialDnsServers {
	if interval == 0 {
		interval = time.Second * 4
	}
	return &SerialDnsServers{name: name, DnsServers: servers, Interval: interval}
}

func (s *SerialDnsServers) Name() string {
	return s.name
}

func (s *SerialDnsServers) Start() error {
	return common.StartAll(s.DnsServers)
}

func (s *SerialDnsServers) Close() error {
	return common.CloseAll(s.DnsServers)
}

func (s *SerialDnsServers) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	if len(s.DnsServers) == 0 {
		return nil, ErrAllServersFailed
	}
	if len(s.DnsServers) == 1 {
		return s.DnsServers[0].HandleQuery(ctx, msg, tcp)
	}

	queryCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan resolveQueryResult, len(s.DnsServers))
	timer := time.NewTimer(s.Interval)
	defer timer.Stop()

	var fallback *dns.Msg
nextServer:
	for i, dnsServer := range s.DnsServers {
		newMsg := msg
		if i != len(s.DnsServers)-1 {
			newMsg = msg.Copy()
		}
		go runDnsQuery(queryCtx, i, dnsServer, newMsg, tcp, resultChan)
		for {
			select {
			case <-timer.C:
				timer.Reset(s.Interval)
				continue nextServer
			case result := <-resultChan:
				if resp := result.resp; resp != nil {
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
				}
				if result.index == i {
					continue nextServer
				}
				// previous server failed, continue to wait for result of this server
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if fallback != nil {
		return fallback, nil
	}
	return nil, ErrAllServersFailed
}

type resolveQueryResult struct {
	index int
	resp  *dns.Msg
}

// runDnsQuery is a package-level function so `go runDnsQuery(...)` does not
// allocate a closure to capture variables. Pushes nil on error so the caller
// can count completions without separate signalling.
func runDnsQuery(ctx context.Context, index int, server DnsServer,
	msg *dns.Msg, tcp bool, results chan<- resolveQueryResult) {
	resp, err := server.HandleQuery(ctx, msg, tcp)
	if err != nil || resp == nil {
		log.Ctx(ctx).Debug().Err(err).Msg("dns server query failed")
		results <- resolveQueryResult{index: index, resp: nil}
		return
	}
	results <- resolveQueryResult{index: index, resp: resp}
}
