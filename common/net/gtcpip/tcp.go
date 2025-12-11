package gtcpip

import (
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type TCP struct {
	header.TCP
}

func (tcp *TCP) ResetChecksum(pseudoSum uint16) {
	tcp.SetChecksum(0)
	tcp.SetChecksum(^checksum.Checksum(tcp.TCP, pseudoSum))
}
