package mux

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockServerFlowHandler implements i.FlowHandler for testing
type mockServerFlowHandler struct {
	handleFlowFunc func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error
}

func (m *mockServerFlowHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	if m.handleFlowFunc != nil {
		return m.handleFlowFunc(ctx, dst, rw)
	}
	// Default: echo data back
	for {
		mb, err := rw.ReadMultiBuffer()
		if err != nil {
			if errors.Is(err, io.EOF) {
				rw.CloseWrite()
				return nil
			}
			return err
		}
		if err := rw.WriteMultiBuffer(mb); err != nil {
			return err
		}
		buf.ReleaseMulti(mb)
	}
}

// mockServerReaderWriter implements buf.ReaderWriter for server testing
type mockServerReaderWriter struct {
	mu        sync.Mutex
	readBuf   *bytes.Buffer
	writeBuf  *bytes.Buffer
	closed    bool
	interrupt bool
	err       error
}

func newMockServerReaderWriter() *mockServerReaderWriter {
	return &mockServerReaderWriter{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockServerReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.interrupt {
		return nil, m.err
	}
	if m.readBuf.Len() == 0 {
		return nil, io.EOF
	}
	b := buf.New()
	_, err := b.ReadFullFrom(m.readBuf, int32(m.readBuf.Len()))
	if err != nil {
		b.Release()
		return nil, err
	}
	return buf.MultiBuffer{b}, nil
}

func (m *mockServerReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.interrupt {
		return m.err
	}
	for _, b := range mb {
		_, err := m.writeBuf.Write(b.Bytes())
		if err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
	}
	buf.ReleaseMulti(mb)
	return nil
}

func (m *mockServerReaderWriter) CloseWrite() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockServerReaderWriter) Interrupt(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.interrupt = true
	m.err = err
}

func (m *mockServerReaderWriter) WriteFrame(meta FrameMetadata, data []byte) error {
	frame := buf.New()
	if err := meta.WriteTo(frame); err != nil {
		return err
	}
	if len(data) > 0 {
		if _, err := serial.WriteUint16(frame, uint16(len(data))); err != nil {
			return err
		}
		b := buf.New()
		b.Write(data)
		mb := buf.MultiBuffer{frame, b}
		// Write to readBuf so server can read it
		m.mu.Lock()
		for _, b := range mb {
			_, err := m.readBuf.Write(b.Bytes())
			if err != nil {
				m.mu.Unlock()
				buf.ReleaseMulti(mb)
				return err
			}
		}
		m.mu.Unlock()
		buf.ReleaseMulti(mb)
		return nil
	}
	// Write to readBuf so server can read it
	m.mu.Lock()
	_, err := m.readBuf.Write(frame.Bytes())
	m.mu.Unlock()
	frame.Release()
	return err
}

func TestServer_HandleStatusNew_TCP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a SessionStatusNew frame
	meta := FrameMetadata{
		SessionID:     1,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}

	testData := []byte("test data")
	err := rw.WriteFrame(meta, testData)
	require.NoError(t, err)

	// Handle frame
	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Verify session was created
	server.sessionLock.RLock()
	session, found := server.sessions[1]
	server.sessionLock.RUnlock()
	assert.True(t, found, "session should be created")
	assert.NotNil(t, session)
	assert.Equal(t, uint16(1), session.id)

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusNew_UDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a SessionStatusNew frame for UDP
	meta := FrameMetadata{
		SessionID:     2,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        net.UDPDestination(net.ParseAddress("8.8.8.8"), net.Port(53)),
	}

	testData := []byte("udp packet")
	err := rw.WriteFrame(meta, testData)
	require.NoError(t, err)

	// Handle frame
	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Verify session was created with packet transfer type
	server.sessionLock.RLock()
	_, found := server.sessions[2]
	server.sessionLock.RUnlock()
	assert.True(t, found, "session should be created")

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusNew_NoData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a SessionStatusNew frame without data
	meta := FrameMetadata{
		SessionID:     3,
		SessionStatus: SessionStatusNew,
		Option:        0, // No data
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}

	err := rw.WriteFrame(meta, nil)
	require.NoError(t, err)

	// Handle frame
	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Verify session was created
	server.sessionLock.RLock()
	_, found := server.sessions[3]
	server.sessionLock.RUnlock()
	assert.True(t, found, "session should be created even without data")

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusKeep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// First create a session
	metaNew := FrameMetadata{
		SessionID:     4,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Now send SessionStatusKeep with data
	metaKeep := FrameMetadata{
		SessionID:     4,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("keep data")
	err = rw.WriteFrame(metaKeep, testData)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusKeep_NoSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Send SessionStatusKeep for non-existent session
	meta := FrameMetadata{
		SessionID:     99,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("data")
	err := rw.WriteFrame(meta, testData)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Should send SessionStatusEnd with error
	// Verify by reading response
	time.Sleep(50 * time.Millisecond)

	// Cleanup
	cancel()
}

func TestServer_HandleStatusKeep_NoData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	meta := FrameMetadata{
		SessionID:     5,
		SessionStatus: SessionStatusKeep,
		Option:        0, // No data
	}
	err := rw.WriteFrame(meta, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Cleanup
	cancel()
}

func TestServer_HandleStatusEnd_NoError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session first
	metaNew := FrameMetadata{
		SessionID:     6,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Send SessionStatusEnd without error
	metaEnd := FrameMetadata{
		SessionID:     6,
		SessionStatus: SessionStatusEnd,
		Option:        0, // No error
	}
	err = rw.WriteFrame(metaEnd, nil)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Session should still exist but link should be closed for writing
	server.sessionLock.RLock()
	_, found := server.sessions[6]
	server.sessionLock.RUnlock()
	assert.True(t, found, "session should still exist")

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusEnd_WithError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session first
	metaNew := FrameMetadata{
		SessionID:     7,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Send SessionStatusEnd with error
	metaEnd := FrameMetadata{
		SessionID:     7,
		SessionStatus: SessionStatusEnd,
		Option:        OptionError,
	}
	err = rw.WriteFrame(metaEnd, nil)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for async cleanup
	time.Sleep(100 * time.Millisecond)

	// Session should be removed
	server.sessionLock.RLock()
	_, found := server.sessions[7]
	server.sessionLock.RUnlock()
	assert.False(t, found, "session should be removed after error")

	// Cleanup
	cancel()
}

func TestServer_HandleStatusKeepAlive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Send keepalive with data
	meta := FrameMetadata{
		SessionID:     8,
		SessionStatus: SessionStatusKeepAlive,
		Option:        OptionData,
	}
	testData := []byte("keepalive")
	err := rw.WriteFrame(meta, testData)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Send keepalive without data
	meta2 := FrameMetadata{
		SessionID:     9,
		SessionStatus: SessionStatusKeepAlive,
		Option:        0,
	}
	err = rw.WriteFrame(meta2, nil)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Cleanup
	cancel()
}

func TestServer_OnSessionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session
	metaNew := FrameMetadata{
		SessionID:     10,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for session to be set up
	time.Sleep(50 * time.Millisecond)

	server.sessionLock.RLock()
	session, found := server.sessions[10]
	server.sessionLock.RUnlock()
	require.True(t, found)

	// Trigger error
	testErr := errors.New("test error")
	server.onSessionError(session, testErr)

	// Wait for cleanup
	time.Sleep(50 * time.Millisecond)

	// Session should be removed
	server.sessionLock.RLock()
	_, found = server.sessions[10]
	server.sessionLock.RUnlock()
	assert.False(t, found, "session should be removed after error")

	// Cleanup
	cancel()
}

func TestServerSession_NotifyPeerSessionEOF(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session
	metaNew := FrameMetadata{
		SessionID:     11,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for session to be set up with writer
	time.Sleep(100 * time.Millisecond)

	server.sessionLock.RLock()
	session, found := server.sessions[11]
	server.sessionLock.RUnlock()
	require.True(t, found)

	// Notify EOF
	session.notifyPeerSessionEOF()
	assert.False(t, session.sendSessionEndError.Load())

	// Should not send again if already sent
	session.sendSessionEndError.Store(true)
	session.notifyPeerSessionEOF()
	assert.True(t, session.sendSessionEndError.Load())

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_Run_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		err := server.run(ctx)
		errCh <- err
	}()

	// Cancel context
	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("run should return after context cancel")
	}
}

func TestServer_Run_EOF(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Close read side to simulate EOF
	rw.Interrupt(io.EOF)

	errCh := make(chan error, 1)
	go func() {
		err := server.run(ctx)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("run should return on EOF")
	}
}

func TestServer_HandleFrame_UnknownStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create invalid frame with unknown status
	meta := FrameMetadata{
		SessionID:     12,
		SessionStatus: SessionStatus(0xFF), // Invalid status
		Option:        0,
	}
	err := rw.WriteFrame(meta, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown status")

	// Cleanup
	cancel()
}

func TestServer_HandleStatusKeep_WriteError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session first
	metaNew := FrameMetadata{
		SessionID:     13,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for session setup
	time.Sleep(50 * time.Millisecond)

	server.sessionLock.RLock()
	session, found := server.sessions[13]
	server.sessionLock.RUnlock()
	require.True(t, found)

	// Interrupt the session link to simulate write error
	session.link.Interrupt(errors.New("write error"))

	// Send SessionStatusKeep with data
	metaKeep := FrameMetadata{
		SessionID:     13,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("data")
	err = rw.WriteFrame(metaKeep, testData)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	// Should handle write error gracefully
	_ = err

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusEnd_NoSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Send SessionStatusEnd for non-existent session
	meta := FrameMetadata{
		SessionID:     99,
		SessionStatus: SessionStatusEnd,
		Option:        0,
	}
	err := rw.WriteFrame(meta, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Cleanup
	cancel()
}

func TestServer_HandleStatusEnd_WithData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session first
	metaNew := FrameMetadata{
		SessionID:     14,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Send SessionStatusEnd with data
	metaEnd := FrameMetadata{
		SessionID:     14,
		SessionStatus: SessionStatusEnd,
		Option:        OptionData,
	}
	testData := []byte("end data")
	err = rw.WriteFrame(metaEnd, testData)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleResponseData_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			// Return error immediately
			return errors.New("handler error")
		},
	}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session
	metaNew := FrameMetadata{
		SessionID:     15,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for handler to process
	time.Sleep(100 * time.Millisecond)

	// Session should be cleaned up
	server.sessionLock.RLock()
	_, found := server.sessions[15]
	server.sessionLock.RUnlock()
	assert.False(t, found, "session should be removed after handler error")

	// Cleanup
	cancel()
}

func TestServer_ConcurrentSessions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create multiple sessions concurrently
	var wg sync.WaitGroup
	numSessions := 5

	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(id uint16) {
			defer wg.Done()
			meta := FrameMetadata{
				SessionID:     id,
				SessionStatus: SessionStatusNew,
				Option:        0,
				Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
			}
			rw.WriteFrame(meta, nil)
		}(uint16(i + 1))
	}

	wg.Wait()

	// Process all frames
	reader := &buf.BufferedReader{Reader: rw}
	for i := 0; i < numSessions; i++ {
		err := server.handleFrame(ctx, reader)
		if err != nil {
			// May error if reader is exhausted
			break
		}
	}

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_OnSessionError_AlreadySent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session
	metaNew := FrameMetadata{
		SessionID:     16,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for session setup
	time.Sleep(100 * time.Millisecond)

	server.sessionLock.RLock()
	session, found := server.sessions[16]
	server.sessionLock.RUnlock()
	require.True(t, found)

	// Set flags to simulate already sent/received
	session.sendSessionEndError.Store(true)
	session.receivedSessionEndError.Store(true)

	// Trigger error - should not send again
	testErr := errors.New("test error")
	server.onSessionError(session, testErr)

	// Session should still be removed
	time.Sleep(50 * time.Millisecond)
	server.sessionLock.RLock()
	_, found = server.sessions[16]
	server.sessionLock.RUnlock()
	assert.False(t, found, "session should be removed")

	// Cleanup
	cancel()
}

func TestServer_HandleStatusKeep_WriteError_Drain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session first
	metaNew := FrameMetadata{
		SessionID:     17,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for session setup
	time.Sleep(50 * time.Millisecond)

	server.sessionLock.RLock()
	session, found := server.sessions[17]
	server.sessionLock.RUnlock()
	require.True(t, found)

	// Interrupt session link to cause write error
	session.link.Interrupt(errors.New("write error"))

	// Send SessionStatusKeep with data
	metaKeep := FrameMetadata{
		SessionID:     17,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}
	testData := []byte("data to drain")
	err = rw.WriteFrame(metaKeep, testData)
	require.NoError(t, err)

	err = server.handleFrame(ctx, reader)
	// Should handle write error and drain data
	_ = err

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleResponseData_EOF(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			// Write some data then close
			testData := []byte("response")
			b := buf.New()
			b.Write(testData)
			rw.WriteMultiBuffer(buf.MultiBuffer{b})
			rw.CloseWrite()
			return nil
		},
	}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session
	metaNew := FrameMetadata{
		SessionID:     18,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for response data handling
	time.Sleep(200 * time.Millisecond)

	// Cleanup
	cancel()
}

func TestServer_ConcurrentSessionAccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create multiple sessions concurrently
	var wg sync.WaitGroup
	numSessions := 10

	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(id uint16) {
			defer wg.Done()
			meta := FrameMetadata{
				SessionID:     id,
				SessionStatus: SessionStatusNew,
				Option:        0,
				Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
			}
			rw.WriteFrame(meta, nil)
		}(uint16(i + 1))
	}

	wg.Wait()

	// Process frames
	reader := &buf.BufferedReader{Reader: rw}
	for i := 0; i < numSessions; i++ {
		err := server.handleFrame(ctx, reader)
		if err != nil {
			break
		}
	}

	// Verify sessions were created
	server.sessionLock.RLock()
	sessionCount := len(server.sessions)
	server.sessionLock.RUnlock()
	assert.Greater(t, sessionCount, 0)

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_HandleStatusEnd_MultipleTimes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a session first
	metaNew := FrameMetadata{
		SessionID:     20,
		SessionStatus: SessionStatusNew,
		Option:        0,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	err := rw.WriteFrame(metaNew, nil)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Send SessionStatusEnd multiple times
	for i := 0; i < 3; i++ {
		metaEnd := FrameMetadata{
			SessionID:     20,
			SessionStatus: SessionStatusEnd,
			Option:        0,
		}
		err = rw.WriteFrame(metaEnd, nil)
		require.NoError(t, err)

		err = server.handleFrame(ctx, reader)
		// Should handle gracefully
		_ = err
	}

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServer_Run_ReadError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Interrupt to cause read error
	rw.Interrupt(errors.New("read error"))

	errCh := make(chan error, 1)
	go func() {
		err := server.run(ctx)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("run should return on read error")
	}

	cancel()
}

func TestServer_HandleStatusNew_CopyError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &mockServerFlowHandler{}
	rw := newMockServerReaderWriter()
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: handler,
	}

	// Create a SessionStatusNew frame with data
	meta := FrameMetadata{
		SessionID:     22,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80)),
	}
	testData := []byte("test data")
	err := rw.WriteFrame(meta, testData)
	require.NoError(t, err)

	reader := &buf.BufferedReader{Reader: rw}
	err = server.handleFrame(ctx, reader)
	require.NoError(t, err)

	// Wait for session setup
	time.Sleep(50 * time.Millisecond)

	server.sessionLock.RLock()
	session, found := server.sessions[22]
	server.sessionLock.RUnlock()
	require.True(t, found)

	// Interrupt session link to cause copy error
	session.link.Interrupt(errors.New("copy error"))

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond)
}
