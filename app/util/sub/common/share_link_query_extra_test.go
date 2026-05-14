// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"strings"
	"testing"
)

func TestApplyShareLinkQueryExtra_vless(t *testing.T) {
	const line = "vless://12345678-1234-1234-1234-123456789012@a.b.com:12345?encryption=none&flow=xtls-rprx-vision&fp=chrome&network=ws&path=%2Fvvv&security=tls&sni=a.b.com#%E6%B5%8B%E8%AF%95vless"
	got := ApplyShareLinkQueryExtra(line, map[string]string{"tx": "10"})
	if !strings.Contains(got, "tx=10") {
		t.Fatalf("expected tx=10 in %q", got)
	}
	if !strings.Contains(got, "encryption=none") {
		t.Fatal("dropped existing query")
	}
	if !strings.Contains(got, "#") {
		t.Fatal("lost fragment")
	}
}

func TestApplyShareLinkQueryExtra_vmessUnchanged(t *testing.T) {
	line := "vmess://eyJhZGRyIjoiMSJ9"
	got := ApplyShareLinkQueryExtra(line, map[string]string{"tx": "1"})
	if got != line {
		t.Fatalf("vmess should be unchanged, got %q", got)
	}
}
