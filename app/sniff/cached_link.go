package sniff

import (
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
)

// Cache, ReadMultiBuffer should be called sequentially.
type CachedReadRw struct {
	buf.ReaderWriter

	readLock sync.Mutex
	waitCh   chan readResult

	cache buf.MultiBuffer

	hasReadInternal bool
}

type readResult struct {
	dataRead bool
	err      error
}

// It runs for at most 10 ms
func (r *CachedReadRw) read(b []byte) (copied bool, len int, err error) {
	success := r.readLock.TryLock()
	if success {
		ch := make(chan readResult)
		r.waitCh = ch
		go func() {
			defer r.readLock.Unlock()
			defer close(ch)
			mb, err := r.ReaderWriter.ReadMultiBuffer()
			if !mb.IsEmpty() {
				r.cache, _ = buf.MergeMulti(r.cache, mb)
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

	timer := time.NewTimer(time.Millisecond * 10)
	defer timer.Stop()
	select {
	case <-timer.C:
		return false, 0, nil
	case result := <-r.waitCh:
		if !result.dataRead {
			return false, 0, result.err
		}
		n := r.cache.Copy(b)
		return true, n, result.err
	}
}

// read from cache
func (r *CachedReadRw) readInternal() buf.MultiBuffer {
	r.hasReadInternal = true
	mb := r.cache
	r.cache = nil
	return mb
}

func (r *CachedReadRw) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if !r.hasReadInternal {
		if success := r.readLock.TryLock(); success {
			r.readLock.Unlock()
		} else {
			<-r.waitCh
		}
		if mb := r.readInternal(); mb != nil && !mb.IsEmpty() {
			return mb, nil
		}
	}
	return r.ReaderWriter.ReadMultiBuffer()
}

func (r *CachedReadRw) OkayToUnwrapReader() int {
	if r.hasReadInternal {
		return 1
	}
	return 0
}

func (r *CachedReadRw) UnwrapReader() any {
	return r.ReaderWriter
}

func (r *CachedReadRw) OkayToUnwrapWriter() int {
	return 1
}

func (r *CachedReadRw) UnwrapWriter() any {
	return r.ReaderWriter
}

func (r *CachedReadRw) returnRw() any {
	return r
}
