package vi_test

// func TestEncodingDecoding(t *testing.T) {
// 	u, err := vi.NewUser("12345678-1234-1234-1234-123456789012", 0, "test")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	header := &protocol.RequestHeader{
// 		Version: 0,
// 		Command: protocol.RequestCommandTCP,
// 		Dst:     net.Tuple2{Address: net.AnyIP, Port: 8},
// 		User:    u,
// 	}

// 	buffer, err := EncodeHeader(header)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	v := vi.Validator{}
// 	v.Add(u)
// 	h, err := DecodeHeader(buffer, &v)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if cmp.Diff(header, h) != "" {
// 		t.Fatal("unexpected output")
// 	}
// }
