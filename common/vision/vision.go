package vision

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"math/rand/v2"
	"reflect"
	"unsafe"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/bytespool"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	mytls "github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/rs/zerolog/log"
)

// This can be used on both client and server side
// It examines TLS traffic, add paddings to initial traffic, and
// direct copy traffic if possible
type vConn struct {
	// the conn given by the caller
	net.Conn
	// initially Conn, if direct copy is on, it will be rawConn
	readerConn net.Conn // the conn used to read from
	writerConn net.Conn // the conn used to write to
	ctx        context.Context
	isClient   bool

	isTls       int
	isTls12Or13 int
	isTls13     int
	// enableDirectCopy    bool

	// visionWrite
	foundAppDataRecord bool
	remainingTimes     int
	hasSentLast        bool
	// protocolHeader                            *buf.Buffer //client only
	headerLen                                 int
	headerWritten                             bool
	visionWriteTimesAfterConfirmingTlsAbove12 int
	// hasReadHeader      bool                    //client only
	// headerReaderFunc   func(*buf.Buffer) error //client only

	// visionRead
	rawInput                                           buf.MultiBuffer
	enableReceivingDirectCopy, enableWritingDirectCopy bool
	remainingContent                                   int32
	paddingLen                                         int32
	paddingGarbage                                     buf.NoOpWriter //allocate one at starting
	hasReceivedLast                                    bool
	inspector                                          *tlsResponseInspector //server only

}

// header is a protocol header and will be written to the [Conn] together with the first payload
func NewVisionConn(ctx context.Context, conn net.Conn, isClient bool, headerLen int) net.Conn {
	c := &vConn{
		Conn:           conn,
		readerConn:     conn,
		writerConn:     conn,
		remainingTimes: 10,
		visionWriteTimesAfterConfirmingTlsAbove12: 3,
		paddingGarbage: buf.DiscardReader,
		ctx:            ctx,
		isClient:       isClient,
		headerLen:      headerLen,
	}
	if !isClient {
		c.inspector = newVlessTls13Inspector()
	} else {
		// c.protocolHeader = header
	}
	// else {
	// 	c.headerReaderFunc = headerReaderFunc
	// }
	return c
}

func (c *vConn) Close() error {
	if c.writerConn == c.Conn {
		return c.Conn.Close()
	} else {
		return c.writerConn.Close()
	}
}

func (c *vConn) Write(b []byte) (int, error) {
	if !c.hasSentLast {
		return c.visionWrite(b)
	}

	return c.writerConn.Write(b)
}

func (c *vConn) OkayToUnwrapReader() int {
	if c.hasReceivedLast && c.remainingContent == 0 && c.rawInput.Len() == 0 {
		return 1
	}
	return -1
	// if c.directCopyRead == 1 && c.rawInput == nil {
	// 	return 1
	// }
	// if c.directCopyRead == -1 {
	// 	return -1
	// }
	// return 0
}

func (c *vConn) UnwrapReader() any {
	return c.readerConn
}

func (c *vConn) OkayToUnwrapWriter() int {
	if c.hasSentLast {
		return 1
	}
	return 0
}

func (c *vConn) UnwrapWriter() any {
	return c.writerConn
}

func (c *vConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	buffer := buf.New()
	n, err := buffer.ReadOnce(c)
	if n == 0 {
		buffer.Release()
		return nil, err
	}
	return buf.MultiBuffer{buffer}, err
}

func (c *vConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if mb.IsEmpty() {
		_, err := c.Write([]byte{})
		return err
	}
	for !mb.IsEmpty() {
		_, err := c.Write(mb[0].Bytes())
		if err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
		mb[0].Release()
		mb = mb[1:]
	}
	return nil
}

func (c *vConn) CloseWrite() error {
	// this means directCopy is not on
	if c.writerConn == c.Conn {
		if cw, ok := c.Conn.(buf.CloseWriter); ok {
			return cw.CloseWrite()
		}
	} else { //direct copy is on
		if cw, ok := c.writerConn.(buf.CloseWriter); ok {
			return cw.CloseWrite()
		}
	}
	return nil
}

func (c *vConn) Read(b []byte) (int, error) {
	if !c.hasReceivedLast || c.remainingContent > 0 {
		n, err := c.visionRead(b)
		// peer send a block with 0 content length, read again
		if n == 0 && err == nil {
			return c.Read(b)
		} else {
			return n, err
		}
	}
	if c.rawInput.Len() > 0 {
		mb, n := buf.SplitBytes(c.rawInput, b)
		c.rawInput = mb
		if c.rawInput.Len() == 0 {
			c.rawInput = nil
		}
		return n, nil
	}
	return c.readerConn.Read(b)
}

func (c *vConn) visionRead(b []byte) (int, error) {
	// buffer := buf.FromBytes(b)
	if c.remainingContent == 0 {
		var hb [5]byte
		n, err := io.ReadFull(c.readerConn, hb[:])
		if err != nil {
			return n, errors.New("cannot read a vision header", err)
		}
		c.decodeHeader(hb[:])
	}

	// read content
	b = b[:min(c.remainingContent, int32(len(b)))]
	n, err := io.ReadFull(c.readerConn, b)
	c.remainingContent -= int32(n)
	if err != nil {
		return n, errors.New("cannot read desired content", err)
	}

	if c.remainingContent == 0 {
		// discard padding
		if c.paddingLen > 0 {
			_, err := c.paddingGarbage.ReadFullFrom(c.readerConn, c.paddingLen)
			if err != nil {
				return n, errors.New("cannot read padding", err)
			}
		}
	}

	if c.hasReceivedLast && c.remainingContent == 0 && c.enableReceivingDirectCopy {
		err = c.startDirectCopy(true)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (c *vConn) visionWrite(b []byte) (int, error) {
	c.remainingTimes--
	// if len(b) == 0 {
	// 	return 0, nil
	// }
	if c.remainingTimes <= 0 {
		return c.encodeHeaderAndWrite(true, b)
	}

	if c.isClient {
		return c.clientVisionWrite(b)
	} else {
		return c.serverVisionWrite(b)
	}
}

func (c *vConn) clientVisionWrite(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, common.Error2(c.encodeHeaderAndWrite(false, b))
	}

	if c.isTls == 0 {
		// the initial data must be larger than the header length
		if len(b) < c.headerLen {
			return 0, errors.New("unexpected length of initial traffic")
		}
		// TODO: handler the case that len is exactly equal to the header length
		isTls := peekClientHello(b[c.headerLen:])
		log.Ctx(c.ctx).Debug().Bool("isTls", isTls).Msg("peekClientHello")
		if isTls {
			c.isTls = 1
		} else {
			c.isTls = -1
			return c.encodeHeaderAndWrite(true, b)
		}
	}

	// at this point, c.isTls == 1
	if c.isTls12Or13 == 0 {
		return c.encodeHeaderAndWrite(false, b)
	} else if c.isTls12Or13 == -1 {
		return c.encodeHeaderAndWrite(true, b)
	}

	buffer := buf.FromBytes(b)
	// at this point, c.isTls12Or13 == 1
	var numBytesWritten int
	found, rcrdBuf, remainigBuf := extractCompleteAppDataRecord(buffer)
	c.foundAppDataRecord = found

	// if found, this app data record will be the last record that is visioned, records following it will
	// be either direct copy(tls13) or normal write(tls12)
	if found {
		// all of the record is in rcrdBuf, this call is the last call of visionWrite
		n, err := c.encodeHeaderAndWrite(true, rcrdBuf.Bytes())
		if err != nil {
			return n, err
		}
		numBytesWritten += n
		// start direct copy
		if c.enableWritingDirectCopy {
			err = c.startDirectCopy(false)
			if err != nil {
				return numBytesWritten, err
			}
		}
		if !remainigBuf.IsEmpty() {
			n, err = c.writerConn.Write(remainigBuf.Bytes())
			numBytesWritten += n
		}
		return numBytesWritten, err
	} else {
		return c.encodeHeaderAndWrite(false, b)
	}
}

func (c *vConn) serverVisionWrite(b []byte) (int, error) {
	if len(b) == 0 {
		// return c.encodeHeaderAndWrite(false, b)
		return 0, common.Error2(c.encodeHeaderAndWrite(false, b))
	}

	if c.isTls == -1 {
		return c.encodeHeaderAndWrite(true, b)
	}

	buffer := buf.FromBytes(b)

	// tls12Or13 == 0 if and only if inspectServerHelloInitial has not been called
	if c.isTls12Or13 == 0 {
		c.isTls12Or13 = c.inspector.inspectServerHelloInitial(buffer)
		log.Ctx(c.ctx).Debug().Bool("isTls12Or13", c.isTls12Or13 == 1).Msg("serverHello")
	}

	if c.isTls12Or13 == -1 {
		return c.encodeHeaderAndWrite(true, b)
	}

	// at this point, tls12Or13 is 1, and tls13 is yet to be determined
	if c.isTls13 == 0 {
		c.isTls13 = c.inspector.serverHelloVersionInspect(buffer)
		// c.isTls13 might be 1, -1, or 0 now. No matter what it is, we sends b
		return c.encodeHeaderAndWrite(false, b)
	}

	// at this point, tls13 is 1 or -1. We need to find application data after two vision writes
	if c.visionWriteTimesAfterConfirmingTlsAbove12 > 0 {
		c.visionWriteTimesAfterConfirmingTlsAbove12--
		return c.encodeHeaderAndWrite(false, b)
	}

	found, rcrdBuf, remainigBuf := extractCompleteAppDataRecord(buffer)
	c.foundAppDataRecord = found
	if found {
		n, err := c.encodeHeaderAndWrite(true, rcrdBuf.Bytes())
		if err != nil {
			return n, err
		}
		numBytesWritten := n
		if c.enableWritingDirectCopy {
			err = c.startDirectCopy(false)
			if err != nil {
				return numBytesWritten, err
			}
		}
		if !remainigBuf.IsEmpty() {
			n, err = c.writerConn.Write(remainigBuf.Bytes())
			numBytesWritten += n
		}
		// // downstream splice copy
		// if c.isTls13 == 1 {
		// 	// info := session.InfoFromContext(c.ctx)
		// 	// info.SpliceCopy.Down = true
		// } else {
		// }
		return numBytesWritten, err
	} else {
		return c.encodeHeaderAndWrite(false, b)
	}
}

// change readerConn or writerConn to rawConn
func (c *vConn) startDirectCopy(reading bool) error {
	log.Ctx(c.ctx).Debug().Bool("isReading", reading).Msg("startDirectCopy")
	if reading {
		conn0 := c.Conn
		var mb buf.MultiBuffer
		if mbConn, ok := conn0.(*net.MbConn); ok {
			conn0 = mbConn.Conn
			if mbConn.Mb.Len() > 0 {
				mb = mbConn.Mb
				mbConn.Mb = nil
			}
		}
		conn0, err := getSecurityConn(conn0)
		if err != nil {
			return err
		}
		// reality
		var rawInput *bytes.Buffer
		if conn, ok := conn0.(*mytls.Conn); ok {
			f, _ := reflect.TypeOf(conn.Conn).Elem().FieldByName("rawInput")
			rawInput = (*bytes.Buffer)(unsafe.Pointer(uintptr(unsafe.Pointer(conn.Conn)) + f.Offset))
		} else if conn, ok := conn0.(*tls.Conn); ok {
			f, _ := reflect.TypeOf(conn).Elem().FieldByName("rawInput")
			rawInput = (*bytes.Buffer)(unsafe.Pointer(uintptr(unsafe.Pointer(conn)) + f.Offset))
		} else if conn, ok := conn0.(*mytls.UConn); ok {
			f, _ := reflect.TypeOf(conn.Conn).Elem().FieldByName("rawInput")
			rawInput = (*bytes.Buffer)(unsafe.Pointer(uintptr(unsafe.Pointer(conn.Conn)) + f.Offset))
		} else if conn, ok := conn0.(*reality.UConn); ok {
			f, _ := reflect.TypeOf(conn.Conn).Elem().FieldByName("rawInput")
			rawInput = (*bytes.Buffer)(unsafe.Pointer(uintptr(unsafe.Pointer(conn.Conn)) + f.Offset))
		} else if conn, ok := conn0.(*reality.Conn); ok {
			f, _ := reflect.TypeOf(conn.Conn).Elem().FieldByName("rawInput")
			rawInput = (*bytes.Buffer)(unsafe.Pointer(uintptr(unsafe.Pointer(conn.Conn)) + f.Offset))
		} else {
			return errors.New("unexpected concrete type of c.Conn")
		}

		if rawInput.Len() != 0 {
			b := buf.NewWithSize(int32(rawInput.Len()))
			b.ReadFullFrom(rawInput, int32(rawInput.Len()))
			mb = append(mb, b)
		}
		c.rawInput = mb

		c.readerConn = conn0.(net.HasNetConn).NetConn()
	} else {
		rawConn, err := getRawConn(c.ctx, c.Conn)
		if err != nil {
			return err
		}
		c.writerConn = rawConn
	}
	return nil
}

func getSecurityConn(conn0 net.Conn) (net.Conn, error) {
	for {
		if mbConn, ok := conn0.(*net.MbConn); ok {
			conn0 = mbConn.NetConn()
		} else {
			if conn, ok := conn0.(*mytls.Conn); ok {
				return conn, nil
			} else if conn, ok := conn0.(*tls.Conn); ok {
				return conn, nil
			} else if conn, ok := conn0.(*mytls.UConn); ok {
				return conn, nil
			} else if conn, ok := conn0.(*reality.UConn); ok {
				return conn, nil
			} else if conn, ok := conn0.(*reality.Conn); ok {
				return conn, nil
			} else {
				return nil, errors.New("unexpected concrete type of conn0")
			}
		}

	}
}

func getRawConn(ctx context.Context, conn net.Conn) (net.Conn, error) {
	conn, err := getSecurityConn(conn)
	if err != nil {
		return nil, err
	}
	netConn, _ := conn.(net.HasNetConn)
	return netConn.NetConn(), nil
}

func (c *vConn) decodeHeader(b []byte) {
	isLst, isTls, isTls12Or13, isTls13, enableReceivingDirectCopy, contentLen, paddingLen := decodeHeader(b)
	c.hasReceivedLast = isLst
	if isLst {
		c.enableReceivingDirectCopy = enableReceivingDirectCopy
	}
	if c.isClient {
		// a server can know tls version when it inspecting the first payloads.
		// so isTls12Or13 and isTls13 values are populated
		if c.isTls12Or13 == 0 && contentLen != 0 {
			if isTls12Or13 {
				c.isTls12Or13 = 1
			} else {
				c.isTls12Or13 = -1
			}
		}
		if c.isTls13 == 0 && contentLen != 0 {
			if isTls13 {
				c.isTls13 = 1
			} else {
				c.isTls13 = -1
			}
		}
		log.Ctx(c.ctx).Debug().Msgf("client got a header: isLast=%v, isTls12Or13=%v, isTls13=%v, enableDC=%v, contentLen=%v, paddingLen=%v",
			isLst, isTls12Or13, isTls13, enableReceivingDirectCopy, contentLen, paddingLen)
	} else {
		if contentLen != 0 && c.isTls == 0 {
			if isTls {
				c.isTls = 1
			} else {
				c.isTls = -1
			}
		}
		log.Ctx(c.ctx).Debug().Msgf("server got a header: isLast=%v, isTls=%v, enableDC=%v, contentLen=%v, paddingLen=%v",
			isLst, isTls, enableReceivingDirectCopy, contentLen, paddingLen)

	}
	c.remainingContent = int32(contentLen)
	c.paddingLen = int32(paddingLen)

}

func (c *vConn) setHasSentLast() {
	_, err := getSecurityConn(c.Conn)
	c.enableWritingDirectCopy = c.isTls13 == 1 && c.foundAppDataRecord && err == nil
	c.hasSentLast = true
}

func (c *vConn) encodeHeaderAndWrite(isLast bool, contentBuffer []byte) (int, error) {
	if isLast {
		c.setHasSentLast()
	}
	padLen := c.makePaddingDecision(len(contentBuffer))
	blockLen := int32(len(contentBuffer) + 5 + padLen)
	b := bytespool.Alloc(blockLen)
	defer bytespool.Free(b)

	if c.isClient {
		if !c.headerWritten {
			c.headerWritten = true
			copy(b[:c.headerLen], contentBuffer[:c.headerLen])
			encodeHeader(b[c.headerLen:c.headerLen+5], isLast, c.isTls == 1, false,
				false, c.enableWritingDirectCopy, uint16(len(contentBuffer)-c.headerLen), uint16(padLen))
			copy(b[c.headerLen+5:], contentBuffer[c.headerLen:])
		} else {
			encodeHeader(b[:5], isLast, c.isTls == 1, false,
				false, c.enableWritingDirectCopy, uint16(len(contentBuffer)), uint16(padLen))
			copy(b[5:], contentBuffer)
		}
	} else {
		encodeHeader(b[:5], isLast, false, c.isTls12Or13 == 1,
			c.isTls13 == 1, c.enableWritingDirectCopy, uint16(len(contentBuffer)), uint16(padLen))
		copy(b[5:], contentBuffer)
	}
	n, err := c.writerConn.Write(b[:blockLen])
	if n >= 5+len(contentBuffer) {
		return len(contentBuffer), err
	} else if n >= 5 {
		return n - 5, err
	} else {
		return 0, err
	}
}

func (c *vConn) makePaddingDecision(len int) (paddingLen int) {
	if len < 900 {
		return 900 - len + rand.IntN(500)
	}
	return 0

}

const (
	last      = 0b10000000
	isTlsBit  = 0b01000000
	tls12Or13 = 0b00100000
	tls13     = 0b00010000
	// This bit is meaningful only when [last] bit is set
	// This bit is to inform peer that future traffic will be directly copied
	directCopy = 0b00001000
)

// b is a 5-byte header
func decodeHeader(b []byte) (isLast bool, isTls bool, isTls12Or13 bool, isTls13 bool, enableDC bool,
	contentLen uint16, paddingLen uint16) {
	firstByte := b[0]
	isLast = last&firstByte > 0
	isTls = isTlsBit&firstByte > 0
	isTls12Or13 = tls12Or13&firstByte > 0
	isTls13 = tls13&firstByte > 0
	enableDC = directCopy&firstByte > 0
	contentLen = (uint16(b[1]) << 8) | uint16(b[2])
	paddingLen = (uint16(b[3]) << 8) | uint16(b[4])
	return
}

// encode a header and write it into b which is slice of length 5
func encodeHeader(b []byte, isLast bool, isTls bool, isTls12Or13 bool, isTls13 bool, enableDC bool,
	contentLen uint16, paddingLen uint16) {
	var firstByte byte
	if isLast {
		firstByte |= last
	}
	if isTls {
		firstByte |= isTlsBit
	}
	if isTls12Or13 {
		firstByte |= tls12Or13
	}
	if isTls13 {
		firstByte |= tls13
	}
	if enableDC {
		firstByte |= directCopy
	}
	b[0] = firstByte
	b[1] = byte(contentLen >> 8)
	b[2] = byte(contentLen)
	b[3] = byte(paddingLen >> 8)
	b[4] = byte(paddingLen)
}
