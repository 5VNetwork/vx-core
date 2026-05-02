// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"net/url"
	"strings"
)

// ApplyShareLinkQueryExtra merges extra query parameters into a single subscription
// line before URI decoding. Keys already present on the link are overwritten.
// vmess:// links are left unchanged. Unparseable lines are returned as-is.
func ApplyShareLinkQueryExtra(line string, extra map[string]string) string {
	if len(extra) == 0 || line == "" {
		return line
	}
	if strings.HasPrefix(line, "vmess://") {
		return line
	}
	u, err := url.Parse(line)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return line
	}
	switch u.Scheme {
	case "vless", "trojan", "ss", "socks5", "socks", "hysteria2", "hy2", "anytls":
	default:
		return line
	}
	q := u.Query()
	for k, v := range extra {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
