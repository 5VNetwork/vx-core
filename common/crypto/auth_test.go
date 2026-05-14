package crypto

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/protocol"

	"github.com/google/go-cmp/cmp"
)

type passthroughAuthenticatorMB struct{}

func (passthroughAuthenticatorMB) NonceSize() int { return 0 }

func (passthroughAuthenticatorMB) Overhead() int { return 0 }

func (passthroughAuthenticatorMB) Open(dst, cipherText []byte) ([]byte, error) {
	return append(dst, cipherText...), nil
}

func (passthroughAuthenticatorMB) Seal(dst, plainText []byte) ([]byte, error) {
	return append(dst, plainText...), nil
}

type failingOpenAuthenticatorMB struct {
	err error
}

func (a failingOpenAuthenticatorMB) NonceSize() int { return 0 }

func (a failingOpenAuthenticatorMB) Overhead() int { return 0 }

func (a failingOpenAuthenticatorMB) Open(dst, cipherText []byte) ([]byte, error) {
	return nil, a.err
}

func (a failingOpenAuthenticatorMB) Seal(dst, plainText []byte) ([]byte, error) {
	return append(dst, plainText...), nil
}

type offsetChunkSizeParserMB struct {
	offset uint16
}

func (p offsetChunkSizeParserMB) SizeBytes() int32 { return 2 }

func (p offsetChunkSizeParserMB) Encode(size uint16, b []byte) []byte {
	binary.BigEndian.PutUint16(b, size-p.offset)
	return b[:2]
}

func (p offsetChunkSizeParserMB) Decode(b []byte) (uint16, error) {
	return binary.BigEndian.Uint16(b), nil
}

func (p offsetChunkSizeParserMB) HasConstantOffset() uint16 { return p.offset }

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

func TestAuthenticationReader_PropagatesOpenError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("authenticate chunk")
	payload := []byte("ciphertext")
	frame := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(frame[:2], uint16(len(payload)))
	copy(frame[2:], payload)

	reader := NewAuthenticationReader(
		context.Background(),
		failingOpenAuthenticatorMB{err: expectedErr},
		PlainChunkSizeParser{},
		bytes.NewReader(frame),
		protocol.TransferTypeStream,
		nil,
	)

	_, err := reader.ReadMultiBuffer()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected read error %v, got %v", expectedErr, err)
	}
}

func TestAuthenticationReader_UsesLargeBufferPathWhenOffsetExceedsBufSize(t *testing.T) {
	t.Parallel()

	parser := offsetChunkSizeParserMB{offset: 1}
	payload := bytes.Repeat([]byte("a"), int(buf.Size)+1)
	frame := make([]byte, 2+len(payload))
	parser.Encode(uint16(len(payload)), frame[:2])
	copy(frame[2:], payload)

	reader := NewAuthenticationReader(
		context.Background(),
		passthroughAuthenticatorMB{},
		parser,
		bytes.NewReader(frame),
		protocol.TransferTypeStream,
		nil,
	)

	mb, err := reader.ReadMultiBuffer()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	defer buf.ReleaseMulti(mb)

	out := make([]byte, len(payload))
	buf.SplitBytes(mb, out)
	if !bytes.Equal(out, payload) {
		t.Fatal("payload mismatch")
	}
}
