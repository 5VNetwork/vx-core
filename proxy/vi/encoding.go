package vi

import (
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
	"github.com/5vnetwork/vx-core/common/uuid"
)

const (
	Version = byte(0)
)

var addrParser = address_parser.VAddressSerializer

type Header struct {
	Version byte
	User    string
	Level   uint32
	Dest    net.Destination
	UdpUuid uuid.UUID
}

// EncodeHeader encodes a header into bytes.
func EncodeHeader(version byte, credential []byte, dest net.Destination, udpUuid uuid.UUID) (*buf.Buffer, error) {
	buffer := buf.New()
	buffer.Extend(2)
	buffer.WriteByte(version)
	// serial.WriteUint16(buffer, uint16(len(credential)))
	buffer.Write(credential)
	err := addrParser.WriteAddressPort(buffer, dest.Address, dest.Port)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	if dest.Network == net.Network_UDP {
		buffer.WriteByte(byte(protocol.RequestCommandUDP))
		buffer.Write(udpUuid.Bytes())
	} else {
		buffer.WriteByte(byte(protocol.RequestCommandTCP))
	}
	l := buffer.Len()
	// write length of the header to the first two bytes
	buffer.SetByte(0, byte((l-2)>>8))
	buffer.SetByte(1, byte(l-2))
	return buffer, nil
}

var (
	InvalidUser    = errors.New("invalid user")
	InvalidVersion = errors.New("invalid version")
)

func DecodeHeader(data *buf.Buffer, um *sync.Map) (*Header, error) {
	header := &Header{}
	// first two bytes are the length of the header
	data.AdvanceStart(2)
	// header := &protocol.RequestHeader{}
	version, _ := data.ReadByte()
	switch version {
	case 0:
		var uid uuid.UUID
		data.Read(uid[:])
		// tokenLen, err := serial.ReadUint16(data)
		// if err != nil {
		// 	return nil, errors.New("failed to read token length").Base(err)
		// }
		//TODO
		um, ok := um.Load(uid)
		if !ok {
			return nil, InvalidUser
		}
		header.User = um.(string)

		addr, port, err := addrParser.ReadAddressPort(nil, data)
		if err != nil {
			return nil, errors.New("failed to read address").Base(err)
		}
		b, _ := data.ReadByte()
		if b == byte(protocol.RequestCommandUDP) {
			header.Dest.Network = net.Network_UDP
			header.Dest.Address = addr
			header.Dest.Port = port
			udpUuid, err := uuid.ParseBytes(data.BytesTo(16))
			if err != nil {
				return nil, errors.New("failed to parse udp uuid").Base(err)
			}
			header.UdpUuid = udpUuid
			return header, nil
		} else {
			header.Dest.Network = net.Network_TCP
			header.Dest.Address = addr
			header.Dest.Port = port
			return header, nil
		}
	default:
		return nil, InvalidVersion
	}
}
