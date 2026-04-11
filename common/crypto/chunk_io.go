package crypto

import (
	"io"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/bytespool"
)

type ChunkStreamReaderIO struct {
	sizeDecoder  ChunkSizeDecoder
	reader       io.Reader
	buffer       []byte
	leftOverSize int32
	maxNumChunk  uint32
	numChunk     uint32
}

func NewChunkStreamReader1(sizeDecoder ChunkSizeDecoder, reader io.Reader, maxNumChunk uint32) *ChunkStreamReaderIO {
	return &ChunkStreamReaderIO{
		sizeDecoder: sizeDecoder,
		reader:      reader,
		buffer:      make([]byte, sizeDecoder.SizeBytes()),
		maxNumChunk: maxNumChunk,
	}
}

func (r *ChunkStreamReaderIO) readSize() (uint16, error) {
	if _, err := io.ReadFull(r.reader, r.buffer); err != nil {
		return 0, err
	}
	return r.sizeDecoder.Decode(r.buffer)
}

func (r *ChunkStreamReaderIO) Read(p []byte) (n int, err error) {
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

func (r *ChunkStreamReaderIO) Close() error {
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type ChunkStreamWriterIO struct {
	sizeEncoder ChunkSizeEncoder
	writer      io.Writer
}

func NewChunkStreamWriterIO(sizeEncoder ChunkSizeEncoder, writer io.Writer) *ChunkStreamWriterIO {
	return &ChunkStreamWriterIO{
		sizeEncoder: sizeEncoder,
		writer:      writer,
	}
}

func (w *ChunkStreamWriterIO) Write(p []byte) (n int, err error) {
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

func (w *ChunkStreamWriterIO) Close() error {
	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
