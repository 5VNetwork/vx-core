package buf

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

type ReaderWriter interface {
	Reader
	Writer
}

// Reader extends io.Reader with MultiBuffer.
type Reader interface {
	// Same as io.Reader.Read, MultiBuffer might be not empty and error is not-nil
	ReadMultiBuffer() (MultiBuffer, error)
}

type UnwrapReader interface {
	// 1 means okay, 0 means not yet, -1 means not supported
	OkayToUnwrapReader() int
	// return either a Reader or io.Reader. Reading from this unwrapped reader
	// get same result as reading from the original Reader.
	UnwrapReader() any
}

type UnwrapWriter interface {
	// 1 means okay, 0 means not yet, -1 means not supported
	OkayToUnwrapWriter() int
	// return either a Writer or io.Writer. Writing to this unwrapped writer
	// is same as writing to the original Writer.
	UnwrapWriter() any
}

// ErrReadTimeout is an error that happens with IO timeout.
var ErrReadTimeout = errors.New("IO timeout")
var ErrWriteTimeout = errors.New("IO timeout")

// TimeoutReader is a reader that returns error if Read() operation takes longer than the given timeout.
type TimeoutReader interface {
	ReadMultiBufferTimeout(time.Duration) (MultiBuffer, error)
}

// Writer extends io.Writer with MultiBuffer.
type Writer interface {
	// WriteMultiBuffer writes a MultiBuffer into underlying writer.
	// Writer releases the MultiBuffer anyway(even if it fails).
	WriteMultiBuffer(MultiBuffer) error
	CloseWrite() error
}

type CloseWriter interface {
	CloseWrite() error
}

// WriteAllBytes ensures all bytes are written into the given writer.
func WriteAllBytes(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		n, err := writer.Write(payload)
		if err != nil {
			return err
		}
		payload = payload[n:]
	}
	return nil
}

func isPacketConn(reader io.Reader) bool {
	_, ok := reader.(net.PacketConn)
	return ok
}

// NewReader creates a new Reader.
// From a io.Reader to a buf.Reader
// The Reader instance doesn't take the ownership of reader.
func NewReader(reader io.Reader) Reader { //net.Conn
	if mr, ok := reader.(Reader); ok {
		return mr
	}

	if isPacketConn(reader) {
		return &PacketReader{
			Reader: reader,
		}
	}

	_, isFile := reader.(*os.File)
	if !isFile && useReadv {
		if sc, ok := reader.(syscall.Conn); ok {
			rawConn, err := sc.SyscallConn()
			if err != nil {
				log.Print("failed to get sysconn", err)
			} else {
				return NewReadVReader(reader, rawConn)
			}
		}
	}

	if useReadv {
		if unwrapReader, ok := reader.(UnwrapReader); ok {
			return &UnwrapReaderReader{
				Reader: &SingleReader{
					Reader: reader,
				},
				UnwrapReader: unwrapReader,
			}
		}
	}

	return &SingleReader{
		Reader: reader,
	}
}

// NewPacketReader creates a new PacketReader based on the given reader.
// read one buffer every time
func NewPacketReader(reader io.Reader) Reader {
	if mr, ok := reader.(Reader); ok {
		return mr
	}

	return &PacketReader{
		Reader: reader,
	}
}

func isPacketWriter(writer io.Writer) bool {
	if _, ok := writer.(net.PacketConn); ok {
		return true
	}

	// If the writer doesn't implement syscall.Conn, it is probably not a TCP connection.
	if _, ok := writer.(syscall.Conn); !ok {
		return true
	}
	return false
}

// NewWriter creates a new Writer.
func NewWriter(writer io.Writer) Writer {
	if mw, ok := writer.(Writer); ok {
		return mw
	}

	if isPacketWriter(writer) {
		return &SequentialWriter{
			Writer: writer,
		}
	}

	return &BufferToBytesWriter{
		Writer: writer,
	}
}

type DdlReaderWriter interface {
	ReaderWriter
	SetReadDeadline(time.Time) error
}

type DeadlineReader interface {
	Reader
	SetReadDeadline(time.Time) error
}

type SecondDdlReaderWriter struct {
	DdlReaderWriter
	Mb MultiBuffer
}

func NewSecondDdl(rw DdlReaderWriter, mb MultiBuffer) *SecondDdlReaderWriter {
	return &SecondDdlReaderWriter{
		DdlReaderWriter: rw,
		Mb:              mb,
	}
}

func (r *SecondDdlReaderWriter) ReadMultiBuffer() (MultiBuffer, error) {
	if r.Mb.Len() > 0 {
		mb := r.Mb
		r.Mb = nil
		return mb, nil
	}
	return r.DdlReaderWriter.ReadMultiBuffer()
}
