package common

import (
	"errors"
	"runtime"
	"strconv"
	"time"

	"github.com/5vnetwork/vx-core/common/units"
	"github.com/rs/zerolog/log"
)

var (
	Version = "debug"
)

var MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
var ErrClosed = errors.New("closed")

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

// Error2 returns the err from the 2nd parameter.
func Error2(v interface{}, err error) error {
	return err
}

func Interrupt(obj interface{}) {
	if i, ok := obj.(Interruptible); ok {
		i.Interrupt()
	}
}

func Start(obj interface{}) error {
	if s, ok := obj.(Startable); ok {
		return s.Start()
	}
	return nil
}

func Close(obj interface{}) error {
	if c, ok := obj.(Closable); ok {
		return c.Close()
	}
	return nil
}

// StartAll starts all the Startable objects in the list.
// It returns the first error encountered, if any.
// And it closes all the started objects when an error occurs.
func StartAll(l ...interface{}) error {
	var started []interface{}
	for _, obj := range l {
		log.Info().Msgf("starting %T", obj)
		if err := Start(obj); err != nil {
			closeErr := CloseAll(started...)
			if closeErr != nil {
				return errors.Join(err, closeErr)
			}
			return err
		}
		started = append(started, obj)
		log.Info().Msgf("started %T", obj)
	}
	return nil
}

func CloseAll(l ...interface{}) error {
	var errs []error
	for _, obj := range l {
		log.Info().Msgf("closing %T", obj)
		err := Close(obj)
		errs = append(errs, err)
		log.Info().Msgf("closed %T, err: %v", obj, err)
	}
	return errors.Join(errs...)
}

// ChainedClosable is a Closable that consists of multiple Closable objects.
type ChainedClosable []Closable

// Close implements Closable.
func (cc ChainedClosable) Close() error {
	var errs []error
	for _, c := range cc {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type CtxKey int

const (
	InstanceKey CtxKey = iota
	InboundManagerKey
	OutboundManagerKey
	PolicyKey
	UserManagerKey
	BufferSizePerConnection
	PiperKey
	DnsKey
	StatsManagerKey
	DispatcherKey
	RouterKey
	FakeDnsKey
	StatsKey
	GeoKey
)

var ErrInvalidConfig = errors.New("invalid config")

func Uint16ToString(n uint16) string {
	return strconv.Itoa(int(n))
}

func GetIndex(list []uint32, value uint32) int {
	for i, v := range list {
		if value <= v {
			return i
		}
	}
	return -1
}

var (
	OneKB uint64 = 1024
	TenKB uint64 = 10 * OneKB
	OneMB uint64 = 1024 * OneKB
	TenMB uint64 = 10 * OneMB
)

func Log() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Debug().
		Int("HeapAlloc", int(units.BytesToMB(m.HeapAlloc))).
		Int("Sys", int(units.BytesToMB(m.Sys))).
		Int("StackSys", int(units.BytesToMB(m.StackSys))).
		Int("NumGC", int(m.NumGC)).
		Int("NumGoroutine", runtime.NumGoroutine()).
		Msg("Memory stats")
}

func IsClosedChan(c <-chan struct{}) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}
