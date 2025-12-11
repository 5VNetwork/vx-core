package tcpip

// func TestIpv4PseudoSum(t *testing.T) {
// 	ipPakcetString := "45000042c45f00008011f47cc0a8007dc0a80001ce780035002e514899ea0100000100000000000009636f6c6c6563746f720667697468756203636f6d0000010001"
// 	ipPacketBytes, err := hex.DecodeString(ipPakcetString)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	ipPacket := header.IPv4(ipPacketBytes)
// 	pseudoSum := header.PseudoHeaderChecksum(ipPacket.TransportProtocol(), ipPacket.SourceAddress(),
// 		ipPacket.DestinationAddress(), ipPacket.PayloadLength())

// }
