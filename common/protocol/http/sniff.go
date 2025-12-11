package http

import (
	"bytes"
	"errors"
	"strings"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"golang.org/x/net/http2"
)

type SniffHeader struct {
	host    string
	version string
}

func (h *SniffHeader) Protocol() string {
	if h.version == "" {
		return "http"
	}
	if h.version == "HTTP/1.1" || h.version == "HTTP/1.0" {
		return "http1"
	}
	return "http2"
}

func (h *SniffHeader) Domain() string {
	return h.host
}

var (
	// refer to https://pkg.go.dev/net/http@master#pkg-constants
	methods = [...]string{"get", "post", "head", "put", "delete", "options", "connect", "patch", "trace"}

	errNotHTTPMethod = errors.New("not an HTTP method")

	// HTTP/2 connection preface (RFC 7540 Section 3.5)
	http2Preface = http2.ClientPreface
)

func beginWithHTTPMethod(b []byte) error {
	for _, m := range &methods {
		if len(b) >= len(m) && strings.EqualFold(string(b[:len(m)]), m) {
			return nil
		}

		if len(b) < len(m) {
			return protocol.ErrNoClue
		}
	}

	return errNotHTTPMethod
}

func SniffHTTP1Host(b []byte) (*SniffHeader, error) {
	if err := beginWithHTTPMethod(b); err != nil {
		return nil, err
	}

	sh := &SniffHeader{}

	headers := bytes.Split(b, []byte{'\n'})

	// Parse headers starting from the second line
	for i := 1; i < len(headers); i++ {
		header := headers[i]
		if len(header) == 0 {
			break
		}
		parts := bytes.SplitN(header, []byte{':'}, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(string(parts[0]))
		if key == "host" {
			rawHost := strings.ToLower(string(bytes.TrimSpace(parts[1])))
			dest, err := ParseHost(rawHost, net.Port(80))
			if err != nil {
				return nil, err
			}
			sh.host = dest.Address.String()
		}
	}

	if len(sh.host) > 0 {
		return sh, nil
	}

	return nil, protocol.ErrNoClue
}

type Http1Header struct {
	host    string
	version string
	method  string
	path    string
	query   string
}

func (h *Http1Header) Host() string {
	return h.host
}

// Method returns the HTTP method (GET, POST, etc.)
func (h *Http1Header) Method() string {
	return h.method
}

// Path returns the request path
func (h *Http1Header) Path() string {
	return h.path
}

// Query returns the query string
func (h *Http1Header) Query() string {
	return h.query
}

// Version returns the HTTP version (e.g., "HTTP/1.1", "HTTP/1.0", "HTTP/2.0")
func (h *Http1Header) Version() string {
	return h.version
}

// FullPath returns the complete path including query string
func (h *Http1Header) FullPath() string {
	if h.query != "" {
		return h.path + "?" + h.query
	}
	return h.path
}

// if whether b is http1 cannot be determined, return nil, protocol.ErrNoClue
// if b is not http1, return nil, errNotHttp1
// if b is http1, returns a Http1Header
func SniffHttp1(b []byte) (*Http1Header, error) {
	if err := beginWithHTTPMethod(b); err != nil {
		return nil, err
	}

	sh := &Http1Header{}

	lines := bytes.Split(b, []byte{'\n'})
	// Parse the request line (first line): METHOD PATH HTTP/VERSION
	if len(lines) > 0 {
		requestLine := bytes.TrimSpace(lines[0])
		parts := bytes.Split(requestLine, []byte{' '})
		if len(parts) >= 2 {
			// Extract method
			sh.method = strings.ToUpper(string(parts[0]))

			// Extract path and query
			pathParts := bytes.SplitN(parts[1], []byte{'?'}, 2)
			sh.path = string(pathParts[0])
			if len(pathParts) == 2 {
				sh.query = string(pathParts[1])
			}

			// Extract version (third part if present)
			if len(parts) >= 3 {
				sh.version = strings.ToUpper(string(parts[2]))
			}
		}
	} else {
		return nil, errNotHTTPMethod
	}

	// Parse headers starting from the second line
	for i := 1; i < len(lines); i++ {
		header := lines[i]
		if len(header) == 0 {
			break
		}
		parts := bytes.SplitN(header, []byte{':'}, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(string(parts[0]))
		if key == "host" {
			rawHost := strings.ToLower(string(bytes.TrimSpace(parts[1])))
			dest, err := ParseHost(rawHost, net.Port(80))
			if err != nil {
				return nil, err
			}
			sh.host = dest.Address.String()
		}
	}

	return sh, nil
}

func IsHttp2(b []byte) bool {
	if len(b) < len(http2Preface) {
		return false
	}
	if !bytes.Equal(b[:len(http2Preface)], []byte(http2Preface)) {
		return false
	}
	return true
}

// func SniffHTTP2(b []byte) (*SniffHeader, error) {
// 	// Check for HTTP/2 connection preface
// 	offset := 0
// 	http2Confirmed := false
// 	if len(b) >= len(http2Preface) && bytes.Equal(b[:len(http2Preface)], []byte(http2Preface)) {
// 		offset = len(http2Preface)
// 		http2Confirmed = true
// 	}

// 	// Scan HTTP/2 frames
// 	for offset+9 <= len(b) {
// 		// Parse HTTP/2 frame header (9 bytes)
// 		// +-----------------------------------------------+
// 		// |                 Length (24 bits)              |
// 		// +---------------+---------------+---------------+
// 		// |   Type (8)    |   Flags (8)   |
// 		// +-+-------------+---------------+-------------------------------+
// 		// |R|                 Stream Identifier (31)                      |
// 		// +=+=============================================================+

// 		length := int(b[offset])<<16 | int(b[offset+1])<<8 | int(b[offset+2])
// 		frameType := b[offset+3]
// 		_ = b[offset+4] // flags - not used in current validation but part of frame header
// 		streamID := binary.BigEndian.Uint32(b[offset+5:offset+9]) & 0x7fffffff

// 		// Validate frame type - valid HTTP/2 frame types are 0x0-0x9
// 		if frameType > 0x9 {
// 			break // Invalid frame type, probably not HTTP/2
// 		}

// 		// Additional validation to avoid false positives with other protocols
// 		// 1. Frame length should be reasonable (< 16MB, which is the max frame size)
// 		if length > 16777215 { // 2^24 - 1
// 			break
// 		}

// 		// 2. SETTINGS frame (0x4) validation - most common first frame after preface
// 		if frameType == 0x4 {
// 			// SETTINGS must be on stream 0
// 			if streamID != 0 {
// 				break // Invalid SETTINGS frame
// 			}
// 			// SETTINGS payload must be multiple of 6 bytes
// 			if length%6 != 0 {
// 				break // Invalid SETTINGS frame
// 			}
// 			// Mark as HTTP/2 only after validating SETTINGS frame
// 			http2Confirmed = true
// 		}

// 		// 3. HEADERS/CONTINUATION frames must have stream ID > 0
// 		if (frameType == 0x1 || frameType == 0x9) && streamID == 0 {
// 			break // Invalid HEADERS/CONTINUATION on stream 0
// 		}

// 		// 4. If we already confirmed HTTP/2 (via preface or valid SETTINGS),
// 		//    then we can trust other frame types
// 		if http2Confirmed {
// 			// Already confirmed via preface or SETTINGS, proceed
// 		} else {
// 			// Not confirmed yet - only trust SETTINGS frames or continue if we have preface
// 			if frameType != 0x4 && offset == 0 {
// 				// First frame is not SETTINGS and we don't have preface
// 				// This is probably not HTTP/2
// 				break
// 			}
// 			if frameType == 0x1 || frameType == 0x9 {
// 				// HEADERS/CONTINUATION can only confirm if we have preface
// 				if offset > 0 {
// 					http2Confirmed = true
// 				}
// 			}
// 		}

// 		// HEADERS frame (0x1) or CONTINUATION frame (0x9) on a stream
// 		if (frameType == 0x1 || frameType == 0x9) && streamID > 0 {
// 			// Check if we have the complete frame
// 			if offset+9+length > len(b) {
// 				// Incomplete frame - try to parse with partial data
// 				availablePayloadLen := len(b) - (offset + 9)
// 				if availablePayloadLen > 0 {
// 					payload := b[offset+9 : offset+9+availablePayloadLen]
// 					if result := parseHTTP2Headers(payload); result != nil {
// 						result.version = "HTTP/2.0"
// 						return result, nil
// 					}
// 				}
// 				// Found HEADERS frame but couldn't parse with partial data
// 				return nil, protocol.ErrProtoNeedMoreData
// 			}

// 			// We have complete frame, parse it fully
// 			payload := b[offset+9 : offset+9+length]
// 			if result := parseHTTP2Headers(payload); result != nil {
// 				result.version = "HTTP/2.0"
// 				return result, nil
// 			}
// 		}

// 		// Check if we can move to next frame
// 		if offset+9+length > len(b) {
// 			break // Not enough data for complete frame
// 		}

// 		offset += 9 + length
// 	}

// 	// If HTTP/2 is confirmed but we don't have enough data for headers
// 	if http2Confirmed {
// 		return nil, protocol.ErrProtoNeedMoreData
// 	}

// 	return nil, protocol.ErrNoClue
// }

// func parseHTTP2Headers(hpackPayload []byte) *SniffHeader {
// 	// HTTP/2 uses HPACK compression for headers
// 	// We'll do a simple string search for common patterns
// 	// This is not a full HPACK decoder, but good enough for sniffing

// 	sh := &SniffHeader{}
// 	s := string(hpackPayload)

// 	// Extract :method pseudo-header
// 	sh.method = extractHTTP2Header(s, ":method")
// 	if sh.method == "" {
// 		sh.method = extractHTTP2Header(s, "method")
// 	}

// 	// Extract :path pseudo-header (contains path + query)
// 	fullPath := extractHTTP2Header(s, ":path")
// 	if fullPath == "" {
// 		fullPath = extractHTTP2Header(s, "path")
// 	}

// 	// If we still don't have path, try to find it by pattern
// 	if fullPath == "" {
// 		fullPath = extractPathPattern(s)
// 	}

// 	// Split path and query
// 	if fullPath != "" {
// 		pathParts := strings.SplitN(fullPath, "?", 2)
// 		sh.path = pathParts[0]
// 		if len(pathParts) == 2 {
// 			sh.query = pathParts[1]
// 		}
// 	}

// 	// Extract :authority pseudo-header (equivalent to Host in HTTP/1.x)
// 	sh.host = extractHTTP2Header(s, ":authority")
// 	if sh.host == "" {
// 		sh.host = extractHTTP2Header(s, "authority")
// 	}
// 	if sh.host == "" {
// 		sh.host = extractHTTP2Header(s, "host")
// 	}

// 	// Validate we have enough information
// 	if sh.host != "" && sh.path != "" {
// 		return sh
// 	}

// 	return nil
// }

// func extractHTTP2Header(s, headerName string) string {
// 	// HPACK can encode headers in various ways, we'll look for common patterns
// 	patterns := []string{
// 		headerName + ":",
// 		headerName + "\x00", // HPACK null separator
// 		headerName + " ",
// 		headerName + "\t",
// 	}

// 	for _, pattern := range patterns {
// 		if idx := strings.Index(s, pattern); idx >= 0 {
// 			start := idx + len(pattern)
// 			// Skip whitespace
// 			for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
// 				start++
// 			}

// 			// Extract until whitespace, newline, null, or control character
// 			end := start
// 			for end < len(s) {
// 				c := s[end]
// 				if c == ' ' || c == '\r' || c == '\n' || c == '\x00' || c == '\t' || c < 0x20 {
// 					break
// 				}
// 				end++
// 			}

// 			if end > start {
// 				return s[start:end]
// 			}
// 		}
// 	}

// 	return ""
// }

// func extractPathPattern(s string) string {
// 	// Look for path-like patterns: starting with /
// 	for i := 0; i < len(s); i++ {
// 		if s[i] == '/' && i+1 < len(s) {
// 			// Check if next character is alphanumeric or another /
// 			next := s[i+1]
// 			if (next >= 'A' && next <= 'Z') || (next >= 'a' && next <= 'z') ||
// 				(next >= '0' && next <= '9') || next == '/' {
// 				// Found potential path start
// 				end := i + 1
// 				for end < len(s) {
// 					c := s[end]
// 					// Valid path characters
// 					if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
// 						(c >= '0' && c <= '9') || c == '/' || c == '-' ||
// 						c == '_' || c == '.' || c == '?' || c == '=' || c == '&' {
// 						end++
// 					} else {
// 						break
// 					}
// 				}

// 				if end > i+1 {
// 					return s[i:end]
// 				}
// 			}
// 		}
// 	}

// 	return ""
// }
