package buf

import (
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/bytespool"
	"github.com/rs/zerolog/log"
)

const (
	Offset = 60
	// default buffer BufferSize
	BufferSize = 1024 * 8
	// available to write.
	Size = BufferSize - Offset
)

var pool = bytespool.GetPool(BufferSize)

// ownership represents the data owner of the buffer.
type ownership uint8

const (
	defaultPool ownership = 0
	unmanaged   ownership = 1
	bytespools  ownership = 2
)

// Buffer is a recyclable allocation of a byte array. Buffer.Release() recycles
// the buffer into an internal buffer pool, in order to recreate a buffer more
// quickly.
type Buffer struct {
	v         []byte
	start     int32
	end       int32
	ownership ownership
	release   func()
	// released is a single-shot guard for Release: it CASes 0 -> 1, so only
	// one goroutine ever runs the release body for a given Buffer even if
	// the same *Buffer is shared and Released concurrently.
	released uint32
}

func New() *Buffer {
	return &Buffer{
		v:     pool.Get().([]byte),
		start: Offset,
		end:   Offset,
	}
}

// NewForMinWritable returns a buffer that can grow by at least minWritable bytes
// from the current write position (same as after New), without reallocation.
// For minWritable <= Size this is equivalent to New().
func NewForMinWritable(minWritable int32) *Buffer {
	if minWritable <= Size {
		return New()
	}
	v := bytespool.Alloc(Offset + minWritable)
	return &Buffer{
		v:         v,
		start:     Offset,
		end:       Offset,
		ownership: bytespools,
	}
}

// NewWithSize creates a Buffer with 0 length and capacity with at least the given size.
func NewWithSize(size int32) *Buffer {
	return &Buffer{
		v:         bytespool.Alloc(size),
		ownership: bytespools,
	}
}

// FromBytes creates a Buffer with an existed bytearray
func FromBytes(data []byte) *Buffer {
	return &Buffer{
		v:         data,
		end:       int32(len(data)),
		ownership: unmanaged,
	}
}

func NewWithRelease(data []byte, release func()) *Buffer {
	return &Buffer{
		v:         data,
		end:       int32(len(data)),
		ownership: unmanaged,
		release:   release,
	}
}

// StackNew creates a new Buffer object on stack.
// This method is for buffers that is released in the same function.
func StackNew() Buffer {
	return Buffer{
		v: pool.Get().([]byte),
	}
}

// Release recycles the buffer.v into an internal buffer pool
// and resets start and end to 0.
func (b *Buffer) Release() {
	if b == nil {
		return
	}
	// Single-shot guard: only the first caller proceeds. Without this, two
	// concurrent Release() calls on a shared *Buffer can both observe b.v
	// as non-nil, both do `p := b.v; b.v = nil`, and either (a) Put a
	// typed-nil []byte into the pool, or (b) Put the same real slice
	// twice. (a) poisons the pool and produces the
	// `slice bounds out of range [60:0]` / `[0:0]` panics in Write/ReadOnce;
	// (b) silently hands the same backing array to two future buffers.
	if !atomic.CompareAndSwapUint32(&b.released, 0, 1) {
		return
	}
	if b.v == nil {
		return
	}

	if b.release != nil {
		b.release()
		b.release = nil
	}

	if b.ownership == unmanaged {
		return
	}

	p := b.v
	b.v = nil
	b.end = 0
	b.start = 0
	switch b.ownership {
	case defaultPool:
		// Belt-and-suspenders: pool only stores BufferSize-cap slices.
		if cap(p) == BufferSize {
			pool.Put(p[:BufferSize])
		} else {
			log.Warn().Int("cap", cap(p)).Msg("buffer is not of size BufferSize")
		}
	case bytespools:
		bytespool.Free(p)
	}
}

// Clear clears the content of the buffer, results an empty buffer with
// Len() = 0.
func (b *Buffer) Clear() {
	b.start = Offset
	b.end = Offset
}

// Byte returns the bytes at index.
func (b *Buffer) Byte(index int32) byte {
	return b.v[b.start+index]
}

// SetByte sets the byte value at index.
func (b *Buffer) SetByte(index int32, value byte) {
	b.v[b.start+index] = value
}

// Bytes returns the content bytes of this Buffer.
// Returns nil if the underlying storage has been released, mirroring Len() == 0.
// Panics with a descriptive message when start/end are inconsistent with v, so
// pool reuse / use-after-Release bugs surface clearly instead of as opaque
// "slice bounds out of range" runtime errors.
func (b *Buffer) Bytes() []byte {
	if b.v == nil {
		return nil
	}
	if b.start < 0 || b.end < b.start || int(b.end) > len(b.v) {
		log.Warn().Int32("start", b.start).Int32("end", b.end).
			Int("cap", len(b.v)).Uint8("ownership", uint8(b.ownership)).
			Msg("buf.Buffer: inconsistent state")
		return nil
	}
	return b.v[b.start:b.end]
}

// Extend increases the buffer size by n bytes, and returns the extended part.
// It panics if result size is larger than buf.Size.
// change b.end
func (b *Buffer) Extend(n int32) []byte {
	newEnd := b.end + n
	if newEnd > int32(len(b.v)) {
		panic("extending out of bound")
	}
	ext := b.v[b.end:newEnd]
	b.end = newEnd
	return ext
}

// BytesRange returns a slice of this buffer with given from and to boundary.
func (b *Buffer) BytesRange(from, to int32) []byte {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	return b.v[b.start+from : b.start+to]
}

// BytesFrom returns a slice of this Buffer starting from the given position.
func (b *Buffer) BytesFrom(from int32) []byte {
	if from < 0 {
		from += b.Len()
	}
	return b.v[b.start+from : b.end]
}

// BytesTo returns a slice of this Buffer from start to the given position.
func (b *Buffer) BytesTo(to int32) []byte {
	if to < 0 {
		to += b.Len()
	}
	return b.v[b.start : b.start+to]
}

// Resize cuts the buffer at the given position.
// [start+from, start+to]
func (b *Buffer) Resize(from, to int32) {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	if to < from {
		panic("Invalid slice")
	}
	b.end = b.start + to
	b.start += from
}

// AdvanceStart cuts the buffer at the given position.
// Changes [b.start]
func (b *Buffer) AdvanceStart(from int32) {
	if from < 0 {
		from += b.Len()
	}
	b.start += from
}

// Retreat moves the buffer start position back by v bytes.
// If retreat cannot be done because of out of bound, a new buffer is created.
// func Retreat(b *Buffer, v int32) *Buffer {
// 	if b.start-v < 0 {
// 		buf := NewWithSize(b.Len() + v)
// 		buf.start = v
// 		buf.end = v
// 		buf.Write(b.Bytes())
// 		buf.start = 0
// 		b.Release()
// 		return buf
// 	} else {
// 		b.start -= v
// 		return b
// 	}
// }

// RetreatStart moves the buffer start position back by v bytes.
// If retreat cannot be done because of out of bound, a new
// slice is allocated from the pool and it replace the exsiting one.
// If v is negative, it will be ignored.
func (b *Buffer) RetreatStart(v int32) {
	if v < 0 {
		return
	}
	if b.start-v < 0 {
		newBuffer := bytespool.Alloc(b.Len() + v)
		copy(newBuffer[v:], b.Bytes())
		l := b.Len()
		b.Release()
		b.v = newBuffer
		b.start = v
		b.end = v + l
		b.ownership = bytespools
	}
	b.start -= v
}

func (b *Buffer) ZeroAll() {
	clear(b.v)
}

func (b *Buffer) Zero(from, to int32) {
	clear(b.v[b.start+from : b.start+to])
}

// Len returns the length of the buffer content.
func (b *Buffer) Len() int32 {
	if b == nil {
		return 0
	}
	if b.v == nil {
		return 0
	}
	return b.end - b.start
}

// Cap returns the len/cap of the underlying slice.
// len - start
func (b *Buffer) Cap() int32 {
	if b == nil {
		return 0
	}
	return int32(len(b.v)) - b.start
}

// WritableSpace returns how many bytes can still be appended (e.g. via Extend)
// without growing the underlying slice.
func (b *Buffer) WritableSpace() int32 {
	if b == nil || b.v == nil {
		return 0
	}
	return int32(len(b.v)) - b.end
}

// IsEmpty returns true if the buffer is empty.
func (b *Buffer) IsEmpty() bool {
	return b.Len() == 0
}

// IsFull returns true if the buffer has no more room to grow.
func (b *Buffer) IsFull() bool {
	return b != nil && b.end == int32(len(b.v))
}

// Write implements Write method in io.Writer.
func (b *Buffer) Write(data []byte) (int, error) {
	nBytes := copy(b.v[b.end:], data)
	b.end += int32(nBytes)
	return nBytes, nil
}

// WriteByte writes a single byte into the buffer.
func (b *Buffer) WriteByte(v byte) error {
	if b.IsFull() {
		return errors.New("buffer full")
	}
	b.v[b.end] = v
	b.end++
	return nil
}

// WriteString implements io.StringWriter.
func (b *Buffer) WriteString(s string) (int, error) {
	return b.Write([]byte(s))
}

// ReadByte implements io.ByteReader
func (b *Buffer) ReadByte() (byte, error) {
	if b.start == b.end {
		return 0, io.EOF
	}

	nb := b.v[b.start]
	b.start++
	return nb, nil
}

// ReadBytes implements bufio.Reader.ReadBytes
func (b *Buffer) ReadBytes(length int32) ([]byte, error) {
	if b.end-b.start < length {
		return nil, io.EOF
	}

	nb := b.v[b.start : b.start+length]
	b.start += length
	return nb, nil
}

// Read implements io.Reader.Read().
func (b *Buffer) Read(data []byte) (int, error) {
	payload := b.Bytes()
	if len(payload) == 0 {
		return 0, io.EOF
	}
	nBytes := copy(data, payload)
	if int32(nBytes) == b.Len() {
		b.Clear()
	} else {
		b.start += int32(nBytes)
	}
	return nBytes, nil
}

func (b *Buffer) ReadOnce(reader io.Reader) (int64, error) {
	n, err := reader.Read(b.v[b.end:])
	b.end += int32(n)
	return int64(n), err
}

func (b *Buffer) ReadOnceWithSize(reader io.Reader, size int32) (int64, error) {
	end := b.end + size
	if end > int32(len(b.v)) {
		v := end
		return 0, fmt.Errorf("out of bound: %d", v)
	}
	n, err := reader.Read(b.v[b.end:end])
	b.end += int32(n)
	return int64(n), err
}

// ReadFullFrom reads exact size of bytes from given reader, or until error occurs.
func (b *Buffer) ReadFullFrom(reader io.Reader, size int32) (int64, error) {
	end := b.end + size
	if end > int32(len(b.v)) {
		v := end
		return 0, fmt.Errorf("out of bound: %d", v)
	}
	n, err := io.ReadFull(reader, b.v[b.end:end])
	b.end += int32(n)
	return int64(n), err
}

// String returns the string form of this Buffer.
func (b *Buffer) String() string {
	return string(b.Bytes())
}

func (b *Buffer) Available() int32 {
	return int32(len(b.v)) - b.end
}

// Clone returns a new Buffer with the same content as the original.
func (b *Buffer) Clone() *Buffer {
	buf := NewWithSize(b.Len())
	copy(buf.v[buf.start:], b.Bytes())
	buf.end = buf.start + b.Len()
	return buf
}
