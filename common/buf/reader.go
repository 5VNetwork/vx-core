package buf

import (
	"io"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/serial"
)

func readOneBuffer(r io.Reader) (*Buffer, error) {
	b := New()
	for i := 0; i < 64; i++ {
		_, err := b.ReadOnce(r)
		if !b.IsEmpty() {
			return b, nil
		}
		if err != nil {
			b.Release()
			return nil, err
		}
	}

	b.Release()
	return nil, errors.New("Reader returns too many empty payloads.")
}

// ReadBuffer reads a Buffer from the given reader.
func ReadBuffer(r io.Reader) (*Buffer, error) {
	b := New()
	n, err := b.ReadOnce(r)
	if n > 0 {
		return b, err
	}
	b.Release()
	return nil, err
}

// BufferedReader is a Reader that keeps its internal buffer.
type BufferedReader struct {
	// Reader is the underlying reader to be read from
	Reader Reader
	// Buffer is the internal buffer to be read from first
	Buffer MultiBuffer
	// Spliter is a function to read bytes from MultiBuffer
	Spliter func(MultiBuffer, []byte) (MultiBuffer, int)
}

func (r *BufferedReader) OkayToUnwrapReader() int {
	if r.Buffer.IsEmpty() {
		return 1
	}
	return 0
}

func (r *BufferedReader) UnwrapReader() any {
	return r.Reader
}

// BufferedBytes returns the number of bytes that is cached in this reader.
func (r *BufferedReader) BufferedBytes() int32 {
	return r.Buffer.Len()
}

// ReadByte implements io.ByteReader.
func (r *BufferedReader) ReadByte() (byte, error) {
	var b [1]byte
	_, err := r.Read(b[:])
	return b[0], err
}

// Read implements io.Reader. It reads from internal buffer first (if available) and then reads from the underlying reader.
// If the internal buffer is empty, it first reads a MultiBuffer from the underlying reader, then split the MultiBuffer based on b
func (r *BufferedReader) Read(b []byte) (int, error) {
	spliter := r.Spliter
	if spliter == nil {
		spliter = SplitBytes
	}

	if !r.Buffer.IsEmpty() {
		buffer, nBytes := spliter(r.Buffer, b)
		r.Buffer = buffer
		if r.Buffer.IsEmpty() {
			r.Buffer = nil
		}
		return nBytes, nil
	}

	mb, err := r.Reader.ReadMultiBuffer()
	if !mb.IsEmpty() {
		mb, nBytes := spliter(mb, b)
		if !mb.IsEmpty() {
			r.Buffer = mb
		}
		return nBytes, err
	}
	return 0, err
}

// ReadMultiBuffer implements Reader.
func (r *BufferedReader) ReadMultiBuffer() (MultiBuffer, error) {
	if !r.Buffer.IsEmpty() {
		mb := r.Buffer
		r.Buffer = nil
		return mb, nil
	}

	return r.Reader.ReadMultiBuffer()
}

// ReadAtMost returns a MultiBuffer with at most size.
func (r *BufferedReader) ReadAtMost(size int32) (MultiBuffer, error) {
	if r.Buffer.IsEmpty() {
		mb, err := r.Reader.ReadMultiBuffer()
		if mb.IsEmpty() && err != nil {
			return nil, err
		}
		r.Buffer = mb
	}

	rb, mb := SplitSize(r.Buffer, size)
	r.Buffer = rb
	if r.Buffer.IsEmpty() {
		r.Buffer = nil
	}
	return mb, nil
}

func (r *BufferedReader) writeToInternal(writer io.Writer) (int64, error) {
	mbWriter := NewWriter(writer)
	var sc SizeCounter
	if r.Buffer != nil {
		sc.Size = int64(r.Buffer.Len())
		if err := mbWriter.WriteMultiBuffer(r.Buffer); err != nil {
			return 0, err
		}
		r.Buffer = nil
	}

	err := Copy(r.Reader, mbWriter, CountSize(&sc))
	return sc.Size, err
}

// WriteTo implements io.WriterTo.
func (r *BufferedReader) WriteTo(writer io.Writer) (int64, error) {
	nBytes, err := r.writeToInternal(writer)
	if errors.As(err, io.EOF) {
		return nBytes, nil
	}
	return nBytes, err
}

func (r *BufferedReader) Close() error {
	return common.Close(r.Reader)
}

// SingleReader is a Reader that read one Buffer every time.
type SingleReader struct {
	io.Reader
}

func (r *SingleReader) OkayToUnwrapReader() int {
	return 1
}
func (r *SingleReader) UnwrapReader() any {
	return r.Reader
}

// ReadMultiBuffer implements Reader.
func (r *SingleReader) ReadMultiBuffer() (MultiBuffer, error) {
	b, err := ReadBuffer(r.Reader)
	if b != nil {
		return MultiBuffer{b}, nil
	}
	return nil, err
}

// PacketReader is a Reader that read one Buffer every time.
// If no pakcet has arrived, it keeps waiting until one arrives.
type PacketReader struct {
	io.Reader
}

// ReadMultiBuffer implements Reader.
func (r *PacketReader) ReadMultiBuffer() (MultiBuffer, error) {
	b, err := readOneBuffer(r.Reader)
	if err != nil {
		return nil, err
	}
	return MultiBuffer{b}, nil
}

func NewLengthPacketReader(reader io.Reader) *LengthPacketReader {
	return &LengthPacketReader{
		Reader: reader,
	}
}

type LengthPacketReader struct {
	io.Reader
}

func (r *LengthPacketReader) ReadMultiBuffer() (MultiBuffer, error) {
	length, err := serial.ReadUint16(r.Reader)
	if err != nil {
		return nil, errors.New("failed to read packet length").Base(err)
	}
	b := NewWithSize(int32(length))
	if _, err := b.ReadFullFrom(r.Reader, int32(length)); err != nil {
		return nil, errors.New("failed to read packet payload").Base(err)
	}
	return MultiBuffer{b}, nil
}

// UnwrapReader is non-nil until unwrapped.
// For using readv
type UnwrapReaderReader struct {
	Reader       Reader
	Unwrapped    bool
	UnwrapReader UnwrapReader
}

func (r *UnwrapReaderReader) ReadMultiBuffer() (MultiBuffer, error) {
	if r.Unwrapped {
		return r.Reader.ReadMultiBuffer()
	}

	reader := Unwrap(r.UnwrapReader)
	if unwrap, ok := reader.(UnwrapReader); ok && unwrap.OkayToUnwrapReader() != -1 {
		// unwrap.OkayToUnwrapReader is 0 at this point
		r.UnwrapReader = unwrap
		return r.Reader.ReadMultiBuffer()
	} else {
		r.Unwrapped = true
		r.UnwrapReader = nil
		// if the innermost reader is still a buf.Reader, nothing to do
		if _, ok := reader.(Reader); ok {
			return r.Reader.ReadMultiBuffer()
		} else {
			rd := NewReader(reader.(io.Reader))
			if rdv, ok := rd.(*ReadVReader); ok {
				r.Reader = rdv
				return r.Reader.ReadMultiBuffer()
			}
			return r.Reader.ReadMultiBuffer()
		}
	}
}

// unwrap reader to maximum extent
func Unwrap(unwrapReader UnwrapReader) any {
	var reader any
	reader = unwrapReader
	for {
		if unwrapReader.OkayToUnwrapReader() == 1 {
			reader = unwrapReader.UnwrapReader()
			if unwrap, ok := reader.(UnwrapReader); ok {
				unwrapReader = unwrap
			} else {
				break
			}
		} else {
			return reader
		}
	}
	return reader
}
