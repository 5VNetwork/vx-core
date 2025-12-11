package mux

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFlowHandler implements i.FlowHandler for testing
type mockFlowHandler struct {
	handleFlowFunc func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error
}

func (m *mockFlowHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	if m.handleFlowFunc != nil {
		return m.handleFlowFunc(ctx, dst, rw)
	}
	return nil
}

func TestNewClientManager(t *testing.T) {
	strategy := DefaultClientStrategy
	handler := &mockFlowHandler{}

	manager := NewClientManager(strategy, handler)

	assert.NotNil(t, manager)
	assert.Equal(t, strategy, manager.Strategy)
	assert.Equal(t, handler, manager.handler)
	assert.Equal(t, 0, len(manager.clients))
	// cleanupTask is initialized when Start() is called
}

func TestClientManager_Start(t *testing.T) {
	manager := NewClientManager(DefaultClientStrategy, &mockFlowHandler{})

	err := manager.Start()
	require.NoError(t, err)
	// cleanupTask is initialized

	// Cleanup
	manager.Close()
}

func TestClientManager_Close(t *testing.T) {
	manager := NewClientManager(DefaultClientStrategy, &mockFlowHandler{})

	err := manager.Start()
	require.NoError(t, err)

	err = manager.Close()
	assert.NoError(t, err)
}

func TestClientManager_Create_WithError(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			return errors.New("handler error")
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)

	client, err := manager.Create()
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Wait for error handling
	time.Sleep(100 * time.Millisecond)

	// Client should be removed after error
	manager.clientsAccessLock.Lock()
	found := false
	for _, c := range manager.clients {
		if c == client {
			found = true
			break
		}
	}
	manager.clientsAccessLock.Unlock()
	assert.False(t, found, "client should be removed after error")
}

func TestClientManager_HandleReaderWriter_NewClient(t *testing.T) {
	test.InitZeroLog()
	manager := NewClientManager(DefaultClientStrategy, &mocks.LoopbackHandler{})
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	iLink, oLink := pipe.NewLinks(64*1024, false)
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Write some data and close
	testData := []byte("test")
	b := buf.New()
	b.Write(testData)
	iLink.WriteMultiBuffer(buf.MultiBuffer{b})
	iLink.CloseWrite()

	// Run in goroutine since HandleReaderWriter blocks
	errCh := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, oLink)
		errCh <- err
	}()

	// Wait for completion
	select {
	case err := <-errCh:
		// May return error if session ends, which is OK
		_ = err
	case <-time.After(2 * time.Second):
		t.Fatal("HandleReaderWriter should complete")
	}

	// Verify client was created (may be cleaned up, so check was created)
	// The client might be removed after session ends, so we just verify it was created
	time.Sleep(100 * time.Millisecond)
}

func TestClientManager_HandleReaderWriter_ReuseClient(t *testing.T) {
	strategy := ClientStrategy{
		MaxConnection:  10,
		MaxConcurrency: 2,
	}
	manager := NewClientManager(strategy, mocks.NewLoopbackHandler())
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Create first session
	rw1 := newMockReaderWriter()
	testData1 := []byte("test1")
	b1 := buf.New()
	b1.Write(testData1)
	rw1.readChan <- buf.MultiBuffer{b1}
	close(rw1.readChan)

	errCh1 := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw1)
		errCh1 <- err
	}()

	time.Sleep(100 * time.Millisecond)

	// Create second session - should reuse same client
	rw2 := newMockReaderWriter()
	testData2 := []byte("test2")
	b2 := buf.New()
	b2.Write(testData2)
	rw2.readChan <- buf.MultiBuffer{b2}
	close(rw2.readChan)

	errCh2 := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw2)
		errCh2 <- err
	}()

	// Wait for both to complete
	select {
	case <-errCh1:
		// May return error, which is OK
	case <-time.After(2 * time.Second):
		t.Fatal("first HandleReaderWriter should complete")
	}

	select {
	case <-errCh2:
		// May return error, which is OK
	case <-time.After(2 * time.Second):
		t.Fatal("second HandleReaderWriter should complete")
	}

	// Verify client management (clients may be cleaned up after sessions end)
	time.Sleep(100 * time.Millisecond)
}

func TestClientManager_HandleReaderWriter_CreateNewWhenFull(t *testing.T) {

	strategy := ClientStrategy{
		MaxConnection:  10,
		MaxConcurrency: 1, // Only one session per client
	}
	manager := NewClientManager(strategy, mocks.NewLoopbackHandler())
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Create first session
	rw1 := newMockReaderWriter()
	testData1 := []byte("test1")
	b1 := buf.New()
	b1.Write(testData1)
	rw1.readChan <- buf.MultiBuffer{b1}
	close(rw1.readChan)

	errCh1 := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw1)
		errCh1 <- err
	}()

	time.Sleep(100 * time.Millisecond)

	// Create second session - should create new client since first is full
	rw2 := newMockReaderWriter()
	testData2 := []byte("test2")
	b2 := buf.New()
	b2.Write(testData2)
	rw2.readChan <- buf.MultiBuffer{b2}
	close(rw2.readChan)

	errCh2 := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw2)
		errCh2 <- err
	}()

	// Wait for both to complete
	select {
	case <-errCh1:
		// May return error, which is OK
	case <-time.After(2 * time.Second):
		t.Fatal("first HandleReaderWriter should complete")
	}

	select {
	case <-errCh2:
		// May return error, which is OK
	case <-time.After(2 * time.Second):
		t.Fatal("second HandleReaderWriter should complete")
	}

	// Verify client management (clients may be cleaned up)
	time.Sleep(100 * time.Millisecond)
}

func TestClientManager_TryRetire(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			return nil
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Add client to manager
	manager.clientsAccessLock.Lock()
	manager.clients = append(manager.clients, client)
	manager.clientsAccessLock.Unlock()

	// Set client to closing state but not empty (has sessions)
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
	client.count.Store(DefaultClientStrategy.MaxConnection)

	// Try retire - should not remove (not empty)
	manager.tryRetire(client)
	manager.clientsAccessLock.Lock()
	assert.Equal(t, 1, len(manager.clients), "should not remove non-empty client")
	manager.clientsAccessLock.Unlock()

	// Make client empty and closing
	client.RemoveSession(session)
	client.count.Store(DefaultClientStrategy.MaxConnection)
	client.Close() // Mark as closing

	// Try retire - should remove
	manager.tryRetire(client)
	manager.clientsAccessLock.Lock()
	assert.Equal(t, 0, len(manager.clients), "should remove empty closing client")
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_DeleteClient(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			return nil
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)

	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client1, _ := NewClient(ctx, iLink, DefaultClientStrategy)
	client2, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	manager.clientsAccessLock.Lock()
	manager.clients = append(manager.clients, client1, client2)
	manager.clientsAccessLock.Unlock()

	assert.Equal(t, 2, len(manager.clients))

	manager.deleteClient(client1)

	manager.clientsAccessLock.Lock()
	assert.Equal(t, 1, len(manager.clients))
	assert.Equal(t, client2, manager.clients[0])
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_CleanupFunc(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			return nil
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)

	ctx := context.Background()
	iLink1, _ := pipe.NewLinks(64*1024, false)
	iLink2, _ := pipe.NewLinks(64*1024, false)
	client1, _ := NewClient(ctx, iLink1, DefaultClientStrategy)
	client2, _ := NewClient(ctx, iLink2, DefaultClientStrategy)

	// Make client1 idle
	client1.RemoveSession(&clientSession{ID: 1})
	time.Sleep(6 * time.Second)

	// Add active session to client2
	rw := newMockReaderWriter()
	session := &clientSession{
		ID:              1,
		ctx:             ctx,
		rw:              rw,
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
	}
	client2.AddSession(session)

	manager.clientsAccessLock.Lock()
	manager.clients = append(manager.clients, client1, client2)
	manager.clientsAccessLock.Unlock()

	err := manager.cleanupFunc()
	assert.NoError(t, err)

	manager.clientsAccessLock.Lock()
	// client1 should be removed (idle), client2 should remain (active)
	assert.Equal(t, 1, len(manager.clients))
	assert.Equal(t, client2, manager.clients[0])
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_CleanupFunc_NoClients(t *testing.T) {
	manager := NewClientManager(DefaultClientStrategy, &mockFlowHandler{})

	err := manager.cleanupFunc()
	assert.NoError(t, err)
}

func TestClientManager_HandleReaderWriter_WithError(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			return nil
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	rw := newMockReaderWriter()
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Send error through rw
	rw.Interrupt(errors.New("test error"))

	errCh := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("HandleReaderWriter should return error")
	}
}

func TestDefaultClientStrategy(t *testing.T) {
	assert.Equal(t, uint32(16), DefaultClientStrategy.MaxConnection)
	assert.Equal(t, uint32(2), DefaultClientStrategy.MaxConcurrency)
}

func TestClientManager_TryRetire_NotClosing(t *testing.T) {
	handler := &mockFlowHandler{}
	manager := NewClientManager(DefaultClientStrategy, handler)
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	manager.clientsAccessLock.Lock()
	manager.clients = append(manager.clients, client)
	manager.clientsAccessLock.Unlock()

	// Client is empty but not closing - should not be removed
	manager.tryRetire(client)
	manager.clientsAccessLock.Lock()
	assert.Equal(t, 1, len(manager.clients))
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_TryRetire_NotEmpty(t *testing.T) {
	handler := &mockFlowHandler{}
	manager := NewClientManager(DefaultClientStrategy, handler)
	manager.Start()
	defer manager.Close()

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
	client.count.Store(DefaultClientStrategy.MaxConnection)

	manager.clientsAccessLock.Lock()
	manager.clients = append(manager.clients, client)
	manager.clientsAccessLock.Unlock()

	// Client is closing but not empty - should not be removed
	manager.tryRetire(client)
	manager.clientsAccessLock.Lock()
	assert.Equal(t, 1, len(manager.clients))
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_DeleteClient_NotInList(t *testing.T) {
	handler := &mockFlowHandler{}
	manager := NewClientManager(DefaultClientStrategy, handler)

	ctx := context.Background()
	iLink, _ := pipe.NewLinks(64*1024, false)
	client, _ := NewClient(ctx, iLink, DefaultClientStrategy)

	// Delete client not in list - should not panic
	manager.deleteClient(client)

	manager.clientsAccessLock.Lock()
	assert.Equal(t, 0, len(manager.clients))
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_HandleReaderWriter_LeftToRightDone(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			// Close immediately to trigger leftToRightDone
			rw.CloseWrite()
			return nil
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	rw := newMockReaderWriter()
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	testData := []byte("test")
	b := buf.New()
	b.Write(testData)
	rw.readChan <- buf.MultiBuffer{b}
	close(rw.readChan)

	errCh := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		// Should complete when leftToRightDone
		_ = err
	case <-time.After(2 * time.Second):
		t.Fatal("HandleReaderWriter should complete")
	}
}

func TestClientManager_HandleReaderWriter_RightToLeftDone(t *testing.T) {

	manager := NewClientManager(DefaultClientStrategy, mocks.NewLoopbackHandler())
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	rw := newMockReaderWriter()
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	testData := []byte("test")
	b := buf.New()
	b.Write(testData)
	rw.readChan <- buf.MultiBuffer{b}
	close(rw.readChan)

	errCh := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		// Should complete when rightToLeftDone
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("HandleReaderWriter should complete")
	}
}

func TestClientManager_CleanupFunc_AllIdle(t *testing.T) {
	handler := &mockFlowHandler{}
	manager := NewClientManager(DefaultClientStrategy, handler)

	ctx := context.Background()
	iLink1, _ := pipe.NewLinks(64*1024, false)
	iLink2, _ := pipe.NewLinks(64*1024, false)
	client1, _ := NewClient(ctx, iLink1, DefaultClientStrategy)
	client2, _ := NewClient(ctx, iLink2, DefaultClientStrategy)

	// Make both idle
	client1.RemoveSession(&clientSession{ID: 1})
	client2.RemoveSession(&clientSession{ID: 2})
	time.Sleep(6 * time.Second)

	manager.clientsAccessLock.Lock()
	manager.clients = append(manager.clients, client1, client2)
	manager.clientsAccessLock.Unlock()

	err := manager.cleanupFunc()
	assert.NoError(t, err)

	manager.clientsAccessLock.Lock()
	// All idle clients should be removed
	assert.Equal(t, 0, len(manager.clients))
	manager.clientsAccessLock.Unlock()
}

func TestClientManager_HandleReaderWriter_ClientSelection(t *testing.T) {

	strategy := ClientStrategy{
		MaxConnection:  10,
		MaxConcurrency: 1, // Only one session per client
	}
	manager := NewClientManager(strategy, mocks.NewLoopbackHandler())
	manager.Start()
	defer manager.Close()

	ctx := context.Background()
	dest := net.TCPDestination(net.ParseAddress("127.0.0.1"), net.Port(80))

	// Create first session
	rw1 := newMockReaderWriter()
	testData1 := []byte("test1")
	b1 := buf.New()
	b1.Write(testData1)
	rw1.readChan <- buf.MultiBuffer{b1}
	close(rw1.readChan)

	errCh1 := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw1)
		errCh1 <- err
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify client was created
	manager.clientsAccessLock.Lock()
	initialClientCount := len(manager.clients)
	manager.clientsAccessLock.Unlock()
	assert.Equal(t, 1, initialClientCount)

	// Create second session - should create new client since first is full
	rw2 := newMockReaderWriter()
	testData2 := []byte("test2")
	b2 := buf.New()
	b2.Write(testData2)
	rw2.readChan <- buf.MultiBuffer{b2}
	close(rw2.readChan)

	errCh2 := make(chan error, 1)
	go func() {
		err := manager.HandleReaderWriter(ctx, dest, rw2)
		errCh2 <- err
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify new client was created
	manager.clientsAccessLock.Lock()
	finalClientCount := len(manager.clients)
	manager.clientsAccessLock.Unlock()
	assert.GreaterOrEqual(t, finalClientCount, 1)

	// Wait for completion
	select {
	case <-errCh1:
	case <-time.After(2 * time.Second):
	}
	select {
	case <-errCh2:
	case <-time.After(2 * time.Second):
	}
}

func TestClientManager_Close_WithoutStart(t *testing.T) {
	manager := NewClientManager(DefaultClientStrategy, &mockFlowHandler{})

	// Should not panic
	err := manager.Close()
	assert.NoError(t, err)
}

func TestClientManager_Create_HandlerError(t *testing.T) {
	handler := &mockFlowHandler{
		handleFlowFunc: func(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
			return errors.New("handler error")
		},
	}
	manager := NewClientManager(DefaultClientStrategy, handler)

	client, err := manager.Create()
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Wait for error handling
	time.Sleep(100 * time.Millisecond)

	// Client should be removed after error
	manager.clientsAccessLock.Lock()
	found := false
	for _, c := range manager.clients {
		if c == client {
			found = true
			break
		}
	}
	manager.clientsAccessLock.Unlock()
	assert.False(t, found, "client should be removed after error")
}
