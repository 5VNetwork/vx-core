package reject

import (
	"github.com/5vnetwork/vx-core/common/buf"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type CombineRejector struct {
	TCPReject *TCPReject
	UDPReject *UdpReject
}

func (r *CombineRejector) Reject(p []byte) *buf.Buffer {
	if header.IPVersion(p) == header.IPv6Version {
		ipv6 := header.IPv6(p)
		if ipv6.TransportProtocol() == header.TCPProtocolNumber {
			return r.TCPReject.Reject(p)
		} else if ipv6.TransportProtocol() == header.UDPProtocolNumber {
			return r.UDPReject.Reject(p)
		}
	}
	return nil
}
