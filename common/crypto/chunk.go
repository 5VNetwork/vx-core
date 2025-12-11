package crypto

import (
	"encoding/binary"
	"io"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/bytespool"
)

// ChunkSizeDecoder is a utility class to decode size value from bytes.
type ChunkSizeDecoder interface {
	// SizeBytes must be stable, return the same value across all calls
	SizeBytes() int32
	Decode([]byte) (uint16, error)
}

type ChunkSizeDecoderWithOffset interface {
	ChunkSizeDecoder
	// HasConstantOffset set the constant offset of Decode
	// The effective size should be HasConstantOffset() + Decode(_).[0](uint64)
	HasConstantOffset() uint16
}

// ChunkSizeEncoder is a utility class to encode size value into bytes.
type ChunkSizeEncoder interface {
	SizeBytes() int32
	Encode(uint16, []byte) []byte
}

type PaddingLengthGenerator interface {
	MaxPaddingLen() uint16
	NextPaddingLen() uint16
}

type PlainChunkSizeParser struct{}

func (PlainChunkSizeParser) SizeBytes() int32 {
	return 2
}

func (PlainChunkSizeParser) Encode(size uint16, b []byte) []byte {
	binary.BigEndian.PutUint16(b, size)
	return b[:2]
}

func (PlainChunkSizeParser) Decode(b []byte) (uint16, error) {
	return binary.BigEndian.Uint16(b), nil
}

type AEADChunkSizeParser struct {
	Auth *AEADAuthenticator
}

func (p *AEADChunkSizeParser) SizeBytes() int32 {
	return 2 + int32(p.Auth.Overhead())
}

func (p *AEADChunkSizeParser) Encode(size uint16, b []byte) []byte {
	binary.BigEndian.PutUint16(b, size-uint16(p.Auth.Overhead()))
	b, err := p.Auth.Seal(b[:0], b[:2])
	common.Must(err)
	return b
}

func (p *AEADChunkSizeParser) Decode(b []byte) (uint16, error) {
	b, err := p.Auth.Open(b[:0], b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b) + uint16(p.Auth.Overhead()), nil
}

type ChunkStreamReader struct {
	sizeDecoder ChunkSizeDecoder
	reader      *buf.BufferedReader

	buffer       []byte
	leftOverSize int32
	maxNumChunk  uint32
	numChunk     uint32
}

func NewChunkStreamReader(sizeDecoder ChunkSizeDecoder, reader io.Reader) *ChunkStreamReader {
	return NewChunkStreamReaderWithChunkCount(sizeDecoder, reader, 0)
}

func NewChunkStreamReaderWithChunkCount(sizeDecoder ChunkSizeDecoder, reader io.Reader, maxNumChunk uint32) *ChunkStreamReader {
	r := &ChunkStreamReader{
		sizeDecoder: sizeDecoder,
		buffer:      make([]byte, sizeDecoder.SizeBytes()),
		maxNumChunk: maxNumChunk,
	}
	if breader, ok := reader.(*buf.BufferedReader); ok {
		r.reader = breader
	} else {
		r.reader = &buf.BufferedReader{Reader: buf.NewReader(reader)}
	}

	return r
}

func (r *ChunkStreamReader) readSize() (uint16, error) {
	if _, err := io.ReadFull(r.reader, r.buffer); err != nil {
		return 0, err
	}
	return r.sizeDecoder.Decode(r.buffer)
}

func (r *ChunkStreamReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	size := r.leftOverSize
	if size == 0 {
		r.numChunk++
		if r.maxNumChunk > 0 && r.numChunk > r.maxNumChunk {
			return nil, io.EOF
		}
		nextSize, err := r.readSize()
		if err != nil {
			return nil, err
		}
		if nextSize == 0 {
			return nil, io.EOF
		}
		size = int32(nextSize)
	}
	r.leftOverSize = size

	mb, err := r.reader.ReadAtMost(size)
	if !mb.IsEmpty() {
		r.leftOverSize -= mb.Len()
		return mb, nil
	}
	return nil, err
}

type ChunkStreamReader1 struct {
	sizeDecoder  ChunkSizeDecoder
	reader       io.Reader
	buffer       []byte
	leftOverSize int32
	maxNumChunk  uint32
	numChunk     uint32
}

func NewChunkStreamReader1(sizeDecoder ChunkSizeDecoder, reader io.Reader, maxNumChunk uint32) *ChunkStreamReader1 {
	return &ChunkStreamReader1{
		sizeDecoder: sizeDecoder,
		reader:      reader,
		buffer:      make([]byte, sizeDecoder.SizeBytes()),
		maxNumChunk: maxNumChunk,
	}
}

func (r *ChunkStreamReader1) readSize() (uint16, error) {
	if _, err := io.ReadFull(r.reader, r.buffer); err != nil {
		return 0, err
	}
	return r.sizeDecoder.Decode(r.buffer)
}

func (r *ChunkStreamReader1) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	size := r.leftOverSize
	if size == 0 {
		r.numChunk++
		if r.maxNumChunk > 0 && r.numChunk > r.maxNumChunk {
			return 0, io.EOF
		}
		nextSize, err := r.readSize()
		if err != nil {
			return 0, err
		}
		if nextSize == 0 {
			return 0, io.EOF
		}
		size = int32(nextSize)
	}
	r.leftOverSize = size

	n, err = io.ReadFull(r.reader, p[:min(size, int32(len(p)))])
	if n > 0 {
		r.leftOverSize -= int32(n)
		return n, nil
	}

	return 0, err
}

func (r *ChunkStreamReader1) Close() error {
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type ChunkStreamWriter struct {
	sizeEncoder ChunkSizeEncoder
	writer      buf.Writer
}

func NewChunkStreamWriter(sizeEncoder ChunkSizeEncoder, writer io.Writer) *ChunkStreamWriter {
	return &ChunkStreamWriter{
		sizeEncoder: sizeEncoder,
		writer:      buf.NewWriter(writer),
	}
}

func (w *ChunkStreamWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	const sliceSize = buf.BufferSize
	mbLen := mb.Len()
	mb2Write := make(buf.MultiBuffer, 0, mbLen/buf.Size+mbLen/sliceSize+2)

	for {
		mb2, slice := buf.SplitSize(mb, sliceSize)
		mb = mb2

		b := buf.New()
		// w.sizeEncoder.SizeBytes() can be 2
		w.sizeEncoder.Encode(uint16(slice.Len()), b.Extend(w.sizeEncoder.SizeBytes()))
		mb2Write = append(mb2Write, b)
		mb2Write = append(mb2Write, slice...)

		if mb.IsEmpty() {
			break
		}
	}

	return w.writer.WriteMultiBuffer(mb2Write)
}

func (*ChunkStreamWriter) CloseWrite() error { return nil }

type ChunkStreamWriter1 struct {
	sizeEncoder ChunkSizeEncoder
	writer      io.Writer
}

func NewChunkStreamWriter1(sizeEncoder ChunkSizeEncoder, writer io.Writer) *ChunkStreamWriter1 {
	return &ChunkStreamWriter1{
		sizeEncoder: sizeEncoder,
		writer:      writer,
	}
}

func (w *ChunkStreamWriter1) Write(p []byte) (n int, err error) {
	const sliceSize = buf.BufferSize
	remaining := p
	written := 0
	sizeHeader := bytespool.Alloc(2048)
	defer bytespool.Free(sizeHeader)

	for {
		// Determine chunk size
		currentSize := len(remaining)
		if currentSize > sliceSize {
			currentSize = sliceSize
		}

		// Create size header
		size := w.sizeEncoder.SizeBytes()
		w.sizeEncoder.Encode(uint16(currentSize), sizeHeader[:size])

		// Write size header
		_, err = w.writer.Write(sizeHeader[:size])
		if err != nil {
			return written, err
		}

		// Write chunk
		n, err = w.writer.Write(remaining[:currentSize])
		written += n
		if err != nil {
			return written, err
		}

		remaining = remaining[currentSize:]
		if len(remaining) == 0 {
			break
		}
	}

	return written, nil
}

func (w *ChunkStreamWriter1) Close() error {
	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
