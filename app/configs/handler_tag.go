// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package configs

// HandlerTag returns the tag from either the outbound or chain variant of HandlerConfig.
func HandlerTag(h *HandlerConfig) string {
	if h == nil {
		return ""
	}
	if o := h.GetOutbound(); o != nil {
		return o.GetTag()
	}
	if c := h.GetChain(); c != nil {
		return c.GetTag()
	}
	return ""
}
