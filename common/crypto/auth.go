package crypto

import (
	"bufio"
	"context"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/bytespool"
	"github.com/5vnetwork/vx-core/common/protocol"
)

type BytesGenerator func() []byte

func GenerateEmptyBytes() BytesGenerator {
	var b [1]byte
	return func() []byte {
		return b[:0]
	}
}

func GenerateStaticBytes(content []byte) BytesGenerator {
	return func() []byte {
		return content
	}
}

func GenerateIncreasingNonce(nonce []byte) BytesGenerator {
	c := append([]byte(nil), nonce...)
	return func() []byte {
		for i := range c {
			c[i]++
			if c[i] != 0 {
				break
			}
		}
		return c
	}
}

func GenerateInitialAEADNonce() BytesGenerator {
	return GenerateIncreasingNonce([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
}

type Authenticator interface {
	NonceSize() int
	Overhead() int
	Open(dst, cipherText []byte) ([]byte, error)
	Seal(dst, plainText []byte) ([]byte, error)
}

type AEADAuthenticator struct {
	cipher.AEAD
	NonceGenerator          BytesGenerator
	AdditionalDataGenerator BytesGenerator
}

func (v *AEADAuthenticator) Open(dst, cipherText []byte) ([]byte, error) {
	iv := v.NonceGenerator()
	if len(iv) != v.AEAD.NonceSize() {
		return nil, fmt.Errorf("invalid AEAD nonce size: %d", len(iv))
	}

	var additionalData []byte
	if v.AdditionalDataGenerator != nil {
		additionalData = v.AdditionalDataGenerator()
	}
	return v.AEAD.Open(dst, iv, cipherText, additionalData)
}

func (v *AEADAuthenticator) Seal(dst, plainText []byte) ([]byte, error) {
	iv := v.NonceGenerator()
	if len(iv) != v.AEAD.NonceSize() {
		return nil, fmt.Errorf("invalid AEAD nonce size: %d", len(iv))
	}

	var additionalData []byte
	if v.AdditionalDataGenerator != nil {
		additionalData = v.AdditionalDataGenerator()
	}
	return v.AEAD.Seal(dst, iv, plainText, additionalData), nil
}

type AuthenticationReader struct {
	auth         Authenticator
	reader       *buf.BufferedReader
	sizeParser   ChunkSizeDecoder
	sizeBytes    []byte
	transferType protocol.TransferType
	padding      PaddingLengthGenerator
	size         uint16
	sizeOffset   uint16
	paddingLen   uint16
	hasSize      bool
	done         bool
}

func NewAuthenticationReader(ctx context.Context, auth Authenticator, sizeParser ChunkSizeDecoder,
	reader io.Reader, transferType protocol.TransferType, paddingLen PaddingLengthGenerator) *AuthenticationReader {
	r := &AuthenticationReader{
		auth:         auth,
		sizeParser:   sizeParser,
		transferType: transferType,
		padding:      paddingLen,
		sizeBytes:    make([]byte, sizeParser.SizeBytes()),
	}
	if chunkSizeDecoderWithOffset, ok := sizeParser.(ChunkSizeDecoderWithOffset); ok {
		r.sizeOffset = chunkSizeDecoderWithOffset.HasConstantOffset()
	}
	if breader, ok := reader.(*buf.BufferedReader); ok {
		r.reader = breader
	} else {
		r.reader = &buf.BufferedReader{Reader: buf.NewReader(reader)}
	}
	return r
}

func (r *AuthenticationReader) readSize() (uint16, uint16, error) {
	if r.hasSize {
		r.hasSize = false
		return r.size, r.paddingLen, nil
	}
	if _, err := io.ReadFull(r.reader, r.sizeBytes); err != nil {
		return 0, 0, err
	}
	var padding uint16
	if r.padding != nil {
		padding = r.padding.NextPaddingLen()
	}
	size, err := r.sizeParser.Decode(r.sizeBytes)
	return size, padding, err
}

var errSoft = errors.New("waiting for more data")

func (r *AuthenticationReader) readBuffer(size int32, padding int32) (*buf.Buffer, error) {
	b := buf.New()
	if _, err := b.ReadFullFrom(r.reader, size); err != nil {
		b.Release()
		return nil, err
	}
	size -= padding
	rb, err := r.auth.Open(b.BytesTo(0), b.BytesTo(size))
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Resize(0, int32(len(rb)))
	return b, nil
}

func (r *AuthenticationReader) readInternal(soft bool, mb *buf.MultiBuffer) error {
	if soft && r.reader.BufferedBytes() < r.sizeParser.SizeBytes() {
		return errSoft
	}

	if r.done {
		return io.EOF
	}

	size, padding, err := r.readSize()
	if err != nil {
		return fmt.Errorf("failed to read size: %w", err)
	}

	if size+r.sizeOffset == uint16(r.auth.Overhead())+padding {
		r.done = true
		return io.EOF
	}

	effectiveSize := int32(size) + int32(r.sizeOffset)

	if soft && effectiveSize > r.reader.BufferedBytes() {
		r.size = size
		r.paddingLen = padding
		r.hasSize = true
		return errSoft
	}

	if size <= buf.Size {
		b, err := r.readBuffer(effectiveSize, int32(padding))
		if err != nil {
			return nil
		}
		*mb = append(*mb, b)
		return nil
	}

	payload := bytespool.Alloc(effectiveSize)
	defer bytespool.Free(payload)

	if _, err := io.ReadFull(r.reader, payload[:effectiveSize]); err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}

	effectiveSize -= int32(padding)

	rb, err := r.auth.Open(payload[:0], payload[:effectiveSize])
	if err != nil {
		return fmt.Errorf("failed to authenticate payload: %w", err)
	}

	*mb = buf.MergeBytes(*mb, rb)
	return nil
}

func (r *AuthenticationReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	const readSize = 16
	mb := make(buf.MultiBuffer, 0, readSize)
	if err := r.readInternal(false, &mb); err != nil {
		buf.ReleaseMulti(mb)
		// log.Ctx(r.ctx).Debug().Msgf("failed to read internal: %v", err)
		return nil, err
	}

	for i := 1; i < readSize; i++ {
		err := r.readInternal(true, &mb)
		if err == errSoft || err == io.EOF {
			// log.Ctx(r.ctx).Debug().Msgf("failed to read internal: %v", err)
			break
		}
		if err != nil {
			// log.Ctx(r.ctx).Debug().Msgf("failed to read internal: %v", err)
			buf.ReleaseMulti(mb)
			return nil, err
		}
	}

	return mb, nil
}

type AuthenticationReader1 struct {
	auth         Authenticator
	r            io.Reader
	reader       *bufio.Reader
	sizeParser   ChunkSizeDecoder
	sizeBytes    []byte
	transferType protocol.TransferType
	padding      PaddingLengthGenerator
	size         uint16
	sizeOffset   uint16
	paddingLen   uint16
	hasSize      bool
	done         bool
	leftOver     buf.MultiBuffer
}

func NewAuthenticationReader1(auth Authenticator, sizeParser ChunkSizeDecoder, reader io.Reader,
	transferType protocol.TransferType, paddingLen PaddingLengthGenerator) *AuthenticationReader1 {
	r := &AuthenticationReader1{
		auth:         auth,
		r:            reader,
		reader:       bufio.NewReader(reader),
		sizeParser:   sizeParser,
		transferType: transferType,
		padding:      paddingLen,
		sizeBytes:    make([]byte, sizeParser.SizeBytes()),
	}
	if chunkSizeDecoderWithOffset, ok := sizeParser.(ChunkSizeDecoderWithOffset); ok {
		r.sizeOffset = chunkSizeDecoderWithOffset.HasConstantOffset()
	}
	return r
}

func (r *AuthenticationReader1) readSize() (uint16, uint16, error) {
	if r.hasSize {
		r.hasSize = false
		return r.size, r.paddingLen, nil
	}
	if _, err := io.ReadFull(r.reader, r.sizeBytes); err != nil {
		return 0, 0, err
	}
	var padding uint16
	if r.padding != nil {
		padding = r.padding.NextPaddingLen()
	}
	size, err := r.sizeParser.Decode(r.sizeBytes)
	return size, padding, err
}

func (r *AuthenticationReader1) readBuffer(size int32, padding int32) (*buf.Buffer, error) {
	b := buf.New()
	if _, err := b.ReadFullFrom(r.reader, size); err != nil {
		b.Release()
		return nil, err
	}
	size -= padding
	rb, err := r.auth.Open(b.BytesTo(0), b.BytesTo(size))
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Resize(0, int32(len(rb)))
	return b, nil
}

func (r *AuthenticationReader1) readInternal(soft bool, mb *buf.MultiBuffer) error {
	if soft && int32(r.reader.Buffered()) < r.sizeParser.SizeBytes() {
		return errSoft
	}

	if r.done {
		return io.EOF
	}

	size, padding, err := r.readSize()
	if err != nil {
		return fmt.Errorf("failed to read size: %w", err)
	}

	if size+r.sizeOffset == uint16(r.auth.Overhead())+padding {
		r.done = true
		return io.EOF
	}

	effectiveSize := int32(size) + int32(r.sizeOffset)

	if soft && effectiveSize > int32(r.reader.Buffered()) {
		r.size = size
		r.paddingLen = padding
		r.hasSize = true
		return errSoft
	}

	if size <= buf.Size {
		b, err := r.readBuffer(effectiveSize, int32(padding))
		if err != nil {
			return nil
		}
		*mb = append(*mb, b)
		return nil
	}

	payload := bytespool.Alloc(effectiveSize)
	defer bytespool.Free(payload)

	if _, err := io.ReadFull(r.reader, payload[:effectiveSize]); err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}

	effectiveSize -= int32(padding)

	rb, err := r.auth.Open(payload[:0], payload[:effectiveSize])
	if err != nil {
		return fmt.Errorf("failed to authenticate payload: %w", err)
	}

	*mb = buf.MergeBytes(*mb, rb)
	return nil
}

func (r *AuthenticationReader1) Read(p []byte) (int, error) {
	if !r.leftOver.IsEmpty() {
		if r.transferType == protocol.TransferTypePacket {
			first := r.leftOver[0]
			defer first.Release()
			if len(r.leftOver) > 1 {
				r.leftOver = r.leftOver[1:]
			} else {
				r.leftOver = nil
			}
			return first.Read(p)
		}
		leftOver, n := buf.SplitBytes(r.leftOver, p)
		r.leftOver = leftOver
		return n, nil
	}

	const readSize = 16
	mb := make(buf.MultiBuffer, 0, readSize)
	if err := r.readInternal(false, &mb); err != nil {
		buf.ReleaseMulti(mb)
		// log.Ctx(r.ctx).Debug().Msgf("failed to read internal: %v", err)
		return 0, err
	}

	for i := 1; i < readSize; i++ {
		err := r.readInternal(true, &mb)
		if err == errSoft || err == io.EOF {
			// log.Ctx(r.ctx).Debug().Msgf("failed to read internal: %v", err)
			break
		}
		if err != nil {
			// log.Ctx(r.ctx).Debug().Msgf("failed to read internal: %v", err)
			buf.ReleaseMulti(mb)
			return 0, err
		}
	}

	if r.transferType == protocol.TransferTypePacket {
		if len(mb) == 0 {
			return 0, nil
		}
		first := mb[0]
		defer first.Release()
		if len(mb) > 1 {
			r.leftOver = mb[1:]
		}
		return first.Read(p)
	}
	leftOver, n := buf.SplitBytes(mb, p)
	r.leftOver = leftOver
	return n, nil
}

func (r *AuthenticationReader1) Close() error {
	if closer, ok := r.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type AuthenticationWriter struct {
	auth         Authenticator
	writer       buf.Writer
	sizeParser   ChunkSizeEncoder
	transferType protocol.TransferType
	padding      PaddingLengthGenerator
}

func NewAuthenticationWriter(auth Authenticator, sizeParser ChunkSizeEncoder, writer io.Writer,
	transferType protocol.TransferType, padding PaddingLengthGenerator) *AuthenticationWriter {
	w := &AuthenticationWriter{
		auth:         auth,
		writer:       buf.NewWriter(writer),
		sizeParser:   sizeParser,
		transferType: transferType,
	}
	if padding != nil {
		w.padding = padding
	}
	return w
}

func (w *AuthenticationWriter) seal(b []byte) (*buf.Buffer, error) {
	encryptedSize := int32(len(b) + w.auth.Overhead())
	var paddingSize int32
	if w.padding != nil {
		paddingSize = int32(w.padding.NextPaddingLen())
	}

	sizeBytes := w.sizeParser.SizeBytes()
	totalSize := sizeBytes + encryptedSize + paddingSize
	if totalSize > buf.Size {
		return nil, fmt.Errorf("size too large: %d", totalSize)
	}

	eb := buf.New()
	w.sizeParser.Encode(uint16(encryptedSize+paddingSize), eb.Extend(sizeBytes))
	if _, err := w.auth.Seal(eb.Extend(encryptedSize)[:0], b); err != nil {
		eb.Release()
		return nil, err
	}
	if paddingSize > 0 {
		// These paddings will send in clear text.
		// To avoid leakage of PRNG internal state, a cryptographically secure PRNG should be used.
		paddingBytes := eb.Extend(paddingSize)
		common.Must2(rand.Read(paddingBytes))
	}

	return eb, nil
}

func (w *AuthenticationWriter) writeStream(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)

	var maxPadding int32
	if w.padding != nil {
		maxPadding = int32(w.padding.MaxPaddingLen())
	}

	payloadSize := buf.Size - int32(w.auth.Overhead()) - w.sizeParser.SizeBytes() - maxPadding
	if len(mb)+10 > 64*1024*1024 {
		return errors.New("value too large")
	}
	sliceSize := len(mb) + 10
	mb2Write := make(buf.MultiBuffer, 0, sliceSize)

	temp := buf.New()
	defer temp.Release()

	rawBytes := temp.Extend(payloadSize)

	for {
		nb, nBytes := buf.SplitBytes(mb, rawBytes)
		mb = nb

		eb, err := w.seal(rawBytes[:nBytes])
		if err != nil {
			buf.ReleaseMulti(mb2Write)
			return err
		}
		mb2Write = append(mb2Write, eb)
		if mb.IsEmpty() {
			break
		}
	}

	return w.writer.WriteMultiBuffer(mb2Write)
}

func (w *AuthenticationWriter) writePacket(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)

	if len(mb)+1 > 64*1024*1024 {
		return errors.New("value too large")
	}
	sliceSize := len(mb) + 1
	mb2Write := make(buf.MultiBuffer, 0, sliceSize)

	for _, b := range mb {
		if b.IsEmpty() {
			continue
		}

		eb, err := w.seal(b.Bytes())
		if err != nil {
			continue
		}

		mb2Write = append(mb2Write, eb)
	}

	if mb2Write.IsEmpty() {
		return nil
	}

	return w.writer.WriteMultiBuffer(mb2Write)
}

// WriteMultiBuffer implements buf.Writer.
func (w *AuthenticationWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if mb.IsEmpty() {
		eb, err := w.seal([]byte{})
		common.Must(err)
		return w.writer.WriteMultiBuffer(buf.MultiBuffer{eb})
	}

	if w.transferType == protocol.TransferTypeStream {
		return w.writeStream(mb)
	}

	return w.writePacket(mb)
}

func (a *AuthenticationWriter) CloseWrite() error { return a.writer.CloseWrite() }

type AuthenticationWriterIO struct {
	auth         Authenticator
	writer       io.Writer
	sizeParser   ChunkSizeEncoder
	transferType protocol.TransferType
	padding      PaddingLengthGenerator
}

func NewAuthenticationWriterIO(auth Authenticator, sizeParser ChunkSizeEncoder, writer io.Writer,
	transferType protocol.TransferType, padding PaddingLengthGenerator) *AuthenticationWriterIO {
	return &AuthenticationWriterIO{
		auth:         auth,
		writer:       writer,
		sizeParser:   sizeParser,
		transferType: transferType,
		padding:      padding,
	}
}

func (w *AuthenticationWriterIO) seal(b []byte) (*buf.Buffer, error) {
	encryptedSize := int32(len(b) + w.auth.Overhead())
	var paddingSize int32
	if w.padding != nil {
		paddingSize = int32(w.padding.NextPaddingLen())
	}

	sizeBytes := w.sizeParser.SizeBytes()
	totalSize := sizeBytes + encryptedSize + paddingSize
	if totalSize > buf.Size {
		return nil, fmt.Errorf("size too large: %d", totalSize)
	}

	eb := buf.New()
	w.sizeParser.Encode(uint16(encryptedSize+paddingSize), eb.Extend(sizeBytes))
	if _, err := w.auth.Seal(eb.Extend(encryptedSize)[:0], b); err != nil {
		eb.Release()
		return nil, err
	}
	if paddingSize > 0 {
		// These paddings will send in clear text.
		// To avoid leakage of PRNG internal state, a cryptographically secure PRNG should be used.
		paddingBytes := eb.Extend(paddingSize)
		common.Must2(rand.Read(paddingBytes))
	}

	return eb, nil
}

func (w *AuthenticationWriterIO) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		sealed, err := w.seal([]byte{})
		if err != nil {
			return 0, err
		}
		_, err = w.writer.Write(sealed.Bytes())
		sealed.Release()
		return 0, err
	}

	if w.transferType == protocol.TransferTypeStream {
		return w.writeStream(p)
	}
	return w.writePacket(p)
}

func (w *AuthenticationWriterIO) writeStream(p []byte) (n int, err error) {
	var maxPadding int32
	if w.padding != nil {
		maxPadding = int32(w.padding.MaxPaddingLen())
	}

	payloadSize := buf.Size - int32(w.auth.Overhead()) - w.sizeParser.SizeBytes() - maxPadding

	temp := buf.New()
	defer temp.Release()
	rawBytes := temp.Extend(payloadSize)

	remaining := p
	written := 0

	for len(remaining) > 0 {
		nBytes := copy(rawBytes, remaining)

		sealed, err := w.seal(rawBytes[:nBytes])
		if err != nil {
			return written, err
		}

		_, err = w.writer.Write(sealed.Bytes())
		sealed.Release()
		if err != nil {
			return written, err
		}

		written += nBytes
		remaining = remaining[nBytes:]
	}

	return written, nil
}

func (w *AuthenticationWriterIO) writePacket(p []byte) (n int, err error) {
	sealed, err := w.seal(p)
	if err != nil {
		return 0, err
	}

	_, err = w.writer.Write(sealed.Bytes())
	sealed.Release()
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *AuthenticationWriterIO) Close() error {
	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
