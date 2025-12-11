package dns

import (
	"encoding/binary"
	"strings"
	"sync"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/serial"

	"golang.org/x/net/dns/dnsmessage"
)

func PackMessage(msg *dnsmessage.Message) (*buf.Buffer, error) {
	buffer := buf.New()
	rawBytes := buffer.Extend(buf.Size)
	packed, err := msg.AppendPack(rawBytes[:0])
	if err != nil {
		buffer.Release()
		return nil, err
	}
	buffer.Resize(0, int32(len(packed)))
	return buffer, nil
}

type MessageReader interface {
	ReadMessage() (*buf.Buffer, error)
}

type PacketReaderToMessageReader struct {
	udp.PacketReader
}

func (r *PacketReaderToMessageReader) ReadMessage() (*buf.Buffer, error) {
	packet, err := r.ReadPacket()
	if err != nil {
		return nil, err
	}
	return packet.Payload, nil
}

type PackerWriterToMessageWriter struct {
	Src net.Destination
	udp.PacketWriter
}

func (w *PackerWriterToMessageWriter) WriteMessage(msg *buf.Buffer) error {
	return w.WritePacket(&udp.Packet{
		Payload: msg,
		Source:  w.Src,
	})
}

type UDPReader struct {
	buf.Reader

	access sync.Mutex
	cache  buf.MultiBuffer
}

func (r *UDPReader) readCache() *buf.Buffer {
	r.access.Lock()
	defer r.access.Unlock()

	mb, b := buf.SplitFirst(r.cache)
	r.cache = mb
	return b
}

func (r *UDPReader) refill() error {
	mb, err := r.Reader.ReadMultiBuffer()
	if err != nil {
		return err
	}
	r.access.Lock()
	r.cache = mb
	r.access.Unlock()
	return nil
}

// ReadMessage implements MessageReader.
func (r *UDPReader) ReadMessage() (*buf.Buffer, error) {
	for {
		b := r.readCache()
		if b != nil {
			return b, nil
		}
		if err := r.refill(); err != nil {
			return nil, err
		}
	}
}

// Close implements common.Closable.
func (r *UDPReader) Close() error {
	defer func() {
		r.access.Lock()
		buf.ReleaseMulti(r.cache)
		r.cache = nil
		r.access.Unlock()
	}()

	return common.Close(r.Reader)
}

type TCPReader struct {
	reader *buf.BufferedReader
}

func NewTCPReader(reader buf.Reader) *TCPReader {
	return &TCPReader{
		reader: &buf.BufferedReader{
			Reader: reader,
		},
	}
}

func (r *TCPReader) ReadMessage() (*buf.Buffer, error) {
	size, err := serial.ReadUint16(r.reader)
	if err != nil {
		return nil, err
	}
	if size > buf.Size {
		return nil, errors.New("message size too large: ", size)
	}
	b := buf.New()
	if _, err := b.ReadFullFrom(r.reader, int32(size)); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *TCPReader) Interrupt() {
	common.Interrupt(r.reader)
}

func (r *TCPReader) Close() error {
	return common.Close(r.reader)
}

type MessageWriter interface {
	WriteMessage(msg *buf.Buffer) error
}

type UDPWriter struct {
	buf.Writer
}

func (w *UDPWriter) WriteMessage(b *buf.Buffer) error {
	return w.WriteMultiBuffer(buf.MultiBuffer{b})
}

type TCPWriter struct {
	buf.Writer
}

func (w *TCPWriter) WriteMessage(b *buf.Buffer) error {
	if b.IsEmpty() {
		return nil
	}

	mb := make(buf.MultiBuffer, 0, 2)

	size := buf.New()
	binary.BigEndian.PutUint16(size.Extend(2), uint16(b.Len()))
	mb = append(mb, size, b)
	return w.WriteMultiBuffer(mb)
}

func RootDomain(domain string) string {
	if domain == "" {
		return ""
	}

	// Remove trailing dot if present
	if len(domain) > 0 && domain[len(domain)-1] == '.' {
		domain = domain[:len(domain)-1]
	}

	// Split by dots
	parts := strings.Split(domain, ".")

	// If we have 1 or fewer parts, return the domain as is
	if len(parts) <= 1 {
		return domain
	}

	// For domains with 2 parts, return as is (e.g., "example.com")
	if len(parts) == 2 {
		return domain
	}

	// For domains with more than 2 parts, return the last 2 parts
	// This handles cases like "www.example.com" -> "example.com"
	// and "sub.sub.example.com" -> "example.com"
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}
