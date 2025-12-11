package sniff

import (
	"context"
	"errors"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/bytespool"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/common/protocol/bittorrent"
	"github.com/5vnetwork/vx-core/common/protocol/http"
	"github.com/5vnetwork/vx-core/common/protocol/quic"
	"github.com/5vnetwork/vx-core/common/protocol/tls"
)

type SniffResult interface {
	Protocol() string
	Domain() string
}

type protocolSniffer func(context.Context, []byte) (SniffResult, error)

type protocolSnifferWithNetwork struct {
	protocolSniffer protocolSniffer
	network         net.Network
	protocol        string
}

// sniffState tracks active sniffers using a fixed-size array (max 8 sniffers)
type sniffState struct {
	activeMask uint8 // Bitmap: bit N = 1 means sniffer N is active
	focusIndex int8  // -1 = check all active; 0-7 = focus on specific sniffer
}

// newSniffState creates initial state with all sniffers active
func newSniffState() sniffState {
	return sniffState{
		activeMask: 0xFF, // All 8 bits set
		focusIndex: -1,
	}
}

type Sniffer struct {
	sniffers []protocolSnifferWithNetwork
}

// Adapter functions - created once at package init
var (
	sniffHTTP1Fn = func(c context.Context, b []byte) (SniffResult, error) { return http.SniffHTTP1Host(b) }
	sniffTLSFn   = func(c context.Context, b []byte) (SniffResult, error) { return tls.SniffTLS(b) }
	sniffQUICFn  = func(c context.Context, b []byte) (SniffResult, error) { return quic.SniffQUIC(b) }
	sniffBTFn    = func(c context.Context, b []byte) (SniffResult, error) { return bittorrent.SniffBittorrent(b) }
	sniffUTPFn   = func(c context.Context, b []byte) (SniffResult, error) { return bittorrent.SniffUTP(b) }
)

// NewSniffer returns the singleton sniffer (zero allocation, thread-safe)
func NewSniffer() *Sniffer {
	return &Sniffer{
		sniffers: []protocolSnifferWithNetwork{
			{sniffTLSFn, net.Network_TCP, "tls"},        // index 0
			{sniffQUICFn, net.Network_UDP, "quic"},      // index 1
			{sniffHTTP1Fn, net.Network_TCP, "http1"},    // index 2
			{sniffBTFn, net.Network_TCP, "bittorrent"},  // index 3
			{sniffUTPFn, net.Network_UDP, "bittorrent"}, // index 4
		},
	}
}

var errUnknownContent = errors.New("unknown content")

type cReader interface {
	// return copied=true if there is new data, and how much data is written into b, and error of the read operation
	// if there is no new data read, copied is false.
	read(b []byte) (copied bool, len int, err error)
	returnRw() any
}

var errSniffingTimeout = errors.New("timeout on sniffing")

// Sniff attempts to identify the protocol from the connection
func (sniffer *Sniffer) Sniff(ctx context.Context, info *session.Info, rw interface{}) (interface{}, error) {
	info.Sniffed = true

	startTime := time.Now()
	defer func() {
		log.Ctx(ctx).Debug().Dur("elapsed", time.Since(startTime)).Msg("SniffFlow")
	}()

	cReader, ok := rw.(cReader)
	if !ok {
		if r, ok := rw.(buf.DdlReaderWriter); ok {
			cReader = &CachedRW{
				DdlReaderWriter: r,
			}
		} else if ddlReaderWriter, ok := rw.(udp.DdlPacketReaderWriter); ok {
			cReader = &CachedDdlPacketConn{
				DdlPacketReaderWriter: ddlReaderWriter,
			}
		} else if readerWriter, ok := rw.(buf.ReaderWriter); ok {
			cReader = &CachedReadRw{
				ReaderWriter: readerWriter,
			}
		} else if packetConn, ok := rw.(udp.PacketReaderWriter); ok {
			cReader = &CachedPacketConn{
				PacketReaderWriter: packetConn,
			}
		} else {
			return nil, errors.New("unsupported type")
		}
	}

	bytes := bytespool.Alloc(8192)
	defer bytespool.Free(bytes)

	// Initialize state on stack
	state := newSniffState()
	totalAttempt := 0

	// Main sniffing loop
	for {
		select {
		case <-ctx.Done():
			return cReader.returnRw(), ctx.Err()
		default:
			copied, n, err := cReader.read(bytes)
			if copied {
				result, err := sniffer.sniff(ctx, bytes[:n], info.Target.Network, &state)
				if err == nil {
					info.Protocol = result.Protocol()
					if domain, err := strmatcher.ToDomain(result.Domain()); err == nil {
						info.SniffedDomain = domain
					} else {
						log.Ctx(ctx).Debug().Err(err).Str("domain", result.Domain()).Msg("strmatcher.ToDomain failed")
					}
					return cReader.returnRw(), nil
				} else if err != protocol.ErrNoClue && err != protocol.ErrProtoNeedMoreData {
					return cReader.returnRw(), err
				} else if err == protocol.ErrNoClue {
					totalAttempt++
				} else {
					// err == protocol.ErrProtoNeedMoreData, do not increase totalAttempt
					info.Protocol = sniffer.sniffers[state.focusIndex].protocol
				}
			} else if err != nil {
				return cReader.returnRw(), err
			} else {
				totalAttempt++
			}
			if n == 8192 {
				return cReader.returnRw(), errUnknownContent
			}
			if totalAttempt >= 2 {
				return cReader.returnRw(), errSniffingTimeout
			}
		}
	}
}

// SniffConn sniffs a TCP connection
func (sniffer *Sniffer) SniffConn(ctx context.Context, conn net.Conn) (net.Conn, SniffResult, error) {
	cachedConn := &CachedConn{
		Conn: conn,
	}

	bytes := bytespool.Alloc(8192)
	defer bytespool.Free(bytes)

	// Initialize state on stack
	state := newSniffState()
	totalAttempt := 0

	// Main sniffing loop
	for {
		select {
		case <-ctx.Done():
			return cachedConn.toConn(), nil, ctx.Err()
		default:
			copied, n, err := cachedConn.cache(bytes)
			if copied {
				result, err := sniffer.sniff(ctx, bytes[:n], net.Network_TCP, &state)
				if err == nil {
					return cachedConn.toConn(), result, nil
				} else if err != protocol.ErrNoClue && err != protocol.ErrProtoNeedMoreData {
					return cachedConn.toConn(), nil, err
				} else if err == protocol.ErrNoClue {
					totalAttempt++
				} else {
					// do not increase totalAttempt if err == protocol.ErrProtoNeedMoreData
				}
			} else if err != nil {
				return cachedConn.toConn(), nil, err
			} else {
				totalAttempt++
			}
			if n == 8192 {
				return cachedConn.toConn(), nil, errUnknownContent
			}
			if totalAttempt >= 2 {
				return cachedConn.toConn(), nil, errSniffingTimeout
			}
		}
	}
}

// sniff performs protocol detection using bitmap state tracking
func (s *Sniffer) sniff(c context.Context, payload []byte, network net.Network, state *sniffState) (SniffResult, error) {
	// If focused on a specific sniffer (after ErrProtoNeedMoreData)
	if state.focusIndex >= 0 {
		si := s.sniffers[state.focusIndex]
		return si.protocolSniffer(c, payload)
	}

	// Check all active sniffers using bitmap
	newMask := uint8(0)

	for i, si := range s.sniffers {
		if i >= 8 {
			break // Max 8 sniffers supported by uint8 bitmap
		}

		// Skip if not active in bitmap
		if state.activeMask&(1<<uint(i)) == 0 {
			continue
		}

		// Skip if network doesn't match
		if si.network != network {
			continue
		}

		result, err := si.protocolSniffer(c, payload)
		if err == nil {
			// Found match!
			return result, nil
		}

		if err == protocol.ErrNoClue {
			// Keep this sniffer active for next packet
			newMask |= (1 << uint(i))
			continue
		} else if err == protocol.ErrProtoNeedMoreData {
			// Focus only on this sniffer
			state.focusIndex = int8(i)
			state.activeMask = (1 << uint(i))
			return nil, err
		}
	}

	// Update active mask for next attempt
	state.activeMask = newMask

	if newMask != 0 {
		// Still have pending sniffers
		return nil, protocol.ErrNoClue
	}

	return nil, errUnknownContent
}
