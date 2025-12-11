package inboundcommon

import (
	"github.com/5vnetwork/vx-core/common/buf"
)

type Rejector interface {
	// return a reject packet or nil. p should contains at least network header and transport header
	Reject(p []byte) *buf.Buffer
}
