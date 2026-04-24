package dns

import (
	"context"
	"errors"
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

	resolver := NewDnsServerToResolver(first, second)

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

	resolver := NewDnsServerToResolver(first, second)

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

	resolver := NewDnsServerToResolver(first, second)

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

	resolver := NewDnsServerToResolver(server)

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

func TestDnsServerToResolverLookupIPv4ResolvesCNAME(t *testing.T) {
	var (
		mu        sync.Mutex
		questions []string
	)

	server := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			mu.Lock()
			questions = append(questions, msg.Question[0].Name)
			mu.Unlock()

			switch msg.Question[0].Name {
			case "example.com.":
				return cnameReply(msg, "alias.example.com"), nil
			case "alias.example.com.":
				return aReply(msg, "2.2.2.2"), nil
			default:
				return nil, errors.New("unexpected question")
			}
		},
	}

	resolver := NewDnsServerToResolver(server)

	ips, err := resolver.LookupIPv4(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPv4 returned error: %v", err)
	}
	if len(ips) != 1 || ips[0].String() != "2.2.2.2" {
		t.Fatalf("expected [2.2.2.2], got %v", ips)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(questions) != 2 || questions[0] != "example.com." || questions[1] != "alias.example.com." {
		t.Fatalf("expected CNAME follow-up queries, got %v", questions)
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

	resolver := NewDnsServerToResolver(server)

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
