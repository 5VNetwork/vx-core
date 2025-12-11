package pipe

import (
	"net"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkConn_BasicReadWrite(t *testing.T) {
	l1, l2 := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	// Test basic write and read
	testData := []byte("Hello, World!")

	// Write data
	n, err := conn.Write(testData)
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)

	// Read data from the other end
	mb, err := l2.Reader.ReadMultiBuffer()
	require.NoError(t, err)
	readBuf := make([]byte, len(testData))
	_, nBytes := buf.SplitBytes(mb, readBuf)
	assert.Equal(t, len(testData), nBytes)
	assert.Equal(t, testData, readBuf)

	// Write data to l2 and read from conn
	writeData := []byte("Response data")
	buffer := buf.NewWithSize(int32(len(writeData)))
	buffer.Write(writeData)
	err = l2.Writer.WriteMultiBuffer(buf.MultiBuffer{buffer})
	require.NoError(t, err)

	readBuf = make([]byte, len(writeData))
	n, err = conn.Read(readBuf)
	require.NoError(t, err)
	assert.Equal(t, len(writeData), n)
	assert.Equal(t, writeData, readBuf)
}

func TestLinkConn_Addresses(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	localAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80}

	conn := &LinkConn{
		link:       l1,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	assert.Equal(t, localAddr, conn.LocalAddr())
	assert.Equal(t, remoteAddr, conn.RemoteAddr())
}

func TestLinkConn_SetDeadline(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	deadline := time.Now().Add(5 * time.Second)
	err := conn.SetDeadline(deadline)
	require.NoError(t, err)

	assert.Equal(t, deadline, conn.readDeadline)
	assert.Equal(t, deadline, conn.writeDeadline)
}

func TestLinkConn_SetReadDeadline(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	deadline := time.Now().Add(5 * time.Second)
	err := conn.SetReadDeadline(deadline)
	require.NoError(t, err)

	assert.Equal(t, deadline, conn.readDeadline)
	assert.True(t, conn.writeDeadline.IsZero())
}

func TestLinkConn_SetWriteDeadline(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	deadline := time.Now().Add(5 * time.Second)
	err := conn.SetWriteDeadline(deadline)
	require.NoError(t, err)

	assert.Equal(t, deadline, conn.writeDeadline)
	assert.True(t, conn.readDeadline.IsZero())
}

func TestLinkConn_ReadDeadlineExceeded(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	// Set deadline in the past
	pastDeadline := time.Now().Add(-1 * time.Second)
	err := conn.SetReadDeadline(pastDeadline)
	require.NoError(t, err)

	// Try to read - should fail immediately
	readBuf := make([]byte, 100)
	_, err = conn.Read(readBuf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestLinkConn_WriteDeadlineExceeded(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	// Set deadline in the past
	pastDeadline := time.Now().Add(-1 * time.Second)
	err := conn.SetWriteDeadline(pastDeadline)
	require.NoError(t, err)

	// Try to write - should fail immediately
	testData := []byte("Hello, World!")
	_, err = conn.Write(testData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestLinkConn_ReadWithTimeout(t *testing.T) {
	l1, l2 := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	// Set read deadline 100ms in the future
	deadline := time.Now().Add(100 * time.Millisecond)
	err := conn.SetReadDeadline(deadline)
	require.NoError(t, err)

	// Try to read without any data available - should timeout
	readBuf := make([]byte, 100)
	start := time.Now()
	_, err = conn.Read(readBuf)
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Should timeout around 100ms (with some tolerance)
	assert.True(t, elapsed >= 90*time.Millisecond && elapsed <= 200*time.Millisecond)

	// Now write some data and read with a future deadline
	deadline = time.Now().Add(5 * time.Second)
	err = conn.SetReadDeadline(deadline)
	require.NoError(t, err)

	testData := []byte("Test data")
	buffer := buf.NewWithSize(int32(len(testData)))
	buffer.Write(testData)
	err = l2.Writer.WriteMultiBuffer(buf.MultiBuffer{buffer})
	require.NoError(t, err)

	readBuf = make([]byte, len(testData))
	n, err := conn.Read(readBuf)
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, readBuf)
}

func TestLinkConn_Close(t *testing.T) {
	l1, _ := NewLinks(8192, false)

	conn := &LinkConn{
		link:       l1,
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 80},
	}

	// Add some data to the buffer
	testData := []byte("Hello, World!")
	buffer := buf.NewWithSize(int32(len(testData)))
	buffer.Write(testData)
	conn.mb = buf.MultiBuffer{buffer}

	err := conn.Close()
	assert.NoError(t, err)

	// After close, the multibuffer should be released (empty)
	assert.True(t, conn.mb.IsEmpty())
}
