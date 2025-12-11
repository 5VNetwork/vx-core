//
//   date  : 2016-05-13
//   author: xjdrew
//

package tcpip

import (
	"encoding/binary"
	"net"
)

type IPProtocol byte

const (
	ICMP IPProtocol = 0x01
	TCP             = 0x06
	UDP             = 0x11
)

type IPv4Packet []byte

func (p IPv4Packet) TotalLen() uint16 {
	return binary.BigEndian.Uint16(p[2:])
}

func (p IPv4Packet) HeaderLen() uint16 {
	return uint16(p[0]&0xf) * 4
}

func (p IPv4Packet) DataLen() uint16 {
	return p.TotalLen() - p.HeaderLen()
}

func (p IPv4Packet) Payload() []byte {
	return p[p.HeaderLen():p.TotalLen()]
}

func (p IPv4Packet) Protocol() IPProtocol {
	return IPProtocol(p[9])
}

func (p IPv4Packet) SourceIP() net.IP {
	return net.IPv4(p[12], p[13], p[14], p[15]).To4()
}

func (p IPv4Packet) SetSourceIP(ip net.IP) {
	ip = ip.To4()
	if ip != nil {
		copy(p[12:16], ip)
	}
}

func (p IPv4Packet) DestinationIP() net.IP {
	return net.IPv4(p[16], p[17], p[18], p[19]).To4()
}

func (p IPv4Packet) SetDestinationIP(ip net.IP) {
	ip = ip.To4()
	if ip != nil {
		copy(p[16:20], ip)
	}
}

func (p IPv4Packet) Checksum() uint16 {
	return binary.BigEndian.Uint16(p[10:])
}

func (p IPv4Packet) SetChecksum(sum [2]byte) {
	p[10] = sum[0]
	p[11] = sum[1]
}

func (p IPv4Packet) ResetChecksum() {
	p.SetChecksum(zeroChecksum)
	p.SetChecksum(Checksum(0, p[:p.HeaderLen()]))
}

// for tcp checksum
func (p IPv4Packet) PseudoSum() uint32 {
	sum := Sum(p[12:20])
	sum += uint32(p.Protocol())
	sum += uint32(p.DataLen())
	return sum
}

var firstByte4 = 0b01000101

func (p IPv4Packet) SetFields(src net.IP, dst net.IP, protocol IPProtocol, identification uint16) {
	if len(p) < 20 {
		return
	}
	p[0] = byte(firstByte4)
	// len
	binary.BigEndian.PutUint16(p[2:], uint16(len(p)))
	p[4] = byte(identification >> 8)
	p[5] = byte(identification)
	// flags and fragment offset
	binary.BigEndian.PutUint16(p[6:], 0)
	p.SetDestinationIP(dst)
	p.SetSourceIP(src)
	// ttl
	p[8] = 64
	// upper layer protocol
	p[9] = byte(protocol)
	p.ResetChecksum()
}
