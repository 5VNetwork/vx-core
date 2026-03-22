package http

import (
	"github.com/5vnetwork/vx-core/common/dice"
)

const protocolName = "http"

func httpHosts(c *HttpConfig) []string {
	if len(c.GetHost()) == 0 {
		return []string{"www.example.com"}
	}
	return c.GetHost()
}

func isValidHTTPHost(c *HttpConfig, host string) bool {
	hosts := httpHosts(c)
	for _, h := range hosts {
		if h == host {
			return true
		}
	}
	return false
}

func randomHTTPHost(c *HttpConfig) string {
	hosts := httpHosts(c)
	return hosts[dice.Roll(len(hosts))]
}

func normalizedHTTPPath(c *HttpConfig) string {
	path := c.GetPath()
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}
