package vision

import (
	"bytes"
	"strconv"

	"github.com/5vnetwork/vx-core/common/buf"
)

// TODO: Write a real Tls inspector
// type tlsRequestInspector interface {
// 		inspectRequest(b *buf.Buffer)
// 		isTls() int
// }

var (
	Tls13SupportedVersions  = []byte{0x00, 0x2b, 0x00, 0x02, 0x03, 0x04}
	TlsClientHandShakeStart = []byte{0x16, 0x03}
	TlsServerHandShakeStart = []byte{0x16, 0x03, 0x03}
	TlsApplicationDataStart = []byte{0x17, 0x03, 0x03}

	Tls13CipherSuiteDic = map[uint16]string{
		0x1301: "TLS_AES_128_GCM_SHA256",
		0x1302: "TLS_AES_256_GCM_SHA384",
		0x1303: "TLS_CHACHA20_POLY1305_SHA256",
		0x1304: "TLS_AES_128_CCM_SHA256",
		0x1305: "TLS_AES_128_CCM_8_SHA256", //
	}
)

// inspect tls reqeust or tls response and find infos
type tlsResponseInspector struct {
	// infos to find
	tls13                     int // 1 for tls1.3, -1 for not tls1.3, 0 for not decided
	tls12Or13                 int
	cipherSuiteString         string
	remainingServerHelloCache *buf.Buffer
	remainingServerHelloLen   int32
}

func newVlessTls13Inspector() *tlsResponseInspector {
	return &tlsResponseInspector{
		remainingServerHelloLen: -1,
	}
}

// to know if it is tls12 or tls13, cipher suite,
// this function set tls12Or13, not isTls13
func (f *tlsResponseInspector) inspectServerHelloInitial(b *buf.Buffer) int {
	if b.IsEmpty() {
		return f.tls12Or13
	}
	if b.Len() >= 79 && bytes.Equal(b.BytesRange(0, 3), TlsServerHandShakeStart) && b.Byte(5) == 0x02 {
		f.tls12Or13 = 1
		shLen := int32(b.Byte(3))<<8 + int32(b.Byte(4)) + 5
		sessionIdLen := int32(b.Byte(43))
		cipherSuite := uint16(b.Byte(sessionIdLen+43+1))<<8 + uint16(b.Byte(43+3+sessionIdLen))
		cs, ok := Tls13CipherSuiteDic[cipherSuite]
		if !ok {
			cs = "Old cipher: " + strconv.FormatUint(uint64(cipherSuite), 16)
		}
		f.cipherSuiteString = cs

		bytesInspected := int32(43 + 3 + sessionIdLen)
		f.remainingServerHelloLen = shLen - int32(bytesInspected)
		b.AdvanceStart(bytesInspected) //after advancing, the cursor is at the beginning of compression method
	} else {
		f.tls12Or13 = -1
	}
	return f.tls12Or13
}

// b will be advanced
// check tls version
func (f *tlsResponseInspector) serverHelloVersionInspect(b *buf.Buffer) int {
	if f.remainingServerHelloCache == nil {
		if b.Len() >= f.remainingServerHelloLen {
			if bytes.Contains(b.Bytes(), Tls13SupportedVersions) {
				f.tls13 = 1
			} else {
				f.tls13 = -1
			}
			b.AdvanceStart(f.remainingServerHelloLen)
		} else {
			f.remainingServerHelloCache = buf.NewWithSize(f.remainingServerHelloLen)
			// cache remaining server hello. When have all remaining server hello, check if it is tls1.3
			n, _ := f.remainingServerHelloCache.ReadFullFrom(b, b.Len())
			b.AdvanceStart(int32(n))
		}
	} else {
		n, _ := f.remainingServerHelloCache.ReadFullFrom(b,
			min(f.remainingServerHelloLen-f.remainingServerHelloCache.Len(), b.Len()))
		b.AdvanceStart(int32(n))
		// now all remaining server hello is cached, okay to check if it is tls1.3
		if f.remainingServerHelloCache.Len() >= f.remainingServerHelloLen {
			if bytes.Contains(f.remainingServerHelloCache.Bytes(), Tls13SupportedVersions) {
				f.tls13 = 1
			} else {
				f.tls13 = -1
			}
			f.remainingServerHelloCache.Release()
		}
	}
	return f.tls13
}

// type tlsRequestInspector struct {
// 	IsTls int
// 	// request
// 	clientHelloDone bool
// 	clientHelloBuf  [6]byte
// 	clientHelloBufSlice []byte
// 	// tls12AppDataRecordFound bool
// }

// the start of the tls app data record header should be at buffer[0].
// todo how to make sure record header start at buffer[0]
// if found, tlsRecord is a complete tls record containing app data. remaining buffer is the rest of the buffer
func extractCompleteAppDataRecord(buffer *buf.Buffer) (found bool, tlsRecord *buf.Buffer, remaningBuffer *buf.Buffer) {
	if buffer.Len() >= 6 && bytes.Equal(buffer.BytesRange(0, 3), TlsApplicationDataStart) {
		recordLen := int32(buffer.Byte(3))<<8 + int32(buffer.Byte(4)) + 5
		if buffer.Len() >= recordLen {
			found = true
			tlsRecord = buffer
			tlsRecord = buf.FromBytes(buffer.BytesRange(0, recordLen))
			remaningBuffer = buf.FromBytes(buffer.BytesRange(recordLen, buffer.Len()))
		} else {
			return
		}
	}
	return
}

func peekClientHello(b []byte) (isTls bool) {
	// not a tls
	if len(b) >= 6 && (bytes.Equal(b[0:2], TlsClientHandShakeStart) && b[5] == 0x01) {
		isTls = true
	} else {
		isTls = false
	}
	return
}

// func readVlessHeader(b *buf.Buffer) bool {
// 	if b.Len() < 26 {
// 		return false
// 	}
// 	b.Advance(17)
// 	addOnLen := int32(b.Byte(0))
// 	if b.Len() < addOnLen+4 {
// 		return false
// 	}
// 	b.Advance(addOnLen + 4)
// 	// now b is at the beginning of a address type
// 	addrType := b.Byte(0)
// 	if addrType == 0x01 {
// 		b.Advance(4 + 1)
// 	} else if addrType == 0x03 {
// 		b.Advance(int32(b.Byte(1)) + 2)
// 	} else if addrType == 0x04 {
// 		b.Advance(16 + 1)
// 	} else {
// 		return false
// 	}
// 	return true
// 	// f.vlessHeaderDone = true
// 	// f.requestProgress++
// 	// if !b.IsEmpty() && !f.hasFinished() {
// 	// 	f.inspectRequest(b)
// 	// }
// }
