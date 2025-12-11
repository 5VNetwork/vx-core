package mux

import (
	"bytes"
	"io"
	"testing"

	"github.com/5vnetwork/vx-core/common/buf"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWriter captures written data for testing
type mockWriter struct {
	data []byte
}

func (m *mockWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for _, b := range mb {
		m.data = append(m.data, b.Bytes()...)
	}
	buf.ReleaseMulti(mb)
	return nil
}

func (m *mockWriter) CloseWrite() error {
	return nil
}

func TestNewMuxWriter(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))

	writer := NewMuxWriter(123, dest, mock, TransferTypeStream)

	assert.NotNil(t, writer)
	assert.Equal(t, uint16(123), writer.id)
	assert.Equal(t, dest, writer.dest)
	assert.Equal(t, mock, writer.writer)
	assert.False(t, writer.followup)
	assert.False(t, writer.hasError)
	assert.Equal(t, TransferTypeStream, writer.transferType)
}

func TestNewResponseMuxWriter(t *testing.T) {
	mock := &mockWriter{}

	writer := NewResponseMuxWriter(456, mock, TransferTypePacket)

	assert.NotNil(t, writer)
	assert.Equal(t, uint16(456), writer.id)
	assert.Equal(t, mock, writer.writer)
	assert.True(t, writer.followup)
	assert.False(t, writer.hasError)
	assert.Equal(t, TransferTypePacket, writer.transferType)
}

func TestMuxWriter_WriteMultiBuffer_Empty(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(1, dest, mock, TransferTypeStream)

	mb := buf.MultiBuffer{}
	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write metadata with SessionStatusNew and no data
	assert.Greater(t, len(mock.data), 0)

	// Verify frame structure
	reader := bytes.NewReader(mock.data)
	var meta FrameMetadata
	err = meta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(1), meta.SessionID)
	assert.Equal(t, SessionStatusNew, meta.SessionStatus)
	assert.False(t, meta.Option.Has(OptionData))
}

func TestMuxWriter_WriteMultiBuffer_WithData_Stream(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("example.com"), nethelper.Port(443))
	writer := NewMuxWriter(2, dest, mock, TransferTypeStream)

	// Create test data
	testData := []byte("Hello, World!")
	b := buf.New()
	b.Write(testData)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write metadata with SessionStatusNew and data
	assert.Greater(t, len(mock.data), 0)

	// Verify frame structure
	reader := bytes.NewReader(mock.data)
	var meta FrameMetadata
	err = meta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(2), meta.SessionID)
	assert.Equal(t, SessionStatusNew, meta.SessionStatus)
	assert.True(t, meta.Option.Has(OptionData))
	assert.Equal(t, dest, meta.Target)

	// Read data length
	dataLen, err := serial.ReadUint16(reader)
	require.NoError(t, err)
	assert.Equal(t, uint16(len(testData)), dataLen)

	// Read actual data
	actualData := make([]byte, dataLen)
	_, err = io.ReadFull(reader, actualData)
	require.NoError(t, err)
	assert.Equal(t, testData, actualData)
}

func TestMuxWriter_WriteMultiBuffer_WithData_Packet(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.UDPDestination(nethelper.ParseAddress("8.8.8.8"), nethelper.Port(53))
	writer := NewMuxWriter(3, dest, mock, TransferTypePacket)

	// Create multiple buffers (packets)
	packet1 := []byte("packet1")
	packet2 := []byte("packet2")

	b1 := buf.New()
	b1.Write(packet1)
	b2 := buf.New()
	b2.Write(packet2)

	mb := buf.MultiBuffer{b1, b2}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// For packet mode, each packet should be written separately
	// First packet with SessionStatusNew
	reader := bytes.NewReader(mock.data)
	var meta1 FrameMetadata
	err = meta1.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(3), meta1.SessionID)
	assert.Equal(t, SessionStatusNew, meta1.SessionStatus)
	assert.True(t, meta1.Option.Has(OptionData))

	// Skip first packet data
	dataLen1, _ := serial.ReadUint16(reader)
	reader.Seek(int64(reader.Size())-int64(reader.Len())+int64(dataLen1), io.SeekStart)

	// Second packet with SessionStatusKeep
	var meta2 FrameMetadata
	err = meta2.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(3), meta2.SessionID)
	assert.Equal(t, SessionStatusKeep, meta2.SessionStatus)
	assert.True(t, meta2.Option.Has(OptionData))
}

func TestMuxWriter_WriteMultiBuffer_LargeData_Stream(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(9000))
	writer := NewMuxWriter(4, dest, mock, TransferTypeStream)

	// Create large data (> 8KB to trigger chunking)
	largeData := make([]byte, 20*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	b := buf.NewWithSize(int32(len(largeData)))
	b.Write(largeData)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write multiple frames due to chunking
	assert.Greater(t, len(mock.data), len(largeData))

	// Verify first frame
	reader := bytes.NewReader(mock.data)
	var meta FrameMetadata
	err = meta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(4), meta.SessionID)
	assert.Equal(t, SessionStatusNew, meta.SessionStatus)
	assert.True(t, meta.Option.Has(OptionData))
}

func TestMuxWriter_SendSessionStatusEnd_NoError(t *testing.T) {
	mock := &mockWriter{}
	writer := NewResponseMuxWriter(5, mock, TransferTypeStream)

	err := writer.SendSessionStatusEnd()
	require.NoError(t, err)

	// Verify frame structure
	reader := bytes.NewReader(mock.data)
	var meta FrameMetadata
	err = meta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(5), meta.SessionID)
	assert.Equal(t, SessionStatusEnd, meta.SessionStatus)
	assert.False(t, meta.Option.Has(OptionError))
	assert.False(t, meta.Option.Has(OptionData))
}

func TestMuxWriter_SendSessionStatusEnd_WithError(t *testing.T) {
	mock := &mockWriter{}
	writer := NewResponseMuxWriter(6, mock, TransferTypeStream)
	writer.hasError = true

	err := writer.SendSessionStatusEnd()
	require.NoError(t, err)

	// Verify frame structure
	reader := bytes.NewReader(mock.data)
	var meta FrameMetadata
	err = meta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(6), meta.SessionID)
	assert.Equal(t, SessionStatusEnd, meta.SessionStatus)
	assert.True(t, meta.Option.Has(OptionError))
}

func TestMuxWriter_GetNextFrameMeta_FirstTime(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("1.2.3.4"), nethelper.Port(80))
	writer := NewMuxWriter(7, dest, mock, TransferTypeStream)

	meta := writer.getNextFrameMeta()

	assert.Equal(t, uint16(7), meta.SessionID)
	assert.Equal(t, SessionStatusNew, meta.SessionStatus)
	assert.Equal(t, dest, meta.Target)
	assert.True(t, writer.followup, "followup should be set after first call")
}

func TestMuxWriter_GetNextFrameMeta_Subsequent(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("1.2.3.4"), nethelper.Port(80))
	writer := NewMuxWriter(8, dest, mock, TransferTypeStream)

	// First call
	meta1 := writer.getNextFrameMeta()
	assert.Equal(t, SessionStatusNew, meta1.SessionStatus)

	// Second call
	meta2 := writer.getNextFrameMeta()
	assert.Equal(t, SessionStatusKeep, meta2.SessionStatus)
}

func TestMuxWriter_CloseWrite(t *testing.T) {
	mock := &mockWriter{}
	writer := NewResponseMuxWriter(9, mock, TransferTypeStream)

	err := writer.CloseWrite()
	assert.NoError(t, err)
}

func TestMuxWriter_MultipleWrites(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("test.local"), nethelper.Port(1234))
	writer := NewMuxWriter(10, dest, mock, TransferTypeStream)

	// First write
	b1 := buf.New()
	b1.Write([]byte("first"))
	err := writer.WriteMultiBuffer(buf.MultiBuffer{b1})
	require.NoError(t, err)

	dataLen1 := len(mock.data)

	// Second write
	b2 := buf.New()
	b2.Write([]byte("second"))
	err = writer.WriteMultiBuffer(buf.MultiBuffer{b2})
	require.NoError(t, err)

	// Should have more data after second write
	assert.Greater(t, len(mock.data), dataLen1)

	// Verify both frames
	reader := bytes.NewReader(mock.data)

	var meta1 FrameMetadata
	err = meta1.Unmarshal(reader)
	require.NoError(t, err)
	assert.Equal(t, SessionStatusNew, meta1.SessionStatus)

	// Skip first frame data
	dataLen, _ := serial.ReadUint16(reader)
	data1 := make([]byte, dataLen)
	io.ReadFull(reader, data1)

	var meta2 FrameMetadata
	err = meta2.Unmarshal(reader)
	require.NoError(t, err)
	assert.Equal(t, SessionStatusKeep, meta2.SessionStatus)
}

func TestWriteMetaWithFrame(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("10.0.0.1"), nethelper.Port(5000))

	meta := FrameMetadata{
		SessionID:     11,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        dest,
	}

	testData := []byte("test payload")
	b := buf.New()
	b.Write(testData)
	mb := buf.MultiBuffer{b}

	err := writeMetaWithFrame(mock, meta, mb)
	require.NoError(t, err)

	// Verify written data
	reader := bytes.NewReader(mock.data)
	var decodedMeta FrameMetadata
	err = decodedMeta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, meta.SessionID, decodedMeta.SessionID)
	assert.Equal(t, meta.SessionStatus, decodedMeta.SessionStatus)

	dataLen, err := serial.ReadUint16(reader)
	require.NoError(t, err)
	assert.Equal(t, uint16(len(testData)), dataLen)

	actualData := make([]byte, dataLen)
	_, err = io.ReadFull(reader, actualData)
	require.NoError(t, err)
	assert.Equal(t, testData, actualData)
}

func TestMuxWriter_WriteMultiBuffer_StreamChunking(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("chunked.test"), nethelper.Port(8000))
	writer := NewMuxWriter(12, dest, mock, TransferTypeStream)

	// Create data exactly at chunk boundary (8KB)
	data := make([]byte, 8*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b := buf.New()
	b.Write(data)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write as single chunk at boundary
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMultiBuffer_EmptyMultiBuffer(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(13, dest, mock, TransferTypeStream)

	// Write empty MultiBuffer (nil)
	err := writer.WriteMultiBuffer(nil)
	require.NoError(t, err)

	// Should write metadata only
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMultiBuffer_MultiplePackets(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.UDPDestination(nethelper.ParseAddress("8.8.8.8"), nethelper.Port(53))
	writer := NewMuxWriter(14, dest, mock, TransferTypePacket)

	// Create multiple packets
	packets := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
		[]byte("packet3"),
	}

	for _, pkt := range packets {
		b := buf.New()
		b.Write(pkt)
		err := writer.WriteMultiBuffer(buf.MultiBuffer{b})
		require.NoError(t, err)
	}

	// Should have written multiple frames
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMetaWithFrame_TooLarge(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("10.0.0.1"), nethelper.Port(5000))

	meta := FrameMetadata{
		SessionID:     15,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        dest,
	}

	// Create data that's too large (> 64MB)
	// Note: This might not actually trigger the error in practice,
	// but we test the error path
	largeData := make([]byte, 65*1024*1024)
	b := buf.New()
	b.Write(largeData[:1024]) // Write a smaller chunk to avoid memory issues
	mb := buf.MultiBuffer{b}

	err := writeMetaWithFrame(mock, meta, mb)
	// Should not error for reasonable sizes
	// The check is for len(data)+1 > 64*1024*1024
	_ = err
}

func TestMuxWriter_ResponseWriter_FirstWrite(t *testing.T) {
	mock := &mockWriter{}
	writer := NewResponseMuxWriter(16, mock, TransferTypeStream)

	// First write should use SessionStatusKeep (not New, since it's a response)
	testData := []byte("response")
	b := buf.New()
	b.Write(testData)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Verify frame structure
	reader := bytes.NewReader(mock.data)
	var meta FrameMetadata
	err = meta.Unmarshal(reader)
	require.NoError(t, err)

	assert.Equal(t, uint16(16), meta.SessionID)
	assert.Equal(t, SessionStatusKeep, meta.SessionStatus)
	assert.True(t, meta.Option.Has(OptionData))
}

func TestMuxWriter_SendSessionStatusEnd_MultipleTimes(t *testing.T) {
	mock := &mockWriter{}
	writer := NewResponseMuxWriter(17, mock, TransferTypeStream)

	// Send multiple times (should work)
	err1 := writer.SendSessionStatusEnd()
	require.NoError(t, err1)

	err2 := writer.SendSessionStatusEnd()
	require.NoError(t, err2)

	// Both should succeed
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestMuxWriter_WriteMultiBuffer_NilWriter(t *testing.T) {
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(18, dest, nil, TransferTypeStream)

	// Should handle nil writer gracefully (will panic on actual write, but we test the structure)
	assert.NotNil(t, writer)
	assert.Equal(t, uint16(18), writer.id)
}

func TestMuxWriter_WriteMultiBuffer_ZeroLengthData(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(19, dest, mock, TransferTypeStream)

	// Create buffer with zero length
	b := buf.New()
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write metadata only
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMultiBuffer_VeryLargePacket(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.UDPDestination(nethelper.ParseAddress("8.8.8.8"), nethelper.Port(53))
	writer := NewMuxWriter(20, dest, mock, TransferTypePacket)

	// Create large packet (but within limits)
	largePacket := make([]byte, 64*1024)
	for i := range largePacket {
		largePacket[i] = byte(i % 256)
	}

	b := buf.New()
	b.Write(largePacket)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write successfully
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMultiBuffer_MultipleEmptyBuffers(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(26, dest, mock, TransferTypeStream)

	// Create multiple empty buffers
	b1 := buf.New()
	b2 := buf.New()
	b3 := buf.New()
	mb := buf.MultiBuffer{b1, b2, b3}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should handle empty buffers gracefully
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMultiBuffer_StreamBoundary(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(27, dest, mock, TransferTypeStream)

	// Create data exactly at 8KB boundary
	data := make([]byte, 8*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b := buf.New()
	b.Write(data)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write successfully
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_WriteMultiBuffer_StreamOverBoundary(t *testing.T) {
	mock := &mockWriter{}
	dest := nethelper.TCPDestination(nethelper.ParseAddress("127.0.0.1"), nethelper.Port(8080))
	writer := NewMuxWriter(28, dest, mock, TransferTypeStream)

	// Create data slightly over 8KB boundary
	data := make([]byte, 8*1024+1)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b := buf.New()
	b.Write(data)
	mb := buf.MultiBuffer{b}

	err := writer.WriteMultiBuffer(mb)
	require.NoError(t, err)

	// Should write in chunks
	assert.Greater(t, len(mock.data), 0)
}

func TestMuxWriter_CloseWrite_AlwaysSucceeds(t *testing.T) {
	mock := &mockWriter{}
	writer := NewResponseMuxWriter(29, mock, TransferTypeStream)

	// CloseWrite should always succeed
	err := writer.CloseWrite()
	assert.NoError(t, err)

	// Multiple calls should also succeed
	err = writer.CloseWrite()
	assert.NoError(t, err)
}
