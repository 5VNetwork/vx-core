package tcpip

import (
	"encoding/binary"
	"net"
)

type IPv6Packet []byte

func (p IPv6Packet) TotalLen() uint16 {
	return binary.BigEndian.Uint16(p[4:]) + 40
}

func (p IPv6Packet) HeaderLen() uint16 {
	return 40
}

func (p IPv6Packet) DataLen() uint16 {
	return binary.BigEndian.Uint16(p[4:])
}

func (p IPv6Packet) Payload() []byte {
	return p[40:p.TotalLen()]
}

func (p IPv6Packet) Protocol() IPProtocol {
	return IPProtocol(p[6])
}

func (p IPv6Packet) SourceIP() net.IP {
	var b []byte
	b = p[8:24]
	return b
}

func (p IPv6Packet) SetSourceIP(ip net.IP) {
	if len(ip) != 16 {
		panic("invalid ip length")
	}
	copy(p[8:24], ip)
}

func (p IPv6Packet) DestinationIP() net.IP {
	var b []byte
	b = p[24:40]
	return b
}

func (p IPv6Packet) SetDestinationIP(ip net.IP) {
	if len(ip) != 16 {
		panic("invalid ip length")
	}
	copy(p[24:40], ip)
}

func (p IPv6Packet) ResetChecksum() {}

// for tcp checksum
func (p IPv6Packet) PseudoSum() uint32 {
	sum := Sum(p[8:40])
	sum += uint32(p.Protocol())
	sum += uint32(p.DataLen())
	return sum
}

var firstByte6 byte = 0b01100000

func (p IPv6Packet) SetFields(src net.IP, dst net.IP, protocol IPProtocol, identification uint16) {
	p[0] = firstByte6
	p[1] = 0
	p[2] = 0
	p[3] = 0
	binary.BigEndian.PutUint16(p[4:], uint16(len(p)-40))
	p.SetDestinationIP(dst)
	p.SetSourceIP(src)
	// upper layer protocol
	p[6] = byte(protocol)
	// ttl
	p[7] = 64
}
