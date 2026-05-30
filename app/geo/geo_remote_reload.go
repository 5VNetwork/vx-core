// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package geo

import (
	"fmt"
	"runtime"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/geo/memloader"
	"github.com/5vnetwork/vx-core/common/geo/stdloader"
)

func defaultGeoFileLoader() loader {
	if runtime.GOOS == "ios" {
		return memloader.New()
	}
	return stdloader.NewStandartLoader()
}

func atomicDomainSetByName(cfg *configs.GeoConfig, name string) *configs.AtomicDomainSetConfig {
	for _, ad := range cfg.GetAtomicDomainSets() {
		if ad.GetName() == name {
			return ad
		}
	}
	return nil
}

func atomicIPSetByName(cfg *configs.GeoConfig, name string) *configs.AtomicIPSetConfig {
	for _, ai := range cfg.GetAtomicIpSets() {
		if ai.GetName() == name {
			return ai
		}
	}
	return nil
}

func greatIPSetReferencesAtoms(c *configs.GreatIPSetConfig, atoms map[string]struct{}) bool {
	for _, n := range c.GetInNames() {
		if _, ok := atoms[n]; ok {
			return true
		}
	}
	for _, n := range c.GetExNames() {
		if _, ok := atoms[n]; ok {
			return true
		}
	}
	return false
}

// ReloadAtomicSetsAfterRemoteGeoFile rebuilds matchers for the given atomic domain and IP set
// names after on-disk geo files were updated. Any great IP set that includes a changed
// atomic IP set is rebuilt as well. Great domain sets resolve atomics by name on each match,
// so they pick up domain atomic updates without rebuilding.
func (g *Geo) ReloadAtomicSetsAfterRemoteGeoFile(cfg *configs.GeoConfig, domainAtomicNames []string, ipAtomicNames []string) error {
	g.Lock()
	defer g.Unlock()
	if cfg == nil {
		return nil
	}
	l := defaultGeoFileLoader()
	for _, name := range domainAtomicNames {
		if name == "" {
			continue
		}
		atomic := atomicDomainSetByName(cfg, name)
		if atomic == nil {
			continue
		}
		m, err := AtomicDomainSetToIndexMatcher(atomic, l)
		if err != nil {
			return fmt.Errorf("reload domain set %q: %w", name, err)
		}
		g.DomainSets[name] = &IndexMatcherToDomainSet{
			IndexMatcher: m, reverseMatch: atomic.Inverse,
		}
	}
	changedAtoms := make(map[string]struct{})
	for _, name := range ipAtomicNames {
		if name == "" {
			continue
		}
		atomic := atomicIPSetByName(cfg, name)
		if atomic == nil {
			continue
		}
		m, err := AtomicIpSetToIPMatcher(atomic, l)
		if err != nil {
			return fmt.Errorf("reload ip set %q: %w", name, err)
		}
		g.IpSets[name] = m
		changedAtoms[name] = struct{}{}
	}
	if len(changedAtoms) == 0 {
		return nil
	}
	for _, greatCfg := range cfg.GetGreatIpSets() {
		if !greatIPSetReferencesAtoms(greatCfg, changedAtoms) {
			continue
		}
		m, err := getGreatIPSet(greatCfg, g.IpSets)
		if err != nil {
			return fmt.Errorf("reload great ip set %q: %w", greatCfg.GetName(), err)
		}
		g.IpSets[greatCfg.GetName()] = m
	}
	return nil
}
