// Package proxy contains all proxies used by Xray.
//
// To implement an inbound or outbound proxy, one needs to do the following:
// 1. Implement the interface(s) below.
// 2. Register a config creator through creator.RegisterConfig.
package vless

import (
	"bytes"
	"context"
	"crypto/rand"
	gotls "crypto/tls"
	"fmt"
	"io"
	"math/big"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/rs/zerolog/log"

	"github.com/pires/go-proxyproto"
)

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
		0x1305: "TLS_AES_128_CCM_8_SHA256",
	}
)

const (
	TlsHandshakeTypeClientHello byte = 0x01
	TlsHandshakeTypeServerHello byte = 0x02

	CommandPaddingContinue byte = 0x00
	CommandPaddingEnd      byte = 0x01
	CommandPaddingDirect   byte = 0x02
)

// TrafficState is used to track uplink and downlink of one connection
// It is used by XTLS to determine if switch to raw copy mode, It is used by Vision to calculate padding
type TrafficState struct {
	UserUUID               []byte
	NumberOfPacketToFilter int
	EnableXtls             bool
	IsTLS12orAbove         bool
	IsTLS                  bool
	Cipher                 uint16
	RemainingServerHello   int32

	// reader link state
	WithinPaddingBuffers     bool
	ReaderSwitchToDirectCopy bool
	RemainingCommand         int32
	RemainingContent         int32
	RemainingPadding         int32
	CurrentCommand           int

	// write link state
	IsPadding                bool
	WriterSwitchToDirectCopy bool
}

func NewTrafficState(userUUID []byte) *TrafficState {
	return &TrafficState{
		UserUUID:                 userUUID,
		NumberOfPacketToFilter:   8,
		EnableXtls:               false,
		IsTLS12orAbove:           false,
		IsTLS:                    false,
		Cipher:                   0,
		RemainingServerHello:     -1,
		WithinPaddingBuffers:     true,
		ReaderSwitchToDirectCopy: false,
		RemainingCommand:         -1,
		RemainingContent:         -1,
		RemainingPadding:         -1,
		CurrentCommand:           0,
		IsPadding:                true,
		WriterSwitchToDirectCopy: false,
	}
}

// VisionReader is used to read xtls vision protocol
// Note Vision probably only make sense as the inner most layer of reader, since it need assess traffic state from origin proxy traffic
type VisionReader struct {
	buf.Reader
	trafficState *TrafficState
	ctx          context.Context
}

func NewVisionReader(reader buf.Reader, state *TrafficState, context context.Context) *VisionReader {
	return &VisionReader{
		Reader:       reader,
		trafficState: state,
		ctx:          context,
	}
}

// The buffer read out does not contain a complete record and a incomplete record. It either contains a complete record or an incomplete record
func (w *VisionReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	buffer, err := w.Reader.ReadMultiBuffer()
	if !buffer.IsEmpty() {
		if w.trafficState.WithinPaddingBuffers || w.trafficState.NumberOfPacketToFilter > 0 {
			mb2 := make(buf.MultiBuffer, 0, len(buffer))
			for _, b := range buffer {
				newbuffer := XtlsUnpadding(b, w.trafficState, w.ctx)
				if newbuffer.Len() > 0 {
					mb2 = append(mb2, newbuffer)
				}
			}
			buffer = mb2
			if w.trafficState.RemainingContent > 0 || w.trafficState.RemainingPadding > 0 || w.trafficState.CurrentCommand == 0 {
				w.trafficState.WithinPaddingBuffers = true
				// The following two cases: last block has been fully read
			} else if w.trafficState.CurrentCommand == 1 {
				w.trafficState.WithinPaddingBuffers = false
			} else if w.trafficState.CurrentCommand == 2 {
				w.trafficState.WithinPaddingBuffers = false
				w.trafficState.ReaderSwitchToDirectCopy = true
			} else {
				// log.Debug().Msg("XtlsRead unknown command")
			}
		}
		if w.trafficState.NumberOfPacketToFilter > 0 {
			XtlsFilterTls(buffer, w.trafficState, w.ctx)
		}
	}
	return buffer, err
}

// VisionWriter is used to write xtls vision protocol
// Note Vision probably only make sense as the inner most layer of writer, since it need assess traffic state from origin proxy traffic
type VisionWriter struct {
	buf.Writer
	trafficState      *TrafficState
	ctx               context.Context
	writeOnceUserUUID []byte
}

func NewVisionWriter(writer buf.Writer, state *TrafficState, context context.Context) *VisionWriter {
	w := make([]byte, len(state.UserUUID))
	copy(w, state.UserUUID)
	return &VisionWriter{
		Writer:            writer,
		trafficState:      state,
		ctx:               context,
		writeOnceUserUUID: w,
	}
}

func (w *VisionWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if w.trafficState.NumberOfPacketToFilter > 0 {
		XtlsFilterTls(mb, w.trafficState, w.ctx)
	}
	if w.trafficState.IsPadding {
		if len(mb) == 1 && mb[0] == nil {
			mb[0] = XtlsPadding(nil, CommandPaddingContinue, &w.writeOnceUserUUID, true, w.ctx) // we do a long padding to hide vless header
			return w.Writer.WriteMultiBuffer(mb)
		}
		mb = ReshapeMultiBuffer(w.ctx, mb)
		longPadding := w.trafficState.IsTLS
		// a b a block. The last b might or might not be the last block
		for i, b := range mb {
			// if app data is found
			if w.trafficState.IsTLS && b.Len() >= 6 && bytes.Equal(TlsApplicationDataStart, b.BytesTo(3)) {
				if w.trafficState.EnableXtls {
					// This WriteMultiBuffer is the last call
					w.trafficState.WriterSwitchToDirectCopy = true
				}
				var command byte = CommandPaddingContinue
				if i == len(mb)-1 {
					command = CommandPaddingEnd
					if w.trafficState.EnableXtls {
						command = CommandPaddingDirect
					}
				}
				mb[i] = XtlsPadding(b, command, &w.writeOnceUserUUID, true, w.ctx)
				w.trafficState.IsPadding = false // padding going to end
				longPadding = false
				continue
			} else if !w.trafficState.IsTLS12orAbove && w.trafficState.NumberOfPacketToFilter <= 1 { // For compatibility with earlier vision receiver, we finish padding 1 packet early
				w.trafficState.IsPadding = false
				mb[i] = XtlsPadding(b, CommandPaddingEnd, &w.writeOnceUserUUID, longPadding, w.ctx)
				break
			}
			var command byte = CommandPaddingContinue
			if i == len(mb)-1 && !w.trafficState.IsPadding {
				command = CommandPaddingEnd
				if w.trafficState.EnableXtls {
					command = CommandPaddingDirect
				}
			}
			mb[i] = XtlsPadding(b, command, &w.writeOnceUserUUID, longPadding, w.ctx)
		}
	}
	return w.Writer.WriteMultiBuffer(mb)
}

// ReshapeMultiBuffer prepare multi buffer for padding structure (max 21 bytes)
func ReshapeMultiBuffer(ctx context.Context, buffer buf.MultiBuffer) buf.MultiBuffer {
	needReshape := 0
	for _, b := range buffer {
		if b.Len() >= buf.Size-21 {
			needReshape += 1
		}
	}
	if needReshape == 0 {
		return buffer
	}
	mb2 := make(buf.MultiBuffer, 0, len(buffer)+needReshape)
	toPrint := ""
	for i, buffer1 := range buffer {
		if buffer1.Len() >= buf.Size-21 {
			index := int32(bytes.LastIndex(buffer1.Bytes(), TlsApplicationDataStart))
			if index < 21 || index > buf.Size-21 {
				index = buf.Size / 2
			}
			buffer2 := buf.New()
			buffer2.Write(buffer1.BytesFrom(index))
			buffer1.Resize(0, index)
			mb2 = append(mb2, buffer1, buffer2)
			toPrint += " " + strconv.Itoa(int(buffer1.Len())) + " " + strconv.Itoa(int(buffer2.Len()))
		} else {
			mb2 = append(mb2, buffer1)
			toPrint += " " + strconv.Itoa(int(buffer1.Len()))
		}
		buffer[i] = nil
	}
	buffer = buffer[:0]
	// log.Debug().Msg("ReshapeMultiBuffer " + toPrint)
	return mb2
}

// XtlsPadding add padding to eliminate length signature during tls handshake
func XtlsPadding(b *buf.Buffer, command byte, userUUID *[]byte, longPadding bool, ctx context.Context) *buf.Buffer {
	var contentLen int32 = 0
	var paddingLen int32 = 0
	if b != nil {
		contentLen = b.Len()
	}
	if contentLen < 900 && longPadding {
		l, err := rand.Int(rand.Reader, big.NewInt(500))
		if err != nil {
			// log.Debug().Msg("failed to generate padding")
		}
		paddingLen = int32(l.Int64()) + 900 - contentLen
	} else {
		l, err := rand.Int(rand.Reader, big.NewInt(256))
		if err != nil {
			// log.Debug().Msg("failed to generate padding")
		}
		paddingLen = int32(l.Int64())
	}
	if paddingLen > buf.Size-21-contentLen {
		paddingLen = buf.Size - 21 - contentLen
	}
	newbuffer := buf.New()
	if userUUID != nil {
		newbuffer.Write(*userUUID)
		*userUUID = nil
	}
	newbuffer.Write([]byte{command, byte(contentLen >> 8), byte(contentLen), byte(paddingLen >> 8), byte(paddingLen)})
	if b != nil {
		newbuffer.Write(b.Bytes())
		b.Release()
		b = nil
	}
	newbuffer.Extend(paddingLen)
	// log.Debug().Int32("content", contentLen).Uint32("padding", uint32(paddingLen)).Uint8("command", command).Msg("XtlsPadding")
	return newbuffer
}

// XtlsUnpadding remove padding and parse command
func XtlsUnpadding(b *buf.Buffer, s *TrafficState, ctx context.Context) *buf.Buffer {
	if s.RemainingCommand == -1 && s.RemainingContent == -1 && s.RemainingPadding == -1 { // inital state
		if b.Len() >= 21 && bytes.Equal(s.UserUUID, b.BytesTo(16)) {
			b.AdvanceStart(16)
			s.RemainingCommand = 5
		} else {
			return b
		}
	}
	newbuffer := buf.New()
	for b.Len() > 0 {
		if s.RemainingCommand > 0 {
			data, err := b.ReadByte()
			if err != nil {
				return newbuffer
			}
			switch s.RemainingCommand {
			case 5:
				s.CurrentCommand = int(data)
			case 4:
				s.RemainingContent = int32(data) << 8
			case 3:
				s.RemainingContent = s.RemainingContent | int32(data)
			case 2:
				s.RemainingPadding = int32(data) << 8
			case 1:
				s.RemainingPadding = s.RemainingPadding | int32(data)
				// log.Ctx(ctx).Debug().Int32("content", s.RemainingContent).Int32("padding", s.RemainingPadding).
				// 	Int("command", s.CurrentCommand).Msg("Xtls Unpadding")
			}
			s.RemainingCommand--
		} else if s.RemainingContent > 0 {
			len := s.RemainingContent
			if b.Len() < len {
				len = b.Len()
			}
			data, err := b.ReadBytes(len)
			if err != nil {
				return newbuffer
			}
			newbuffer.Write(data)
			s.RemainingContent -= len
		} else { // remainingPadding > 0
			len := s.RemainingPadding
			if b.Len() < len {
				len = b.Len()
			}
			b.AdvanceStart(len)
			s.RemainingPadding -= len
		}
		if s.RemainingCommand <= 0 && s.RemainingContent <= 0 && s.RemainingPadding <= 0 { // this block done
			if s.CurrentCommand == 0 {
				s.RemainingCommand = 5
			} else {
				s.RemainingCommand = -1 // set to initial state
				s.RemainingContent = -1
				s.RemainingPadding = -1
				if b.Len() > 0 { // shouldn't happen
					newbuffer.Write(b.Bytes())
				}
				break
			}
		}
	}
	b.Release()
	b = nil
	return newbuffer
}

// XtlsFilterTls filter and recognize tls 1.3 and other Debug
func XtlsFilterTls(buffer buf.MultiBuffer, trafficState *TrafficState, ctx context.Context) {
	for _, b := range buffer {
		if b == nil {
			continue
		}
		trafficState.NumberOfPacketToFilter--
		if b.Len() >= 6 {
			startsBytes := b.BytesTo(6)
			if bytes.Equal(TlsServerHandShakeStart, startsBytes[:3]) && startsBytes[5] == TlsHandshakeTypeServerHello {
				trafficState.RemainingServerHello = (int32(startsBytes[3])<<8 | int32(startsBytes[4])) + 5
				trafficState.IsTLS12orAbove = true
				trafficState.IsTLS = true
				if b.Len() >= 79 && trafficState.RemainingServerHello >= 79 {
					sessionIdLen := int32(b.Byte(43))
					cipherSuite := b.BytesRange(43+sessionIdLen+1, 43+sessionIdLen+3)
					trafficState.Cipher = uint16(cipherSuite[0])<<8 | uint16(cipherSuite[1])
				} else {
					// log.Ctx(ctx).Debug().Msg("XtlsFilterTls short server hello, tls 1.2 or older? " + strconv.Itoa(int(b.Len())) + " " + strconv.Itoa(int(trafficState.RemainingServerHello)))
				}
			} else if bytes.Equal(TlsClientHandShakeStart, startsBytes[:2]) && startsBytes[5] == TlsHandshakeTypeClientHello {
				trafficState.IsTLS = true
				// log.Ctx(ctx).Debug().Msg("XtlsFilterTls found tls client hello! " + strconv.Itoa(int(b.Len())))
			}
		}
		if trafficState.RemainingServerHello > 0 {
			end := trafficState.RemainingServerHello
			if end > b.Len() {
				end = b.Len()
			}
			trafficState.RemainingServerHello -= b.Len()
			if bytes.Contains(b.BytesTo(end), Tls13SupportedVersions) {
				v, ok := Tls13CipherSuiteDic[trafficState.Cipher]
				if !ok {
					v = "Old cipher: " + strconv.FormatUint(uint64(trafficState.Cipher), 16)
				} else if v != "TLS_AES_128_CCM_8_SHA256" {
					trafficState.EnableXtls = true
				}
				// log.Ctx(ctx).Debug().Msg("XtlsFilterTls found tls 1.3! " + strconv.Itoa(int(b.Len())) + " " + v)
				trafficState.NumberOfPacketToFilter = 0
				return
			} else if trafficState.RemainingServerHello <= 0 {
				// log.Ctx(ctx).Debug().Msg("XtlsFilterTls inconclusive server hello! " + strconv.Itoa(int(b.Len())) + " " + strconv.Itoa(int(trafficState.RemainingServerHello)))
				trafficState.NumberOfPacketToFilter = 0
				return
			}
			// log.Ctx(ctx).Debug().Msg("XtlsFilterTls inconclusive server hello! " + strconv.Itoa(int(b.Len())) + " " + strconv.Itoa(int(trafficState.RemainingServerHello)))
		}
		if trafficState.NumberOfPacketToFilter <= 0 {
			// log.Ctx(ctx).Debug().Msg("XtlsFilterTls stop filtering! " + strconv.Itoa(int(b.Len())))
		}
	}
}

// UnwrapRawConn support unwrap stats, tls, utls, reality and proxyproto conn and get raw tcp conn from it
func UnwrapRawConn(conn net.Conn) (net.Conn, *atomic.Uint64, *atomic.Uint64) {
	var readCounter, writerCounter *atomic.Uint64
	if conn != nil {
		// statConn, ok := conn.(*stat.CounterConnection)
		// if ok {
		// 	conn = statConn.Connection
		// 	readCounter = statConn.ReadCounter
		// 	writerCounter = statConn.WriteCounter
		// }
		if xc, ok := conn.(*tls.Conn); ok {
			conn = xc.NetConn()
		} else if goTlsConn, ok := conn.(*gotls.Conn); ok {
			conn = goTlsConn.NetConn()
		} else if utlsConn, ok := conn.(*tls.UConn); ok {
			conn = utlsConn.NetConn()
		} else if realityUConn, ok := conn.(*reality.UConn); ok {
			conn = realityUConn.NetConn()
		} else if realityConn, ok := conn.(*reality.Conn); ok {
			conn = realityConn.NetConn()
		}
		if pc, ok := conn.(*proxyproto.Conn); ok {
			conn = pc.Raw()
			// 8192 > 4096, there is no need to process pc's bufReader
		}
	}
	return conn, readCounter, writerCounter
}

// CopyRawConnIfExist use the most efficient copy method.
// - If caller don't want to turn on splice, do not pass in both reader conn and writer conn
// - writer are from *transport.Link
func CopyRawConnIfExist(ctx context.Context, readerConn net.Conn, writerConn net.Conn, writer buf.Writer, timer *signal.ActivityChecker, inTimer *signal.ActivityChecker) error {
	readerConn, readCounter, _ := UnwrapRawConn(readerConn)
	writerConn, _, writeCounter := UnwrapRawConn(writerConn)
	reader := buf.NewReader(readerConn)
	if runtime.GOOS != "linux" && runtime.GOOS != "android" {
		return readV(ctx, reader, writer, timer, readCounter)
	}
	tc, ok := writerConn.(*net.TCPConn)
	if !ok || readerConn == nil || writerConn == nil {
		return readV(ctx, reader, writer, timer, readCounter)
	}
	inbound := InboundFromContext(ctx)
	if inbound == nil || inbound.CanSpliceCopy == 3 {
		return readV(ctx, reader, writer, timer, readCounter)
	}
	outbounds := OutboundsFromContext(ctx)
	if len(outbounds) == 0 {
		return readV(ctx, reader, writer, timer, readCounter)
	}
	for _, ob := range outbounds {
		if ob.CanSpliceCopy == 3 {
			return readV(ctx, reader, writer, timer, readCounter)
		}
	}

	for {
		inbound := InboundFromContext(ctx)
		outbounds := OutboundsFromContext(ctx)
		var splice = inbound.CanSpliceCopy == 1
		for _, ob := range outbounds {
			if ob.CanSpliceCopy != 1 {
				splice = false
			}
		}
		if splice {
			log.Debug().Msg("CopyRawConn splice")
			//runtime.Gosched() // necessary
			time.Sleep(time.Millisecond)    // without this, there will be a rare ssl error for freedom splice
			timer.SetTimeout(8 * time.Hour) // prevent leak, just in case
			if inTimer != nil {
				inTimer.SetTimeout(8 * time.Hour)
			}
			w, err := tc.ReadFrom(readerConn)
			if readCounter != nil {
				readCounter.Add(uint64(w)) // outbound stats
			}
			if writeCounter != nil {
				writeCounter.Add(uint64(w)) // inbound stats
			}
			if inbound.DownCounter != nil {
				inbound.DownCounter.DownTraffic(uint64(w)) // user stats
			}
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		}
		buffer, err := reader.ReadMultiBuffer()
		if !buffer.IsEmpty() {
			if readCounter != nil {
				readCounter.Add(uint64(buffer.Len()))
			}
			timer.Update()
			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
				return werr
			}
		}
		if err != nil {
			return err
		}
	}
}

func readV(ctx context.Context, reader buf.Reader, writer buf.Writer, timer *signal.ActivityChecker, readCounter *atomic.Uint64) error {
	if err := buf.Copy(reader, writer, buf.UpdateActivityCopyOption(timer), buf.AddToStatCounter(readCounter)); err != nil {
		return fmt.Errorf("failed to copy, %w", err)
	}
	return nil
}
