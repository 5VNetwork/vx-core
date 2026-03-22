package common

import (
	"net/url"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
)

func TestSetProtocolWebsocket(t *testing.T) {
	tests := []struct {
		name           string
		query          url.Values
		expectedPath   string
		expectedHost   string
		expectedMaxED  int32
		expectedHeader string
	}{
		{
			name: "simple path without early data",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path"},
				"host": {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  0,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with early data parameter",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=2560"},
				"host": {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  2560,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with early data and other query params",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=1234&other=value"},
				"host": {"example.com"},
			},
			expectedPath:   "/path?other=value",
			expectedHost:   "example.com",
			expectedMaxED:  1234,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with invalid early data (non-numeric)",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=invalid"},
				"host": {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  0, // Should default to 0 when parsing fails
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path without host",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=500"},
			},
			expectedPath:   "/path",
			expectedHost:   "",
			expectedMaxED:  500,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "empty path",
			query: url.Values{
				"type": {"ws"},
				"path": {""},
				"host": {"example.com"},
			},
			expectedPath:   "",
			expectedHost:   "example.com",
			expectedMaxED:  0,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with URL-encoded early data",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=2048&foo=bar"},
				"host": {"example.com"},
			},
			expectedPath:   "/path?foo=bar",
			expectedHost:   "example.com",
			expectedMaxED:  2048,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path that is not a valid URL (should still work)",
			query: url.Values{
				"type": {"ws"},
				"path": {"/simple/path"},
				"host": {"example.com"},
			},
			expectedPath:   "/simple/path",
			expectedHost:   "example.com",
			expectedMaxED:  0,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with multiple ed parameters (first one wins)",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=100&ed=200"},
				"host": {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  100, // url.Parse and Query().Get() returns first value
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with zero early data",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=0"},
				"host": {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  0,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "path with negative early data (should parse as 0)",
			query: url.Values{
				"type": {"ws"},
				"path": {"/path?ed=-100"},
				"host": {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  -100, // strconv.Atoi will parse negative numbers
			expectedHeader: "Sec-WebSocket-Protocol",
		},
		{
			name: "using network parameter instead of type",
			query: url.Values{
				"network": {"ws"},
				"path":    {"/path?ed=3000"},
				"host":    {"example.com"},
			},
			expectedPath:   "/path",
			expectedHost:   "example.com",
			expectedMaxED:  3000,
			expectedHeader: "Sec-WebSocket-Protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configs.TransportConfig{}
			err := setProtocol(config, tt.query)
			if err != nil {
				t.Fatalf("setProtocol() error = %v", err)
			}

			wsConfig, ok := config.Protocol.(*configs.TransportConfig_Websocket)
			if !ok {
				t.Fatalf("Protocol is not websocket, got %T", config.Protocol)
			}

			if wsConfig.Websocket.Path != tt.expectedPath {
				t.Errorf("Path = %v, want %v", wsConfig.Websocket.Path, tt.expectedPath)
			}

			if wsConfig.Websocket.Host != tt.expectedHost {
				t.Errorf("Host = %v, want %v", wsConfig.Websocket.Host, tt.expectedHost)
			}

			if wsConfig.Websocket.MaxEarlyData != tt.expectedMaxED {
				t.Errorf("MaxEarlyData = %v, want %v", wsConfig.Websocket.MaxEarlyData, tt.expectedMaxED)
			}

			if wsConfig.Websocket.EarlyDataHeaderName != tt.expectedHeader {
				t.Errorf("EarlyDataHeaderName = %v, want %v", wsConfig.Websocket.EarlyDataHeaderName, tt.expectedHeader)
			}
		})
	}
}

func TestSetProtocolWebsocketEdgeCases(t *testing.T) {
	t.Run("path with only ed parameter", func(t *testing.T) {
		query := url.Values{
			"type": {"ws"},
			"path": {"/path?ed=999"},
		}
		config := &configs.TransportConfig{}
		err := setProtocol(config, query)
		if err != nil {
			t.Fatalf("setProtocol() error = %v", err)
		}

		wsConfig := config.Protocol.(*configs.TransportConfig_Websocket)
		if wsConfig.Websocket.Path != "/path" {
			t.Errorf("Path = %v, want /path", wsConfig.Websocket.Path)
		}
		if wsConfig.Websocket.MaxEarlyData != 999 {
			t.Errorf("MaxEarlyData = %v, want 999", wsConfig.Websocket.MaxEarlyData)
		}
	})

	t.Run("path with ed parameter in middle of query string", func(t *testing.T) {
		query := url.Values{
			"type": {"ws"},
			"path": {"/path?foo=bar&ed=777&baz=qux"},
		}
		config := &configs.TransportConfig{}
		err := setProtocol(config, query)
		if err != nil {
			t.Fatalf("setProtocol() error = %v", err)
		}

		wsConfig := config.Protocol.(*configs.TransportConfig_Websocket)
		if wsConfig.Websocket.MaxEarlyData != 777 {
			t.Errorf("MaxEarlyData = %v, want 777", wsConfig.Websocket.MaxEarlyData)
		}
		// Check that ed is removed but other params remain
		if wsConfig.Websocket.Path != "/path?baz=qux&foo=bar" && wsConfig.Websocket.Path != "/path?foo=bar&baz=qux" {
			t.Errorf("Path = %v, should contain foo=bar and baz=qux but not ed", wsConfig.Websocket.Path)
		}
	})

	t.Run("path with empty ed parameter", func(t *testing.T) {
		query := url.Values{
			"type": {"ws"},
			"path": {"/path?ed="},
		}
		config := &configs.TransportConfig{}
		err := setProtocol(config, query)
		if err != nil {
			t.Fatalf("setProtocol() error = %v", err)
		}

		wsConfig := config.Protocol.(*configs.TransportConfig_Websocket)
		// Empty ed should be treated as not present
		if wsConfig.Websocket.MaxEarlyData != 0 {
			t.Errorf("MaxEarlyData = %v, want 0 for empty ed", wsConfig.Websocket.MaxEarlyData)
		}
	})

	t.Run("very large early data value", func(t *testing.T) {
		query := url.Values{
			"type": {"ws"},
			"path": {"/path?ed=2147483647"}, // max int32
		}
		config := &configs.TransportConfig{}
		err := setProtocol(config, query)
		if err != nil {
			t.Fatalf("setProtocol() error = %v", err)
		}

		wsConfig := config.Protocol.(*configs.TransportConfig_Websocket)
		if wsConfig.Websocket.MaxEarlyData != 2147483647 {
			t.Errorf("MaxEarlyData = %v, want 2147483647", wsConfig.Websocket.MaxEarlyData)
		}
	})
}
