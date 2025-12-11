package crypto

import (
	"bufio"
	"io"
)

type ChunkStreamReaderIO struct {
	sizeDecoder ChunkSizeDecoder
	reader      *bufio.Reader

	buffer       []byte
	leftOverSize int32
	maxNumChunk  uint32
	numChunk     uint32
}

func NewChunkStreamReaderIO(sizeDecoder ChunkSizeDecoder, reader io.Reader) *ChunkStreamReader {
	return NewChunkStreamReaderWithChunkCount(sizeDecoder, reader, 0)
}

func NewChunkStreamReaderWithChunkCountIO(sizeDecoder ChunkSizeDecoder,
	reader io.Reader, maxNumChunk uint32) *ChunkStreamReaderIO {
	r := &ChunkStreamReaderIO{
		sizeDecoder: sizeDecoder,
		buffer:      make([]byte, sizeDecoder.SizeBytes()),
		maxNumChunk: maxNumChunk,
	}
	if breader, ok := reader.(*bufio.Reader); ok {
		r.reader = breader
	} else {
		r.reader = bufio.NewReader(reader)
	}
	return r
}

func (r *ChunkStreamReaderIO) readSize() (uint16, error) {
	if _, err := io.ReadFull(r.reader, r.buffer); err != nil {
		return 0, err
	}
	return r.sizeDecoder.Decode(r.buffer)
}

func (r *ChunkStreamReaderIO) Read(p []byte) (int, error) {
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

	// Read the minimum of leftOverSize or len(p)
	toRead := int(r.leftOverSize)
	if toRead > len(p) {
		toRead = len(p)
	}

	n, err := r.reader.Read(p[:toRead])
	if n > 0 {
		r.leftOverSize -= int32(n)
	}
	return n, err
}
