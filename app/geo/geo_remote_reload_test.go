// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package geo

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	cgeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/i"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReloadAtomicSetsAfterRemoteGeoFile_NilConfig(t *testing.T) {
	g := &Geo{}
	require.NoError(t, g.ReloadAtomicSetsAfterRemoteGeoFile(nil, []string{"a"}, []string{"b"}))
}

func TestReloadAtomicSetsAfterRemoteGeoFile_SkipsEmptyAndUnknownNames(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{Name: "domains", Domains: []*cgeo.Domain{{Type: cgeo.Domain_Full, Value: "a.example"}}},
		},
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{Name: "ips", Cidrs: []*cgeo.CIDR{{Ip: []byte{10, 0, 0, 0}, Prefix: 8}}},
		},
	}
	g, err := NewGeo(cfg)
	require.NoError(t, err)

	beforeDomain := g.DomainSets["domains"]
	beforeIP := g.IpSets["ips"]

	require.NoError(t, g.ReloadAtomicSetsAfterRemoteGeoFile(cfg, []string{"", "missing"}, []string{"", "missing"}))

	assert.Same(t, beforeDomain, g.DomainSets["domains"])
	assert.Same(t, beforeIP, g.IpSets["ips"])
}

func TestReloadAtomicSetsAfterRemoteGeoFile_ReloadsDomainAtomic(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "custom",
				Domains: []*cgeo.Domain{
					{Type: cgeo.Domain_Full, Value: "old.example"},
				},
			},
		},
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{Name: "proxy", InNames: []string{"custom"}},
		},
	}
	g, err := NewGeo(cfg)
	require.NoError(t, err)

	ok, err := g.MatchDomain("old.example", "custom")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = g.MatchDomain("new.example", "custom")
	require.NoError(t, err)
	assert.False(t, ok)

	cfg.AtomicDomainSets[0].Domains = []*cgeo.Domain{
		{Type: cgeo.Domain_Full, Value: "new.example"},
	}
	require.NoError(t, g.ReloadAtomicSetsAfterRemoteGeoFile(cfg, []string{"custom"}, nil))

	ok, err = g.MatchDomain("old.example", "custom")
	require.NoError(t, err)
	assert.False(t, ok)
	ok, err = g.MatchDomain("new.example", "custom")
	require.NoError(t, err)
	assert.True(t, ok)

	// Great domain sets resolve atomics by name; no rebuild needed.
	ok, err = g.MatchDomain("new.example", "proxy")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = g.MatchDomain("old.example", "proxy")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestReloadAtomicSetsAfterRemoteGeoFile_ReloadsIPAtomicAndGreatIPSet(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name:  "base",
				Cidrs: []*cgeo.CIDR{{Ip: []byte{10, 0, 0, 0}, Prefix: 8}},
			},
			{
				Name:  "other",
				Cidrs: []*cgeo.CIDR{{Ip: []byte{192, 168, 0, 0}, Prefix: 16}},
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{Name: "combined", InNames: []string{"base"}},
			{Name: "untouched", InNames: []string{"other"}},
		},
	}
	g, err := NewGeo(cfg)
	require.NoError(t, err)

	ok, err := g.MatchIP(net.IP{10, 1, 2, 3}, "base")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = g.MatchIP(net.IP{172, 16, 0, 1}, "base")
	require.NoError(t, err)
	assert.False(t, ok)
	ok, err = g.MatchIP(net.IP{10, 1, 2, 3}, "combined")
	require.NoError(t, err)
	assert.True(t, ok)

	beforeUntouched := g.IpSets["untouched"]

	cfg.AtomicIpSets[0].Cidrs = []*cgeo.CIDR{{Ip: []byte{172, 16, 0, 0}, Prefix: 12}}
	require.NoError(t, g.ReloadAtomicSetsAfterRemoteGeoFile(cfg, nil, []string{"base"}))

	ok, err = g.MatchIP(net.IP{10, 1, 2, 3}, "base")
	require.NoError(t, err)
	assert.False(t, ok)
	ok, err = g.MatchIP(net.IP{172, 16, 0, 1}, "base")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = g.MatchIP(net.IP{172, 16, 0, 1}, "combined")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = g.MatchIP(net.IP{10, 1, 2, 3}, "combined")
	require.NoError(t, err)
	assert.False(t, ok)

	assert.Same(t, beforeUntouched, g.IpSets["untouched"], "great IP sets that do not reference changed atoms must not be rebuilt")
}

func TestReloadAtomicSetsAfterRemoteGeoFile_RebuildsGreatIPSetViaExNames(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name:  "all",
				Cidrs: []*cgeo.CIDR{{Ip: []byte{0, 0, 0, 0}, Prefix: 0}},
			},
			{
				Name:  "private",
				Cidrs: []*cgeo.CIDR{{Ip: []byte{192, 168, 0, 0}, Prefix: 16}},
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{Name: "public", InNames: []string{"all"}, ExNames: []string{"private"}},
		},
	}
	g, err := NewGeo(cfg)
	require.NoError(t, err)

	ok, err := g.MatchIP(net.IP{192, 168, 1, 1}, "public")
	require.NoError(t, err)
	assert.False(t, ok)

	cfg.AtomicIpSets[1].Cidrs = []*cgeo.CIDR{{Ip: []byte{10, 0, 0, 0}, Prefix: 8}}
	require.NoError(t, g.ReloadAtomicSetsAfterRemoteGeoFile(cfg, nil, []string{"private"}))

	ok, err = g.MatchIP(net.IP{192, 168, 1, 1}, "public")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = g.MatchIP(net.IP{10, 0, 0, 1}, "public")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestReloadAtomicSetsAfterRemoteGeoFile_DomainReloadError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-domains.yaml")
	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "broken",
				RemoteGeoFiles: []*configs.GeoRemoteFile{
					{Filepath: missing},
				},
			},
		},
	}
	geo := &Geo{
		DomainSets: make(map[string]i.DomainSet),
		IpSets:     make(map[string]i.IPSet),
	}
	err := geo.ReloadAtomicSetsAfterRemoteGeoFile(cfg, []string{"broken"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `reload domain set "broken"`)
}

func TestReloadAtomicSetsAfterRemoteGeoFile_IPReloadError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-ips.yaml")
	cfg := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "broken-ip",
				RemoteGeoFiles: []*configs.GeoRemoteFile{
					{Filepath: missing},
				},
			},
		},
	}
	geo := &Geo{
		DomainSets: make(map[string]i.DomainSet),
		IpSets:     make(map[string]i.IPSet),
	}
	err := geo.ReloadAtomicSetsAfterRemoteGeoFile(cfg, nil, []string{"broken-ip"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `reload ip set "broken-ip"`)
}

func TestReloadAtomicSetsAfterRemoteGeoFile_FromUpdatedRemoteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	require.NoError(t, os.WriteFile(path, []byte("payload:\n  - DOMAIN,first.example\n"), 0o644))

	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "remote-domains",
				RemoteGeoFiles: []*configs.GeoRemoteFile{
					{Filepath: path},
				},
			},
		},
	}
	g, err := NewGeo(cfg)
	require.NoError(t, err)

	ok, err := g.MatchDomain("first.example", "remote-domains")
	require.NoError(t, err)
	assert.True(t, ok)

	require.NoError(t, os.WriteFile(path, []byte("payload:\n  - DOMAIN,second.example\n"), 0o644))
	require.NoError(t, g.ReloadAtomicSetsAfterRemoteGeoFile(cfg, []string{"remote-domains"}, nil))

	ok, err = g.MatchDomain("first.example", "remote-domains")
	require.NoError(t, err)
	assert.False(t, ok)
	ok, err = g.MatchDomain("second.example", "remote-domains")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAtomicDomainSetByName(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{Name: "a"},
			{Name: "b"},
		},
	}
	assert.Equal(t, "b", atomicDomainSetByName(cfg, "b").GetName())
	assert.Nil(t, atomicDomainSetByName(cfg, "missing"))
}

func TestAtomicIPSetByName(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{Name: "a"},
			{Name: "b"},
		},
	}
	assert.Equal(t, "b", atomicIPSetByName(cfg, "b").GetName())
	assert.Nil(t, atomicIPSetByName(cfg, "missing"))
}

func TestGreatIPSetReferencesAtoms(t *testing.T) {
	atoms := map[string]struct{}{"x": {}}
	assert.True(t, greatIPSetReferencesAtoms(&configs.GreatIPSetConfig{InNames: []string{"x"}}, atoms))
	assert.True(t, greatIPSetReferencesAtoms(&configs.GreatIPSetConfig{ExNames: []string{"x"}}, atoms))
	assert.False(t, greatIPSetReferencesAtoms(&configs.GreatIPSetConfig{InNames: []string{"y"}}, atoms))
}
