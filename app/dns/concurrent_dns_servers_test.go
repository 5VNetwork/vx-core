// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"errors"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

// Shared helpers for concurrent and serial dns server tests.

func newTestQuery() *mdns.Msg {
	msg := new(mdns.Msg)
	msg.SetQuestion(mdns.Fqdn("example.com"), mdns.TypeA)
	return msg
}

func nxdomainReply(msg *mdns.Msg) *mdns.Msg {
	resp := new(mdns.Msg)
	resp.SetReply(msg)
	resp.Rcode = mdns.RcodeNameError
	return resp
}

func nodataReply(msg *mdns.Msg) *mdns.Msg {
	resp := new(mdns.Msg)
	resp.SetReply(msg)
	resp.Rcode = mdns.RcodeSuccess
	return resp
}

func servfailReply(msg *mdns.Msg) *mdns.Msg {
	resp := new(mdns.Msg)
	resp.SetReply(msg)
	resp.Rcode = mdns.RcodeServerFailure
	return resp
}

func TestConcurrentDnsServers_EmptyReturnsErr(t *testing.T) {
	s := NewConcurrentDnsServers("test", nil)
	if _, err := s.HandleQuery(context.Background(), newTestQuery(), false); !errors.Is(err, ErrAllServersFailed) {
		t.Fatalf("expected ErrAllServersFailed, got %v", err)
	}
}

func TestConcurrentDnsServers_SingleServerDelegates(t *testing.T) {
	called := false
	srv := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			called = true
			return aReply(msg, "1.1.1.1"), nil
		},
	}
	resp, err := NewConcurrentDnsServers("test", srv).HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !called || len(resp.Answer) != 1 {
		t.Fatalf("expected delegation with one answer, got called=%v resp=%v", called, resp)
	}
}

func TestConcurrentDnsServers_FirstAnswerWinsAndCancelsOthers(t *testing.T) {
	fast := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return aReply(msg, "1.2.3.4"), nil
		},
	}
	slowCanceled := make(chan struct{})
	slow := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			<-ctx.Done()
			close(slowCanceled)
			return nil, ctx.Err()
		},
	}

	resp, err := NewConcurrentDnsServers("test", fast, slow).HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("expected one answer, got %d", len(resp.Answer))
	}

	select {
	case <-slowCanceled:
	case <-time.After(time.Second):
		t.Fatal("slow server was not canceled after fast one won")
	}
}

func TestConcurrentDnsServers_PositivePreferredOverNxdomain(t *testing.T) {
	fastNX := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nxdomainReply(msg), nil
		},
	}
	slowAnswer := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			time.Sleep(50 * time.Millisecond)
			return aReply(msg, "1.2.3.4"), nil
		},
	}

	resp, err := NewConcurrentDnsServers("test", fastNX, slowAnswer).HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != mdns.RcodeSuccess || len(resp.Answer) != 1 {
		t.Fatalf("expected positive answer to beat NXDOMAIN, got rcode=%d answers=%d", resp.Rcode, len(resp.Answer))
	}
}

func TestConcurrentDnsServers_FallbackToNxdomain(t *testing.T) {
	nx := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nxdomainReply(msg), nil
		},
	}
	failing := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nil, errors.New("network error")
		},
	}

	resp, err := NewConcurrentDnsServers("test", nx, failing).HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != mdns.RcodeNameError {
		t.Fatalf("expected NXDOMAIN fallback, got rcode=%d", resp.Rcode)
	}
}

func TestConcurrentDnsServers_FallbackToNodata(t *testing.T) {
	nodata := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nodataReply(msg), nil
		},
	}
	servfail := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return servfailReply(msg), nil
		},
	}

	resp, err := NewConcurrentDnsServers("test", nodata, servfail).HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != mdns.RcodeSuccess || len(resp.Answer) != 0 {
		t.Fatalf("expected NODATA fallback, got rcode=%d answers=%d", resp.Rcode, len(resp.Answer))
	}
}

func TestConcurrentDnsServers_AllFailReturnsErr(t *testing.T) {
	transportErr := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nil, errors.New("transport failed")
		},
	}
	servfail := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return servfailReply(msg), nil
		},
	}

	if _, err := NewConcurrentDnsServers("test", transportErr, servfail).HandleQuery(context.Background(), newTestQuery(), false); !errors.Is(err, ErrAllServersFailed) {
		t.Fatalf("expected ErrAllServersFailed, got %v", err)
	}
}

func TestConcurrentDnsServers_ContextCancellation(t *testing.T) {
	blocked := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	if _, err := NewConcurrentDnsServers("test", blocked, blocked).HandleQuery(ctx, newTestQuery(), false); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
