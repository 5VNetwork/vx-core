package tcpip

// func TestIsIPv4(t *testing.T) {
// 	h := ipv4.Header{
// 		Version:  4,
// 		Src:      []byte{1, 1, 1, 1},
// 		Dst:      []byte{2, 2, 2, 2},
// 		Protocol: 6,
// 		Len:      20,
// 		TTL:      64,
// 	}
// 	headerBytes, err := h.Marshal()
// 	common.Must(err)
// 	b := buf.New()
// 	b.Write(headerBytes)
// 	b.WriteString("hello")
// 	if IsIPv4(b.Bytes()) == false {
// 		t.Errorf("IsIPv4() = %v; want true", IsIPv4(b.Bytes()))
// 	}
// }
