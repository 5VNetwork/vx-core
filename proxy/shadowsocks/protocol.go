package shadowsocks

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"io"
	mrand "math/rand"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/drain"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
)

const (
	Version = 1
)

var addrParser = address_parser.SocksAddressSerializer

// ReadTCPSession reads a Shadowsocks TCP session from the given reader, returns its header and remaining parts.
func ReadTCPSession(account *MemoryAccount, reader io.Reader, drainOnError bool) (*protocol.RequestHeader, buf.Reader, error) {

	hashkdf := hmac.New(sha256.New, []byte("SSBSKDF"))
	hashkdf.Write(account.Key)

	behaviorSeed := crc32.ChecksumIEEE(hashkdf.Sum(nil))

	var drainer drain.Drainer
	if drainOnError {
		var err error
		drainer, err = drain.NewBehaviorSeedLimitedDrainer(int64(behaviorSeed), 16+38, 3266, 64)
		if err != nil {
			return nil, nil, errors.New("failed to initialize drainer").Base(err)
		}
	} else {
		drainer = drain.NewNopDrainer()
	}

	buffer := buf.New()
	defer buffer.Release()

	ivLen := account.Cipher.IVSize()
	var iv []byte
	if ivLen > 0 {
		if drainOnError {
			if _, err := buffer.ReadFullFrom(reader, ivLen); err != nil {
				drainer.AcknowledgeReceive(int(buffer.Len()))
				return nil, nil, drain.WithError(drainer, reader, errors.New("failed to read IV").Base(err))
			}
		} else {
			// in this case, this function is called by fallback process
			if _, err := buffer.ReadOnceWithSize(reader, ivLen); err != nil {
				return nil, nil, drain.WithError(drainer, reader, errors.New("failed to read IV").Base(err))
			}
			if buffer.Len() != ivLen {
				return nil, nil, errors.New("failed to read IV: expected %d bytes, got %d", ivLen, buffer.Len())
			}
		}

		iv = append([]byte(nil), buffer.BytesTo(ivLen)...)
	}

	r, err := account.Cipher.NewDecryptionReader(account.Key, iv, reader)
	if err != nil {
		drainer.AcknowledgeReceive(int(buffer.Len()))
		return nil, nil, drain.WithError(drainer, reader, errors.New("failed to initialize decoding stream").Base(err))
	}
	br := &buf.BufferedReader{Reader: r}

	request := &protocol.RequestHeader{
		Version: Version,
		User:    account.Uid,
		Account: account,
		Command: protocol.RequestCommandTCP,
	}

	drainer.AcknowledgeReceive(int(buffer.Len()))
	buffer.Clear()

	addr, port, err := addrParser.ReadAddressPort(buffer, br)
	if err != nil {
		drainer.AcknowledgeReceive(int(buffer.Len()))
		return nil, nil, drain.WithError(drainer, reader, fmt.Errorf("failed to read address, %w", err))
	}

	request.Address = addr
	request.Port = port

	if request.Address == nil {
		drainer.AcknowledgeReceive(int(buffer.Len()))
		return nil, nil, drain.WithError(drainer, reader, errors.New("invalid remote address."))
	}

	if ivError := account.CheckIV(iv); ivError != nil {
		drainer.AcknowledgeReceive(int(buffer.Len()))
		return nil, nil, drain.WithError(drainer, reader, errors.New("failed iv check").Base(ivError))
	}

	return request, br, nil
}

// WriteTCPRequest writes Shadowsocks request into the given writer, and returns a writer for body.
func WriteTCPRequest(request *protocol.RequestHeader, writer io.Writer) (buf.Writer, error) {
	account := request.Account.(*MemoryAccount)

	var iv []byte
	if account.Cipher.IVSize() > 0 {
		iv = make([]byte, account.Cipher.IVSize())
		common.Must2(rand.Read(iv))
		if account.ReducedIVEntropy {
			remapToPrintable(iv[:6])
		}
		if ivError := account.CheckIV(iv); ivError != nil {
			return nil, errors.New("failed to mark outgoing iv").Base(ivError)
		}
		if err := buf.WriteAllBytes(writer, iv); err != nil {
			return nil, errors.New("failed to write IV")
		}
	}

	w, err := account.Cipher.NewEncryptionWriter(account.Key, iv, writer)
	if err != nil {
		return nil, errors.New("failed to create encoding stream").Base(err)
	}

	header := buf.New()

	if err := addrParser.WriteAddressPort(header, request.Address, request.Port); err != nil {
		return nil, errors.New("failed to write address").Base(err)
	}

	if err := w.WriteMultiBuffer(buf.MultiBuffer{header}); err != nil {
		return nil, errors.New("failed to write header").Base(err)
	}

	return w, nil
}

// WriteTCPRequest writes Shadowsocks request into the given writer, and returns a writer for body.
func WriteTCPRequestIO(request *protocol.RequestHeader, writer io.Writer) (io.Writer, error) {
	account := request.Account.(*MemoryAccount)

	var iv []byte
	if account.Cipher.IVSize() > 0 {
		iv = make([]byte, account.Cipher.IVSize())
		common.Must2(rand.Read(iv))
		if account.ReducedIVEntropy {
			remapToPrintable(iv[:6])
		}
		if ivError := account.CheckIV(iv); ivError != nil {
			return nil, errors.New("failed to mark outgoing iv").Base(ivError)
		}
		if err := buf.WriteAllBytes(writer, iv); err != nil {
			return nil, errors.New("failed to write IV")
		}
	}

	w, err := account.Cipher.NewEncryptionWriterIO(account.Key, iv, writer)
	if err != nil {
		return nil, errors.New("failed to create encoding stream").Base(err)
	}

	header := buf.New()

	if err := addrParser.WriteAddressPort(header, request.Address, request.Port); err != nil {
		return nil, errors.New("failed to write address").Base(err)
	}

	_, err = io.Writer.Write(w, header.Bytes())
	if err != nil {
		return nil, errors.New("failed to write header").Base(err)
	}

	return w, nil
}

func ReadTCPResponse(account *MemoryAccount, reader io.Reader) (buf.Reader, error) {

	hashkdf := hmac.New(sha256.New, []byte("SSBSKDF"))
	hashkdf.Write(account.Key)

	behaviorSeed := crc32.ChecksumIEEE(hashkdf.Sum(nil))

	drainer, err := drain.NewBehaviorSeedLimitedDrainer(int64(behaviorSeed), 16+38, 3266, 64)
	if err != nil {
		return nil, errors.New("failed to initialize drainer").Base(err)
	}

	var iv []byte
	if account.Cipher.IVSize() > 0 {
		iv = make([]byte, account.Cipher.IVSize())
		if n, err := io.ReadFull(reader, iv); err != nil {
			return nil, errors.New("failed to read IV").Base(err)
		} else { // nolint: revive
			drainer.AcknowledgeReceive(n)
		}
	}

	if ivError := account.CheckIV(iv); ivError != nil {
		return nil, drain.WithError(drainer, reader, errors.New("failed iv check").Base(ivError))
	}

	return account.Cipher.NewDecryptionReader(account.Key, iv, reader)
}

func ReadTCPResponseIO(account *MemoryAccount, reader io.Reader) (io.Reader, error) {

	hashkdf := hmac.New(sha256.New, []byte("SSBSKDF"))
	hashkdf.Write(account.Key)

	behaviorSeed := crc32.ChecksumIEEE(hashkdf.Sum(nil))

	drainer, err := drain.NewBehaviorSeedLimitedDrainer(int64(behaviorSeed), 16+38, 3266, 64)
	if err != nil {
		return nil, errors.New("failed to initialize drainer").Base(err)
	}

	var iv []byte
	if account.Cipher.IVSize() > 0 {
		iv = make([]byte, account.Cipher.IVSize())
		if n, err := io.ReadFull(reader, iv); err != nil {
			return nil, errors.New("failed to read IV").Base(err)
		} else { // nolint: revive
			drainer.AcknowledgeReceive(n)
		}
	}

	if ivError := account.CheckIV(iv); ivError != nil {
		return nil, drain.WithError(drainer, reader, errors.New("failed iv check").Base(ivError))
	}

	return account.Cipher.NewDecryptionReaderIO(account.Key, iv, reader)
}

func WriteTCPResponse(request *protocol.RequestHeader, writer io.Writer) (buf.Writer, error) {
	account := request.Account.(*MemoryAccount)

	var iv []byte
	if account.Cipher.IVSize() > 0 {
		iv = make([]byte, account.Cipher.IVSize())
		common.Must2(rand.Read(iv))
		if ivError := account.CheckIV(iv); ivError != nil {
			return nil, errors.New("failed to mark outgoing iv").Base(ivError)
		}
		if err := buf.WriteAllBytes(writer, iv); err != nil {
			return nil, errors.New("failed to write IV.").Base(err)
		}
	}

	return account.Cipher.NewEncryptionWriter(account.Key, iv, writer)
}

func EncodeUDPPacket(request *protocol.RequestHeader, payload []byte) (*buf.Buffer, error) {
	account := request.Account.(*MemoryAccount)

	buffer := buf.New()
	ivLen := account.Cipher.IVSize()
	if ivLen > 0 {
		common.Must2(buffer.ReadFullFrom(rand.Reader, ivLen))
	}

	if err := addrParser.WriteAddressPort(buffer, request.Address, request.Port); err != nil {
		return nil, errors.New("failed to write address").Base(err)
	}

	buffer.Write(payload)

	if err := account.Cipher.EncodePacket(account.Key, buffer); err != nil {
		return nil, errors.New("failed to encrypt UDP payload").Base(err)
	}

	return buffer, nil
}

func DecodeUDPPacket(account *MemoryAccount, payload *buf.Buffer) (*protocol.RequestHeader, *buf.Buffer, error) {
	var iv []byte
	if !account.Cipher.IsAEAD() && account.Cipher.IVSize() > 0 {
		// Keep track of IV as it gets removed from payload in DecodePacket.
		iv = make([]byte, account.Cipher.IVSize())
		copy(iv, payload.BytesTo(account.Cipher.IVSize()))
	}

	if err := account.Cipher.DecodePacket(account.Key, payload); err != nil {
		return nil, nil, errors.New("failed to decrypt UDP payload").Base(err)
	}

	request := &protocol.RequestHeader{
		Version: Version,
		User:    account.Uid,
		Account: account,
		Command: protocol.RequestCommandUDP,
	}

	payload.SetByte(0, payload.Byte(0)&0x0F)

	addr, port, err := addrParser.ReadAddressPort(nil, payload)
	if err != nil {
		return nil, nil, errors.New("failed to parse address").Base(err)
	}

	request.Address = addr
	request.Port = port

	return request, payload, nil
}

type UDPReader struct {
	Reader io.Reader
	User   *MemoryAccount
}

func (v *UDPReader) Read(p []byte) (n int, err error) {
	buffer := buf.FromBytes(p)
	_, err = buffer.ReadOnce(v.Reader)
	if err != nil {
		return
	}
	var payload *buf.Buffer
	_, payload, err = DecodeUDPPacket(v.User, buffer)
	if err != nil {
		return
	}
	return int(payload.Len()), nil
}

func (v *UDPReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	buffer := buf.New()
	_, err := buffer.ReadOnce(v.Reader)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	_, payload, err := DecodeUDPPacket(v.User, buffer)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	return buf.MultiBuffer{payload}, nil
}

// func (v *UDPReader) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
// 	buffer := buf.FromBytes(p)
// 	_, err = buffer.ReadOnce(v.Reader)
// 	if err != nil {
// 		return 0, nil, err
// 	}
// 	vaddr, payload, err := DecodeUDPPacket(v.User, buffer)
// 	if err != nil {
// 		return 0, nil, err
// 	}
// 	return int(payload.Len()), &net.UDPAddr{IP: vaddr.Address.IP(), Port: int(vaddr.Port)}, nil
// }

func (v *UDPReader) ReadPacket() (*udp.Packet, error) {
	buffer := buf.New()
	_, err := buffer.ReadOnce(v.Reader)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	vaddr, payload, err := DecodeUDPPacket(v.User, buffer)
	if err != nil {
		buffer.Release()
		return nil, err
	}
	return &udp.Packet{
		Source:  net.UDPDestination(vaddr.Address, vaddr.Port),
		Payload: payload,
	}, nil
}

type UDPWriter struct {
	Writer  io.Writer
	Request *protocol.RequestHeader
}

// Write implements io.Writer.
func (w *UDPWriter) Write(payload []byte) (int, error) {
	packet, err := EncodeUDPPacket(w.Request, payload)
	if err != nil {
		return 0, err
	}
	_, err = w.Writer.Write(packet.Bytes())
	packet.Release()
	return len(payload), err
}

func (w *UDPWriter) WritePacket(packet *udp.Packet) error {
	defer packet.Release()
	request := *w.Request
	request.Command = protocol.RequestCommandUDP
	request.Address = packet.Target.Address
	request.Port = packet.Target.Port
	b, err := EncodeUDPPacket(&request, packet.Payload.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Writer.Write(b.Bytes())
	b.Release()
	return err
}

func (w *UDPWriter) WriteTo(payload []byte, addr net.Addr) (n int, err error) {
	request := *w.Request
	udpAddr := addr.(*net.UDPAddr)
	request.Command = protocol.RequestCommandUDP
	request.Address = net.IPAddress(udpAddr.IP)
	request.Port = net.Port(udpAddr.Port)
	packet, err := EncodeUDPPacket(&request, payload)
	if err != nil {
		return 0, err
	}
	_, err = w.Writer.Write(packet.Bytes())
	packet.Release()
	return len(payload), err
}

func remapToPrintable(input []byte) {
	const charSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!#$%&()*+,./:;<=>?@[]^_`{|}~\\\""
	seed := mrand.New(mrand.NewSource(int64(crc32.ChecksumIEEE(input))))
	for i := range input {
		input[i] = charSet[seed.Intn(len(charSet))]
	}
}
