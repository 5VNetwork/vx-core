package dns

import (
	"context"
	"errors"
	"log"
	stdnet "net"
	"sync"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

type fakeResolverDnsServer struct {
	handle func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error)
}

func (f *fakeResolverDnsServer) Start() error { return nil }

func (f *fakeResolverDnsServer) Close() error { return nil }

func (f *fakeResolverDnsServer) HandleQuery(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
	return f.handle(ctx, msg, tcp)
}

func aReply(msg *mdns.Msg, ip string) *mdns.Msg {
	resp := new(mdns.Msg)
	resp.SetReply(msg)
	resp.Answer = []mdns.RR{
		&mdns.A{
			Hdr: mdns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: mdns.TypeA,
				Class:  mdns.ClassINET,
				Ttl:    60,
			},
			A: stdnet.ParseIP(ip).To4(),
		},
	}
	return resp
}

func truncatedReply(msg *mdns.Msg) *mdns.Msg {
	resp := new(mdns.Msg)
	resp.SetReply(msg)
	resp.Truncated = true
	return resp
}

func cnameReply(msg *mdns.Msg, target string) *mdns.Msg {
	resp := new(mdns.Msg)
	resp.SetReply(msg)
	resp.Answer = []mdns.RR{
		&mdns.CNAME{
			Hdr: mdns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: mdns.TypeCNAME,
				Class:  mdns.ClassINET,
				Ttl:    60,
			},
			Target: mdns.Fqdn(target),
		},
	}
	return resp
}

func TestDnsServerToResolverLookupIPv4ReturnsLateEarlierResultAndCancelsLaterRequests(t *testing.T) {
	secondStarted := make(chan struct{})
	secondCanceled := make(chan struct{})

	first := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			timer := time.NewTimer(4200 * time.Millisecond)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timer.C:
				return aReply(msg, "1.1.1.1"), nil
			}
		},
	}
	second := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			close(secondStarted)
			<-ctx.Done()
			close(secondCanceled)
			return nil, ctx.Err()
		},
	}

	resolver := NewDnsServerToResolver(DnsServerToResolverOption{DnsServers: []DnsServer{first, second}})

	start := time.Now()
	ips, err := resolver.LookupIPv4(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPv4 returned error: %v", err)
	}
	if len(ips) != 1 || ips[0].String() != "1.1.1.1" {
		t.Fatalf("expected [1.1.1.1], got %v", ips)
	}

	elapsed := time.Since(start)
	if elapsed < 4*time.Second {
		t.Fatalf("expected late response after second server launch, got %v", elapsed)
	}

	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("second server was not started")
	}

	select {
	case <-secondCanceled:
	case <-time.After(time.Second):
		t.Fatal("second server was not canceled after first result won")
	}
}

func TestDnsServerToResolverLookupIPv4FallsBackToNextServer(t *testing.T) {
	first := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nil, errors.New("first failed")
		},
	}
	second := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return aReply(msg, "8.8.8.8"), nil
		},
	}

	resolver := NewDnsServerToResolver(
		DnsServerToResolverOption{DnsServers: []DnsServer{first, second}})

	ips, err := resolver.LookupIPv4(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPv4 returned error: %v", err)
	}
	if len(ips) != 1 || ips[0].String() != "8.8.8.8" {
		t.Fatalf("expected [8.8.8.8], got %v", ips)
	}
}

func TestDnsServerToResolverLookupIPv4ReturnsErrAllServersFailed(t *testing.T) {
	first := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nil, errors.New("first failed")
		},
	}
	second := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nil, nil
		},
	}

	resolver := NewDnsServerToResolver(DnsServerToResolverOption{DnsServers: []DnsServer{first, second}})

	_, err := resolver.LookupIPv4(context.Background(), "example.com")
	if !errors.Is(err, ErrAllServersFailed) {
		t.Fatalf("expected %v, got %v", ErrAllServersFailed, err)
	}
}

func TestDnsServerToResolverLookupIPv4RetriesTruncatedResponseWithTCP(t *testing.T) {
	var (
		mu    sync.Mutex
		calls []bool
	)

	server := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			mu.Lock()
			calls = append(calls, tcp)
			mu.Unlock()

			if !tcp {
				return truncatedReply(msg), nil
			}
			return aReply(msg, "9.9.9.9"), nil
		},
	}

	resolver := NewDnsServerToResolver(DnsServerToResolverOption{DnsServers: []DnsServer{server}})

	ips, err := resolver.LookupIPv4(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPv4 returned error: %v", err)
	}
	if len(ips) != 1 || ips[0].String() != "9.9.9.9" {
		t.Fatalf("expected [9.9.9.9], got %v", ips)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 2 || calls[0] || !calls[1] {
		t.Fatalf("expected UDP then TCP calls, got %v", calls)
	}
}

func TestDnsServerToResolverLookupIPv4ReturnsTruncatedUDPAnswerWithoutTCPRetry(t *testing.T) {
	var (
		mu    sync.Mutex
		calls []bool
	)

	server := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			mu.Lock()
			calls = append(calls, tcp)
			mu.Unlock()

			resp := aReply(msg, "3.3.3.3")
			resp.Truncated = true
			return resp, nil
		},
	}

	resolver := NewDnsServerToResolver(DnsServerToResolverOption{DnsServers: []DnsServer{server}})

	ips, err := resolver.LookupIPv4(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPv4 returned error: %v", err)
	}
	if len(ips) != 1 || ips[0].String() != "3.3.3.3" {
		t.Fatalf("expected [3.3.3.3], got %v", ips)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 1 || calls[0] {
		t.Fatalf("expected only one UDP call, got %v", calls)
	}
}

func TestFindCNAMEChainTail(t *testing.T) {
	t.Run("follows chain to tail", func(t *testing.T) {
		answers := []mdns.RR{
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "a.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "b.example.com.",
			},
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "b.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "c.example.com.",
			},
		}

		tail := findCNAMEChainTail(answers, "a.example.com.", 8)
		if tail != "c.example.com." {
			t.Fatalf("expected tail c.example.com., got %q", tail)
		}
	})

	t.Run("returns start when no cname match", func(t *testing.T) {
		answers := []mdns.RR{
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "x.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "y.example.com.",
			},
		}

		tail := findCNAMEChainTail(answers, "a.example.com.", 8)
		if tail != "a.example.com." {
			t.Fatalf("expected unchanged qname, got %q", tail)
		}
	})

	t.Run("stops on self loop", func(t *testing.T) {
		answers := []mdns.RR{
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "loop.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "loop.example.com.",
			},
		}

		tail := findCNAMEChainTail(answers, "loop.example.com.", 8)
		if tail != "loop.example.com." {
			t.Fatalf("expected self-loop name, got %q", tail)
		}
	})

	t.Run("returns last hop when max hops reached", func(t *testing.T) {
		answers := []mdns.RR{
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "a.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "b.example.com.",
			},
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "b.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "c.example.com.",
			},
			&mdns.CNAME{
				Hdr:    mdns.RR_Header{Name: "c.example.com.", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: 60},
				Target: "d.example.com.",
			},
		}

		tail := findCNAMEChainTail(answers, "a.example.com.", 2)
		if tail != "c.example.com." {
			t.Fatalf("expected hop-limited tail c.example.com., got %q", tail)
		}
	})
}

func TestDnsServerToResolverLookupEch(t *testing.T) {
	t.Skip()
	resolver := DefaultCfResolver()
	ech, err := resolver.LookupECH(context.Background(), "")
	if err != nil {
		t.Fatalf("LookupECH returned error: %v", err)
	}
	if len(ech) == 0 {
		t.Fatal("expected ech, got empty")
	}
	log.Println(ech)
}
