package http

import (
	"github.com/5vnetwork/vx-core/common/dice"
)

const protocolName = "http"

func (c *HttpConfig) getHosts() []string {
	if len(c.Host) == 0 {
		return []string{"www.example.com"}
	}
	return c.Host
}

func (c *HttpConfig) isValidHost(host string) bool {
	hosts := c.getHosts()
	for _, h := range hosts {
		if h == host {
			return true
		}
	}
	return false
}

func (c *HttpConfig) getRandomHost() string {
	hosts := c.getHosts()
	return hosts[dice.Roll(len(hosts))]
}

func (c *HttpConfig) getNormalizedPath() string {
	if c.Path == "" {
		return "/"
	}
	if c.Path[0] != '/' {
		return "/" + c.Path
	}
	return c.Path
}
