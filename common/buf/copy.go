package buf

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Copy dumps all payload from reader to writer or stops when an error occurs or
// EOF in which case it returns nil. An error is either failure to read or failure to write
func Copy(reader Reader, writer Writer, datahandlers ...DataHandler) error {
	return copyRW(reader, writer, datahandlers...)
}

func transformReadError(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return ReadError{err}
}

func copyRW(reader Reader, writer Writer, datahandlers ...DataHandler) error {
	for {
		buffer, err := reader.ReadMultiBuffer()
		if !buffer.IsEmpty() {
			for _, handler := range datahandlers {
				handler.HandleData(buffer)
			}
			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
				return WriteError{werr}
			}
		}
		if err != nil {
			return transformReadError(err)
		}
	}
}

type CopySetting struct {
	DataHandlers []DataHandler //functions to apply to data before writing
	OnEOF        func()
	OnEnd        func()
}

type DataHandler interface {
	HandleData(MultiBuffer)
}

// SizeCounter is for counting bytes copied by Copy().
type SizeCounter struct {
	Size int64
}

// CountSize is a CopyOption that sums the total size of data copied into the given SizeCounter.
func (sc *SizeCounter) HandleData(mb MultiBuffer) {
	sc.Size += int64(mb.Len())
}

// AddToStatCounter a CopyOption add to stat counter
// func AddToStatCounter(sc *atomic.Uint64) CopyOption {
// 	return func(handler *CopySetting) {
// 		handler.DataHandlers = append(handler.DataHandlers, func(b MultiBuffer) {
// 			if sc != nil {
// 				sc.Add(uint64(b.Len()))
// 			}
// 		})
// 	}
// }

// func UpdateActivityCopyOption(ac *signal.ActivityChecker) CopyOption {
// 	return func(handler *CopySetting) {
// 		handler.DataHandlers = append(handler.DataHandlers, func(MultiBuffer) {
// 			ac.Update()
// 		})
// 	}
// }

// func OnEOFCopyOption(f func()) CopyOption {
// 	return func(handler *CopySetting) {
// 		handler.OnEOF = f
// 	}
// }

// func OnEndCopyOption(f func()) CopyOption {
// 	return func(handler *CopySetting) {
// 		handler.OnEnd = f
// 	}
// }

type ReadError struct {
	error
}

func (e ReadError) Error() string {
	return "readError: " + e.error.Error()
}

func (e ReadError) Inner() error {
	return e.error
}

func (e ReadError) Unwrap() error {
	return e.error
}

// IsReadError returns true if the error in Copy() comes from reading.
func IsReadError(err error) bool {
	_, ok := err.(ReadError)
	return ok
}

type WriteError struct {
	error
}

func (e WriteError) Error() string {
	return "writeError: " + e.error.Error()
}

func (e WriteError) Inner() error {
	return e.error
}

func (e WriteError) UnWrap() error {
	return e.error
}

// IsWriteError returns true if the error in Copy() comes from writing.
func IsWriteError(err error) bool {
	_, ok := err.(WriteError)
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
