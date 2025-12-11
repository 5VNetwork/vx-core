package tester

import (
	context "context"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type ResultReporter interface {
	UsableResult(tag string, ok bool)
	SpeedResult(tag string, speed int64)
	IPv6Result(tag string, ok bool)
	PingResult(tag string, ping int)
}

type Tester struct {
	SpeedTestFunc  func(ctx context.Context, h i.Outbound) (int64, error)
	UsableTestFunc func(ctx context.Context, h i.Outbound) (bool, error)
	PingTestFunc   func(ctx context.Context, h i.Outbound) (int, error)
	ResultReporter
	// track ongoing tests
	ongoingUsableTests sync.Map // key is handler, value is *testUsableResult
}

// testUsableResult holds the result of a test and a channel to signal completion
type testUsableResult struct {
	result bool
	done   chan struct{}
}

func (t *Tester) TestPing(ctx context.Context, h i.Outbound) int {
	ping, err := t.PingTestFunc(ctx, h)
	if err != nil {
		log.Debug().Err(err).Str("handler", h.Tag()).Msg("ping test func error")
		ping = -1
	}
	t.PingResult(h.Tag(), ping)
	return ping
}

func (t *Tester) TestIPv6(ctx context.Context, h i.Outbound) bool {
	yes, err := util.TestIpv6(ctx, h, util.GoogleDNS6)
	if err != nil {
		log.Debug().Err(err).Str("handler", h.Tag()).Msg("ipv6 test func error")
	}
	t.IPv6Result(h.Tag(), yes)
	log.Debug().Bool("yes", yes).Str("handler", h.Tag()).Msg("ipv6 test done")
	return yes
}

// return -1 if test failed
func (t *Tester) TestSpeed(ctx context.Context, h i.Outbound, rtry bool) int64 {
	var speed int64
	var err error

	times := 1
	if rtry {
		times = 5
	}
	err = retry.Timed(times, 1000).On(func() error {
		speed, err = t.SpeedTestFunc(ctx, h)
		if err != nil {
			log.Debug().Err(err).Str("handler", h.Tag()).Int64("speed", speed).Msg("speed test func error")
		}
		return err
	})
	if err != nil {
		log.Debug().Str("handler", h.Tag()).Msgf("speedtest err: %v", err)
		speed = -1
	}

	t.SpeedResult(h.Tag(), speed)
	return speed
}

func (t *Tester) TestUsable(ctx context.Context, h i.Outbound, retry bool) bool {
	// Check if there's an ongoing test for this handler
	if existing, ok := t.ongoingUsableTests.Load(h); ok {
		log.Debug().Str("handler", h.Tag()).Msg("concurrent handler usable test")
		// Wait for the existing test to complete
		result := existing.(*testUsableResult)
		<-result.done
		return result.result
	}

	// Create a new test result
	result := &testUsableResult{
		done: make(chan struct{}),
	}
	t.ongoingUsableTests.Store(h, result)

	// Run the test
	var ok bool
	var err error
	times := 1
	if retry {
		times = 3
	}
	for i := 0; i < times; i++ {
		ok, err = t.UsableTestFunc(ctx, h)
		if err != nil {
			log.Debug().Str("handler", h.Tag()).Msgf("usable test err: %v", err)
		}
		if ok {
			break
		}
		time.Sleep(time.Millisecond * 1000)
	}

	t.UsableResult(h.Tag(), ok)

	// Update result and signal completion
	result.result = ok
	close(result.done)
	t.ongoingUsableTests.Delete(h)

	return ok
}
