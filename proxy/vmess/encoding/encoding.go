package encoding

import (
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
)

const (
	Version = byte(1)
)

var addrParser = address_parser.VAddressSerializer
