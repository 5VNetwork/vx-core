package protocol

import (
	"github.com/5vnetwork/vx-core/common/bitmask"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/os"
	"github.com/5vnetwork/vx-core/common/uuid"
)

// RequestCommand is a custom command in a proxy request.
type RequestCommand byte

const (
	RequestCommandTCP = RequestCommand(0x01)
	RequestCommandUDP = RequestCommand(0x02)
	RequestCommandMux = RequestCommand(0x03)
)

func (c RequestCommand) TransferType() TransferType {
	switch c {
	case RequestCommandTCP, RequestCommandMux:
		return TransferTypeStream
	case RequestCommandUDP:
		return TransferTypePacket
	default:
		return TransferTypeStream
	}
}

const (
	// RequestOptionChunkStream indicates request payload is chunked. Each chunk consists of length, authentication and payload.
	RequestOptionChunkStream bitmask.Byte = 0x01

	// RequestOptionConnectionReuse indicates client side expects to reuse the connection.
	RequestOptionConnectionReuse bitmask.Byte = 0x02

	RequestOptionChunkMasking bitmask.Byte = 0x04

	RequestOptionGlobalPadding bitmask.Byte = 0x08

	RequestOptionAuthenticatedLength bitmask.Byte = 0x10
)

type RequestHeader struct {
	Version  byte
	Command  RequestCommand
	Port     net.Port
	Address  net.Address
	Option   bitmask.Byte
	Security SecurityType
	User     string      // used in server
	Account  interface{} // used in client
}

func (h *RequestHeader) Destination() net.Destination {
	if h.Command == RequestCommandUDP {
		return net.UDPDestination(h.Address, h.Port)
	}
	return net.TCPDestination(h.Address, h.Port)
}

const (
	ResponseOptionConnectionReuse bitmask.Byte = 0x01
)

type ResponseCommand interface{}

type ResponseHeader struct {
	Option  bitmask.Byte
	Command ResponseCommand
}

type CommandSwitchAccount struct {
	Host     net.Address
	Port     net.Port
	ID       uuid.UUID
	Level    uint32
	AlterIds uint16
	ValidMin byte
}

func IsDomainTooLong(domain string) bool {
	return len(domain) > 255
}

func (sc SecurityType) GetSecurityType() SecurityType {
	if sc == SecurityType_AUTO || sc == SecurityType_UNKNOWN {
		if os.HasAESGCMHardwareSupport {
			return SecurityType_AES128_GCM
		}
		return SecurityType_CHACHA20_POLY1305
	}
	return sc
}
