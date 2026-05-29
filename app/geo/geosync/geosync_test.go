// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package geosync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/stretchr/testify/require"
)

func TestCollectJobs_nil(t *testing.T) {
	require.Nil(t, CollectJobs(nil))
}

func TestCollectJobs_collectAndSort(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "d1",
				Geosite: &configs.GeositeConfig{
					Filepath:    "/data/geosite-d1.dat",
					Codes:       []string{"cn"},
					RemoteUrl:   "https://a.example/a",
					RefreshCron: "0 * * * *",
				},
			},
			{
				Name: "d2",
				Geosite: &configs.GeositeConfig{
					Filepath:    "/data/geosite-d2.dat",
					Codes:       []string{"cn"},
					RemoteUrl:   "https://b.example/b",
					RefreshCron: "30 * * * *",
				},
			},
		},
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "ip1",
				Geoip: &configs.GeoIPConfig{
					Filepath:    "/data/geoip.dat",
					Codes:       []string{"cn"},
					RemoteUrl:   "https://c.example/c",
					RefreshCron: "0 0 * * *",
				},
			},
		},
	}
	jobs := CollectJobs(cfg)
	require.Len(t, jobs, 3)
	require.Equal(t, "/data/geoip.dat", jobs[0].Filepath)
	require.Equal(t, "ip1", jobs[0].IPAtomicName)
	require.Equal(t, "/data/geosite-d1.dat", jobs[1].Filepath)
	require.Equal(t, "d1", jobs[1].DomainAtomicName)
	require.Equal(t, "https://a.example/a", jobs[1].URL)
	require.Equal(t, "0 * * * *", jobs[1].CronExpr)
	require.Equal(t, "/data/geosite-d2.dat", jobs[2].Filepath)
	require.Equal(t, "d2", jobs[2].DomainAtomicName)
}

func TestCollectJobs_geositesAndLegacyRemoteGeoFiles(t *testing.T) {
	cfg := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "multi",
				Geosites: []*configs.GeositeConfig{
					{
						Filepath:    "/data/extra.dat",
						Codes:       []string{"ads"},
						RemoteUrl:   "https://example.com/extra",
						RefreshCron: "0 0 * * *",
					},
				},
				RemoteGeoFiles: []*configs.GeoRemoteFile{
					{Filepath: "/data/legacy.dat", SourceUrl: "https://example.com/legacy", RefreshCron: "0 1 * * *"},
				},
			},
		},
	}
	jobs := CollectJobs(cfg)
	require.Len(t, jobs, 2)
	require.Equal(t, "/data/extra.dat", jobs[0].Filepath)
	require.Equal(t, "multi", jobs[0].DomainAtomicName)
	require.Equal(t, "/data/legacy.dat", jobs[1].Filepath)
	require.Empty(t, jobs[1].DomainAtomicName)
}

func TestFetchViaDirectHTTP_rejectsHTTP(t *testing.T) {
	_, err := FetchViaDirectHTTP(context.Background(), "http://example.com/x")
	require.Error(t, err)
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "sub", "out.bin")
	payload := []byte("hello-geo")
	require.NoError(t, writeFileAtomic(dest, payload))
	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	require.Equal(t, payload, got)
	_, err = os.Stat(dest + ".tmp")
	require.True(t, os.IsNotExist(err))
}
