// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package memmon

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Monitor struct {
	monitorConfig *MonitorConfig
	Dispatcher    *dispatcher.Dispatcher
	done          *done.Instance
}

type MonitorConfig struct {
	Interval      time.Duration
	Path          string
	ListenAddress string
}

func NewMonitor(config *MonitorConfig) *Monitor {
	return &Monitor{
		done:          done.New(),
		monitorConfig: config,
	}
}
func (m *Monitor) Start() error {
	go m.log()
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		go func() {
			log.Debug().Str("listen_address", m.monitorConfig.ListenAddress).Msg("starting pprof server")
			http.ListenAndServe(m.monitorConfig.ListenAddress, nil)
		}()
	}
	return nil
}

func (m *Monitor) Close() error {
	m.done.Close()
	return nil
}

func (mon *Monitor) TakeHeapSnapshot() error {
	_, err := TakeHeapSnapshot(mon.monitorConfig.Path)
	return err
}

const (
	memoryThreshold = 25 * 1024 * 1024
	// memoryRearmThreshold provides hysteresis: after a snapshot we only re-arm
	// once memory falls back below this lower watermark, so the forced GC
	// dropping just under memoryThreshold doesn't cause repeated snapshots when
	// usage oscillates around the boundary.
	memoryRearmThreshold = 22 * 1024 * 1024
)

func (mon *Monitor) log() {
	log.Debug().Msg("start monitor memory")
	var m runtime.MemStats
	// overThreshold tracks whether we were already above the threshold on the
	// previous tick, so the heap snapshot is taken once per crossing instead of
	// on every interval while we stay above it.
	var overThreshold bool

	for {
		select {
		case <-mon.done.Wait():
			return
		case <-time.After(mon.monitorConfig.Interval):
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

			if (m.Alloc+m.StackInuse > memoryThreshold) && runtime.GOOS == "ios" {
				log.Debug().Msg("Memory threshold exceeded, forcing GC")
				runtime.GC()
				// Only snapshot on the transition into the over-threshold
				// state; taking it every tick would itself churn memory.
				if !overThreshold && zerolog.GlobalLevel() == zerolog.DebugLevel {
					TakeHeapSnapshot(mon.monitorConfig.Path)
				}
				overThreshold = true
			} else if m.Alloc+m.StackInuse < memoryRearmThreshold {
				overThreshold = false
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
}

// TakeHeapSnapshot saves current heap profile to file
func TakeHeapSnapshot(dir string) (string, error) {
	filename := filepath.Join(dir, "heap-"+time.Now().Format("20060102-150405")+".prof")
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return "", err
	}

	log.Info().Str("file", filename).Msg("Heap snapshot saved")
	return filename, nil
}

// TakeGoroutineSnapshot saves current goroutine stacks. Use alongside heap
// snapshots to find what still references live objects after Close.
func TakeGoroutineSnapshot(dir string) (string, error) {
	filename := filepath.Join(dir, "goroutine-"+time.Now().Format("20060102-150405")+".prof")
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return "", os.ErrNotExist
	}
	if err := profile.WriteTo(f, 0); err != nil {
		return "", err
	}

	log.Info().Str("file", filename).Msg("Goroutine snapshot saved")
	return filename, nil
}
