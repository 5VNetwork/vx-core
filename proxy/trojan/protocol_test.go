package trojan

import (
	"crypto/rand"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/uuid"
)

func TestTCPRequest(t *testing.T) {
	ma := NewMemoryAccount(uuid.New().String(), "password")

	payload := []byte("test string")
	data := buf.New()
	common.Must2(data.Write(payload))

	buffer := buf.New()
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_TCP, Address: net.LocalHostIP, Port: 1234}
	writer := &ConnWriter{Writer: buffer, Target: destination, Account: ma}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	buffer.AdvanceStart(58)
	dst, err := ParseHeader(buffer)
	common.Must(err)

	if r := cmp.Diff(dst, destination); r != "" {
		t.Error("destination: ", r)
	}

	if r := cmp.Diff(buffer.Bytes(), payload); r != "" {
		t.Error("data: ", r)
	}
}

func TestUDPRequest(t *testing.T) {
	ma := NewMemoryAccount(uuid.New().String(), "password")
	payload := []byte("test string")
	data := buf.New()
	common.Must2(data.Write(payload))

	buffer := buf.New()
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_UDP, Address: net.LocalHostIP, Port: 1234}
	writer := &PacketWriter{writer: &ConnWriter{Writer: buffer, Target: destination, Account: ma}, Dest: destination}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	buffer.AdvanceStart(58)
	_, err := ParseHeader(buffer)
	common.Must(err)

	packetReader := &PacketReader{reader: buffer}
	p, err := packetReader.ReadPacket()
	common.Must(err)

	if p.Payload.IsEmpty() {
		t.Error("no request data")
	}

	if r := cmp.Diff(p.Target, destination); r != "" {
		t.Error("destination: ", r)
	}

	if r := cmp.Diff(p.Payload.Bytes(), payload); r != "" {
		t.Error("data: ", r)
	}
}

func TestLargeUDPRequest(t *testing.T) {
	ma := NewMemoryAccount(uuid.New().String(), "password")

	payload := make([]byte, 4096)
	common.Must2(rand.Read(payload))
	data := buf.NewWithSize(int32(len(payload)))
	common.Must2(data.Write(payload))

	buffer := buf.NewWithSize(2*data.Len() + 1)
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_UDP, Address: net.LocalHostIP, Port: 1234}
	writer := &PacketWriter{writer: &ConnWriter{Writer: buffer, Target: destination, Account: ma}, Dest: destination}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data, data}))

	buffer.AdvanceStart(58)
	_, err := ParseHeader(buffer)
	common.Must(err)

	packetReader := &PacketReader{reader: buffer}
	for i := 0; i < 2; i++ {
		p, err := packetReader.ReadPacket()
		common.Must(err)

		if p.Payload.IsEmpty() {
			t.Error("no request data")
		}

		if r := cmp.Diff(p.Target, destination); r != "" {
			t.Error("destination: ", r)
		}

		if r := cmp.Diff(p.Payload.Bytes(), payload); r != "" {
			t.Error("data: ", r)
		}
	}
}
