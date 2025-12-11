package trojan

import (
	"encoding/binary"
	"io"
	gonet "net"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
)

var (
	crlf = []byte{'\r', '\n'}

	addrParser = address_parser.SocksAddressSerializer
)

const (
	commandTCP byte = 1
	commandUDP byte = 3
)

// ConnWriter is TCP Connection Writer Wrapper for trojan protocol
type ConnWriter struct {
	io.Writer
	Target     net.Destination
	Account    *MemoryAccount
	headerSent bool
}

func (c *ConnWriter) CloseWrite() error {
	return nil
}

func (c *ConnWriter) OkayToUnwrapWriter() int {
	if c.headerSent {
		return 1
	}
	return 0
}

func (c *ConnWriter) UnwrapWriter() any {
	return c.Writer
}

// Write implements io.Writer
func (c *ConnWriter) Write(p []byte) (n int, err error) {
	if !c.headerSent {
		if err := c.writeHeader(); err != nil {
			return 0, errors.New("failed to write request header")
		}
	}

	return c.Writer.Write(p)
}

// WriteMultiBuffer implements buf.Writer
func (c *ConnWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)

	for _, b := range mb {
		if !b.IsEmpty() {
			if _, err := c.Write(b.Bytes()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *ConnWriter) WriteHeader() error {
	if !c.headerSent {
		if err := c.writeHeader(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ConnWriter) writeHeader() error {
	buffer := buf.StackNew()
	defer buffer.Release()

	command := commandTCP
	if c.Target.Network == net.Network_UDP {
		command = commandUDP
	}

	if _, err := buffer.Write(c.Account.Key); err != nil {
		return err
	}
	if _, err := buffer.Write(crlf); err != nil {
		return err
	}
	if err := buffer.WriteByte(command); err != nil {
		return err
	}
	if err := addrParser.WriteAddressPort(&buffer, c.Target.Address, c.Target.Port); err != nil {
		return err
	}
	if _, err := buffer.Write(crlf); err != nil {
		return err
	}

	_, err := c.Writer.Write(buffer.Bytes())
	if err == nil {
		c.headerSent = true
	}

	return err
}

// PacketWriter UDP Connection Writer Wrapper for trojan protocol
type PacketWriter struct {
	writer io.Writer
	client bool
	Dest   net.Destination
}

func (w *PacketWriter) CloseWrite() error {
	return nil
}

// WriteMultiBuffer implements buf.Writer
func (w *PacketWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for _, b := range mb {
		if b.IsEmpty() {
			continue
		}
		if _, err := w.writePacket(b.Bytes(), w.Dest); err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
	}

	return nil
}

func (w *PacketWriter) Write(p []byte) (n int, err error) {
	return w.writePacket(p, w.Dest)
}

// WritePacket writes udp packet with destination specified
func (w *PacketWriter) WritePacket(p *udp.Packet) error {
	if p.Payload.IsEmpty() {
		return nil
	}
	var dest net.Destination
	if w.client {
		dest = p.Target
	} else {
		dest = p.Source
	}
	if _, err := w.writePacket(p.Payload.Bytes(), dest); err != nil {
		p.Payload.Release()
		return err
	}

	return nil
}

func (w *PacketWriter) WriteTo(payload []byte, addr gonet.Addr) (int, error) {
	dest := net.DestinationFromAddr(addr)
	return w.writePacket(payload, dest)
}

func (w *PacketWriter) writePacket(payload []byte, dest net.Destination) (int, error) { // nolint: unparam
	var addrPortLen int32
	switch dest.Address.Family() {
	case net.AddressFamilyDomain:
		if protocol.IsDomainTooLong(dest.Address.Domain()) {
			return 0, errors.New("Super long domain is not supported: ", dest.Address.Domain())
		}
		addrPortLen = 1 + 1 + int32(len(dest.Address.Domain())) + 2
	case net.AddressFamilyIPv4:
		addrPortLen = 1 + 4 + 2
	case net.AddressFamilyIPv6:
		addrPortLen = 1 + 16 + 2
	default:
		panic("Unknown address type.")
	}

	length := len(payload)
	lengthBuf := [2]byte{}
	binary.BigEndian.PutUint16(lengthBuf[:], uint16(length))

	buffer := buf.NewWithSize(addrPortLen + 2 + 2 + int32(length))
	defer buffer.Release()

	if err := addrParser.WriteAddressPort(buffer, dest.Address, dest.Port); err != nil {
		return 0, err
	}
	if _, err := buffer.Write(lengthBuf[:]); err != nil {
		return 0, err
	}
	if _, err := buffer.Write(crlf); err != nil {
		return 0, err
	}
	if _, err := buffer.Write(payload); err != nil {
		return 0, err
	}
	_, err := w.writer.Write(buffer.Bytes())
	if err != nil {
		return 0, err
	}

	return length, nil
}

// ParseHeader parses the trojan protocol header
func ParseHeader(reader io.Reader) (dst net.Destination, err error) {
	var crlf [2]byte
	var command [1]byte
	// var hash [56]byte
	// if _, err := io.ReadFull(c.Reader, hash[:]); err != nil {
	// 	return errors.New("failed to read user hash").Base(err)
	// }

	// if _, err := io.ReadFull(c.Reader, crlf[:]); err != nil {
	// 	return errors.New("failed to read crlf").Base(err)
	// }

	if _, err := io.ReadFull(reader, command[:]); err != nil {
		return dst, errors.New("failed to read command").Base(err)
	}
	network := net.Network_TCP
	if command[0] == commandUDP {
		network = net.Network_UDP
	}

	addr, port, err := addrParser.ReadAddressPort(nil, reader)
	if err != nil {
		return dst, errors.New("failed to read address and port").Base(err)
	}
	dst = net.Destination{Network: network, Address: addr, Port: port}

	if _, err := io.ReadFull(reader, crlf[:]); err != nil {
		return dst, errors.New("failed to read crlf").Base(err)
	}

	return dst, nil
}

// PacketPayload combines udp payload and destination
type PacketPayload struct {
	Target net.Destination
	Buffer buf.MultiBuffer
}

// PacketReader is UDP Connection Reader Wrapper for trojan protocol
type PacketReader struct {
	client bool
	reader io.Reader
}

// ReadMultiBuffer implements buf.Reader
func (r *PacketReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	p, err := r.ReadPacket()
	if p != nil {
		return buf.MultiBuffer{p.Payload}, err
	}
	return nil, err
}

func (r *PacketReader) Read(p []byte) (n int, err error) {
	var lengthBuf [2]byte
	if _, err := io.ReadFull(r.reader, lengthBuf[:]); err != nil {
		return 0, errors.New("failed to read payload length").Base(err)
	}
	length := binary.BigEndian.Uint16(lengthBuf[:])

	var crlf [2]byte
	if _, err := io.ReadFull(r.reader, crlf[:]); err != nil {
		return 0, errors.New("failed to read crlf").Base(err)
	}

	if len(p) < int(length) {
		return 0, io.ErrShortBuffer
	}

	b := buf.FromBytes(p)
	_, err = b.ReadFullFrom(r.reader, int32(length))
	if err != nil {
		return 0, errors.New("failed to read payload").Base(err)
	}
	return int(length), nil
}

// ReadMultiBufferWithMetadata reads udp packet with destination
func (r *PacketReader) ReadPacket() (*udp.Packet, error) {
	addr, port, err := addrParser.ReadAddressPort(nil, r.reader)
	if err != nil {
		return nil, errors.New("failed to read address and port").Base(err)
	}

	var lengthBuf [2]byte
	if _, err := io.ReadFull(r.reader, lengthBuf[:]); err != nil {
		return nil, errors.New("failed to read payload length").Base(err)
	}

	length := binary.BigEndian.Uint16(lengthBuf[:])

	var crlf [2]byte
	if _, err := io.ReadFull(r.reader, crlf[:]); err != nil {
		return nil, errors.New("failed to read crlf").Base(err)
	}

	dest := net.UDPDestination(addr, port)

	b := buf.NewWithSize(int32(length))
	_, err = b.ReadFullFrom(r.reader, int32(length))
	if err != nil {
		return nil, errors.New("failed to read payload").Base(err)
	}

	if r.client {
		return &udp.Packet{Source: dest, Payload: b}, nil
	}

	return &udp.Packet{Target: dest, Payload: b}, nil
}
