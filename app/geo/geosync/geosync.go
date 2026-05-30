// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

// Package geosync downloads geoip / geosite data files from HTTPS URLs on a schedule
// and reapplies GeoConfig via GeoWrapper.UpdateGeo after each successful download.
package geosync

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/util/downloader"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// GeoSync runs periodic HTTPS downloads for geo data files (cron from GeoRemoteFile).
type GeoSync struct {
	geo *geo.Geo
	dl  *downloader.Downloader

	mu         sync.Mutex
	geoConfig  atomic.Pointer[configs.GeoConfig]
	jobs       []Job
	cronRunner *cron.Cron
	started    bool
	closed     bool
}

// New returns a GeoSync that uses the given GeoWrapper for periodic refreshes.
func New(gw *geo.Geo) *GeoSync {
	return &GeoSync{geo: gw}
}

// PrefetchGeoDataFiles downloads every configured HTTPS URL into filepath before the first UpdateGeo
func PrefetchGeoDataFiles(ctx context.Context, cfg *configs.GeoConfig) error {
	for _, j := range CollectJobs(cfg) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		body, err := FetchViaDirectHTTP(ctx, j.URL)
		if err != nil {
			return err
		}
		if err := writeFileAtomic(j.Filepath, body); err != nil {
			return err
		}
	}
	return nil
}

// Reconfigure stores a clone of cfg for reloads after downloads, replaces download jobs,
// and restarts periodic tasks if GeoSync has already been started.
func (s *GeoSync) Reconfigure(cfg *configs.GeoConfig) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg != nil {
		s.geoConfig.Store(proto.Clone(cfg).(*configs.GeoConfig))
	} else {
		s.geoConfig.Store(nil)
	}
	s.stopTasksLocked()
	s.jobs = CollectJobs(cfg)
	if s.started && !s.closed {
		s.startTasksLocked()
	}
}

// Start begins periodic refresh tasks (no-op if already started or closed).
func (s *GeoSync) Start() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.started {
		return nil
	}
	s.started = true
	s.startTasksLocked()
	return nil
}

// Close stops periodic tasks.
func (s *GeoSync) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.started = false
	s.stopTasksLocked()
	return nil
}

func (s *GeoSync) stopTasksLocked() {
	if s.cronRunner != nil {
		ctx := s.cronRunner.Stop()
		<-ctx.Done()
		s.cronRunner = nil
	}
}

func (s *GeoSync) startTasksLocked() {
	if s.dl == nil {
		log.Warn().Msg("geosync: no downloader; periodic geo URL refresh disabled")
		return
	}
	c := cron.New()
	has := false
	for _, j := range s.jobs {
		expr := strings.TrimSpace(j.CronExpr)
		if expr == "" {
			continue
		}
		has = true
		jj := j
		if _, err := c.AddFunc(expr, func() { s.runJob(jj) }); err != nil {
			log.Err(err).Str("cron", expr).Str("filepath", j.Filepath).Msg("geosync: invalid cron")
		}
	}
	if !has {
		return
	}
	c.Start()
	s.cronRunner = c
}

func (s *GeoSync) runJob(j Job) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	body, err := FetchViaDirectHTTP(ctx, j.URL)
	if err != nil {
		log.Err(err).Str("url", j.URL).Str("filepath", j.Filepath).Msg("geosync: download failed")
		return
	}
	if err := writeFileAtomic(j.Filepath, body); err != nil {
		log.Err(err).Str("filepath", j.Filepath).Msg("geosync: atomic write failed")
		return
	}
	cfg := s.geoConfig.Load()
	if cfg == nil {
		log.Warn().Msg("geosync: no geo config; skipping reload after download")
		return
	}
	if j.DomainAtomicName == "" && j.IPAtomicName == "" {
		log.Warn().Str("filepath", j.Filepath).Msg("geosync: no atomic geo sets use this filepath; skipping reload")
		return
	}
	var domainNames, ipNames []string
	if j.DomainAtomicName != "" {
		domainNames = []string{j.DomainAtomicName}
	}
	if j.IPAtomicName != "" {
		ipNames = []string{j.IPAtomicName}
	}
	if err := s.geo.ReloadAtomicSetsAfterRemoteGeoFile(cfg, domainNames, ipNames); err != nil {
		log.Err(err).Msg("geosync: partial geo reload after download failed")
	}
}
