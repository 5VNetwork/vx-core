package mux

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/bitmask"
	"github.com/5vnetwork/vx-core/common/buf"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
)

/* how to read a FrameMetadata out from a reader (unmarshal)
or write a FrameMetadata into a writer (marshal)*/

var addrParser = address_parser.NewAddressParser(
	map[byte]nethelper.AddressFamily{
		byte(AddressTypeIPv4):   nethelper.AddressFamilyIPv4,
		byte(AddressTypeDomain): nethelper.AddressFamilyDomain,
		byte(AddressTypeIPv6):   nethelper.AddressFamilyIPv6,
	}, true, nil,
)

type AddressType byte

const (
	AddressTypeIPv4   AddressType = 1
	AddressTypeDomain AddressType = 2
	AddressTypeIPv6   AddressType = 3
)

type SessionStatus byte

const (
	SessionStatusNew       SessionStatus = 0x01
	SessionStatusKeep      SessionStatus = 0x02
	SessionStatusEnd       SessionStatus = 0x03
	SessionStatusKeepAlive SessionStatus = 0x04
)

const (
	OptionData  bitmask.Byte = 0x01
	OptionError bitmask.Byte = 0x02
)

type TargetNetwork byte

const (
	TargetNetworkTCP TargetNetwork = 0x01
	TargetNetworkUDP TargetNetwork = 0x02
)

type FrameMetadata struct {
	GlobalID [8]byte //only present when xudp
	// UdpUuid       uuid.UUID //only present when udp
	Target        nethelper.Destination
	SessionID     uint16
	Option        bitmask.Byte
	SessionStatus SessionStatus
}

func (f FrameMetadata) WriteTo(b *buf.Buffer) error {
	lenBytes := b.Extend(2)

	len0 := b.Len()
	sessionBytes := b.Extend(2)
	binary.BigEndian.PutUint16(sessionBytes, f.SessionID)

	common.Must(b.WriteByte(byte(f.SessionStatus)))
	common.Must(b.WriteByte(byte(f.Option)))

	if f.SessionStatus == SessionStatusNew {
		switch f.Target.Network {
		case nethelper.Network_TCP:
			common.Must(b.WriteByte(byte(TargetNetworkTCP)))
		case nethelper.Network_UDP:
			common.Must(b.WriteByte(byte(TargetNetworkUDP)))
		}

		if err := addrParser.WriteAddressPort(b, f.Target.Address, f.Target.Port); err != nil {
			return err
		}

		// if f.Target.Network == nethelper.Network_UDP && f.UdpUuid.IsSet() {
		// 	if _, err := b.Write(f.UdpUuid.Bytes()); err != nil {
		// 		return err
		// 	}
		// }
	}

	len1 := b.Len()
	binary.BigEndian.PutUint16(lenBytes, uint16(len1-len0))
	if len1-len0 > 512 {
		return fmt.Errorf("metadata too long: %d", len1-len0)
	}
	return nil
}

// Unmarshal reads FrameMetadata from the given reader.
func (f *FrameMetadata) Unmarshal(reader io.Reader) error {
	metaLen, err := serial.ReadUint16(reader)
	if err != nil {
		return err
	}
	if metaLen > 512 {
		return fmt.Errorf("invalid metalen %d", metaLen)
	}

	b := buf.New()
	defer b.Release()

	if _, err := b.ReadFullFrom(reader, int32(metaLen)); err != nil {
		return err
	}
	return f.unmarshalFromBuffer(b)
}

// UnmarshalFromBuffer reads a FrameMetadata from the given buffer.
// Visible for testing only.
func (f *FrameMetadata) unmarshalFromBuffer(b *buf.Buffer) error {
	if b.Len() < 4 {
		return fmt.Errorf("insufficient buffer: %d", b.Len())
	}

	f.SessionID = binary.BigEndian.Uint16(b.BytesTo(2))
	f.SessionStatus = SessionStatus(b.Byte(2))
	f.Option = bitmask.Byte(b.Byte(3))
	f.Target.Network = nethelper.Network_Unknown

	if f.SessionStatus == SessionStatusNew {
		if b.Len() < 8 {
			return fmt.Errorf("insufficient buffer: %d", b.Len())
		}
		network := TargetNetwork(b.Byte(4))
		b.AdvanceStart(5)

		addr, port, err := addrParser.ReadAddressPort(nil, b)
		if err != nil {
			return fmt.Errorf("failed to parse target address and port")
		}

		switch network {
		case TargetNetworkTCP:
			f.Target = nethelper.TCPDestination(addr, port)
		case TargetNetworkUDP:
			f.Target = nethelper.UDPDestination(addr, port)
			// if b.Len() == 16 {
			// 	uuid, err := uuid.ParseBytes(b.BytesTo(16))
			// 	if err != nil {
			// 		return fmt.Errorf("failed to parse udp uuid")
			// 	}
			// 	f.UdpUuid = uuid
			// }
		default:
			return fmt.Errorf("unknown network type: %d", network)
		}

	}

	return nil
}
