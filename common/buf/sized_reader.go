package buf

import (
	"io"

	"github.com/5vnetwork/vx-core/common/serial"
)

// implement buf.Reader
// read 2 bytes (uint16, the num of bytes) from the reader first,
// then read "the num of bytes" from the reader, after this,
// further ReadMultiBuffer() will return io.EOF
type SizedReader struct {
	reader           *BufferedReader
	numOfBytesToRead int
}

func NewSizedReader(reader *BufferedReader) Reader {
	return &SizedReader{
		reader:           reader,
		numOfBytesToRead: -1,
	}
}

func (r *SizedReader) ReadMultiBuffer() (MultiBuffer, error) {
	if r.numOfBytesToRead == 0 {
		return nil, io.EOF
	}

	if r.numOfBytesToRead == -1 {
		size, err := serial.ReadUint16(r.reader)
		if err != nil {
			return nil, err
		}
		r.numOfBytesToRead = int(size)
	}

	mb, err := r.reader.ReadAtMost(int32(r.numOfBytesToRead))
	if !mb.IsEmpty() {
		r.numOfBytesToRead -= int(mb.Len())
		return mb, nil
	}

	return nil, err
}
