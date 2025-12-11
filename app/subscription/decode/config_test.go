package decode

import (
	"testing"

	mynet "github.com/5vnetwork/vx-core/common/net"
)

func TestTryParsePorts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []mynet.PortRange
		wantNil  bool
	}{
		{
			name:  "single port",
			input: "123",
			expected: []mynet.PortRange{
				{From: 123, To: 123},
			},
			wantNil: false,
		},
		{
			name:  "port range",
			input: "5000-6000",
			expected: []mynet.PortRange{
				{From: 5000, To: 6000},
			},
			wantNil: false,
		},
		{
			name:  "multiple ports and ranges",
			input: "123,5000-6000,8080",
			expected: []mynet.PortRange{
				{From: 123, To: 123},
				{From: 5000, To: 6000},
				{From: 8080, To: 8080},
			},
			wantNil: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantNil: true,
		},
		{
			name:    "invalid port number",
			input:   "abc",
			wantNil: true,
		},
		{
			name:    "invalid port range",
			input:   "5000-",
			wantNil: true,
		},
		{
			name:    "invalid port range format",
			input:   "5000-6000-7000",
			wantNil: true,
		},
		{
			name:    "invalid port range values",
			input:   "abc-def",
			wantNil: true,
		},
		{
			name:    "invalid multiple ports",
			input:   "123,abc,456",
			wantNil: true,
		},
		{
			name:    "invalid multiple ranges",
			input:   "123-456,abc-def,789-012",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TryParsePorts(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("TryParsePorts() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("TryParsePorts() = nil, want %v", tt.expected)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("TryParsePorts() length = %v, want %v", len(got), len(tt.expected))
				return
			}

			for i := range got {
				if got[i].From != tt.expected[i].From || got[i].To != tt.expected[i].To {
					t.Errorf("TryParsePorts()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
