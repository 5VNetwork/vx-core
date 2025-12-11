package memmon

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Monitor struct {
	Dispatcher *dispatcher.Dispatcher
	done       *done.Instance
	Interval   time.Duration
}

func NewMonitor(interval time.Duration) *Monitor {

	return &Monitor{
		done:     done.New(),
		Interval: interval,
	}
}

func (m *Monitor) Start() error {
	go m.log()
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		go func() {
			log.Debug().Msg("starting pprof server on port 6060")
			http.ListenAndServe("0.0.0.0:6060", nil)
		}()
	}
	return nil
}

func (m *Monitor) Close() error {
	m.done.Close()
	return nil
}

func (mon *Monitor) log() {
	log.Debug().Msg("start monitor memory")
	var m runtime.MemStats

	for {
		select {
		case <-mon.done.Wait():
			return
		case <-time.After(mon.Interval):
			runtime.ReadMemStats(&m)
			log.Debug().
				Int("HeapAlloc", int(units.BytesToMB(m.HeapAlloc))).
				Int("HeapInuse", int(units.BytesToMB(m.HeapInuse))).
				Int("HeapIdle", int(units.BytesToMB(m.HeapIdle))).
				Int("HeapReleased", int(units.BytesToMB(m.HeapReleased))).
				Int("HeapObjects", int(m.HeapObjects)).
				Int("Sys", int(units.BytesToMB(m.Sys))).
				Int("StackSys", int(units.BytesToMB(m.StackSys))).
				Int("NumGC", int(m.NumGC)).
				Uint64("TotalAlloc", m.TotalAlloc/1024/1024).
				Uint32("live objects", uint32(m.Mallocs-m.Frees)).
				Int("NumGoroutine", runtime.NumGoroutine()).
				Int32("Flow", mon.Dispatcher.Flows.Load()).
				Int32("Conn", mon.Dispatcher.PacketConns.Load()).
				Msg("Memory stats")

			if (m.Alloc+m.StackInuse > 25*1024*1024) && runtime.GOOS == "ios" {
				log.Debug().Msg("Memory threshold exceeded, forcing GC")
				runtime.GC()
			}
		}
	}
}

func Log() {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)
	log.Debug().
		Int("HeapAlloc", int(units.BytesToMB(m.HeapAlloc))).
		Int("HeapInuse", int(units.BytesToMB(m.HeapInuse))).
		Int("HeapIdle", int(units.BytesToMB(m.HeapIdle))).
		Int("HeapReleased", int(units.BytesToMB(m.HeapReleased))).
		Int("HeapObjects", int(m.HeapObjects)).
		Int("Sys", int(units.BytesToMB(m.Sys))).
		Int("StackSys", int(units.BytesToMB(m.StackSys))).
		Int("NumGC", int(m.NumGC)).
		Uint64("TotalAlloc", m.TotalAlloc/1024/1024).
		Uint32("live objects", uint32(m.Mallocs-m.Frees)).
		Int("NumGoroutine", runtime.NumGoroutine()).
		Msg("Memory stats")

	if (m.Alloc+m.StackInuse > 25*1024*1024) && runtime.GOOS == "ios" {
		log.Debug().Msg("Memory threshold exceeded, forcing GC")
		runtime.GC()
	}
}
