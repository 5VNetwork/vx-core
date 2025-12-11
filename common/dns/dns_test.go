package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{
			name:     "empty domain",
			domain:   "",
			expected: "",
		},
		{
			name:     "single part domain",
			domain:   "localhost",
			expected: "localhost",
		},
		{
			name:     "root domain with 2 parts",
			domain:   "example.com",
			expected: "example.com",
		},
		{
			name:     "root domain with 2 parts and trailing dot",
			domain:   "example.com.",
			expected: "example.com",
		},
		{
			name:     "subdomain with 3 parts",
			domain:   "www.example.com",
			expected: "example.com",
		},
		{
			name:     "subdomain with 3 parts and trailing dot",
			domain:   "www.example.com.",
			expected: "example.com",
		},
		{
			name:     "deep subdomain with 4 parts",
			domain:   "sub.sub.example.com",
			expected: "example.com",
		},
		{
			name:     "deep subdomain with 5 parts",
			domain:   "a.b.c.example.com",
			expected: "example.com",
		},
		{
			name:     "domain with single character parts",
			domain:   "a.b.c.d",
			expected: "c.d",
		},
		{
			name:     "domain with numbers",
			domain:   "api.v1.example.com",
			expected: "example.com",
		},
		{
			name:     "domain with hyphens",
			domain:   "my-api.example.com",
			expected: "example.com",
		},
		{
			name:     "domain with underscores",
			domain:   "my_api.example.com",
			expected: "example.com",
		},
		{
			name:     "just trailing dot",
			domain:   ".",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RootDomain(tt.domain)
			assert.Equal(t, tt.expected, result, "RootDomain(%q) = %q, want %q", tt.domain, result, tt.expected)
		})
	}
}
