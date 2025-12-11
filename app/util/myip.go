package util

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const testUrl = "https://blog.cloudflare.com/cdn-cgi/trace"

func GetMyIPv4() (string, error) {
	// Create HTTP client with timeout
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial("tcp4", addr)
	}
	client := &http.Client{
		Transport: httpTransport,
		Timeout:   10 * time.Second,
	}

	// Make GET request to testUrl
	resp, err := client.Get(testUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	pairs := ParseKeyValueText(string(body))
	return pairs["ip"], nil
}

func ParseKeyValueText(text string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result
}
