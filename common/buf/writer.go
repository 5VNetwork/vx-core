package buf

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/5vnetwork/vx-core/common"
)

// BufferToBytesWriter is a Writer that writes alloc.Buffer into underlying writer.
// used in tcp. Writer is typically a net.Conn
type BufferToBytesWriter struct {
	io.Writer

	cache [][]byte
}

func (w *BufferToBytesWriter) OkayToUnwrapWriter() int {
	return 1
}
func (w *BufferToBytesWriter) UnwrapWriter() any {
	return w.Writer
}

func (w *BufferToBytesWriter) CloseWrite() error {
	conn, ok := w.Writer.(CloseWriter)
	if ok {
		return conn.CloseWrite()
	}
	return nil
}

// WriteMultiBuffer implements Writer. This method takes ownership of the given buffer.
// transfer multi buffer to net.buffers and use net.Buffers.WriteTo to write to the underlying io.writer
func (w *BufferToBytesWriter) WriteMultiBuffer(mb MultiBuffer) error {
	defer ReleaseMulti(mb)

	size := mb.Len() //total bytes to write
	if size == 0 {
		return nil
	}

	if len(mb) == 1 {
		return WriteAllBytes(w.Writer, mb[0].Bytes())
	}

	if cap(w.cache) < len(mb) {
		w.cache = make([][]byte, 0, len(mb))
	}

	bs := w.cache
	for _, b := range mb {
		bs = append(bs, b.Bytes())
	}

	defer func() {
		for idx := range bs {
			bs[idx] = nil
		}
	}()

	nb := net.Buffers(bs)

	for size > 0 {
		n, err := nb.WriteTo(w.Writer)
		if err != nil {
			return err
		}
		size -= int32(n)
	}

	return nil
}

// ReadFrom implements io.ReaderFrom.
func (w *BufferToBytesWriter) ReadFrom(reader io.Reader) (int64, error) {
	var sc SizeCounter
	err := Copy(NewReader(reader), w, CountSize(&sc))
	return sc.Size, err
}

// BufferedWriter is a Writer with internal buffer.
type BufferedWriter struct {
	sync.Mutex
	Writer
	buffer   *Buffer
	buffered bool
}

// NewBufferedWriter creates a new BufferedWriter.
func NewBufferedWriter(writer Writer) *BufferedWriter {
	return &BufferedWriter{
		Writer:   writer,
		buffer:   New(),
		buffered: true,
	}
}

func (w *BufferedWriter) OkayToUnwrapWriter() int {
	if w.buffered {
		return 0
	}
	return 1
}
func (w *BufferedWriter) UnwrapWriter() any {
	return w.Writer
}

func (w *BufferedWriter) Buffered() bool {
	return w.buffered
}

// WriteByte implements io.ByteWriter.
func (w *BufferedWriter) WriteByte(c byte) error {
	_, err := w.Write([]byte{c})
	return err
}

// Write implements io.Writer.
// If w.buffered is true, the given bytes will be written into internal buffer,
// if after writing the buffer is full,
// it will be flushed into underlying writer.
func (w *BufferedWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	w.Lock()
	defer w.Unlock()

	if !w.buffered {
		if writer, ok := w.Writer.(io.Writer); ok {
			return writer.Write(b)
		}
	}

	totalBytes := 0
	for len(b) > 0 {
		if w.buffer == nil {
			w.buffer = New()
		}

		nBytes, err := w.buffer.Write(b)
		totalBytes += nBytes
		if err != nil {
			return totalBytes, err
		}
		if !w.buffered || w.buffer.IsFull() {
			if err := w.flushInternal(); err != nil {
				return totalBytes, err
			}
		}
		b = b[nBytes:]
	}

	return totalBytes, nil
}

// WriteMultiBuffer implements Writer. It takes ownership of the given MultiBuffer.
// If w.buffered is true, the given MultiBuffer will be written into internal buffer,
// if after writing the buffer is full, it will be flushed into underlying writer;
// if the buffer is not full, that's it, the buffer is not written into the underlying writer.
func (w *BufferedWriter) WriteMultiBuffer(b MultiBuffer) error {
	if b.IsEmpty() {
		return nil
	}

	w.Lock()
	defer w.Unlock()

	if !w.buffered {
		return w.Writer.WriteMultiBuffer(b)
	}

	reader := MultiBufferContainer{
		MultiBuffer: b,
	}
	defer reader.Close()

	for !reader.MultiBuffer.IsEmpty() {
		if w.buffer == nil {
			w.buffer = New()
		}
		common.Must2(w.buffer.ReadOnce(&reader))
		if w.buffer.IsFull() {
			if err := w.flushInternal(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flush flushes buffered content into underlying writer.
func (w *BufferedWriter) Flush() error {
	w.Lock()
	defer w.Unlock()

	return w.flushInternal()
}

func (w *BufferedWriter) flushInternal() error {
	if w.buffer.IsEmpty() {
		return nil
	}

	b := w.buffer
	w.buffer = nil

	if writer, ok := w.Writer.(io.Writer); ok {
		err := WriteAllBytes(writer, b.Bytes())
		b.Release()
		return err
	}

	return w.Writer.WriteMultiBuffer(MultiBuffer{b})
}

// SetBuffered sets whether the internal buffer is used. If set to false, Flush() will be called to clear the buffer.
func (w *BufferedWriter) SetBuffered(f bool) error {
	w.Lock()
	defer w.Unlock()

	w.buffered = f
	if !f {
		return w.flushInternal()
	}
	return nil
}

// ReadFrom implements io.ReaderFrom.
func (w *BufferedWriter) ReadFrom(reader io.Reader) (int64, error) {
	if err := w.SetBuffered(false); err != nil {
		return 0, err
	}

	var sc SizeCounter
	err := Copy(NewReader(reader), w, CountSize(&sc))
	return sc.Size, err
}

func (w *BufferedWriter) CloseWrite() error {
	return w.Close()
}

// Close implements io.Closable.
func (w *BufferedWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	return w.Writer.CloseWrite()
}

// SequentialWriter is a Writer that writes MultiBuffer sequentially into the underlying io.Writer.
// Write buffer one by one
type SequentialWriter struct {
	io.Writer
}

func (w *SequentialWriter) OkayToUnwrapWriter() int {
	return 1
}
func (w *SequentialWriter) UnwrapWriter() any {
	return w.Writer
}

func (w *SequentialWriter) CloseWrite() error { return nil }

// WriteMultiBuffer implements Writer.
func (w *SequentialWriter) WriteMultiBuffer(mb MultiBuffer) error {
	mb, err := WriteMultiBuffer(w.Writer, mb)
	ReleaseMulti(mb)
	return err
}

type NoOpWriter byte

func (NoOpWriter) CloseWrite() error { return nil }

func (NoOpWriter) WriteMultiBuffer(b MultiBuffer) error {
	ReleaseMulti(b)
	return nil
}

func (NoOpWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (NoOpWriter) ReadFrom(reader io.Reader) (int64, error) {
	b := New()
	defer b.Release()

	totalBytes := int64(0)
	for {
		b.Clear()
		_, err := b.ReadOnce(reader)
		totalBytes += int64(b.Len())
		if err != nil {
			if errors.Is(err, io.EOF) {
				return totalBytes, nil
			}
			return totalBytes, err
		}
	}
}

func (NoOpWriter) ReadFullFrom(reader io.Reader, size int32) (int64, error) {
	b := New()
	defer b.Release()

	totalBytes := int64(0)
	for {
		b.Clear()
		n, err := b.ReadFullFrom(reader, min(b.Cap(), size))
		totalBytes += n
		size -= int32(n)
		if err != nil {
			return int64(totalBytes), err
		}
		if size <= 0 {
			return int64(totalBytes), nil
		}
	}
}

func NewMultiLengthPacketWriter(writer Writer) *MultiLengthPacketWriter {
	return &MultiLengthPacketWriter{
		Writer: writer,
	}
}

// create a new mb
// for each buffer in the old multi buffer, write the length of the buffer first, then write the buffer
type MultiLengthPacketWriter struct {
	Writer
}

func (w *MultiLengthPacketWriter) WriteMultiBuffer(mb MultiBuffer) error {
	for _, b := range mb {
		length := b.Len()
		if length == 0 {
			continue
		}
		b.RetreatStart(2)
		b.SetByte(0, byte(length>>8))
		b.SetByte(1, byte(length))
	}
	mb = RemoveEmptyBuffer(mb)
	return w.Writer.WriteMultiBuffer(mb)
}

var (
	// Discard is a Writer that swallows all contents written in.
	Discard Writer = NoOpWriter(0)

	// DiscardBytes is an io.Writer that swallows all contents written in.
	DiscardBytes io.Writer = NoOpWriter(0)

	DiscardReader NoOpWriter = NoOpWriter(0)
)

// RemoveEmptyBuffer removes empty buffers from the given MultiBuffer.
// The removed buffers are released.
func RemoveEmptyBuffer(mb MultiBuffer) MultiBuffer {
	n := 0
	for _, b := range mb {
		if b.Len() == 0 {
			b.Release()
			continue
		}
		mb[n] = b
		n++
	}
	return mb[:n]
}
