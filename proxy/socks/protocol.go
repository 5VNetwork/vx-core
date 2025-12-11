package socks

import (
	"encoding/binary"
	"io"
	gonet "net"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
)

const (
	socks5Version = 0x05
	socks4Version = 0x04

	cmdTCPConnect    = 0x01
	cmdTCPBind       = 0x02
	cmdUDPAssociate  = 0x03
	cmdTorResolve    = 0xF0
	cmdTorResolvePTR = 0xF1

	socks4RequestGranted  = 90
	socks4RequestRejected = 91

	authNotRequired = 0x00
	// authGssAPI           = 0x01
	authPassword         = 0x02
	authNoMatchingMethod = 0xFF

	statusSuccess       = 0x00
	statusCmdNotSupport = 0x07
)

var addrParser = address_parser.SocksAddressSerializer

type ServerSession struct {
	serverConfig   *Server
	gatewayAddress net.Address
	gatewayPort    net.Port
	clientAddress  net.Address
}

func (s *ServerSession) handshake4(cmd byte, reader io.Reader, writer io.Writer) (*protocol.RequestHeader, error) {
	if s.serverConfig.authType == proxy.AuthType_PASSWORD {
		writeSocks4Response(writer, socks4RequestRejected, net.AnyIP, net.Port(0))
		return nil, errors.New("socks 4 is not allowed when auth is required.")
	}

	var port net.Port
	var address net.Address

	{
		buffer := buf.StackNew()
		if _, err := buffer.ReadFullFrom(reader, 6); err != nil {
			buffer.Release()
			return nil, errors.New("insufficient header").Base(err)
		}
		port = net.PortFromBytes(buffer.BytesRange(0, 2))
		address = net.IPAddress(buffer.BytesRange(2, 6))
		buffer.Release()
	}

	if _, err := ReadUntilNull(reader); /* user id */ err != nil {
		return nil, err
	}
	if address.IP()[0] == 0x00 {
		domain, err := ReadUntilNull(reader)
		if err != nil {
			return nil, errors.New("failed to read domain for socks 4a").Base(err)
		}
		address = net.DomainAddress(domain)
	}

	switch cmd {
	case cmdTCPConnect:
		request := &protocol.RequestHeader{
			Command: protocol.RequestCommandTCP,
			Address: address,
			Port:    port,
			Version: socks4Version,
		}
		if err := writeSocks4Response(writer, socks4RequestGranted, net.AnyIP, net.Port(0)); err != nil {
			return nil, err
		}
		return request, nil
	default:
		writeSocks4Response(writer, socks4RequestRejected, net.AnyIP, net.Port(0))
		return nil, errors.New("unsupported command: ", cmd)
	}
}

// returns username if auth is successful, otherwise returns empty string
func (s *ServerSession) auth5(nMethod byte, reader io.Reader, writer io.Writer) (string, error) {
	buffer := buf.StackNew()
	defer buffer.Release()

	if _, err := buffer.ReadFullFrom(reader, int32(nMethod)); err != nil {
		return "", errors.New("failed to read auth methods").Base(err)
	}

	var expectedAuth byte = authNotRequired
	if s.serverConfig.authType == proxy.AuthType_PASSWORD {
		expectedAuth = authPassword
	}

	if !hasAuthMethod(expectedAuth, buffer.BytesRange(0, int32(nMethod))) {
		writeSocks5AuthenticationResponse(writer, socks5Version, authNoMatchingMethod)
		return "", errors.New("no matching auth method")
	}

	if err := writeSocks5AuthenticationResponse(writer, socks5Version, expectedAuth); err != nil {
		return "", errors.New("failed to write auth response").Base(err)
	}

	if expectedAuth == authPassword {
		username, password, err := ReadUsernamePassword(reader)
		if err != nil {
			return "", errors.New("failed to read username and password for authentication").Base(err)
		}

		if !s.serverConfig.HasAccount(username, password) {
			writeSocks5AuthenticationResponse(writer, 0x01, 0xFF)
			return "", errors.New("invalid username or password")
		}

		if err := writeSocks5AuthenticationResponse(writer, 0x01, 0x00); err != nil {
			return "", errors.New("failed to write auth response").Base(err)
		}
		return username, nil
	}

	return "", nil
}

func (s *ServerSession) handshake5(nMethod byte, reader io.Reader, writer io.Writer) (*protocol.RequestHeader, error) {
	username, err := s.auth5(nMethod, reader, writer)
	if err != nil {
		return nil, err
	}

	var cmd byte
	{
		buffer := buf.StackNew()
		if _, err := buffer.ReadFullFrom(reader, 3); err != nil {
			buffer.Release()
			return nil, errors.New("failed to read request").Base(err)
		}
		cmd = buffer.Byte(1)
		buffer.Release()
	}

	request := new(protocol.RequestHeader)
	request.User = username

	switch cmd {
	case cmdTCPConnect, cmdTorResolve, cmdTorResolvePTR:
		// We don't have a solution for Tor case now. Simply treat it as connect command.
		request.Command = protocol.RequestCommandTCP
	case cmdUDPAssociate:
		if !s.serverConfig.udpEnabled {
			writeSocks5Response(writer, statusCmdNotSupport, net.AnyIP, net.Port(0))
			return nil, errors.New("UDP is not enabled.")
		}
		request.Command = protocol.RequestCommandUDP
	case cmdTCPBind:
		writeSocks5Response(writer, statusCmdNotSupport, net.AnyIP, net.Port(0))
		return nil, errors.New("TCP bind is not supported.")
	default:
		writeSocks5Response(writer, statusCmdNotSupport, net.AnyIP, net.Port(0))
		return nil, errors.New("unknown command ", cmd)
	}

	request.Version = socks5Version

	addr, port, err := addrParser.ReadAddressPort(nil, reader)
	if err != nil {
		return nil, errors.New("failed to read address").Base(err)
	}
	request.Address = addr
	request.Port = port

	responseAddress := s.gatewayAddress
	responsePort := s.gatewayPort
	//nolint:gocritic // Use if else chain for clarity
	if request.Command == protocol.RequestCommandUDP {
		if s.serverConfig.address != nil {
			// Use configured IP as remote address in the response to UdpAssociate
			responseAddress = s.serverConfig.address
		} else if s.clientAddress == net.LocalHostIP || s.clientAddress == net.LocalHostIPv6 {
			// For localhost clients use loopback IP
			responseAddress = s.clientAddress
		} else {
			// For non-localhost clients use inbound listening address
			responseAddress = s.gatewayAddress
		}
	}
	if err := writeSocks5Response(writer, statusSuccess, responseAddress, responsePort); err != nil {
		return nil, err
	}

	return request, nil
}

// Handshake performs a Socks4/4a/5 handshake.
func (s *ServerSession) Handshake(reader io.Reader, writer io.Writer) (*protocol.RequestHeader, error) {
	buffer := buf.StackNew()
	if _, err := buffer.ReadFullFrom(reader, 2); err != nil {
		buffer.Release()
		return nil, errors.New("insufficient header").Base(err)
	}

	version := buffer.Byte(0)
	cmd := buffer.Byte(1)
	buffer.Release()

	switch version {
	case socks4Version:
		return s.handshake4(cmd, reader, writer)
	case socks5Version:
		return s.handshake5(cmd, reader, writer)
	default:
		return nil, errors.New("unknown Socks version: ", version)
	}
}

// ReadUsernamePassword reads Socks 5 username/password message from the given reader.
// +----+------+----------+------+----------+
// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
// +----+------+----------+------+----------+
// | 1  |  1   | 1 to 255 |  1   | 1 to 255 |
// +----+------+----------+------+----------+
func ReadUsernamePassword(reader io.Reader) (string, string, error) {
	buffer := buf.StackNew()
	defer buffer.Release()

	if _, err := buffer.ReadFullFrom(reader, 2); err != nil {
		return "", "", err
	}
	nUsername := int32(buffer.Byte(1))

	buffer.Clear()
	if _, err := buffer.ReadFullFrom(reader, nUsername); err != nil {
		return "", "", err
	}
	username := buffer.String()

	buffer.Clear()
	if _, err := buffer.ReadFullFrom(reader, 1); err != nil {
		return "", "", err
	}
	nPassword := int32(buffer.Byte(0))

	buffer.Clear()
	if _, err := buffer.ReadFullFrom(reader, nPassword); err != nil {
		return "", "", err
	}
	password := buffer.String()
	return username, password, nil
}

// ReadUntilNull reads content from given reader, until a null (0x00) byte.
func ReadUntilNull(reader io.Reader) (string, error) {
	b := buf.StackNew()
	defer b.Release()

	for {
		_, err := b.ReadFullFrom(reader, 1)
		if err != nil {
			return "", err
		}
		if b.Byte(b.Len()-1) == 0x00 {
			b.Resize(0, b.Len()-1)
			return b.String(), nil
		}
		if b.IsFull() {
			return "", errors.New("buffer overrun")
		}
	}
}

func hasAuthMethod(expectedAuth byte, authCandidates []byte) bool {
	for _, a := range authCandidates {
		if a == expectedAuth {
			return true
		}
	}
	return false
}

func writeSocks5AuthenticationResponse(writer io.Writer, version byte, auth byte) error {
	return buf.WriteAllBytes(writer, []byte{version, auth})
}

func writeSocks5Response(writer io.Writer, errCode byte, address net.Address, port net.Port) error {
	buffer := buf.New()
	defer buffer.Release()

	common.Must2(buffer.Write([]byte{socks5Version, errCode, 0x00 /* reserved */}))
	if err := addrParser.WriteAddressPort(buffer, address, port); err != nil {
		return err
	}

	return buf.WriteAllBytes(writer, buffer.Bytes())
}

func writeSocks4Response(writer io.Writer, errCode byte, address net.Address, port net.Port) error {
	buffer := buf.StackNew()
	defer buffer.Release()

	common.Must(buffer.WriteByte(0x00))
	common.Must(buffer.WriteByte(errCode))
	portBytes := buffer.Extend(2)
	binary.BigEndian.PutUint16(portBytes, port.Value())
	common.Must2(buffer.Write(address.IP()))
	return buf.WriteAllBytes(writer, buffer.Bytes())
}

func DecodeUDPPacket(packet *buf.Buffer) (*protocol.RequestHeader, error) {
	if packet.Len() < 5 {
		return nil, errors.New("insufficient length of packet.")
	}
	request := &protocol.RequestHeader{
		Version: socks5Version,
		Command: protocol.RequestCommandUDP,
	}

	// packet[0] and packet[1] are reserved
	if packet.Byte(2) != 0 /* fragments */ {
		return nil, errors.New("discarding fragmented payload.")
	}

	packet.AdvanceStart(3)

	addr, port, err := addrParser.ReadAddressPort(nil, packet)
	if err != nil {
		return nil, errors.New("failed to read UDP header").Base(err)
	}
	request.Address = addr
	request.Port = port
	return request, nil
}

func EncodeUDPPacket(request *protocol.RequestHeader, data []byte) (*buf.Buffer, error) {
	b := buf.New()
	common.Must2(b.Write([]byte{0, 0, 0 /* Fragment */}))
	if err := addrParser.WriteAddressPort(b, request.Address, request.Port); err != nil {
		b.Release()
		return nil, err
	}
	common.Must2(b.Write(data))
	return b, nil
}

func EncodeUDPPacketFromAddress(address net.Destination, data []byte) (*buf.Buffer, error) {
	b := buf.New()
	common.Must2(b.Write([]byte{0, 0, 0 /* Fragment */}))
	if err := addrParser.WriteAddressPort(b, address.Address, address.Port); err != nil {
		b.Release()
		return nil, err
	}
	common.Must2(b.Write(data))
	return b, nil
}

type UDPReader struct {
	reader io.Reader
}

func NewUDPReader(reader io.Reader) *UDPReader {
	return &UDPReader{reader: reader}
}
func (r *UDPReader) Read(p []byte) (int, error) {
	buffer := buf.FromBytes(p)
	_, err := buffer.ReadOnce(r.reader)
	if err != nil {
		return 0, err
	}
	_, err = DecodeUDPPacket(buffer)
	if err != nil {
		return 0, err
	}
	return int(buffer.Len()), nil
}
func (r *UDPReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	b := buf.New()
	if _, err := b.ReadOnce(r.reader); err != nil {
		b.Release()
		return nil, err
	}
	if _, err := DecodeUDPPacket(b); err != nil {
		b.Release()
		return nil, err
	}
	return buf.MultiBuffer{b}, nil
}
func (r *UDPReader) ReadPacket() (*udp.Packet, error) {
	buffer := buf.New()
	_, err := buffer.ReadOnce(r.reader)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	req, err := DecodeUDPPacket(buffer)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	return &udp.Packet{Source: net.UDPDestination(req.Address, req.Port),
		Payload: buffer}, nil
}

type UDPWriter struct {
	request *protocol.RequestHeader
	writer  io.Writer
}

func NewUDPWriter(request *protocol.RequestHeader, writer io.Writer) *UDPWriter {
	return &UDPWriter{
		request: request,
		writer:  writer,
	}
}

// Write implements io.Writer.
func (w *UDPWriter) Write(b []byte) (int, error) {
	eb, err := EncodeUDPPacket(w.request, b)
	if err != nil {
		return 0, err
	}
	defer eb.Release()
	if _, err := w.writer.Write(eb.Bytes()); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *UDPWriter) WriteTo(payload []byte, addr gonet.Addr) (n int, err error) {
	request := *w.request
	udpAddr := addr.(*gonet.UDPAddr)
	request.Command = protocol.RequestCommandUDP
	request.Address = net.IPAddress(udpAddr.IP)
	request.Port = net.Port(udpAddr.Port)
	packet, err := EncodeUDPPacket(&request, payload)
	if err != nil {
		return 0, err
	}
	_, err = w.writer.Write(packet.Bytes())
	packet.Release()
	return len(payload), err
}

func ClientHandshake(request *protocol.RequestHeader, reader io.Reader, writer io.Writer, delayAuthWrite bool) (*net.Destination, error) {
	authByte := byte(authNotRequired)
	if request.Account != nil {
		authByte = byte(authPassword)
	}

	b := buf.New()
	defer b.Release()

	common.Must2(b.Write([]byte{socks5Version, 0x01, authByte}))
	if !delayAuthWrite {
		if authByte == authPassword {
			account := request.Account.(*User)
			common.Must(b.WriteByte(0x01))
			common.Must(b.WriteByte(byte(len(account.Name))))
			common.Must2(b.WriteString(account.Name))
			common.Must(b.WriteByte(byte(len(account.Secret))))
			common.Must2(b.WriteString(account.Secret))
		}
	}

	if err := buf.WriteAllBytes(writer, b.Bytes()); err != nil {
		return nil, err
	}

	b.Clear()
	if _, err := b.ReadFullFrom(reader, 2); err != nil {
		return nil, err
	}

	if b.Byte(0) != socks5Version {
		return nil, errors.New("unexpected server version: ", b.Byte(0))
	}
	if b.Byte(1) != authByte {
		return nil, errors.New("auth method not supported.")
	}

	if authByte == authPassword {
		b.Clear()
		if delayAuthWrite {
			account := request.Account.(*User)
			common.Must(b.WriteByte(0x01))
			common.Must(b.WriteByte(byte(len(account.Name))))
			common.Must2(b.WriteString(account.Name))
			common.Must(b.WriteByte(byte(len(account.Secret))))
			common.Must2(b.WriteString(account.Secret))
			if err := buf.WriteAllBytes(writer, b.Bytes()); err != nil {
				return nil, err
			}
			b.Clear()
		}
		if _, err := b.ReadFullFrom(reader, 2); err != nil {
			return nil, err
		}
		if b.Byte(1) != 0x00 {
			return nil, errors.New("server rejects account: ", b.Byte(1))
		}
	}

	b.Clear()

	command := byte(cmdTCPConnect)
	if request.Command == protocol.RequestCommandUDP {
		command = byte(cmdUDPAssociate)
	}
	common.Must2(b.Write([]byte{socks5Version, command, 0x00 /* reserved */}))
	if err := addrParser.WriteAddressPort(b, request.Address, request.Port); err != nil {
		return nil, err
	}

	if err := buf.WriteAllBytes(writer, b.Bytes()); err != nil {
		return nil, err
	}

	b.Clear()
	if _, err := b.ReadFullFrom(reader, 3); err != nil {
		return nil, err
	}

	resp := b.Byte(1)
	if resp != 0x00 {
		return nil, errors.New("server rejects request: ", resp)
	}

	b.Clear()

	address, port, err := addrParser.ReadAddressPort(b, reader)
	if err != nil {
		return nil, err
	}

	if request.Command == protocol.RequestCommandUDP {
		return &net.Destination{
			Address: address,
			Port:    port,
			Network: net.Network_UDP,
		}, nil
	}

	return nil, nil
}

func ClientHandshake4(request *protocol.RequestHeader, reader io.Reader, writer io.Writer) error {
	b := buf.New()
	defer b.Release()

	common.Must2(b.Write([]byte{socks4Version, cmdTCPConnect}))
	portBytes := b.Extend(2)
	binary.BigEndian.PutUint16(portBytes, request.Port.Value())
	switch request.Address.Family() {
	case net.AddressFamilyIPv4:
		common.Must2(b.Write(request.Address.IP()))
	case net.AddressFamilyDomain:
		common.Must2(b.Write([]byte{0x00, 0x00, 0x00, 0x01}))
	case net.AddressFamilyIPv6:
		return errors.New("ipv6 is not supported in socks4")
	default:
		panic("Unknown family type.")
	}
	if request.Account != nil {
		account := request.Account.(*User)
		common.Must2(b.WriteString(account.Name))
	}
	common.Must(b.WriteByte(0x00))
	if request.Address.Family() == net.AddressFamilyDomain {
		common.Must2(b.WriteString(request.Address.Domain()))
		common.Must(b.WriteByte(0x00))
	}
	if err := buf.WriteAllBytes(writer, b.Bytes()); err != nil {
		return err
	}

	b.Clear()
	if _, err := b.ReadFullFrom(reader, 8); err != nil {
		return err
	}
	if b.Byte(0) != 0x00 {
		return errors.New("unexpected version of the reply code: ", b.Byte(0))
	}
	if b.Byte(1) != socks4RequestGranted {
		return errors.New("server rejects request: ", b.Byte(1))
	}
	return nil
}
