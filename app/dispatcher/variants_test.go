package dispatcher

import (
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
)

// mockReaderWriter is a simple mock that records reads and writes
type mockReaderWriter struct {
	readData  buf.MultiBuffer
	writeData buf.MultiBuffer
	mu        sync.Mutex
}

func (m *mockReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mb := m.readData
	m.readData = nil
	return mb, nil
}

func (m *mockReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeData, _ = buf.MergeMulti(m.writeData, mb)
	return nil
}

func (m *mockReaderWriter) CloseWrite() error {
	return nil
}

func createMultiBuffer(size int) buf.MultiBuffer {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return buf.MergeBytes(nil, data)
}
