package address_parser

import (
	"errors"
	"fmt"
	"io"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
)

// vmess,vless
var VAddressSerializer = NewAddressParser(
	map[byte]net.AddressFamily{
		0x01: net.AddressFamilyIPv4,
		0x02: net.AddressFamilyDomain,
		0x03: net.AddressFamilyIPv6,
	}, true, nil,
)

var SocksAddressSerializer = NewAddressParser(
	map[byte]net.AddressFamily{
		0x01: net.AddressFamilyIPv4,
		0x03: net.AddressFamilyDomain,
		0x04: net.AddressFamilyIPv6,
	}, false, nil,
)

// read address and port from bytes or write address and port to bytes, used in mux
type AddressParser interface {
	// read from input, and put address and port into buffer, buffer can be nil, if so, a new buffer will be created
	// if input is a buf.Buffer, its internal end will change
	ReadAddressPort(buffer *buf.Buffer, input io.Reader) (net.Address, net.Port, error)
	WriteAddressPort(writer io.Writer, addr net.Address, port net.Port) error
	PeekAddressPort(buffer *buf.Buffer) (net.Address, net.Port, error)
}

// NewAddressParser creates a new AddressParser
func NewAddressParser(byteToAddressFamily map[byte]net.AddressFamily, portFirst bool, typeParser AddressTypeParser) AddressParser {
	var addrRW addressParser

	for i := range addrRW.addressFamilyToByte {
		addrRW.addressFamilyToByte[i] = afInvalid
	}
	for i := range addrRW.byteToAddressFamily {
		addrRW.byteToAddressFamily[i] = net.AddressFamily(afInvalid)
	}

	for b, af := range byteToAddressFamily {
		if b > 16 {
			panic("address family byte too big")
		}
		addrRW.byteToAddressFamily[b] = af
		addrRW.addressFamilyToByte[af] = b
	}

	addrRW.portFirst = portFirst
	addrRW.typeParser = typeParser

	return &addrRW
}

type addressParser struct {
	byteToAddressFamily [16]net.AddressFamily
	addressFamilyToByte [16]byte
	typeParser          AddressTypeParser
	portFirst           bool
}

func (rw *addressParser) PeekAddressPort(buffer *buf.Buffer) (net.Address, net.Port, error) {
	bytes := buffer.Bytes()
	b := buf.FromBytes(bytes)
	addr, port, err := rw.ReadAddressPort(nil, b)
	if err != nil {
		return nil, 0, err
	}
	return addr, port, nil
}

func (rw *addressParser) ReadAddressPort(buffer *buf.Buffer, input io.Reader) (net.Address, net.Port, error) {
	var port net.Port
	var err error

	if buffer == nil {
		buffer = buf.New()
		defer buffer.Release()
	}

	if rw.portFirst {
		port, err = readPort(buffer, input)
		if err != nil {
			return nil, 0, err
		}
	}

	addr, err := rw.readAddress(buffer, input)
	if err != nil {
		return nil, 0, err
	}

	if !rw.portFirst {
		port, err = readPort(buffer, input)
		if err != nil {
			return nil, 0, err
		}
	}

	return addr, port, nil
}

func (rw *addressParser) WriteAddressPort(writer io.Writer, addr net.Address, port net.Port) error {
	if rw.portFirst {
		if err := writePort(writer, port); err != nil {
			return err
		}
	}

	if err := rw.writeAddress(writer, addr); err != nil {
		return err
	}

	if !rw.portFirst {
		return writePort(writer, port)
	}

	return nil
}

func (p *addressParser) readAddress(b *buf.Buffer, reader io.Reader) (net.Address, error) {
	if _, err := b.ReadFullFrom(reader, 1); err != nil {
		return nil, err
	}

	addrType := b.Byte(b.Len() - 1)
	if p.typeParser != nil {
		addrType = p.typeParser(addrType)
	}

	if addrType >= 16 {
		return nil, fmt.Errorf("unknown address type: %b", addrType)
	}

	addrFamily := p.byteToAddressFamily[addrType]
	if addrFamily == net.AddressFamily(afInvalid) {
		return nil, fmt.Errorf("unknown address type: %b", addrType)
	}

	switch addrFamily {
	case net.AddressFamilyIPv4:
		if _, err := b.ReadFullFrom(reader, 4); err != nil {
			return nil, err
		}
		return net.IPAddress(b.BytesFrom(-4)), nil
	case net.AddressFamilyIPv6:
		if _, err := b.ReadFullFrom(reader, 16); err != nil {
			return nil, err
		}
		return net.IPAddress(b.BytesFrom(-16)), nil
	case net.AddressFamilyDomain:
		if _, err := b.ReadFullFrom(reader, 1); err != nil {
			return nil, err
		}
		domainLength := int32(b.Byte(b.Len() - 1))
		if _, err := b.ReadFullFrom(reader, domainLength); err != nil {
			return nil, err
		}
		domain := string(b.BytesFrom(-domainLength))
		if maybeIPPrefix(domain[0]) {
			addr := net.ParseAddress(domain)
			if addr.Family().IsIP() {
				return addr, nil
			}
		}
		if !isValidDomain(domain) {
			return nil, fmt.Errorf("invalid domain name: %s", domain)
		}
		return net.DomainAddress(domain), nil
	default:
		panic("impossible case")
	}
}

func (p *addressParser) writeAddress(writer io.Writer, address net.Address) error {
	tb := p.addressFamilyToByte[address.Family()]
	if tb == afInvalid {
		return fmt.Errorf("unknown address family %b", address.Family())
	}

	switch address.Family() {
	case net.AddressFamilyIPv4, net.AddressFamilyIPv6:
		if _, err := writer.Write([]byte{tb}); err != nil {
			return err
		}
		if _, err := writer.Write(address.IP()); err != nil {
			return err
		}
	case net.AddressFamilyDomain:
		domain := address.Domain()
		if net.IsDomainTooLong(domain) {
			return errors.New("Super long domain is not supported: " + domain)
		}

		if _, err := writer.Write([]byte{tb, byte(len(domain))}); err != nil {
			return err
		}
		if _, err := writer.Write([]byte(domain)); err != nil {
			return err
		}
	default:
		panic("Unknown family type.")
	}

	return nil
}

func readPort(b *buf.Buffer, reader io.Reader) (net.Port, error) {
	if _, err := b.ReadFullFrom(reader, 2); err != nil {
		return 0, err
	}
	return net.PortFromBytes(b.BytesFrom(-2)), nil
}

func writePort(writer io.Writer, port net.Port) error {
	_, err := serial.WriteUint16(writer, port.Value())
	return err
}

func maybeIPPrefix(b byte) bool {
	return b == '[' || (b >= '0' && b <= '9')
}

func isValidDomain(d string) bool {
	for _, c := range d {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '-' || c == '.' || c == '_') {
			return false
		}
	}
	return true
}

type AddressTypeParser func(byte) byte

const afInvalid = 255
