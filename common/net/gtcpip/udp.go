package gtcpip

import (
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type UDP struct {
	header.UDP // UDP contains both header and payload
}

// recalculate checksum and set it
func (udp *UDP) ResetChecksum(pseudoSum uint16) {
	udp.SetChecksum(0)
	udp.SetChecksum(^checksum.Checksum(udp.UDP, pseudoSum))
}
