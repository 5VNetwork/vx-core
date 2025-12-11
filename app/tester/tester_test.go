package tester

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNamedHandler implements i.Outbound for testing
type MockNamedHandler struct {
	tag string
}

func (m *MockNamedHandler) Tag() string {
	return m.tag
}

func (m *MockNamedHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	return nil
}

func (m *MockNamedHandler) HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error {
	return nil
}

// MockResultReporter implements ResultReporter for testing
type MockResultReporter struct {
	mock.Mock
}

func (m *MockResultReporter) UsableResult(tag string, ok bool) {
	m.Called(tag, ok)
}

func (m *MockResultReporter) SpeedResult(tag string, speed int64) {
	m.Called(tag, speed)
}

func (m *MockResultReporter) IPv6Result(tag string, ok bool) {
	m.Called(tag, ok)
}

func (m *MockResultReporter) PingResult(tag string, ping int) {
	m.Called(tag, ping)
}

func TestTester_TestUsable_Concurrent(t *testing.T) {
	// Setup
	mockReporter := new(MockResultReporter)
	tester := &Tester{
		ResultReporter: mockReporter,
		UsableTestFunc: func(ctx context.Context, h i.Outbound) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return true, nil
		},
	}

	// Create a mock handler
	handler := &MockNamedHandler{tag: "1"}

	// Mock the result reporter
	mockReporter.On("UsableResult", "1", true).Return()

	// Test concurrent calls
	var wg sync.WaitGroup
	results := make([]bool, 3)

	// Start multiple concurrent tests
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = tester.TestUsable(context.Background(), handler, true)
		}(i)
	}

	// Wait for all tests to complete
	wg.Wait()

	// Verify all results are the same
	firstResult := results[0]
	for i := 1; i < len(results); i++ {
		assert.Equal(t, firstResult, results[i], "All concurrent test results should be the same")
	}

	// Verify the result reporter was called exactly once
	mockReporter.AssertCalled(t, "UsableResult", "1", true)
	mockReporter.AssertNumberOfCalls(t, "UsableResult", 1)
}

func TestTester_TestUsable_Failure(t *testing.T) {
	// Setup
	mockReporter := new(MockResultReporter)
	tester := &Tester{
		ResultReporter: mockReporter,
		UsableTestFunc: func(ctx context.Context, h i.Outbound) (bool, error) {
			time.Sleep(100 * time.Millisecond)
			return false, nil
		},
	}

	// Create a mock handler
	handler := &MockNamedHandler{tag: "1"}

	// Mock the result reporter
	mockReporter.On("UsableResult", "1", false).Return()

	// Test concurrent calls
	var wg sync.WaitGroup
	results := make([]bool, 3)

	// Start multiple concurrent tests
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = tester.TestUsable(context.Background(), handler, true)
		}(i)
	}

	// Wait for all tests to complete
	wg.Wait()

	// Verify all results are the same
	firstResult := results[0]
	for i := 1; i < len(results); i++ {
		assert.Equal(t, firstResult, results[i], "All concurrent test results should be the same")
	}

	// Verify the result reporter was called exactly once
	mockReporter.AssertCalled(t, "UsableResult", "1", false)
	mockReporter.AssertNumberOfCalls(t, "UsableResult", 1)
}
