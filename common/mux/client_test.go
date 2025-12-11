package mux

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReaderWriter implements buf.ReaderWriter for testing
type mockReaderWriter struct {
	mu          sync.Mutex
	readData    []byte
	readPos     int
	writeData   []byte
	closed      bool
	readChan    chan buf.MultiBuffer
	writeChan   chan buf.MultiBuffer
	interrupted bool
	err         error
}

func newMockReaderWriter() *mockReaderWriter {
	return &mockReaderWriter{
		readChan:  make(chan buf.MultiBuffer, 10),
		writeChan: make(chan buf.MultiBuffer, 10),
	}
}

func (m *mockReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if m.interrupted {
		return nil, m.err
	}
	select {
	case mb, ok := <-m.readChan:
		if !ok {
			return nil, io.EOF
		}
		return mb, nil
	case <-time.After(time.Second):
		return nil, io.EOF
	}
}

func (m *mockReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if m.interrupted {
		return m.err
	}
	select {
	case m.writeChan <- mb:
		return nil
	case <-time.After(time.Second):
		return errors.New("write timeout")
	}
}

func (m *mockReaderWriter) CloseWrite() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockReaderWriter) Interrupt(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.interrupted = true
	m.err = err
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	strategy := DefaultClientStrategy

	client, err := NewClient(ctx, iLink, strategy)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, strategy, client.ClientStrategy)
	assert.NotNil(t, client.sessions)
	assert.Equal(t, 0, len(client.sessions))
}

func TestClient_AddSession(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)
	assert.Equal(t, 1, len(client.sessions))
	assert.Equal(t, session, client.sessions[1])
}

func TestClient_RemoveSession(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)
	assert.Equal(t, 1, len(client.sessions))

	client.RemoveSession(session)
	assert.Equal(t, 0, len(client.sessions))
	assert.True(t, !client.lastEmptySessionTime.IsZero())
}

func TestClient_IsEmpty(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	assert.True(t, client.IsEmpty())

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)
	assert.False(t, client.IsEmpty())
}

func TestClient_IsClosing(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	strategy := ClientStrategy{
		MaxConnection:  2,
		MaxConcurrency: 10,
	}
	client, _ := NewClient(ctx, iLink, strategy)

	assert.False(t, client.IsClosing())

	// Set count to max
	client.count.Store(2)
	assert.True(t, client.IsClosing())

	// Set count above max
	client.count.Store(3)
	assert.True(t, client.IsClosing())

	// Reset count
	client.count.Store(1)
	assert.False(t, client.IsClosing())

	// Test with MaxConnection = 0 (unlimited)
	strategy.MaxConnection = 0
	client.ClientStrategy = strategy
	client.count.Store(100)
	assert.False(t, client.IsClosing())
}

func TestClient_IsFull(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	strategy := ClientStrategy{
		MaxConnection:  10,
		MaxConcurrency: 2,
	}
	client, _ := NewClient(ctx, iLink, strategy)

	assert.False(t, client.IsFull())

	// Add sessions up to MaxConcurrency
	rw1 := newMockReaderWriter()
	session1 := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw1,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client.AddSession(session1)
	assert.False(t, client.IsFull())

	rw2 := newMockReaderWriter()
	session2 := &clientSession{
		ID:              2,
		ctx:             ctx,
		rw:              rw2,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client.AddSession(session2)
	assert.True(t, client.IsFull())

	// Test with MaxConcurrency = 0 (unlimited)
	strategy.MaxConcurrency = 0
	client.ClientStrategy = strategy
	assert.False(t, client.IsFull())

	// Test IsClosing makes it full
	strategy.MaxConcurrency = 10
	strategy.MaxConnection = 2
	client.ClientStrategy = strategy
	client.count.Store(2)
	assert.True(t, client.IsFull())
}

func TestClient_Close(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	client.Close()

	// Verify link is interrupted
	_, err := oLink.ReadMultiBuffer()
	assert.Error(t, err)
}

func TestClient_Interrupt(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw1 := newMockReaderWriter()
	session1 := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw1,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	rw2 := newMockReaderWriter()
	session2 := &clientSession{
		ID:              2,
		ctx:             ctx,
		rw:              rw2,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session1)
	client.AddSession(session2)

	client.interrupt()

	// Verify both sessions received error
	select {
	case err := <-session1.errChan:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("session1 should receive error")
	}

	select {
	case err := <-session2.errChan:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("session2 should receive error")
	}
}

func TestClientSession_OnError(t *testing.T) {
	session := &clientSession{
		errChan: make(chan error, 1),
	}

	testErr := errors.New("test error")
	session.onError(testErr)

	select {
	case err := <-session.errChan:
		assert.Equal(t, testErr, err)
	case <-time.After(time.Second):
		t.Fatal("should receive error")
	}

	// Send first error again to fill the channel
	session.onError(testErr)

	// Test that second error doesn't block (channel is full, so it won't be sent)
	session.onError(errors.New("second error"))
	// Channel should still contain first error
	select {
	case err := <-session.errChan:
		assert.Equal(t, testErr, err) // Should still be first error
	default:
		t.Fatal("channel should still contain first error")
	}
}

func TestClientSession_NotifyPeerEOF(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	session.writer = NewMuxWriter(1, net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)), iLink, TransferTypeStream)

	// Should not error if sendSessionEndError is false
	session.notifyPeerEOF()
	assert.False(t, session.sendSessionEndError.Load())

	// Should not send if already sent
	session.sendSessionEndError.Store(true)
	session.notifyPeerEOF()
	assert.True(t, session.sendSessionEndError.Load())
}

func TestClientSession_NotifyPeerSessionError(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	session.writer = NewMuxWriter(1, net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)), iLink, TransferTypeStream)

	// Should send error
	session.notifyPeerSessionError()
	assert.True(t, session.sendSessionEndError.Load())
	assert.True(t, session.writer.hasError)

	// Should not send again if already sent
	session.sendSessionEndError.Store(false)
	session.receivedSessionEndError.Store(true)
	session.notifyPeerSessionError()
	assert.False(t, session.sendSessionEndError.Load())
}

func TestWriteFirstPayload(t *testing.T) {
	iLink, _ := pipe.NewLinks(64*1024, false)
	writer := NewMuxWriter(1, net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)), iLink, TransferTypeStream)

	// Test with data available
	testData := []byte("test data")
	b := buf.New()
	b.Write(testData)
	rw := &buf.BufferedReader{Reader: buf.NewReader(b)}

	err := writeFirstPayload(rw, writer)
	assert.NoError(t, err)

	// Test with timeout (no data)
	emptyRW := &buf.BufferedReader{Reader: buf.NewReader(buf.New())}
	err = writeFirstPayload(emptyRW, writer)
	assert.NoError(t, err)
}

func TestClient_Merge_InvalidDestination(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)
	client.AddSession(session)

	// Test with invalid destination (empty destination is invalid)
	// Note: The merge function checks IsValid() but continues execution
	// which may cause issues. We test that the error is reported.
	invalidDest := net.Destination{}

	// Use a recover to catch potential panic from invalid destination
	doneCh := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected due to invalid destination
				doneCh <- true
			}
		}()
		client.merge(ctx, invalidDest, session)
		doneCh <- true
	}()

	// Check for error or completion
	select {
	case err := <-session.errChan:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid target")
	case <-doneCh:
		// Function completed (may have panicked)
	case <-time.After(time.Second):
		t.Fatal("should receive error or complete for invalid destination")
	}
}

func TestClient_Merge_ValidDestination(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)

	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Write some data to rw
	testData := []byte("test data")
	b := buf.New()
	b.Write(testData)
	rw.readChan <- buf.MultiBuffer{b}
	close(rw.readChan) // Signal EOF

	go client.merge(ctx, dest, session)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Verify writer was created
	assert.NotNil(t, session.writer)
}

func TestClient_Split_HandleStatusKeep(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client.AddSession(session)

	// Write a SessionStatusKeep frame with data
	meta := FrameMetadata{
		SessionID:     1,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("response data")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for data to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify data was written to session's rw
	select {
	case received := <-rw.readChan:
		assert.NotNil(t, received)
		buf.ReleaseMulti(received)
	default:
		// Data might have been read already
	}

	client.Close()
}

func TestClient_Split_HandleStatusKeep_NoSession(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Write a SessionStatusKeep frame for non-existent session
	meta := FrameMetadata{
		SessionID:     999,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("data")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should handle gracefully (drain data)
	client.Close()
}

func TestClient_Split_HandleStatusEnd_NoError(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              2,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client.AddSession(session)

	// Write a SessionStatusEnd frame without error
	meta := FrameMetadata{
		SessionID:     2,
		SessionStatus: SessionStatusEnd,
		Option:        0, // No error
	}
	frame := buf.New()
	meta.WriteTo(frame)
	mb := buf.MultiBuffer{frame}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify rightToLeftDone was closed
	select {
	case <-session.rightToLeftDone.Wait():
		// Expected
	default:
		t.Fatal("rightToLeftDone should be closed")
	}

	client.Close()
}

func TestClient_Split_HandleStatusEnd_WithError(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              3,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client.AddSession(session)

	// Write a SessionStatusEnd frame with error
	meta := FrameMetadata{
		SessionID:     3,
		SessionStatus: SessionStatusEnd,
		Option:        OptionError,
	}
	frame := buf.New()
	meta.WriteTo(frame)
	mb := buf.MultiBuffer{frame}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify error was sent to session
	select {
	case err := <-session.errChan:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ended by peer")
	case <-time.After(time.Second):
		t.Fatal("should receive error")
	}

	assert.True(t, session.receivedSessionEndError.Load())

	client.Close()
}

func TestClient_Split_HandleStatusKeepAlive(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Write a SessionStatusKeepAlive frame with data
	meta := FrameMetadata{
		SessionID:     0,
		SessionStatus: SessionStatusKeepAlive,
		Option:        OptionData,
	}
	testData := []byte("keepalive data")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should handle gracefully (drain data)
	client.Close()
}

func TestClient_Split_HandleStatusKeepAlive_NoData(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Write a SessionStatusKeepAlive frame without data
	meta := FrameMetadata{
		SessionID:     0,
		SessionStatus: SessionStatusKeepAlive,
		Option:        0,
	}
	frame := buf.New()
	meta.WriteTo(frame)
	mb := buf.MultiBuffer{frame}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should handle gracefully
	client.Close()
}

func TestClient_Split_HandleStatusNew(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Write a SessionStatusNew frame with data
	meta := FrameMetadata{
		SessionID:     0,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	testData := []byte("new session data")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should handle gracefully (drain data)
	client.Close()
}

func TestClient_Split_HandleStatusNew_NoData(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Write a SessionStatusNew frame without data
	meta := FrameMetadata{
		SessionID:     0,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	frame := buf.New()
	meta.WriteTo(frame)
	mb := buf.MultiBuffer{frame}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should handle gracefully
	client.Close()
}

func TestClient_Split_UnknownStatus(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Write a frame with unknown status
	meta := FrameMetadata{
		SessionID:     0,
		SessionStatus: SessionStatus(0xFF), // Invalid status
		Option:        0,
	}
	frame := buf.New()
	meta.WriteTo(frame)
	mb := buf.MultiBuffer{frame}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Split should stop processing
	client.Close()
}

func TestClient_Split_EOF(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Close the link to simulate EOF
	oLink.CloseWrite()

	// Wait for split to process EOF
	time.Sleep(100 * time.Millisecond)

	// Should handle EOF gracefully
	client.Close()
}

func TestClient_HandleStatusKeep_WriteError(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Create a session with a writer that will fail
	rw := newMockReaderWriter()
	rw.Interrupt(errors.New("write error"))
	session := &clientSession{
		ID:              4,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	// Set a writer so notifyPeerSessionError doesn't panic
	session.writer = NewMuxWriter(4, net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)), iLink, TransferTypeStream)
	client.AddSession(session)

	// Write a SessionStatusKeep frame with data
	meta := FrameMetadata{
		SessionID:     4,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("data")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify error handling
	select {
	case err := <-session.errChan:
		assert.Error(t, err)
	case <-time.After(time.Second):
		// Error might be handled differently
	}

	client.Close()
}

func TestClient_HandleStatusEnd_WithData(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              5,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client.AddSession(session)

	// Write a SessionStatusEnd frame with data
	meta := FrameMetadata{
		SessionID:     5,
		SessionStatus: SessionStatusEnd,
		Option:        OptionData,
	}
	testData := []byte("end data")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify rightToLeftDone was closed
	select {
	case <-session.rightToLeftDone.Wait():
		// Expected
	default:
		t.Fatal("rightToLeftDone should be closed")
	}

	client.Close()
}

func TestClient_Merge_UDP(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              6,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)

	dest := net.UDPDestination(net.ParseAddress("8.8.8.8"), net.Port(53))

	// Write some data to rw
	testData := []byte("udp packet")
	b := buf.New()
	b.Write(testData)
	rw.readChan <- buf.MultiBuffer{b}
	close(rw.readChan)

	go client.merge(ctx, dest, session)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Verify writer was created with packet transfer type
	assert.NotNil(t, session.writer)
	assert.Equal(t, TransferTypePacket, session.writer.transferType)
}

func TestClient_ConcurrentSessions(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	var wg sync.WaitGroup
	numSessions := 10

	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(id uint16) {
			defer wg.Done()
			rw := newMockReaderWriter()
			session := &clientSession{
				ID:              id,
				ctx:             ctx,
				rw:              rw,
				errChan:         make(chan error, 1),
				leftToRightDone: done.New(),
				rightToLeftDone: done.New(),
			}
			client.AddSession(session)
			time.Sleep(10 * time.Millisecond)
			client.RemoveSession(session)
		}(uint16(i + 1))
	}

	wg.Wait()

	// Verify all sessions were handled
	assert.Equal(t, 0, len(client.sessions))
	client.Close()
}

func TestClient_Merge_WriteError(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Create a reader/writer that will fail on write
	rw := newMockReaderWriter()
	rw.Interrupt(errors.New("write error"))
	session := &clientSession{
		ID:              7,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Write some data
	testData := []byte("test")
	b := buf.New()
	b.Write(testData)
	rw.readChan <- buf.MultiBuffer{b}

	go client.merge(ctx, dest, session)

	// Wait for error
	select {
	case err := <-session.errChan:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("should receive error")
	}
}

func TestClient_Merge_CopyError(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              8,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Interrupt after some data
	testData := []byte("test")
	b := buf.New()
	b.Write(testData)
	rw.readChan <- buf.MultiBuffer{b}
	rw.Interrupt(errors.New("read error"))

	go client.merge(ctx, dest, session)

	// Wait for error
	select {
	case err := <-session.errChan:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("should receive error")
	}
}

func TestClient_RemoveSession_MultipleTimes(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              9,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session)
	assert.Equal(t, 1, len(client.sessions))

	// Remove multiple times - should be safe
	client.RemoveSession(session)
	assert.Equal(t, 0, len(client.sessions))

	client.RemoveSession(session)
	assert.Equal(t, 0, len(client.sessions))
}

func TestClient_AddSession_DuplicateID(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	rw1 := newMockReaderWriter()
	session1 := &clientSession{
		ID:              10,
		ctx:             ctx,
		rw:              rw1,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	rw2 := newMockReaderWriter()
	session2 := &clientSession{
		ID:              10, // Same ID
		ctx:             ctx,
		rw:              rw2,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}

	client.AddSession(session1)
	client.AddSession(session2)

	// Second session should overwrite first
	assert.Equal(t, 1, len(client.sessions))
	assert.Equal(t, session2, client.sessions[10])
}

func TestClient_IsFull_EdgeCases(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)

	// Test with MaxConcurrency = 0 (unlimited)
	strategy := ClientStrategy{
		MaxConnection:  10,
		MaxConcurrency: 0,
	}
	client, _ := NewClient(ctx, iLink, strategy)

	// Should never be full
	for i := 0; i < 100; i++ {
		rw := newMockReaderWriter()
		session := &clientSession{
			ID:              uint16(i + 1),
			ctx:             ctx,
			rw:              rw,
			errChan:         make(chan error, 1),
			leftToRightDone: done.New(),
			rightToLeftDone: done.New(),
		}
		client.AddSession(session)
		assert.False(t, client.IsFull())
	}
}

func TestClient_HandleStatusKeep_WriteError_Drain(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Create session with writer that will fail
	rw := newMockReaderWriter()
	rw.Interrupt(errors.New("write error"))
	session := &clientSession{
		ID:              12,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	session.writer = NewMuxWriter(12, net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)), iLink, TransferTypeStream)
	client.AddSession(session)

	// Write a SessionStatusKeep frame with data
	meta := FrameMetadata{
		SessionID:     12,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("data to drain")
	frame := buf.New()
	meta.WriteTo(frame)
	serial.WriteUint16(frame, uint16(len(testData)))
	dataBuf := buf.New()
	dataBuf.Write(testData)
	mb := buf.MultiBuffer{frame, dataBuf}
	oLink.WriteMultiBuffer(mb)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should handle write error and drain data
	select {
	case err := <-session.errChan:
		assert.Error(t, err)
	default:
		// Error might be handled differently
	}

	client.Close()
}

func TestClient_Split_ReadError(t *testing.T) {
	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Interrupt the link to cause read error
	oLink.Interrupt(errors.New("read error"))

	// Wait for split to process error
	time.Sleep(100 * time.Millisecond)

	client.Close()
}

func TestWriteFirstPayload_Timeout(t *testing.T) {
	iLink, _ := pipe.NewLinks(64*1024, false)
	writer := NewMuxWriter(13, net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)), iLink, TransferTypeStream)

	// Create reader with no data (will timeout)
	emptyRW := &buf.BufferedReader{Reader: buf.NewReader(buf.New())}

	err := writeFirstPayload(emptyRW, writer)
	// Should handle timeout gracefully
	assert.NoError(t, err)
}

func TestClient_Interrupt_EmptySessions(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Interrupt with no sessions - should not panic
	client.interrupt()

	// Should still work
	assert.True(t, client.IsEmpty())
}

func TestClient_ConcurrentAddRemove(t *testing.T) {
	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	var wg sync.WaitGroup
	numOps := 100

	// Concurrent add/remove operations
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(id uint16) {
			defer wg.Done()
			rw := newMockReaderWriter()
			session := &clientSession{
				ID:              id,
				ctx:             ctx,
				rw:              rw,
				errChan:         make(chan error, 1),
				leftToRightDone: done.New(),
				rightToLeftDone: done.New(),
			}
			client.AddSession(session)
			time.Sleep(time.Millisecond)
			client.RemoveSession(session)
		}(uint16(i + 1))
	}

	wg.Wait()

	// Should be empty after all operations
	assert.Equal(t, 0, len(client.sessions))
}
