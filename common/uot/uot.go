package uot

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
	"github.com/5vnetwork/vx-core/i"
)

var Addr = net.DomainAddress("sp.v2.udp-over-tcp.arpa")

// UoT address parser: type byte (0x00=IPv4, 0x01=IPv6, 0x02=Domain), then address, then port
var addrParser = address_parser.NewAddressParser(
	map[byte]net.AddressFamily{
		0x00: net.AddressFamilyIPv4,
		0x01: net.AddressFamilyIPv6,
		0x02: net.AddressFamilyDomain,
	}, false, nil,
)

// readAddress reads an address from the buf.Reader in UoT format
func readAddress(r io.Reader) (net.Destination, error) {
	addr, port, err := addrParser.ReadAddressPort(nil, r)
	if err != nil {
		return net.Destination{}, err
	}
	return net.Destination{
		Address: addr,
		Port:    port,
		Network: net.Network_UDP,
	}, nil
}

// uotPacketConn wraps the TCP stream for packet mode (multi-destination)
type uotReader struct {
	io.Reader
}

func (c *uotReader) ReadPacket() (*udp.Packet, error) {
	target, err := readAddress(c)
	if err != nil {
		return nil, err
	}
	length, err := serial.ReadUint16(c)
	if err != nil {
		return nil, err
	}
	b := buf.NewWithSize(int32(length))
	_, err = b.ReadFullFrom(c, int32(length))
	if err != nil {
		return nil, err
	}
	return &udp.Packet{
		Payload: b,
		Target:  target,
	}, nil
}

type uotWriter struct {
	buf.Writer
}

func getPacketHeaderLen(addr net.Address) (int32, error) {
	var addrPortLen int32
	switch addr.Family() {
	case net.AddressFamilyDomain:
		if protocol.IsDomainTooLong(addr.Domain()) {
			return 0, fmt.Errorf("super long domain is not supported: %s", addr.Domain())
		}
		addrPortLen = 1 + 1 + int32(len(addr.Domain()))
	case net.AddressFamilyIPv4:
		addrPortLen = 1 + 4
	case net.AddressFamilyIPv6:
		addrPortLen = 1 + 16
	default:
		panic("Unknown address type.")
	}
	return addrPortLen + 4, nil
}

func (c *uotWriter) WritePacket(p *udp.Packet) error {
	pakcetheaderLen, err := getPacketHeaderLen(p.Source.Address)
	if err != nil {
		p.Release()
		return err
	}
	b := p.Payload
	payloadLen := b.Len()

	b.RetreatStart(pakcetheaderLen)
	// make payload start and end at the same position
	b.Resize(0, 0)
	err = addrParser.WriteAddressPort(b, p.Source.Address, p.Source.Port)
	if err != nil {
		p.Release()
		return err
	}
	_, err = serial.WriteUint16(b, uint16(payloadLen))
	if err != nil {
		p.Release()
		return err
	}

	b.Resize(0, pakcetheaderLen+payloadLen)
	return c.WriteMultiBuffer(buf.MultiBuffer{b})
}

func Serve(ctx context.Context, rw buf.ReaderWriter, d i.Handler) error {
	// Read the UoT request
	br := &buf.BufferedReader{Reader: rw}

	var isConnect bool
	if err := binary.Read(br, binary.BigEndian, &isConnect); err != nil {
		return fmt.Errorf("failed to read isConnect flag: %w", err)
	}
	destination, err := readAddress(br)
	if err != nil {
		return fmt.Errorf("failed to read destination: %w", err)
	}

	if isConnect {
		return d.HandleFlow(ctx, destination,
			buf.NewRW(buf.NewLengthPacketReader(br),
				buf.NewMultiLengthPacketWriter(rw)))
	} else {
		return d.HandlePacketConn(ctx, destination, udp.PacketRW{
			PacketReader: &uotReader{Reader: br},
			PacketWriter: &uotWriter{Writer: rw},
		})
	}
}

type UotReaderWriter struct {
	buf.ReaderWriter
	target     net.Destination
	headerDone bool

	packetLen uint16
	left      buf.MultiBuffer
}

func NewUotReaderWriter(rw buf.ReaderWriter, target net.Destination) *UotReaderWriter {
	return &UotReaderWriter{
		ReaderWriter: rw,
		target:       target,
	}
}

func (u *UotReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := u.ReaderWriter.ReadMultiBuffer()
	for _, b := range mb {
		len := b.Len()
		b.RetreatStart(2)
		binary.BigEndian.PutUint16(b.BytesRange(0, 2), uint16(len))
	}
	if !u.headerDone {
		u.headerDone = true
		ret := buf.MultiBuffer{header(u.target, true)}
		ret = append(ret, mb...)
		return ret, err
	}
	return mb, err
}

func header(dst net.Destination, connect bool) *buf.Buffer {
	header := buf.New()
	if connect {
		header.WriteByte(1)
	} else {
		header.WriteByte(0)
	}
	addrParser.WriteAddressPort(header, dst.Address, dst.Port)
	return header
}

func (u *UotReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	var newMb buf.MultiBuffer
	var length [2]byte

	mb, _ = buf.MergeMulti(u.left, mb)
	for {
		if mb.Len() == 0 {
			break
		}
		// read length
		if u.packetLen == 0 && mb.Len() >= 2 {
			mb, _ = buf.SplitBytes(mb, length[:])
			l := binary.BigEndian.Uint16(length[:])
			u.packetLen = l
		} else if u.packetLen <= uint16(mb.Len()) {
			// read a whole packet
			b := buf.New()
			mb, _ = buf.SplitBytes(mb, b.Extend(int32(u.packetLen)))
			newMb = append(newMb, b)
			u.packetLen = 0
		} else {
			u.left = mb
			break
		}
	}

	if newMb.Len() > 0 {
		return u.ReaderWriter.WriteMultiBuffer(newMb)
	}
	return nil
}

type UotPacketReaderWriter struct {
	rw         udp.PacketReaderWriter
	left       buf.MultiBuffer
	headerSent bool
	target     net.Destination

	addrType  byte
	address   net.Address
	port      net.Port
	packetLen uint16
}

func NewUotPacketReaderWriter(rw udp.PacketReaderWriter, target net.Destination) *UotPacketReaderWriter {
	return &UotPacketReaderWriter{
		rw:       rw,
		target:   target,
		addrType: 8,
	}
}

func (u *UotPacketReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	var mb buf.MultiBuffer

	if !u.headerSent {
		u.headerSent = true
		mb = append(mb, header(u.target, false))
	}

	p, err := u.rw.ReadPacket()
	if err != nil {
		return nil, err
	}

	retreatLen, err := getPacketHeaderLen(p.Target.Address)
	if err != nil {
		p.Release()
		return nil, err
	}
	packetLen := p.Payload.Len()
	p.Payload.RetreatStart(retreatLen)

	// make payload start and end at the same position
	p.Payload.Resize(0, 0)
	err = addrParser.WriteAddressPort(p.Payload, p.Target.Address, p.Target.Port)
	if err != nil {
		p.Release()
		return nil, err
	}
	serial.WriteUint16(p.Payload, uint16(packetLen))
	// reset end position
	p.Payload.Resize(0, packetLen+retreatLen)

	mb = append(mb, p.Payload)
	return mb, nil
}

func (u *UotPacketReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	var bytes [256]byte

	mb, _ = buf.MergeMulti(u.left, mb)
	for {
		if mb.Len() == 0 {
			break
		}
		// read addr type
		if u.addrType >= 3 {
			mb, _ = buf.SplitBytes(mb, bytes[:1])
			u.addrType = bytes[0]
		} else if u.address == nil {
			switch net.AddressFamily(u.addrType) {
			case net.AddressFamilyDomain:
				l := uint8(mb[0].Byte(0))
				if int32(l)+1 <= mb.Len() {
					mb, _ = buf.SplitBytes(mb, bytes[:1])
					mb, _ = buf.SplitBytes(mb, bytes[:l])
					u.address = net.DomainAddress(string(bytes[:l]))
				} else {
					break
				}
			case net.AddressFamilyIPv4:
				if mb.Len() >= 4 {
					mb, _ = buf.SplitBytes(mb, bytes[:4])
					u.address = net.IPAddress(bytes[:4])
				} else {
					break
				}
			case net.AddressFamilyIPv6:
				if mb.Len() >= 16 {
					mb, _ = buf.SplitBytes(mb, bytes[:16])
					u.address = net.IPAddress(bytes[:16])
				} else {
					break
				}
			default:
				panic("Unknown address type.")
			}
		} else if u.port == 0 && mb.Len() >= 2 {
			mb, _ = buf.SplitBytes(mb, bytes[:2])
			l := binary.BigEndian.Uint16(bytes[:2])
			u.port = net.Port(l)
		} else if u.packetLen == 0 && mb.Len() >= 2 {
			mb, _ = buf.SplitBytes(mb, bytes[:2])
			u.packetLen = binary.BigEndian.Uint16(bytes[:2])
		} else if u.packetLen <= uint16(mb.Len()) {
			// read a whole packet
			b := buf.New()
			mb, _ = buf.SplitBytes(mb, b.Extend(int32(u.packetLen)))
			if err := u.rw.WritePacket(&udp.Packet{
				Payload: b,
				Source: net.Destination{Address: u.address,
					Port: u.port, Network: net.Network_UDP},
			}); err != nil {
				return err
			}
			u.packetLen = 0
			u.address = nil
			u.port = 0
			u.addrType = 8
		} else {
			u.left = mb
			break
		}
	}

	return nil
}

func (u *UotPacketReaderWriter) CloseWrite() error {
	return nil
}
