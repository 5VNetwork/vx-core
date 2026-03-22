package quic

import (
	protocol "github.com/5vnetwork/vx-core/common/protocol"
	"google.golang.org/protobuf/types/known/anypb"
)

// QuicConfig is the transport-layer QUIC settings (not generated in Buf transport protos).
type QuicConfig struct {
	Security protocol.SecurityType
	Key      string
	Header   *anypb.Any
}
