// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package router_test

import (
	"context"
	"reflect"
	"testing"

	. "github.com/5vnetwork/vx-core/app/router"
	cnet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

type ipSetStub struct {
	match func(ip cnet.IP) bool
}

func (s *ipSetStub) Match(ip cnet.IP) bool {
	if s.match != nil {
		return s.match(ip)
	}
	return false
}

type resolverStub struct {
	lookupIP func(ctx context.Context, domain string) ([]cnet.IP, error)
}

func (r *resolverStub) LookupIP(ctx context.Context, domain string) ([]cnet.IP, error) {
	if r.lookupIP != nil {
		return r.lookupIP(ctx, domain)
	}
	return nil, nil
}

func (r *resolverStub) LookupIPv4(ctx context.Context, domain string) ([]cnet.IP, error) {
	return nil, nil
}

func (r *resolverStub) LookupIPv6(ctx context.Context, domain string) ([]cnet.IP, error) {
	return nil, nil
}

func (r *resolverStub) LookupIPSpeed(ctx context.Context, domain string) ([]cnet.IP, error) {
	return nil, nil
}

var _ i.IPSet = (*ipSetStub)(nil)
var _ i.IPResolver = (*resolverStub)(nil)

func TestIpMatcher_MatchSourceIp_sourceMatches(t *testing.T) {
	src := cnet.ParseIP("10.0.0.5")
	m := &IpMatcher{
		MatchSourceIp: true,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return ip != nil && ip.Equal(src)
		}},
	}
	info := &session.Info{
		Source: cnet.TCPDestination(cnet.IPAddress(src), 12345),
		Target: cnet.TCPDestination(cnet.DomainAddress("example.com"), 443),
	}
	rw := "handler"
	gotRW, ok := m.Apply(context.Background(), info, rw)
	if !ok || gotRW != rw {
		t.Fatalf("expected match with rw %q, ok=%v gotRW=%v", rw, ok, gotRW)
	}
}

func TestIpMatcher_MatchSourceIp_sourceDoesNotMatch(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp: true,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return ip != nil && ip.Equal(cnet.ParseIP("10.0.0.99"))
		}},
	}
	info := &session.Info{
		Source: cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("10.0.0.5")), 12345),
		Target: cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("8.8.8.8")), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("expected no match when source IP not in set")
	}
}

func TestIpMatcher_MatchSourceIp_domainSourceNoIP(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp: true,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return len(ip) > 0
		}},
	}
	info := &session.Info{
		Source: cnet.TCPDestination(cnet.DomainAddress("client.internal"), 12345),
		Target: cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("8.8.8.8")), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("expected no match when source is domain (GetSourceIPs is nil)")
	}
}

func TestIpMatcher_MatchSourceIp_ignoresMatchingTarget(t *testing.T) {
	eight := cnet.ParseIP("8.8.8.8")
	m := &IpMatcher{
		MatchSourceIp: true,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return ip != nil && ip.Equal(eight)
		}},
	}
	info := &session.Info{
		Source: cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("192.168.1.1")), 12345),
		Target: cnet.TCPDestination(cnet.IPAddress(eight), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("must use source IP only when MatchSourceIp is true")
	}
}

func TestIpMatcher_targetIPMatch(t *testing.T) {
	dst := cnet.ParseIP("203.0.113.10")
	m := &IpMatcher{
		MatchSourceIp: false,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return ip != nil && ip.Equal(dst)
		}},
	}
	info := &session.Info{
		Source: cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("10.0.0.1")), 12345),
		Target: cnet.TCPDestination(cnet.IPAddress(dst), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if !ok {
		t.Fatal("expected target IP to match")
	}
}

func TestIpMatcher_targetIPNoMatch(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp: false,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return false
		}},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("198.51.100.1")), 80),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("expected no match")
	}
}

func TestIpMatcher_domainWithoutResolve(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp: false,
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return true
		}},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.DomainAddress("example.com"), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("without ResolveHard/ResolveSoft, domain target should not match")
	}
}

func TestIpMatcher_resolveSoft_firstMatchingIP_updatesTarget(t *testing.T) {
	first := cnet.ParseIP("192.0.2.1")
	second := cnet.ParseIP("192.0.2.2")
	m := &IpMatcher{
		MatchSourceIp:         false,
		ResolveSoftAndRewrite: true,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				if domain != "example.com" {
					t.Fatalf("unexpected domain %q", domain)
				}
				return []cnet.IP{first, second}, nil
			},
		},
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return ip != nil && ip.Equal(second)
		}},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.DomainAddress("example.com"), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if !ok {
		t.Fatal("expected ResolveSoft to match on second resolved IP")
	}
	if info.Target.Address == nil || !info.Target.Address.Family().IsIP() {
		t.Fatalf("expected Target rewritten to IP, got %#v", info.Target.Address)
	}
	if !info.Target.Address.IP().Equal(second) {
		t.Fatalf("expected target IP %v, got %v", second, info.Target.Address.IP())
	}
}

func TestIpMatcher_resolveSoft_noMatchingIP(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp:         false,
		ResolveSoftAndRewrite: true,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				return []cnet.IP{cnet.ParseIP("192.0.2.1")}, nil
			},
		},
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool { return false }},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.DomainAddress("x.test"), 80),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("expected no match when no resolved IP is in set")
	}
}

func TestIpMatcher_resolveHard_rejectsOnFirstNonMatch(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp: false,
		ResolveHard:   true,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				return []cnet.IP{cnet.ParseIP("192.0.2.1"), cnet.ParseIP("192.0.2.2")}, nil
			},
		},
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool {
			return ip != nil && ip.Equal(cnet.ParseIP("192.0.2.2"))
		}},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.DomainAddress("x.test"), 80),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("ResolveHard should fail when an earlier resolved IP is not in set")
	}
}

func TestIpMatcher_resolveEmpty(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp:         false,
		ResolveSoftAndRewrite: true,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				return nil, nil
			},
		},
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool { return true }},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.DomainAddress("nxdomain.test"), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("expected no match when resolver returns no IPs")
	}
}

func TestIpMatcher_targetIP_skipsResolver_evenWithSniffedDomain(t *testing.T) {
	lookupCalled := false
	m := &IpMatcher{
		MatchSourceIp:         false,
		ResolveSoftAndRewrite: true,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				lookupCalled = true
				return []cnet.IP{cnet.ParseIP("198.51.100.50")}, nil
			},
		},
		IpSet: &ipSetStub{match: func(got cnet.IP) bool {
			return got != nil && got.Equal(cnet.ParseIP("127.0.0.1"))
		}},
	}
	info := &session.Info{
		Target:        cnet.TCPDestination(cnet.IPAddress(cnet.ParseIP("127.0.0.1")), 443),
		SniffedDomain: "sniffed.example",
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if lookupCalled {
		t.Fatal("resolver must not run when GetTargetIP() is already set")
	}
	if !ok {
		t.Fatal("expected match on target IP; SniffedDomain does not trigger resolution in this path")
	}
}

func TestIpMatcher_resolveAllMatchWithoutSoft(t *testing.T) {
	m := &IpMatcher{
		MatchSourceIp:         false,
		ResolveHard:           true,
		ResolveSoftAndRewrite: false,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				return []cnet.IP{cnet.ParseIP("192.0.2.1"), cnet.ParseIP("192.0.2.2")}, nil
			},
		},
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool { return true }},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(cnet.DomainAddress("allmatch.test"), 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if !ok {
		t.Fatal("when every resolved IP matches and ResolveSoft is false, expect true at end of loop")
	}
}

func TestIpMatcher_targetAddressUnchangedWhenResolveSoftNoHit(t *testing.T) {
	orig := cnet.DomainAddress("keep.example")
	m := &IpMatcher{
		MatchSourceIp:         false,
		ResolveSoftAndRewrite: true,
		IpResolver: &resolverStub{
			lookupIP: func(ctx context.Context, domain string) ([]cnet.IP, error) {
				return []cnet.IP{cnet.ParseIP("192.0.2.1")}, nil
			},
		},
		IpSet: &ipSetStub{match: func(ip cnet.IP) bool { return false }},
	}
	info := &session.Info{
		Target: cnet.TCPDestination(orig, 443),
	}
	_, ok := m.Apply(context.Background(), info, "rw")
	if ok {
		t.Fatal("no match")
	}
	if !reflect.DeepEqual(info.Target.Address, orig) {
		t.Fatal("Target.Address should stay domain when no resolved IP matched")
	}
}
