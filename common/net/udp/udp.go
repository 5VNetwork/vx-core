package udp

import (
	"fmt"
	"io"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"
	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type UdpConn interface {
	PacketReaderWriter
	Close() error
}

type DdlPacketReaderWriter interface {
	PacketReaderWriter
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
	SetDeadline(time.Time) error
}

type PacketReaderWriter interface {
	PacketReader
	PacketWriter
}

type PacketReader interface {
	// Only one of the returned value is nil
	ReadPacket() (*Packet, error)
}

type PacketWriter interface {
	// Owner of the p is transfered. Caller should not use p after calling this
	WritePacket(p *Packet) error
}

type SecondDdlReaderWriter struct {
	DdlPacketReaderWriter
	Packets []*Packet
}

func (r *SecondDdlReaderWriter) ReadPacket() (*Packet, error) {
	if len(r.Packets) > 0 {
		p := r.Packets[0]
		r.Packets = r.Packets[1:]
		return p, nil
	}
	return r.DdlPacketReaderWriter.ReadPacket()
}

// adapts a PacketConn to a buf.ReaderWriter
type ReaderWriterAdaptor struct {
	PacketReaderWriter
	Addr net.Destination
}

func (p *ReaderWriterAdaptor) ReadMultiBuffer() (buf.MultiBuffer, error) {
	pa, err := p.ReadPacket()
	if err != nil {
		return nil, err
	}

	return buf.MultiBuffer{pa.Payload}, nil
}

func (p *ReaderWriterAdaptor) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for len(mb) > 0 {
		b := mb[0]
		mb = mb[1:]
		err := p.WritePacket(&Packet{
			Payload: b,
			Source:  p.Addr,
		})
		if err != nil {
			b.Release()
			buf.ReleaseMulti(mb)
			return fmt.Errorf("failed to write all multibuffer, %w", err)
		}
	}
	return nil
}

// implements PacketConn
type PacketRW struct {
	PacketReader
	PacketWriter
	OnClose func() error
}

func (p *PacketRW) Close() error {
	if p.OnClose != nil {
		return p.OnClose()
	}
	return nil
}

type ReadFromer interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
}

type WriteToer interface {
	WriteTo(p []byte, addr net.Addr) (n int, err error)
}

// used to adapt a net.PacketConn to a PacketConn
type ReaderFromerToPacketReader struct {
	ReadFromer
}

func (rw *ReaderFromerToPacketReader) ReadPacket() (*Packet, error) {
	b := buf.New()
	n, from, err := rw.ReadFromer.ReadFrom(b.BytesTo(b.Cap()))
	if err != nil {
		b.Release()
		return nil, fmt.Errorf("failed to read from PacketConn, %w", err)
	}
	b.Extend(int32(n))
	return &Packet{
		Payload: b,
		Source:  net.DestinationFromAddr(from),
	}, nil
}

type WriteToerToPacketWriter struct {
	WriteToer
}

func (rw *WriteToerToPacketWriter) WritePacket(p *Packet) error {
	defer p.Release()
	_, err := rw.WriteToer.WriteTo(p.Payload.Bytes(), p.Target.Addr())
	if err != nil {
		return fmt.Errorf("failed to write to PacketConn, %w", err)
	}
	return nil
}

type PacketWriterToWriteToer struct {
	PacketWriter
}

func (rw *PacketWriterToWriteToer) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	err = rw.PacketWriter.WritePacket(&Packet{
		Payload: buf.FromBytes(p),
		Target:  net.DestinationFromAddr(addr),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to write to PacketConn, %w", err)
	}
	return len(p), nil
}

type PacketReaderToReadFromer struct {
	PacketReader
}

func (rw *PacketReaderToReadFromer) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	packet, err := rw.PacketReader.ReadPacket()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read from PacketConn, %w", err)
	}
	defer packet.Release()
	n = copy(p, packet.Payload.Bytes())
	if n != int(packet.Payload.Len()) {
		return 0, nil, io.ErrShortBuffer
	}
	return n, packet.Source.Addr(), nil
}

type Packet struct {
	Payload *buf.Buffer
	Source  net.Destination
	Target  net.Destination
}

func (p *Packet) Clone() *Packet {
	return &Packet{
		Payload: p.Payload.Clone(),
		Source:  p.Source,
		Target:  p.Target,
	}
}

func (p *Packet) Release() {
	if p == nil {
		return
	}
	p.Payload.Release()
}

// This function will modify the payload of p, make it an ip packet
// p should not be used after this function is called
func UdpPacketToIpPacket(p *Packet) *buf.Buffer {
	var ipPacket gtcpip.IPPacket
	if p.Target.Address == nil {
		log.Fatal().Str("src", p.Source.String()).Str("dst", p.Target.String()).Hex("udpPayload", p.Payload.Bytes()).Msg("invalid udp packet")
	}
	if p.Target.Address.Family() == net.AddressFamilyIPv4 {
		p.Payload.RetreatStart(header.IPv4MinimumSize + header.UDPMinimumSize)
		ipPacket = &gtcpip.IPv4{
			IPv4: header.IPv4(p.Payload.BytesRange(0, header.IPv4MinimumSize)),
		}
		ipv4Field := header.IPv4Fields{
			TotalLength: uint16(p.Payload.Len()),
			Protocol:    uint8(header.UDPProtocolNumber),
			SrcAddr:     tcpip.AddrFromSlice(p.Source.Address.IP()),
			DstAddr:     tcpip.AddrFromSlice(p.Target.Address.IP()),
			TTL:         64,
			ID:          uint16(time.Now().UnixNano()),
		}
		ipPacket.(*gtcpip.IPv4).Encode(&ipv4Field)
		ipPacket.ResetChecksum()
	} else {
		p.Payload.RetreatStart(header.IPv6MinimumSize + header.UDPMinimumSize)
		ipPacket = &gtcpip.IPv6{
			IPv6: header.IPv6(p.Payload.BytesRange(0, header.IPv6MinimumSize)),
		}
		ipv6Field := header.IPv6Fields{
			PayloadLength:     uint16(p.Payload.Len() - header.IPv6MinimumSize),
			TransportProtocol: header.UDPProtocolNumber,
			SrcAddr:           tcpip.AddrFromSlice(p.Source.Address.IP()),
			DstAddr:           tcpip.AddrFromSlice(p.Target.Address.IP()),
			HopLimit:          64,
		}
		ipPacket.(*gtcpip.IPv6).Encode(&ipv6Field)
	}
	udpPacket := gtcpip.UDP{
		UDP: header.UDP(ipPacket.Payload()),
	}
	udpPacket.SetSourcePort(uint16(p.Source.Port))
	udpPacket.SetDestinationPort(p.Target.Port.Value())
	udpPacket.SetLength(uint16(len(ipPacket.Payload())))
	udpPacket.ResetChecksum(ipPacket.PseudoHeaderChecksum())

	if !udpPacket.IsChecksumValid(ipPacket.SourceAddress(), ipPacket.DestinationAddress(), checksum.Checksum(udpPacket.Payload(), 0)) {
		log.Fatal().Str("src", p.Source.String()).Str("dst", p.Target.String()).Hex("udpPayload", udpPacket.Payload()).Hex("ipPacket", p.Payload.Bytes()).Msg("invalid udp checksum")
	}
	if !ipPacket.IsValid(int(p.Payload.Len())) {
		log.Fatal().Str("src", p.Source.String()).Str("dst", p.Target.String()).Hex("ipPacket", p.Payload.Bytes()).Msg("invalid ip packet")
	}
	return p.Payload
}
