package websocket

import (
	"net/http"
)

const protocolName = "websocket"

func (c *WebsocketConfig) GetNormalizedPath() string {
	path := c.Path
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

func (c *WebsocketConfig) GetRequestHeader() http.Header {
	header := http.Header{}
	for _, h := range c.Header {
		header.Add(h.Key, h.Value)
	}
	header.Add("Host", c.Host)
	return header
}
