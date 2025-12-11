package sniff

import (
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/rs/zerolog/log"
)

type CachedPacketConn struct {
	udp.PacketReaderWriter

	readLock sync.Mutex
	waitCh   chan readResult

	cache           []*udp.Packet
	hasReadAllCache bool
}

func (r *CachedPacketConn) read(b []byte) (copied bool, len int, err error) {
	success := r.readLock.TryLock()
	if success {
		ch := make(chan readResult)
		r.waitCh = ch
		go func() {
			defer r.readLock.Unlock()
			defer close(ch)
			p, err := r.PacketReaderWriter.ReadPacket()
			if p != nil {
				log.Info().Int32("len", p.Payload.Len()).Msg("cache packet")
				r.cache = append(r.cache, p)
				ch <- readResult{
					dataRead: true,
					err:      err,
				}
			} else {
				ch <- readResult{
					dataRead: false,
					err:      err,
				}
			}
		}()
	}

	timer := time.NewTimer(time.Millisecond * 100)
	defer timer.Stop()
	select {
	case <-timer.C:
		return false, 0, nil
	case result := <-r.waitCh:
		if !result.dataRead {
			return false, 0, result.err
		}

		copied := 0
		for _, p := range r.cache {
			n := copy(b, p.Payload.Bytes())
			b = b[n:]
			copied += n
		}
		return true, copied, result.err
	}
}

func (r *CachedPacketConn) returnRw() any {
	return r
}

// read from cache
func (r *CachedPacketConn) readInternal() *udp.Packet {
	var ret *udp.Packet
	if len(r.cache) > 0 {
		ret = r.cache[0]
		r.cache = r.cache[1:]
	}
	if len(r.cache) == 0 {
		r.hasReadAllCache = true
	}
	return ret
}

func (r *CachedPacketConn) ReadPacket() (*udp.Packet, error) {
	if !r.hasReadAllCache {
		if success := r.readLock.TryLock(); success {
			r.readLock.Unlock()
		} else {
			<-r.waitCh
		}
		if p := r.readInternal(); p != nil {
			log.Info().Int32("len", p.Payload.Len()).Msg("read packet from cache")
			return p, nil
		}
	}
	return r.PacketReaderWriter.ReadPacket()
}
