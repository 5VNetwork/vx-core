package gtcpip

import "gvisor.dev/gvisor/pkg/tcpip/header"

type IPv6 struct {
	header.IPv6 // IPv6 contains both header and payload
}

func (ipv6 *IPv6) ResetChecksum() {
}

func (ipv6 *IPv6) PseudoHeaderChecksum() uint16 {
	return header.PseudoHeaderChecksum(ipv6.TransportProtocol(),
		ipv6.SourceAddress(), ipv6.DestinationAddress(), ipv6.PayloadLength())
}

func (ipv6 *IPv6) HeaderLength() uint8 {
	return 40
}
