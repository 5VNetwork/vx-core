// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package geosync

import (
	"sort"
	"strings"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
)

// Job describes one on-disk geo data file synced from an HTTPS URL.
type Job struct {
	URL      string
	Filepath string
	// Standard 5-field cron; empty means prefetch only at startup / reconfigure.
	CronExpr string
	// Atomic domain set to reload after download (partial reload).
	DomainAtomicName string
	// Atomic IP set to reload after download (partial reload).
	IPAtomicName string
}

// CollectJobs walks GeoConfig and returns one job per remote geosite/geoip entry.
func CollectJobs(cfg *configs.GeoConfig) []Job {
	if cfg == nil {
		return nil
	}
	var jobs []Job
	for _, ad := range cfg.GetAtomicDomainSets() {
		geositePaths := make(map[string]struct{})
		for _, gs := range geo.AtomicDomainGeosites(ad) {
			if fp := gs.GetFilepath(); fp != "" {
				geositePaths[fp] = struct{}{}
			}
			if url := strings.TrimSpace(gs.GetRemoteUrl()); url != "" {
				if fp := gs.GetFilepath(); fp != "" {
					jobs = append(jobs, Job{
						URL:              url,
						Filepath:         fp,
						CronExpr:         strings.TrimSpace(gs.GetRefreshCron()),
						DomainAtomicName: ad.GetName(),
					})
				}
			}
		}
		for _, rf := range ad.GetRemoteGeoFiles() {
			url := strings.TrimSpace(rf.GetSourceUrl())
			if url == "" {
				continue
			}
			fp := rf.GetFilepath()
			if fp == "" {
				continue
			}
			domainAtomic := ""
			if _, ok := geositePaths[fp]; ok {
				domainAtomic = ad.GetName()
			}
			jobs = append(jobs, Job{
				URL:              url,
				Filepath:         fp,
				CronExpr:         strings.TrimSpace(rf.GetRefreshCron()),
				DomainAtomicName: domainAtomic,
			})
		}
	}
	for _, ai := range cfg.GetAtomicIpSets() {
		var ipPath string
		if gip := ai.GetGeoip(); gip != nil {
			ipPath = gip.GetFilepath()
			if url := strings.TrimSpace(gip.GetRemoteUrl()); url != "" && ipPath != "" {
				jobs = append(jobs, Job{
					URL:          url,
					Filepath:     ipPath,
					CronExpr:     strings.TrimSpace(gip.GetRefreshCron()),
					IPAtomicName: ai.GetName(),
				})
			}
		}
		for _, rf := range ai.GetRemoteGeoFiles() {
			url := strings.TrimSpace(rf.GetSourceUrl())
			if url == "" {
				continue
			}
			fp := rf.GetFilepath()
			if fp == "" {
				continue
			}
			ipAtomic := ""
			if ipPath != "" && ipPath == fp {
				ipAtomic = ai.GetName()
			}
			jobs = append(jobs, Job{
				URL:          url,
				Filepath:     fp,
				CronExpr:     strings.TrimSpace(rf.GetRefreshCron()),
				IPAtomicName: ipAtomic,
			})
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].Filepath < jobs[j].Filepath })
	return jobs
}
