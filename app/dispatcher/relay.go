package dispatcher

import (
	"context"
	"io"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/rs/zerolog/log"
)

func Relay(ctx context.Context, info *session.Info, left, right any) error {
	leftToRight := func() error {
		err := copyReaderToWriter(ctx, info, left, right, true)
		if closeWriter, ok := right.(buf.CloseWriter); ok && err == nil {
			closeWriter.CloseWrite()
		}
		if err != nil {
			if context.Cause(ctx) != errors.ErrIdle {
				err = errors.NewLeftToRightError(err)
			}
		}
		return err
	}

	rightToLeft := func() error {
		err := copyReaderToWriter(ctx, info, right, left, false)
		if closeWriter, ok := left.(buf.CloseWriter); ok && err == nil {
			closeWriter.CloseWrite()
		}
		if err != nil {
			err = errors.NewRightToLeftError(err)
		}
		return err
	}

	return task.Run(ctx, leftToRight, rightToLeft)
}

// reader is either a buf.Reader or a io.Reader
// writer is either a buf.Writer or a io.Writer
func copyReaderToWriter(ctx context.Context, info *session.Info, reader, writer any, up bool) error {
	// unwrap both reader and writer until they are not Unwrapper
	unwrapReader, unwrapWriter := true, true

	for {
		if unwrapReader {
			if unwrapper, ok := reader.(buf.UnwrapReader); ok {
				if unwrapper.OkayToUnwrapReader() == 1 {
					log.Debug().Type("reader", reader).Msg("unwrap reader")
					reader = unwrapper.UnwrapReader()
				} else if unwrapper.OkayToUnwrapReader() == -1 {
					unwrapReader = false
				}
			} else {
				unwrapReader = false
			}
		}
		mb, err := readFromReader(reader)
		if mb.Len() > 0 {
			if unwrapWriter {
				for {
					if unwrapper, ok := writer.(buf.UnwrapWriter); ok {
						if unwrapper.OkayToUnwrapWriter() == 1 {
							log.Debug().Type("writer", writer).Msg("unwrap writer")
							writer = unwrapper.UnwrapWriter()
							log.Debug().Type("writer", writer).Msg("unwraped writer")
						} else if unwrapper.OkayToUnwrapWriter() == -1 {
							unwrapWriter = false
							break
						}
						break
					} else {
						unwrapWriter = false
						break
					}
				}
			}
			if err := writeToWriter(writer, mb); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		if !unwrapReader && !unwrapWriter {
			break
		}
	}

	if readFromer, ok := writer.(io.ReaderFrom); ok {
		if ioReader, ok := reader.(io.Reader); ok {
			if info.ActivityChecker != nil {
				info.ActivityChecker.Cancel()
			}
			n, err := readFromer.ReadFrom(ioReader)
			if info != nil {
				if up {
					info.UpCounter.UpTraffic(uint64(n))
				} else {
					info.DownCounter.DownTraffic(uint64(n))
				}
			}
			return err
		}
	}
	var bufReader buf.Reader
	if ioReader, ok := reader.(io.Reader); ok {
		bufReader = buf.NewReader(ioReader)
	} else if bufReader, ok = reader.(buf.Reader); !ok {
		return errors.New("invalid reader")
	}
	var bufWriter buf.Writer
	if ioWriter, ok := writer.(io.Writer); ok {
		bufWriter = buf.NewWriter(ioWriter)
	} else if bufWriter, ok = writer.(buf.Writer); !ok {
		return errors.New("invalid writer")
	}
	return buf.Copy(bufReader, bufWriter)
}

func writeToWriter(writer any, mb buf.MultiBuffer) error {
	if bufWriter, ok := writer.(buf.Writer); ok {
		return bufWriter.WriteMultiBuffer(mb)
	} else if ioWriter, ok := writer.(io.Writer); ok {
		_, err := buf.WriteMultiBuffer(ioWriter, mb)
		return err
	} else {
		return errors.New("invalid writer")
	}
}

func readFromReader(reader any) (buf.MultiBuffer, error) {
	if bufReader, ok := reader.(buf.Reader); ok {
		return bufReader.ReadMultiBuffer()
	} else if ioReader, ok := reader.(io.Reader); ok {
		b := buf.New()
		n, err := b.ReadOnce(ioReader)
		if n > 0 {
			return buf.MultiBuffer{b}, nil
		}
		b.Release()
		return nil, err
	} else {
		return nil, errors.New("invalid reader")
	}
}
