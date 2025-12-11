package logger

import (
	"regexp"
	"testing"
)

func TestRedactSensitiveData(t *testing.T) {
	logger := &Logger{
		redactionRegex: regexp.MustCompile(ipv4Pattern + "|" + ipv6Pattern + "|" + domainPattern),
	}

	normalValues := []string{
		"abcd",
		"1234567890",
		"An Error Occured",
		"An Error Occurred: 192.168.1.1",
		"An Error Occurred: 2001:cdba::3257:9652",
		"An Error Occurred: example.com",
		"An Error Occurred: subdomain.example.com",
		"An Error Occurred: api.v1.service.example.org",
		"An Error Occurred: www.google.com",
		"Connection failed to 8.8.8.8 via proxy.example.com",
		"Multiple IPs: 192.168.1.1, 10.0.0.1, and domain api.service.com",
		"IPv6 address 2001:db8::1 and domain subdomain.example.org in same message",
	}
	for _, value := range normalValues {
		redacted := logger.RedactSensitiveData(value)
		t.Logf("Normal: %s -> %s", value, redacted)
	}

	// Test IPv4 addresses
	ipv4Tests := []string{
		"192.168.1.1",
		"91.108.255.254",
		"8.8.8.8",
		"127.0.0.1",
	}

	for _, ip := range ipv4Tests {
		redacted := logger.RedactSensitiveData(ip)
		t.Logf("IPv4: %s -> %s", ip, redacted)
	}

	// Test IPv6 addresses
	ipv6Tests := []string{
		"2001:cdba::3257:9652",                    // compressed
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334", // full
		"::1",              // localhost
		"2001:db8::",       // compressed with empty right side
		"::ffff:192.0.2.1", // IPv4-mapped
		"fe80::1",          // link-local
	}

	for _, ip := range ipv6Tests {
		redacted := logger.RedactSensitiveData(ip)
		t.Logf("IPv6: %s -> %s", ip, redacted)
	}

	// Test domain names
	domainTests := []string{
		"example.com",
		"subdomain.example.com",
		"api.v1.service.example.org",
		"www.google.com",
		"mail.server.company.co.uk",
		"test.local",
	}

	for _, domain := range domainTests {
		redacted := logger.RedactSensitiveData(domain)
		t.Logf("Domain: %s -> %s", domain, redacted)
	}
}
