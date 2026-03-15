// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/rs/zerolog/log"
)

type cacheReaderWriter struct {
	buf.DdlReaderWriter
	mb buf.MultiBuffer

	lock     sync.RWMutex
	readLock sync.Mutex

	// when this value is true, no fallback
	stopCaching      bool
	maximumCacheSize int
	// set to true when handler failed to handle
	done bool
	ctx  context.Context
}

// if fallbackable, make sure no reading and no writing of DdlReaderWriter and
// return all cached request data. Else, return nil, err
func (f *cacheReaderWriter) Done(ctx context.Context) (buf.MultiBuffer, error) {
	f.lock.Lock()
	f.done = true
	if f.stopCaching {
		f.lock.Unlock()
		return nil, errors.New("unfallbackable")
	}
	f.lock.Unlock()

	err := f.DdlReaderWriter.SetReadDeadline(time.Now().Add(-100 * time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("unable to clean read ddl, %w", err)
	}
	log.Ctx(ctx).Debug().Msg("set read deadline to -100ms")

	f.readLock.Lock()
	defer f.readLock.Unlock()

	return f.mb, nil
}

func (f *cacheReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	f.lock.Lock()
	if f.done {
		f.lock.Unlock()
		return nil, errors.New("done")
	}
	if f.stopCaching {
		f.lock.Unlock()
		return f.DdlReaderWriter.ReadMultiBuffer()
	}
	f.lock.Unlock()

	f.readLock.Lock()
	mb, err := f.DdlReaderWriter.ReadMultiBuffer()
	f.lock.Lock()
	defer f.lock.Unlock()
	defer f.readLock.Unlock()

	if f.stopCaching {
		return mb, err
	}

	// if request data is too large, no fallback
	if len(f.mb)+len(mb) > f.maximumCacheSize && !f.done {
		f.stopCaching = true
		buf.ReleaseMulti(f.mb)
		f.mb = nil
	} else {
		clone := mb.Clone()
		f.mb = append(f.mb, clone...)
	}

	return mb, err
}

func (f *cacheReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	f.lock.Lock()
	if f.done {
		f.lock.Unlock()
		return errors.New("done")
	}
	// any write will make this session unfallbackable
	if !f.stopCaching {
		f.stopCaching = true
		buf.ReleaseMulti(f.mb)
		f.mb = nil
	}
	f.lock.Unlock()
	return f.DdlReaderWriter.WriteMultiBuffer(mb)
}
