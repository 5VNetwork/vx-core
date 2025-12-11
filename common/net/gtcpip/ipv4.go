package gtcpip

import (
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type IPv4 struct {
	header.IPv4 // IPv4 contains both header and payload
}

func (ipv4 *IPv4) ResetChecksum() {
	ipv4.SetChecksum(0)
	ipv4.SetChecksum(^ipv4.CalculateChecksum())
}

func (ipv4 *IPv4) PseudoHeaderChecksum() uint16 {
	return header.PseudoHeaderChecksum(ipv4.TransportProtocol(),
		ipv4.SourceAddress(), ipv4.DestinationAddress(), ipv4.PayloadLength())
}
