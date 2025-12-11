//
//   date  : 2016-05-13
//   author: xjdrew
//

package tcpip

import (
	"net"
)

type IPPacket interface {
	TotalLen() uint16
	HeaderLen() uint16
	DataLen() uint16
	Payload() []byte
	Protocol() IPProtocol
	SourceIP() net.IP
	SetSourceIP(ip net.IP)
	DestinationIP() net.IP
	SetDestinationIP(ip net.IP)
	ResetChecksum()
	PseudoSum() uint32
	SetFields(src net.IP, dst net.IP, protocol IPProtocol, identification uint16)
}

func IsIPPacket(packet []byte) bool {
	return IsIPv4(packet) || IsIPv6(packet)
}

func NewIPPacket(packet []byte) IPPacket {
	if IsIPv4(packet) {
		return IPv4Packet(packet)
	}
	if IsIPv6(packet) {
		return IPv6Packet(packet)
	}
	return nil
}

func IsIPv4(packet []byte) bool {
	return (packet[0]>>4) == 4 && len(packet) >= 20
}

func IsIPv6(packet []byte) bool {
	return (packet[0]>>4) == 6 && len(packet) >= 40
}

func ConvertIPv4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}

	v := uint32(ip[0]) << 24
	v += uint32(ip[1]) << 16
	v += uint32(ip[2]) << 8
	v += uint32(ip[3])
	return v
}

func ConvertUint32ToIPv4(v uint32) net.IP {
	return net.IPv4(byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
