package buf

import (
	"io"
)

// all data read from reader will be saved to history until stop
type MemoryReader struct {
	reader io.Reader

	history MultiBuffer
	stop    bool
}

func NewMemoryReader(reader io.Reader) *MemoryReader {
	return &MemoryReader{
		reader: reader,
	}
}

func (c *MemoryReader) History() MultiBuffer {
	return c.history
}

func (c *MemoryReader) StopMemorize() {
	c.stop = true
	ReleaseMulti(c.history)
}

func (c *MemoryReader) Read(p []byte) (int, error) {
	if c.stop {
		return c.reader.Read(p)
	}

	n, err := c.reader.Read(p)
	if err != nil {
		return n, err
	}

	c.history = MergeBytes(c.history, p[:n])
	return n, nil
}

func (c *MemoryReader) OkayToUnwrapReader() int {
	if c.stop {
		return 1
	}
	return 0
}
