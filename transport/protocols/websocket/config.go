package websocket

import (
	"net/http"
)

const protocolName = "websocket"

func normalizedWebsocketPath(c *WebsocketConfig) string {
	path := c.GetPath()
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

func websocketRequestHeader(c *WebsocketConfig) http.Header {
	header := http.Header{}
	for _, h := range c.GetHeader() {
		header.Add(h.GetKey(), h.GetValue())
	}
	header.Add("Host", c.GetHost())
	return header
}
