package helper

import (
	"context"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"

	"github.com/stretchr/testify/require"
)

type mockPacketConn struct {
	readChan  chan *udp.Packet
	writeChan chan *udp.Packet
}

func newMockPacketConn() *mockPacketConn {
	return &mockPacketConn{
		readChan:  make(chan *udp.Packet, 16),
		writeChan: make(chan *udp.Packet, 16),
	}
}

func (m *mockPacketConn) ReadPacket() (*udp.Packet, error) {
	p := <-m.readChan
	return p, nil
}

func (m *mockPacketConn) WritePacket(p *udp.Packet) error {
	m.writeChan <- p
	return nil
}

func (m *mockPacketConn) Close() error {
	return nil
}

func TestRelayUDPPacketConn(t *testing.T) {
	left := newMockPacketConn()
	right := newMockPacketConn()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start relay
	errCh := make(chan error, 1)
	go func() {
		errCh <- RelayUDPPacketConn(ctx, left, right)
	}()

	// Test left to right relay
	addr1 := net.Destination{
		Address: net.IPAddress([]byte{1, 2, 3, 4}),
		Port:    1234,
	}
	data1 := []byte("test packet 1")
	buffer1 := buf.New()
	buffer1.Write(data1)
	packet1 := &udp.Packet{
		Target:  addr1,
		Payload: buffer1,
	}
	left.readChan <- packet1

	// Verify packet received on right side
	select {
	case p := <-right.writeChan:
		require.Equal(t, addr1, p.Target)
		require.Equal(t, data1, p.Payload.Bytes())
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for packet on right side")
	}

	// Test right to left relay
	addr2 := net.Destination{
		Address: net.IPAddress([]byte{5, 6, 7, 8}),
		Port:    5678,
	}
	data2 := []byte("test packet 2")
	buffer2 := buf.New()
	buffer2.Write(data2)
	packet2 := &udp.Packet{
		Source:  addr2,
		Payload: buffer2,
	}
	right.readChan <- packet2

	// Verify packet received on left side
	select {
	case p := <-left.writeChan:
		require.Equal(t, addr2, p.Source)
		require.Equal(t, data2, p.Payload.Bytes())
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for packet on left side")
	}

	// Test cleanup
	cancel()
	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for relay to stop")
	}
}
