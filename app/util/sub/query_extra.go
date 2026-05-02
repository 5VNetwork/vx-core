// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package sub

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ShareLinkQueryExtraFromStored parses the subscription DB text field into a map.
// Accepts JSON objects (e.g. {"tx":"10"}) or query strings (e.g. tx=10&foo=bar).
func ShareLinkQueryExtraFromStored(s string) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if strings.HasPrefix(s, "{") {
		var strMap map[string]string
		if err := json.Unmarshal([]byte(s), &strMap); err == nil && len(strMap) > 0 {
			return strMap
		}
		var anyMap map[string]any
		if err := json.Unmarshal([]byte(s), &anyMap); err == nil && len(anyMap) > 0 {
			out := make(map[string]string, len(anyMap))
			for k, v := range anyMap {
				out[k] = fmt.Sprintf("%v", v)
			}
			return out
		}
	}
	q, err := url.ParseQuery(strings.TrimPrefix(s, "?"))
	if err != nil || len(q) == 0 {
		return nil
	}
	out := make(map[string]string, len(q))
	for k, vals := range q {
		if len(vals) == 0 {
			continue
		}
		out[k] = vals[0]
	}
	return out
}
