package buf

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/signal"
)

// Copy dumps all payload from reader to writer or stops when an error occurs or
// EOF in which case it returns nil. An error is either failure to read or failure to write
func Copy(reader Reader, writer Writer, options ...CopyOption) error {
	var setting copySetting
	for _, option := range options {
		option(&setting)
	}
	if setting.onEnd != nil {
		defer setting.onEnd()
	}

	return copyRW(reader, writer, &setting)
}

func transformReadError(err error, setting *copySetting) error {
	if errors.Is(err, io.EOF) {
		if setting.onEOF != nil {
			setting.onEOF()
		}
		return nil
	}
	return readError{err}
}

func copyRW(reader Reader, writer Writer, setting *copySetting) error {
	for {
		buffer, err := reader.ReadMultiBuffer()
		if !buffer.IsEmpty() {
			for _, handler := range setting.dataHandlers {
				handler(buffer)
			}
			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
				return writeError{werr}
			}
		}
		if err != nil {
			return transformReadError(err, setting)
		}
	}
}

type copySetting struct {
	dataHandlers []dataHandler //functions to apply to data before writing
	onEOF        func()
	onEnd        func()
}

type CopyOption func(*copySetting)

type dataHandler func(MultiBuffer)

func DataHandler(dh dataHandler) CopyOption {
	return func(handler *copySetting) {
		handler.dataHandlers = append(handler.dataHandlers, dh)
	}
}

// SizeCounter is for counting bytes copied by Copy().
type SizeCounter struct {
	Size int64
}

// CountSize is a CopyOption that sums the total size of data copied into the given SizeCounter.
func CountSize(sc *SizeCounter) CopyOption {
	return func(handler *copySetting) {
		handler.dataHandlers = append(handler.dataHandlers, func(b MultiBuffer) {
			sc.Size += int64(b.Len())
		})
	}
}

// AddToStatCounter a CopyOption add to stat counter
func AddToStatCounter(sc *atomic.Uint64) CopyOption {
	return func(handler *copySetting) {
		handler.dataHandlers = append(handler.dataHandlers, func(b MultiBuffer) {
			if sc != nil {
				sc.Add(uint64(b.Len()))
			}
		})
	}
}

func UpdateActivityCopyOption(ac *signal.ActivityChecker) CopyOption {
	return func(handler *copySetting) {
		handler.dataHandlers = append(handler.dataHandlers, func(MultiBuffer) {
			ac.Update()
		})
	}
}

func OnEOFCopyOption(f func()) CopyOption {
	return func(handler *copySetting) {
		handler.onEOF = f
	}
}

func OnEndCopyOption(f func()) CopyOption {
	return func(handler *copySetting) {
		handler.onEnd = f
	}
}

type readError struct {
	error
}

func (e readError) Error() string {
	return "readError: " + e.error.Error()
}

func (e readError) Inner() error {
	return e.error
}

func (e readError) Unwrap() error {
	return e.error
}

// IsReadError returns true if the error in Copy() comes from reading.
func IsReadError(err error) bool {
	_, ok := err.(readError)
	return ok
}

type writeError struct {
	error
}

func (e writeError) Error() string {
	return "writeError: " + e.error.Error()
}

func (e writeError) Inner() error {
	return e.error
}

func (e writeError) UnWrap() error {
	return e.error
}

// IsWriteError returns true if the error in Copy() comes from writing.
func IsWriteError(err error) bool {
	_, ok := err.(writeError)
	return ok
}

var ErrNotTimeoutReader = errors.New("not a TimeoutReader")

func CopyOnceTimeout(reader Reader, writer Writer, timeout time.Duration) error {
	if timeoutReader, ok := reader.(TimeoutReader); ok {
		mb, err := timeoutReader.ReadMultiBufferTimeout(timeout)
		if mb.Len() > 0 {
			return writer.WriteMultiBuffer(mb)
		}
		return err
	} else if ddl, ok := reader.(DeadlineReader); ok {
		err := ddl.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}
		mb, err := reader.ReadMultiBuffer()
		ddl.SetReadDeadline(time.Time{})
		if mb.Len() > 0 {
			return writer.WriteMultiBuffer(mb)
		}
		if err != nil && strings.Contains(err.Error(), "i/o timeout") {
			return nil
		}
		return err
	} else {
		return ErrNotTimeoutReader
	}
}
