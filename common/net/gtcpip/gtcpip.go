package gtcpip

import "gvisor.dev/gvisor/pkg/tcpip/header"

type IPPacket interface {
	header.Network
	IsValid(int) bool
	HeaderLength() uint8
	PayloadLength() uint16
	ResetChecksum()
	PseudoHeaderChecksum() uint16
}

// return nil if b is invalid
func NewIPPacket(b []byte) IPPacket {
	ipVersion := header.IPVersion(b)
	var IPPacket IPPacket
	if ipVersion == 4 {
		IPPacket = &IPv4{IPv4: header.IPv4(b)}
	} else if ipVersion == 6 {
		IPPacket = &IPv6{IPv6: header.IPv6(b)}
	} else {
		return nil
	}
	if !IPPacket.IsValid(len(b)) {
		return nil
	}
	return IPPacket
}
