// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package geosync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const maxGeoDownloadBytes = 64 << 20

func validateHTTPS(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("geosync: only https URLs are allowed")
	}
	return u, nil
}

// FetchViaDirectHTTP downloads using the process default HTTP transport (no outbound routing).
func FetchViaDirectHTTP(ctx context.Context, rawURL string) ([]byte, error) {
	if _, err := validateHTTPS(rawURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geosync: http %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxGeoDownloadBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(body) > maxGeoDownloadBytes {
		return nil, fmt.Errorf("geosync: response exceeds size limit")
	}
	return body, nil
}

func writeFileAtomic(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dest)
}
