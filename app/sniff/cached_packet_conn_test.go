package sniff

import (
	"crypto/rand"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/google/go-cmp/cmp"
)

// mockPacketConn is a mock implementation of PacketConn for testing
type mockPacketConn struct {
	readFunc  func() (*udp.Packet, error)
	writeFunc func(p *udp.Packet) error
	packets   []*udp.Packet
	readIndex int
	readDelay time.Duration
	mu        sync.Mutex
}

func newMockPacketConn() *mockPacketConn {
	return &mockPacketConn{
		packets: make([]*udp.Packet, 0),
	}
}

func (m *mockPacketConn) addPacket(data []byte, source, target net.Destination) {
	buffer := buf.NewWithSize(int32(len(data)))
	buffer.Write(data)
	packet := &udp.Packet{
		Payload: buffer,
		Source:  source,
		Target:  target,
	}
	m.mu.Lock()
	m.packets = append(m.packets, packet)
	m.mu.Unlock()
}

func (m *mockPacketConn) setReadDelay(delay time.Duration) {
	m.readDelay = delay
}

func (m *mockPacketConn) setReadFunc(f func() (*udp.Packet, error)) {
	m.readFunc = f
}

func (m *mockPacketConn) setWriteFunc(f func(p *udp.Packet) error) {
	m.writeFunc = f
}

func (m *mockPacketConn) Close() error {
	return nil
}

func (m *mockPacketConn) ReadPacket() (*udp.Packet, error) {
	if m.readFunc != nil {
		return m.readFunc()
	}

	if m.readDelay > 0 {
		time.Sleep(m.readDelay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readIndex >= len(m.packets) {
		return nil, errors.New("no more packets")
	}

	packet := m.packets[m.readIndex]
	m.readIndex++
	return packet.Clone(), nil
}

func (m *mockPacketConn) WritePacket(p *udp.Packet) error {
	if m.writeFunc != nil {
		return m.writeFunc(p)
	}
	// Default behavior: just release the packet
	p.Release()
	return nil
}

func TestCachedPacketConn_Cache_Success(t *testing.T) {
	mock := newMockPacketConn()
	testData := []byte("test packet data")
	mock.addPacket(testData, net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	buffer := make([]byte, 1024)
	copied, count, err := cached.read(buffer)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}

	if count != len(testData) {
		t.Errorf("Expected cache count to be %d, got %d", len(testData), count)
	}

	if string(buffer[:len(testData)]) != string(testData) {
		t.Errorf("Expected cached data to be %q, got %q", string(testData), string(buffer[:len(testData)]))
	}

	// Verify cache contains the packet
	if len(cached.cache) != 1 {
		t.Errorf("Expected cache to contain 1 packet, got %d", len(cached.cache))
	}
}

func TestCachedPacketConn_Cache_Timeout(t *testing.T) {
	mock := newMockPacketConn()
	// Set a long delay to trigger timeout
	mock.setReadDelay(200 * time.Millisecond)
	mock.addPacket([]byte("test"), net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	buffer := make([]byte, 1024)
	start := time.Now()
	copied, count, err := cached.read(buffer)
	if err != nil {
		t.Error(err)
	}
	if copied {
		t.Error("Expected data not to be copied")
	}
	elapsed := time.Since(start)

	if count != 0 {
		t.Errorf("Expected cache count to be 0 (timeout), got %d", count)
	}

	// Should timeout after ~100ms
	if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", elapsed)
	}
}

func TestCachedPacketConn_ReadPacket_FromCache(t *testing.T) {
	mock := newMockPacketConn()
	testData := []byte("cached packet")
	mock.addPacket(testData, net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	// First cache a packet
	buffer := make([]byte, 1024)
	cached.read(buffer)

	// Now read should return from cache
	packet, err := cached.ReadPacket()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(packet.Payload.Bytes()) != string(testData) {
		t.Errorf("Expected packet data %q, got %q", string(testData), string(packet.Payload.Bytes()))
	}

	// read should now be empty and hasReadAllCache should be true
	if !cached.hasReadAllCache {
		t.Error("Expected hasReadAllCache to be true after reading all cached packets")
	}

	if len(cached.cache) != 0 {
		t.Errorf("Expected cache to be empty after reading, got %d packets", len(cached.cache))
	}

	packet.Release()
}

func TestCachedPacketConn_ReadPacket_FromUnderlying(t *testing.T) {
	mock := newMockPacketConn()
	testData := []byte("direct packet")
	mock.addPacket(testData, net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
		hasReadAllCache:    true, // Simulate cache already exhausted
	}

	packet, err := cached.ReadPacket()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(packet.Payload.Bytes()) != string(testData) {
		t.Errorf("Expected packet data %q, got %q", string(testData), string(packet.Payload.Bytes()))
	}

	packet.Release()
}

func TestCachedPacketConn_ReadPacket_Error(t *testing.T) {
	mock := newMockPacketConn()
	expectedError := errors.New("read error")
	mock.setReadFunc(func() (*udp.Packet, error) {
		return nil, expectedError
	})

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
		hasReadAllCache:    true,
	}

	_, err := cached.ReadPacket()
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestCachedPacketConn_WritePacket(t *testing.T) {
	mock := newMockPacketConn()
	writeCallCount := 0
	mock.setWriteFunc(func(p *udp.Packet) error {
		writeCallCount++
		p.Release()
		return nil
	})

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	buffer := buf.NewWithSize(32)
	buffer.Write([]byte("test write"))
	packet := &udp.Packet{
		Payload: buffer,
		Source:  net.UDPDestination(net.LocalHostIP, 1234),
		Target:  net.UDPDestination(net.LocalHostIP, 5678),
	}

	err := cached.WritePacket(packet)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if writeCallCount != 1 {
		t.Errorf("Expected WritePacket to be called once, got %d calls", writeCallCount)
	}
}

func TestCachedPacketConn_Cache_MultipleCalls(t *testing.T) {
	mock := newMockPacketConn()

	// Add multiple packets
	for i := 0; i < 3; i++ {
		data := []byte("packet " + string(rune('A'+i)))
		mock.addPacket(data, net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))
	}

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	// First cache call should succeed and cache one packet
	buffer1 := make([]byte, 1024)
	copied, count1, err := cached.read(buffer1)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if count1 != 8 {
		t.Errorf("Expected first cache count to be 8, got %d", count1)
	}

	// Second cache call may return existing cache or read another packet
	// depending on timing and implementation behavior
	buffer2 := make([]byte, 1024)
	copied, count2, err := cached.read(buffer2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}

	// Could be 0 (timeout), 8 (existing cache), or 8+8 (new read + existing cache)
	if count2 < 0 || count2 > 16 {
		t.Errorf("Expected second cache count to be 0-16, got %d", count2)
	}
}

func TestCachedPacketConn_ReadInternal(t *testing.T) {
	mock := newMockPacketConn()

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	// Initially should return nil
	packet := cached.readInternal()
	if packet != nil {
		t.Error("Expected readInternal to return nil when cache is empty")
	}
	if !cached.hasReadAllCache {
		t.Error("Expected hasReadAllCache to be true when cache is empty")
	}

	// Add packets to cache manually
	buffer1 := buf.NewWithSize(32)
	buffer1.Write([]byte("packet1"))
	packet1 := &udp.Packet{Payload: buffer1}

	buffer2 := buf.NewWithSize(32)
	buffer2.Write([]byte("packet2"))
	packet2 := &udp.Packet{Payload: buffer2}

	cached.cache = []*udp.Packet{packet1, packet2}
	cached.hasReadAllCache = false

	// Read first packet
	readPacket1 := cached.readInternal()
	if readPacket1 != packet1 {
		t.Error("Expected to read first packet from cache")
	}
	if len(cached.cache) != 1 {
		t.Errorf("Expected cache length to be 1 after reading one packet, got %d", len(cached.cache))
	}
	if cached.hasReadAllCache {
		t.Error("Expected hasReadAllCache to be false when cache is not empty")
	}

	// Read second packet
	readPacket2 := cached.readInternal()
	if readPacket2 != packet2 {
		t.Error("Expected to read second packet from cache")
	}
	if len(cached.cache) != 0 {
		t.Errorf("Expected cache to be empty after reading all packets, got %d", len(cached.cache))
	}
	if !cached.hasReadAllCache {
		t.Error("Expected hasReadAllCache to be true after reading all cached packets")
	}

	// Clean up
	packet1.Release()
	packet2.Release()
}

func TestCachedPacketConn_ConcurrentReadPacket(t *testing.T) {
	mock := newMockPacketConn()
	mock.setReadDelay(50 * time.Millisecond)

	// Add some packets to mock
	for i := 0; i < 5; i++ {
		data := []byte("packet " + string(rune('0'+i)))
		mock.addPacket(data, net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))
	}

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	// First cache one packet
	buffer := make([]byte, 1024)
	cached.read(buffer)

	var wg sync.WaitGroup
	results := make([]*udp.Packet, 3)
	errors := make([]error, 3)

	// Start concurrent reads
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			packet, err := cached.ReadPacket()
			results[index] = packet
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All reads should succeed
	for i, err := range errors {
		if err != nil {
			t.Errorf("Read %d failed with error: %v", i, err)
		}
	}

	// Clean up
	for _, packet := range results {
		if packet != nil {
			packet.Release()
		}
	}
}

func TestCachedPacketConn_Cache_WithNilWaitCh(t *testing.T) {
	mock := newMockPacketConn()
	mock.addPacket([]byte("test"), net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
		waitCh:             nil, // Explicitly set to nil
	}

	buffer := make([]byte, 1024)
	copied, count, err := cached.read(buffer)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}

	// Even with nil waitCh, if TryLock() succeeds, a new waitCh is created
	// and the operation should complete successfully
	if count != 4 {
		t.Errorf("Expected cache count to be 1 (new waitCh created), got %d", count)
	}

	if len(cached.cache) != 1 {
		t.Errorf("Expected cache to contain 1 packet, got %d", len(cached.cache))
	}
}

func TestCachedPacketConn_Cache_ReadError(t *testing.T) {
	mock := newMockPacketConn()
	expectedError := errors.New("read error")
	mock.setReadFunc(func() (*udp.Packet, error) {
		return nil, expectedError
	})

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	buffer := make([]byte, 1024)
	copied, count, err := cached.read(buffer)
	if err == nil {
		t.Error(err)
	}
	if copied {
		t.Error("Expected data not to be copied")
	}

	// When read fails, cache should remain empty
	if count != 0 {
		t.Errorf("Expected cache count to be 0 when read fails, got %d", count)
	}

	if len(cached.cache) != 0 {
		t.Errorf("Expected cache to be empty when read fails, got %d packets", len(cached.cache))
	}
}

func TestCachedPacketConn_Cache_WaitOnNilChannel(t *testing.T) {
	mock := newMockPacketConn()
	mock.setReadDelay(200 * time.Millisecond)
	mock.addPacket([]byte("test"), net.UDPDestination(net.LocalHostIP, 1234), net.UDPDestination(net.LocalHostIP, 5678))

	cached := &CachedPacketConn{
		PacketReaderWriter: mock,
	}

	// Manually acquire the lock to simulate another goroutine holding it
	cached.readLock.Lock()
	defer cached.readLock.Unlock()

	// Now when read() is called, TryLock() will fail but waitCh is nil
	buffer := make([]byte, 1024)
	start := time.Now()
	copied, count, err := cached.read(buffer)
	if err != nil {
		t.Error(err)
	}
	if copied {
		t.Error("Expected data not to be copied")
	}
	elapsed := time.Since(start)

	// Should timeout since waitCh is nil and can't wait
	if count != 0 {
		t.Errorf("Expected cache count to be 0 (timeout on nil waitCh), got %d", count)
	}

	// Should timeout after ~100ms
	if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", elapsed)
	}
}

func TestCachedUdpLink(t *testing.T) {
	iLink, oLink := udp.NewLink(10)
	r := &CachedPacketConn{
		PacketReaderWriter: oLink,
	}

	b1 := make([]byte, 100)
	rand.Read(b1)

	iLink.WritePacket(&udp.Packet{
		Payload: buf.FromBytes(b1),
		Source:  net.UDPDestination(net.LocalHostIP, 1234),
		Target:  net.UDPDestination(net.LocalHostIP, 5678),
	})

	b2 := make([]byte, 8192)
	copied, n, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if d := cmp.Diff(b1, b2[:n]); d != "" {
		t.Error(d)
	}

	mb, err := r.ReadPacket()
	if err != nil {
		t.Error(err)
	}
	if d := cmp.Diff(b1, mb.Payload.Bytes()); d != "" {
		t.Error(d)
	}

	for i := 0; i < 5; i++ {
		rand.Read(b1)
		iLink.WritePacket(&udp.Packet{
			Payload: buf.FromBytes(b1),
			Source:  net.UDPDestination(net.LocalHostIP, 1234),
			Target:  net.UDPDestination(net.LocalHostIP, 5678),
		})
		mb, err := r.ReadPacket()
		if err != nil {
			t.Error(err)
		}
		if d := cmp.Diff(b1, mb.Payload.Bytes()); d != "" {
			t.Error(d)
		}
	}
}

func TestCachedPacketConnBasic(t *testing.T) {
	iLink, oLink := udp.NewLink(10)
	r := &CachedPacketConn{
		PacketReaderWriter: oLink,
	}

	b1 := make([]byte, 100)
	rand.Read(b1)

	iLink.WritePacket(&udp.Packet{
		Payload: buf.FromBytes(b1),
		Source:  net.UDPDestination(net.LocalHostIP, 1234),
		Target:  net.UDPDestination(net.LocalHostIP, 5678),
	})

	b2 := make([]byte, 8192)
	copied, n, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if d := cmp.Diff(b1, b2[:n]); d != "" {
		t.Error(d)
	}

	mb, err := r.ReadPacket()
	if err != nil {
		t.Error(err)
	}
	if d := cmp.Diff(b1, mb.Payload.Bytes()); d != "" {
		t.Error(d)
	}
}

func TestCachedPacketConnMultipleReads(t *testing.T) {
	iLink, oLink := udp.NewLink(10)
	r := &CachedPacketConn{
		PacketReaderWriter: oLink,
	}

	for i := 0; i < 5; i++ {
		testData := make([]byte, 100)
		rand.Read(testData)
		iLink.WritePacket(&udp.Packet{
			Payload: buf.FromBytes(testData),
			Source:  net.UDPDestination(net.LocalHostIP, 1234),
			Target:  net.UDPDestination(net.LocalHostIP, 5678),
		})
		mb, err := r.ReadPacket()
		if err != nil {
			t.Fatal(err)
		}
		if d := cmp.Diff(testData, mb.Payload.Bytes()); d != "" {
			t.Error(d)
		}
	}
}

func TestCachedPacketConnCacheTwice(t *testing.T) {
	iLink, oLink := udp.NewLink(10)
	r := &CachedPacketConn{
		PacketReaderWriter: oLink,
	}

	b1 := make([]byte, 100)
	rand.Read(b1)

	iLink.WritePacket(&udp.Packet{
		Payload: buf.FromBytes(b1),
		Source:  net.UDPDestination(net.LocalHostIP, 1234),
		Target:  net.UDPDestination(net.LocalHostIP, 5678),
	})

	b2 := make([]byte, 8192)
	copied, n1, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if !copied {
		t.Error("Expected data to be copied")
	}
	if d := cmp.Diff(b1, b2[:n1]); d != "" {
		t.Error(d)
	}

	copied, n2, err := r.read(b2)
	if err != nil {
		t.Error(err)
	}
	if copied {
		t.Error("Expected data not to be copied")
	}
	if n2 != 0 {
		t.Errorf("Expected second cache call to return 100, got %d", n2)
	}

	iLink.Close()
	mb, err := r.ReadPacket()
	if err != nil {
		t.Error(err)
	}
	if d := cmp.Diff(b1, mb.Payload.Bytes()); d != "" {
		t.Error(d)
	}
}
