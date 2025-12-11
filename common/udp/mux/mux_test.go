package udpmux

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"

	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/uuid"

	"github.com/stretchr/testify/require"
)

// mockPacketConn implements net.MyPacketConn for testing
type mockPacketConn struct {
	readChan  chan readResult
	writeChan chan writeResult
	closed    bool
	mu        sync.Mutex
}

type readResult struct {
	data []byte
	addr net.Addr
	err  error
}

type writeResult struct {
	data []byte
	addr net.Addr
}

func newMockPacketConn() *mockPacketConn {
	return &mockPacketConn{
		readChan:  make(chan readResult, 16),
		writeChan: make(chan writeResult, 16),
	}
}

func (m *mockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	result := <-m.readChan
	if result.err != nil {
		return 0, nil, result.err
	}
	n = copy(p, result.data)
	return n, result.addr, nil
}

func (m *mockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return 0, net.ErrClosed
	}
	m.mu.Unlock()

	data := make([]byte, len(p))
	copy(data, p)
	m.writeChan <- writeResult{data, addr}
	return len(p), nil
}

func (m *mockPacketConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// mockDNS implements i.DNS for testing
type mockDNS struct{}

func (m *mockDNS) LookupIPSpeed(ctx context.Context, domain string) ([]net.IP, error) {
	return []net.IP{net.IPv4(1, 2, 3, 4)}, nil
}

func (m *mockDNS) LookupIP(ctx context.Context, domain string) ([]net.IP, error) {
	return []net.IP{net.IPv4(1, 2, 3, 4)}, nil
}

func (m *mockDNS) LookupIPv4(ctx context.Context, domain string) ([]net.IP, error) {
	return []net.IP{net.IPv4(1, 2, 3, 4)}, nil
}

func (m *mockDNS) LookupIPv6(ctx context.Context, domain string) ([]net.IP, error) {
	return []net.IP{}, nil
}

func (m *mockDNS) LookupIPPrefer4(ctx context.Context, domain string) ([]net.IP, error) {
	return []net.IP{net.IPv4(1, 2, 3, 4)}, nil
}

// mockReaderWriter implements buf.ReaderWriter for testing
type mockReaderWriter struct {
	readChan  chan buf.MultiBuffer
	writeChan chan buf.MultiBuffer
	closed    bool
	mu        sync.Mutex
}

func newMockReaderWriter() *mockReaderWriter {
	return &mockReaderWriter{
		readChan:  make(chan buf.MultiBuffer, 16),
		writeChan: make(chan buf.MultiBuffer, 16),
	}
}

func (m *mockReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb := <-m.readChan
	return mb, nil
}

func (m *mockReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return net.ErrClosed
	}
	m.mu.Unlock()
	m.writeChan <- mb
	return nil
}

func (m *mockReaderWriter) CloseWrite() error {
	return nil
}

func TestMuxerSingleFlow(t *testing.T) {
	muxer := NewMuxer(&mockDNS{})
	mockConn := newMockPacketConn()
	mockRW := newMockReaderWriter()

	// Test data
	sourceID := uuid.New()
	destAddr := mynet.AddressPort{
		Address: mynet.IPAddress([]byte{8, 8, 8, 8}),
		Port:    53,
	}

	// Start handling the flow
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dialCalled := false
	errCh := make(chan error, 1)
	go func() {
		err := muxer.Handle(ctx, sourceID, destAddr, mockRW, func(ctx context.Context) (MyPacketConn, error) {
			dialCalled = true
			return mockConn, nil
		})
		errCh <- err
	}()

	// Wait for dialer to be called
	time.Sleep(100 * time.Millisecond)
	require.True(t, dialCalled, "Dialer should be called")

	// Test outbound flow (client -> destination)
	testData := []byte("test packet")
	buffer := buf.New()
	buffer.Write(testData)
	mockRW.readChan <- buf.MultiBuffer{buffer}

	// Verify packet was sent to destination
	select {
	case result := <-mockConn.writeChan:
		require.Equal(t, testData, result.data)
		require.Equal(t, &net.UDPAddr{IP: net.IP{8, 8, 8, 8}, Port: 53}, result.addr)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for packet to be sent")
	}

	// Test inbound flow (destination -> client)
	responseData := []byte("response packet")
	mockConn.readChan <- readResult{
		data: responseData,
		addr: &net.UDPAddr{IP: net.IP{8, 8, 8, 8}, Port: 53},
	}

	// Verify response was forwarded to client
	select {
	case mb := <-mockRW.writeChan:
		require.Equal(t, responseData, mb[0].Bytes())
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for response")
	}

	// Test cleanup
	cancel()
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for handler to stop")
	}

	if len(muxer.sessions) != 0 {
		t.Fatal("Num of sessions should be 0 now")
	}
}

func TestMuxerMultipleFlows(t *testing.T) {
	muxer := NewMuxer(&mockDNS{})
	mockConn := newMockPacketConn()

	sourceID := uuid.New()
	destinations := []mynet.AddressPort{
		{
			Address: mynet.IPAddress([]byte{8, 8, 8, 8}),
			Port:    53,
		},
		{
			Address: mynet.IPAddress([]byte{1, 1, 1, 1}),
			Port:    53,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var rwChannels []*mockReaderWriter
	var errChannels []chan error

	// Start multiple flows
	for _, dest := range destinations {
		mockRW := newMockReaderWriter()
		rwChannels = append(rwChannels, mockRW)
		errCh := make(chan error, 1)
		errChannels = append(errChannels, errCh)

		go func(dest mynet.AddressPort, rw *mockReaderWriter, errCh chan error) {
			err := muxer.Handle(ctx, sourceID, dest, rw, func(ctx context.Context) (MyPacketConn, error) {
				return mockConn, nil
			})
			errCh <- err
		}(dest, mockRW, errCh)
	}

	time.Sleep(100 * time.Millisecond)

	// Test sending data from each flow
	for i, rw := range rwChannels {
		testData := []byte(fmt.Sprintf("%d: test packet", i))
		buffer := buf.New()
		buffer.Write(testData)
		rw.readChan <- buf.MultiBuffer{buffer}

		select {
		case result := <-mockConn.writeChan:
			require.Equal(t, testData, result.data)
			require.Equal(t, &net.UDPAddr{IP: destinations[i].Address.IP(), Port: int(destinations[i].Port)}, result.addr)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for packet to be sent")
		}
	}

	// Test receiving responses for each flow
	for i, rw := range rwChannels {
		responseData := []byte(fmt.Sprintf("%d: response", i))
		mockConn.readChan <- readResult{
			data: responseData,
			addr: &net.UDPAddr{IP: destinations[i].Address.IP(), Port: int(destinations[i].Port)},
		}

		select {
		case mb := <-rw.writeChan:
			require.Equal(t, responseData, mb[0].Bytes())
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for response")
		}
	}

	// Test cleanup
	cancel()
	for _, errCh := range errChannels {
		select {
		case err := <-errCh:
			require.ErrorIs(t, err, context.Canceled)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for handler to stop")
		}
	}

	if len(muxer.sessions) != 0 {
		t.Fatal("Num of sessions should be 0 now")
	}
}
