// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

func TestSerialDnsServers_EmptyReturnsErr(t *testing.T) {
	s := NewSerialDnsServers("test", 50*time.Millisecond)
	if _, err := s.HandleQuery(context.Background(), newTestQuery(), false); !errors.Is(err, ErrAllServersFailed) {
		t.Fatalf("expected ErrAllServersFailed, got %v", err)
	}
}

func TestSerialDnsServers_FirstSucceedsImmediatelyDoesNotStartSecond(t *testing.T) {
	var secondStarted atomic.Bool

	first := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return aReply(msg, "1.1.1.1"), nil
		},
	}
	second := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			secondStarted.Store(true)
			return aReply(msg, "2.2.2.2"), nil
		},
	}

	s := NewSerialDnsServers("test", 100*time.Millisecond, first, second)
	resp, err := s.HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("expected one answer, got %d", len(resp.Answer))
	}

	time.Sleep(150 * time.Millisecond)
	if secondStarted.Load() {
		t.Fatal("second server should not be started when first succeeded immediately")
	}
}

func TestSerialDnsServers_KicksOffSecondAfterInterval(t *testing.T) {
	blocking := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}
	second := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return aReply(msg, "2.2.2.2"), nil
		},
	}

	interval := 50 * time.Millisecond
	start := time.Now()
	resp, err := NewSerialDnsServers("test", interval, blocking, second).
		HandleQuery(context.Background(), newTestQuery(), false)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("expected one answer, got %d", len(resp.Answer))
	}
	if elapsed < interval {
		t.Fatalf("expected at least %v before second server kicks in, got %v", interval, elapsed)
	}
}

// Slow-earlier-wins works if the earlier server returns its definitive answer
// while we are still waiting on a later server (before the timer fires again).
func TestSerialDnsServers_SlowEarlierBeatsLaterStart(t *testing.T) {
	const interval = 50 * time.Millisecond

	slowCorrect := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			select {
			case <-time.After(80 * time.Millisecond):
				return aReply(msg, "1.1.1.1"), nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}
	blocking := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	resp, err := NewSerialDnsServers("test", interval, slowCorrect, blocking).
		HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("expected one answer from earlier server, got %d", len(resp.Answer))
	}
	if got := resp.Answer[0].(*mdns.A).A.String(); got != "1.1.1.1" {
		t.Fatalf("expected slow earlier server's IP 1.1.1.1, got %s", got)
	}
}

func TestSerialDnsServers_PositivePreferredOverNxdomain(t *testing.T) {
	nxFirst := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nxdomainReply(msg), nil
		},
	}
	answerSecond := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return aReply(msg, "8.8.8.8"), nil
		},
	}

	resp, err := NewSerialDnsServers("test", 50*time.Millisecond, nxFirst, answerSecond).
		HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != mdns.RcodeSuccess || len(resp.Answer) != 1 {
		t.Fatalf("expected positive answer to beat NXDOMAIN, got rcode=%d answers=%d", resp.Rcode, len(resp.Answer))
	}
}

func TestSerialDnsServers_FallbackToNxdomain(t *testing.T) {
	nx := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return nxdomainReply(msg), nil
		},
	}
	servfail := &fakeResolverDnsServer{
		handle: func(ctx context.Context, msg *mdns.Msg, tcp bool) (*mdns.Msg, error) {
			return servfailReply(msg), nil
		},
	}

	resp, err := NewSerialDnsServers("test", 20*time.Millisecond, nx, servfail).
		HandleQuery(context.Background(), newTestQuery(), false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != mdns.RcodeNameError {
		t.Fatalf("expected NXDOMAIN fallback, got rcode=%d", resp.Rcode)
	}
}

func TestSerialDnsServers_AllFailReturnsErr(t *testing.T) {
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

	if _, err := NewSerialDnsServers("test", 20*time.Millisecond, transportErr, servfail).
		HandleQuery(context.Background(), newTestQuery(), false); !errors.Is(err, ErrAllServersFailed) {
		t.Fatalf("expected ErrAllServersFailed, got %v", err)
	}
}

func TestSerialDnsServers_ContextCancellation(t *testing.T) {
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

	if _, err := NewSerialDnsServers("test", 50*time.Millisecond, blocked, blocked).
		HandleQuery(ctx, newTestQuery(), false); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
