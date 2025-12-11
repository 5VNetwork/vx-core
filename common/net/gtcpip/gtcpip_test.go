package gtcpip

import (
	"encoding/hex"
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func TestTcpResetChecksum(t *testing.T) {
	ipPacketS := "450000280000400040061daac0a800667916e301c21e01bb2ce098b5c7b4d2525010ffff6f370000"
	ipPacketBytes, err := hex.DecodeString(ipPacketS)
	if err != nil {
		t.Fatal(err)
	}
	ipPacket := NewIPPacket(ipPacketBytes)
	tcpPacket := TCP{
		TCP: header.TCP(ipPacket.Payload()),
	}
	oldChecksum := tcpPacket.Checksum()
	tcpPacket.ResetChecksum(ipPacket.PseudoHeaderChecksum())
	newChecksum := tcpPacket.Checksum()
	if oldChecksum != newChecksum {
		t.Errorf("Checksum mismatch: got %x, want %x", newChecksum, oldChecksum)
	}
}

func TestUdpResetChecksum(t *testing.T) {
	ipPakcetString := "45000042c45f00008011f47cc0a8007dc0a80001ce780035002e514899ea0100000100000000000009636f6c6c6563746f720667697468756203636f6d0000010001"
	ipPacketBytes, err := hex.DecodeString(ipPakcetString)
	if err != nil {
		t.Fatal(err)
	}
	ipPacket := NewIPPacket(ipPacketBytes)
	udpPacket := UDP{
		UDP: header.UDP(ipPacket.Payload()),
	}
	oldChecksum := udpPacket.Checksum()
	udpPacket.ResetChecksum(ipPacket.PseudoHeaderChecksum())
	newChecksum := udpPacket.Checksum()

	if oldChecksum != newChecksum {
		t.Errorf("Checksum mismatch: got %x, want %x", newChecksum, oldChecksum)
	}
}

func TestIpv4ResetChecksum(t *testing.T) {
	ipPakcetString := "45000042c45f00008011f47cc0a8007dc0a80001ce780035002e514899ea0100000100000000000009636f6c6c6563746f720667697468756203636f6d0000010001"
	ipPacketBytes, err := hex.DecodeString(ipPakcetString)
	if err != nil {
		t.Fatal(err)
	}
	ipPacket := NewIPPacket(ipPacketBytes)
	oldChecksum := ipPacket.Checksum()
	ipPacket.ResetChecksum()
	newChecksum := ipPacket.Checksum()
	if oldChecksum != newChecksum {
		t.Errorf("Checksum mismatch: got %x, want %x", newChecksum, oldChecksum)
	}
}
