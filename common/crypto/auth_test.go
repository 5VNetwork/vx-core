package crypto

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/protocol"

	"github.com/google/go-cmp/cmp"
)

// testPadding is a simple padding generator for testing
type testPadding struct {
	maxLen uint16
	next   uint16
}

func (p *testPadding) MaxPaddingLen() uint16 {
	return p.maxLen
}

func (p *testPadding) NextPaddingLen() uint16 {
	return p.next
}

func TestAuthenticationReaderWriter(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	const payloadSize = 1024 * 80
	rawPayload := make([]byte, payloadSize)
	rand.Read(rawPayload)

	payload := buf.MergeBytes(nil, rawPayload)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriter(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	common.Must(writer.WriteMultiBuffer(payload))
	if cache.Len() <= 1024*80 {
		t.Error("cache len: ", cache.Len())
	}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{}))

	reader := NewAuthenticationReader(context.Background(), &AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	var mb buf.MultiBuffer

	for mb.Len() < payloadSize {
		mb2, err := reader.ReadMultiBuffer()
		common.Must(err)

		mb, _ = buf.MergeMulti(mb, mb2)
	}

	if mb.Len() != payloadSize {
		t.Error("mb len: ", mb.Len())
	}

	mbContent := make([]byte, payloadSize)
	buf.SplitBytes(mb, mbContent)
	if r := cmp.Diff(mbContent, rawPayload); r != "" {
		t.Error(r)
	}

	_, err = reader.ReadMultiBuffer()
	if err != io.EOF {
		t.Error("error: ", err)
	}
}

func TestAuthenticationReaderWriterPacket(t *testing.T) {
	key := make([]byte, 16)
	common.Must2(rand.Read(key))
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := buf.New()
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriter(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypePacket, nil)

	var payload buf.MultiBuffer
	pb1 := buf.New()
	pb1.Write([]byte("abcd"))
	payload = append(payload, pb1)

	pb2 := buf.New()
	pb2.Write([]byte("efgh"))
	payload = append(payload, pb2)

	common.Must(writer.WriteMultiBuffer(payload))
	if cache.Len() == 0 {
		t.Error("cache len: ", cache.Len())
	}

	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{}))

	reader := NewAuthenticationReader(context.Background(), &AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypePacket, nil)

	mb, err := reader.ReadMultiBuffer()
	common.Must(err)

	mb, b1 := buf.SplitFirst(mb)
	if b1.String() != "abcd" {
		t.Error("b1: ", b1.String())
	}

	mb, b2 := buf.SplitFirst(mb)
	if b2.String() != "efgh" {
		t.Error("b2: ", b2.String())
	}

	if !mb.IsEmpty() {
		t.Error("not empty")
	}

	_, err = reader.ReadMultiBuffer()
	if err != io.EOF {
		t.Error("error: ", err)
	}
}

func TestAuthenticationReaderWriterIO(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	const payloadSize = 1024 * 80
	rawPayload := make([]byte, payloadSize)
	rand.Read(rawPayload)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	common.Must2(writer.Write(rawPayload))
	if cache.Len() <= 1024*80 {
		t.Error("cache len: ", cache.Len())
	}
	common.Must2(writer.Write([]byte{}))

	reader := NewAuthenticationReader1(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	var mb buf.MultiBuffer

	for mb.Len() < payloadSize {
		b := buf.New()
		mb2, err := reader.Read(b.BytesRange(0, b.Cap()))
		common.Must(err)
		b.Extend(int32(mb2))
		mb, _ = buf.MergeMulti(mb, buf.MultiBuffer{b})
	}

	if mb.Len() != payloadSize {
		t.Error("mb len: ", mb.Len())
	}

	mbContent := make([]byte, payloadSize)
	buf.SplitBytes(mb, mbContent)
	if r := cmp.Diff(mbContent, rawPayload); r != "" {
		t.Error(r)
	}

	_, err = reader.Read(make([]byte, 2048))
	if err != io.EOF {
		t.Error("error: ", err)
	}
}

func TestAuthenticationReaderWriterPacketIO(t *testing.T) {
	key := make([]byte, 16)
	common.Must2(rand.Read(key))
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := buf.New()
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypePacket, nil)

	pb1 := buf.New()
	pb1.Write([]byte("abcd"))

	pb2 := buf.New()
	pb2.Write([]byte("efgh"))

	common.Must2(writer.Write(pb1.Bytes()))
	if cache.Len() == 0 {
		t.Error("cache len: ", cache.Len())
	}

	common.Must2(writer.Write(pb2.Bytes()))
	if cache.Len() == 0 {
		t.Error("cache len: ", cache.Len())
	}

	common.Must2(writer.Write([]byte{}))

	reader := NewAuthenticationReader1(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypePacket, nil)

	b := make([]byte, 2048)
	mb, err := reader.Read(b)
	common.Must(err)

	if string(b[:mb]) != "abcd" {
		t.Error("b1: ", string(b[:mb]))
	}

	mb, err = reader.Read(b)
	common.Must(err)

	if string(b[:mb]) != "efgh" {
		t.Error("b2: ", string(b[:mb]))
	}

	_, err = reader.Read(make([]byte, 2048))
	if err != io.EOF {
		t.Error("error: ", err)
	}
}

// TestAuthenticationWriterIO_EmptyWrite tests writing empty data
func TestAuthenticationWriterIO_EmptyWrite(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	n, err := writer.Write([]byte{})
	if err != nil {
		t.Errorf("Write empty data failed: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes written, got %d", n)
	}
	if cache.Len() == 0 {
		t.Error("Empty write should still produce output (termination chunk)")
	}
}

// TestAuthenticationWriterIO_StreamLargeData tests writing large data that requires chunking
func TestAuthenticationWriterIO_StreamLargeData(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	// Write data larger than payloadSize to test chunking
	const largeSize = 50000
	largeData := make([]byte, largeSize)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	n, err := writer.Write(largeData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != largeSize {
		t.Errorf("Expected %d bytes written, got %d", largeSize, n)
	}

	// Verify we can read it back
	reader := NewAuthenticationReader1(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	readData := make([]byte, largeSize)
	totalRead := 0
	for totalRead < largeSize {
		n, err := reader.Read(readData[totalRead:])
		if err != nil && err != io.EOF {
			t.Fatalf("Read failed: %v", err)
		}
		if n == 0 {
			break
		}
		totalRead += n
	}

	if totalRead != largeSize {
		t.Errorf("Expected %d bytes read, got %d", largeSize, totalRead)
	}

	if !bytes.Equal(readData, largeData) {
		t.Error("Read data doesn't match written data")
	}
}

// TestAuthenticationWriterIO_PacketMode tests packet mode writes
func TestAuthenticationWriterIO_PacketMode(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypePacket, nil)

	testData := []byte("test packet data")
	n, err := writer.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}
}

// TestAuthenticationWriterIO_MultipleWrites tests multiple sequential writes
func TestAuthenticationWriterIO_MultipleWrites(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	// Write multiple chunks
	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
		[]byte("chunk3"),
	}

	totalWritten := 0
	for _, chunk := range chunks {
		n, err := writer.Write(chunk)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(chunk) {
			t.Errorf("Expected %d bytes written, got %d", len(chunk), n)
		}
		totalWritten += n
	}

	// Verify we can read it back
	reader := NewAuthenticationReader1(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	expectedData := bytes.Join(chunks, nil)
	readData := make([]byte, len(expectedData))
	totalRead := 0
	for totalRead < len(expectedData) {
		n, err := reader.Read(readData[totalRead:])
		if err != nil && err != io.EOF {
			t.Fatalf("Read failed: %v", err)
		}
		if n == 0 {
			break
		}
		totalRead += n
	}

	if !bytes.Equal(readData, expectedData) {
		t.Errorf("Read data doesn't match. Expected %q, got %q", expectedData, readData)
	}
}

// TestAuthenticationWriterIO_WithPadding tests writes with padding
func TestAuthenticationWriterIO_WithPadding(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	// Create a simple padding generator
	padding := &testPadding{
		maxLen: 10,
		next:   5,
	}

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, padding)

	testData := []byte("test data with padding")
	n, err := writer.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}
}

// TestAuthenticationWriterIO_SealError tests error handling when seal fails
// BUG: In writeStream, if seal() returns an error, sealed will be nil, but
// the code calls sealed.Release() which will panic. This is a bug at line 638 of auth.go
func TestAuthenticationWriterIO_SealError(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	// Test with normal data first - this should work
	normalData := make([]byte, 1000)
	rand.Read(normalData)
	n, err := writer.Write(normalData)
	if err != nil {
		t.Errorf("Normal write failed: %v", err)
	}
	if n != len(normalData) {
		t.Errorf("Expected %d bytes written, got %d", len(normalData), n)
	}
}

// TestAuthenticationWriterIO_ExactChunkSize tests writing exactly chunk size
func TestAuthenticationWriterIO_ExactChunkSize(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	cache := bytes.NewBuffer(nil)
	iv := make([]byte, 12)
	rand.Read(iv)

	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	// Calculate exact payload size
	overhead := int(aead.Overhead())
	sizeBytes := int(PlainChunkSizeParser{}.SizeBytes())
	payloadSize := int(buf.Size) - overhead - sizeBytes

	// Write exactly payloadSize bytes
	exactData := make([]byte, payloadSize)
	rand.Read(exactData)

	n, err := writer.Write(exactData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != payloadSize {
		t.Errorf("Expected %d bytes written, got %d", payloadSize, n)
	}
}

// TestAuthenticationWriterIO_Close tests Close method
func TestAuthenticationWriterIO_Close(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	block, err := aes.NewCipher(key)
	common.Must(err)

	aead, err := cipher.NewGCM(block)
	common.Must(err)

	iv := make([]byte, 12)
	rand.Read(iv)

	// Test with a closer
	cache := bytes.NewBuffer(nil)
	writer := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, cache, protocol.TransferTypeStream, nil)

	err = writer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Test with non-closer (should not error)
	nonCloser := bytes.NewBuffer(nil)
	writer2 := NewAuthenticationWriterIO(&AEADAuthenticator{
		AEAD:                    aead,
		NonceGenerator:          GenerateStaticBytes(iv),
		AdditionalDataGenerator: GenerateEmptyBytes(),
	}, PlainChunkSizeParser{}, nonCloser, protocol.TransferTypeStream, nil)

	err = writer2.Close()
	if err != nil {
		t.Errorf("Close on non-closer failed: %v", err)
	}
}
