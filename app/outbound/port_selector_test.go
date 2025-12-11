package outbound

import (
	"testing"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRandomPortSelector(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 1000, To: 2000},
		{From: 3000, To: 4000},
	}

	selector := NewRandomPortSelector(ranges)
	require.NotNil(t, selector)
	assert.Equal(t, ranges, selector.ranges)
}

func TestRandomPortSelector_SelectPort_SingleRange(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 1000, To: 1005},
	}

	selector := NewRandomPortSelector(ranges)

	// Test multiple selections to ensure randomness
	ports := make(map[uint16]bool)
	for i := 0; i < 100; i++ {
		port := selector.SelectPort()
		assert.GreaterOrEqual(t, port, uint16(1000))
		assert.LessOrEqual(t, port, uint16(1005))
		ports[port] = true
	}

	// Should have selected all ports in range
	assert.Equal(t, 6, len(ports)) // 1000, 1001, 1002, 1003, 1004, 1005
}

func TestRandomPortSelector_SelectPort_MultipleRanges(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 1000, To: 1002}, // 3 ports
		{From: 2000, To: 2001}, // 2 ports
		{From: 3000, To: 3000}, // 1 port
	}

	selector := NewRandomPortSelector(ranges)

	// Test multiple selections
	ports := make(map[uint16]bool)
	for i := 0; i < 200; i++ {
		port := selector.SelectPort()
		assert.True(t, port == 1000 || port == 1001 || port == 1002 ||
			port == 2000 || port == 2001 ||
			port == 3000)
		ports[port] = true
	}

	// Should have selected all ports from all ranges
	expectedPorts := map[uint16]bool{
		1000: true, 1001: true, 1002: true,
		2000: true, 2001: true,
		3000: true,
	}
	assert.Equal(t, expectedPorts, ports)
}

func TestRandomPortSelector_SelectPort_SinglePortRange(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 5000, To: 5000},
	}

	selector := NewRandomPortSelector(ranges)

	// Should always return the same port
	for i := 0; i < 10; i++ {
		port := selector.SelectPort()
		assert.Equal(t, uint16(5000), port)
	}
}

func TestRandomPortSelector_SelectPort_EmptyRanges(t *testing.T) {
	selector := NewRandomPortSelector(nil)

	port := selector.SelectPort()
	assert.Equal(t, uint16(0), port)
}

func TestRandomPortSelector_SelectPort_EmptyRangesSlice(t *testing.T) {
	selector := NewRandomPortSelector([]*net.PortRange{})

	port := selector.SelectPort()
	assert.Equal(t, uint16(0), port)
}

func TestRandomPortSelector_SelectPort_ConcurrentAccess(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 1000, To: 2000},
		{From: 3000, To: 4000},
	}

	selector := NewRandomPortSelector(ranges)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				port := selector.SelectPort()
				assert.GreaterOrEqual(t, port, uint16(1000))
				assert.LessOrEqual(t, port, uint16(4000))
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRandomPortSelector_SelectPort_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		ranges   []*net.PortRange
		expected uint16
	}{
		{
			name:     "nil ranges",
			ranges:   nil,
			expected: 0,
		},
		{
			name:     "empty ranges",
			ranges:   []*net.PortRange{},
			expected: 0,
		},
		{
			name: "single port range",
			ranges: []*net.PortRange{
				{From: 8080, To: 8080},
			},
			expected: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRandomPortSelector(tt.ranges)
			port := selector.SelectPort()

			assert.Equal(t, tt.expected, port)
		})
	}
}

func TestRandomPortSelector_SelectPort_Distribution(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 1000, To: 1001}, // 2 ports
		{From: 2000, To: 2002}, // 3 ports
	}

	selector := NewRandomPortSelector(ranges)

	// Count selections from each range
	range1Count := 0
	range2Count := 0
	totalSelections := 1000

	for i := 0; i < totalSelections; i++ {
		port := selector.SelectPort()

		if port == 1000 || port == 1001 {
			range1Count++
		} else if port == 2000 || port == 2001 || port == 2002 {
			range2Count++
		} else {
			t.Fatalf("Unexpected port: %d", port)
		}
	}

	// With random selection, we expect roughly equal distribution
	// Allow for some variance (±20%)
	expectedRange1 := totalSelections / 2
	expectedRange2 := totalSelections / 2

	assert.Greater(t, range1Count, expectedRange1*4/5)
	assert.Less(t, range1Count, expectedRange1*6/5)
	assert.Greater(t, range2Count, expectedRange2*4/5)
	assert.Less(t, range2Count, expectedRange2*6/5)
}

func TestRandomPortSelector_SelectPort_AllPortsSelected(t *testing.T) {
	ranges := []*net.PortRange{
		{From: 1000, To: 1002},
	}

	selector := NewRandomPortSelector(ranges)

	// Select all possible ports
	selectedPorts := make(map[uint16]bool)
	attempts := 0
	maxAttempts := 1000

	for len(selectedPorts) < 3 && attempts < maxAttempts {
		port := selector.SelectPort()
		selectedPorts[port] = true
		attempts++
	}

	// Should have selected all 3 ports (1000, 1001, 1002)
	assert.Equal(t, 3, len(selectedPorts))
	assert.True(t, selectedPorts[1000])
	assert.True(t, selectedPorts[1001])
	assert.True(t, selectedPorts[1002])
}
