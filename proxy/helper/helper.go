package helper

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

// read from leftReader, write to rightWriter, read from rightReader, write to leftWriter.
// When any direction met an error, returns the error.
// When both directions return nil, the function returns nil.
func Relay(ctx context.Context, leftReader buf.Reader, leftWriter buf.Writer,
	rightReader buf.Reader, rightWriter buf.Writer) error {
	leftToRight := func() error {
		var err error
		if runtime.GOOS == "linux" {
			err = spliceCopy(ctx, leftReader, rightWriter, true)
			if err == nil {
				rightWriter.CloseWrite()
			}
		} else {
			err = buf.Copy(leftReader, rightWriter,
				buf.OnEOFCopyOption(func() {
					rightWriter.CloseWrite()
				}),
			)
		}
		if err != nil {
			if context.Cause(ctx) != errors.ErrIdle {
				err = errors.NewLeftToRightError(err)
			}
		}
		return err
	}

	rightToLeft := func() error {
		var err error
		if runtime.GOOS == "linux" {
			err = spliceCopy(ctx, rightReader, leftWriter, false)
			if err == nil {
				leftWriter.CloseWrite()
			}
		} else {
			err = buf.Copy(rightReader, leftWriter,
				buf.OnEOFCopyOption(func() {
					leftWriter.CloseWrite()
				}),
			)
		}
		if err != nil {
			err = errors.NewRightToLeftError(err)
		}
		return err
	}

	return task.Run(ctx, leftToRight, rightToLeft)
}

func spliceCopy(ctx context.Context, reader, writer any, up bool) error {
	originalReader, originalWriter := reader, writer
	// unwrap both reader and writer until they are not Unwrapper
	unwrapReader, unwrapWriter := true, true
	info := session.InfoFromContext(ctx)

	for {
		if unwrapReader {
			if unwrapper, ok := reader.(buf.UnwrapReader); ok {
				if unwrapper.OkayToUnwrapReader() == 1 {
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
				if unwrapper, ok := writer.(buf.UnwrapWriter); ok {
					if unwrapper.OkayToUnwrapWriter() == 1 {
						writer = unwrapper.UnwrapWriter()
					} else if unwrapper.OkayToUnwrapWriter() == -1 {
						unwrapWriter = false
					}
				} else {
					unwrapWriter = false
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
		// splice copy
		if ioReader, ok := reader.(io.Reader); ok {
			log.Ctx(ctx).Debug().Bool("up", up).Msg("readFrom")
			if info != nil && info.ActivityChecker != nil {
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
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
	// var bufReader buf.Reader
	// if ioReader, ok := reader.(io.Reader); ok {
	// 	bufReader = buf.NewReader(ioReader)
	// } else if bufReader, ok = reader.(buf.Reader); !ok {
	// 	return errors.New("invalid reader")
	// }
	// var bufWriter buf.Writer
	// if ioWriter, ok := writer.(io.Writer); ok {
	// 	bufWriter = buf.NewWriter(ioWriter)
	// } else if bufWriter, ok = writer.(buf.Writer); !ok {
	// 	return errors.New("invalid writer")
	// }
	// splice copy is not feasible. so use original reader and writer
	return buf.Copy(originalReader.(buf.Reader), originalWriter.(buf.Writer))
}

func writeToWriter(writer any, mb buf.MultiBuffer) error {
	if bufWriter, ok := writer.(buf.Writer); ok {
		return bufWriter.WriteMultiBuffer(mb)
	} else if ioWriter, ok := writer.(io.Writer); ok {
		defer buf.ReleaseMulti(mb)
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

func GetLinks(network net.Network, userLevel uint32, policy i.BufferPolicy) (*pipe.Link, *pipe.Link) {
	bufferSize := policy.UserBufferSize(0)
	if userLevel != 0 {
		bufferSize = policy.UserBufferSize(userLevel)
	}

	linkA, linkB := pipe.NewLinks(bufferSize, network == net.Network_UDP)

	var iLink, oLink *pipe.Link
	iLink = linkA
	oLink = linkB

	return iLink, oLink
}

func RelayUDPPacketConn(ctx context.Context, left udp.PacketReaderWriter, right udp.PacketReaderWriter) error {

	leftToRight := func() error {
		for {
			p, err := left.ReadPacket()
			if err != nil {
				return fmt.Errorf("failed to read from left: %w", err)
			}
			if err := right.WritePacket(p); err != nil {
				return fmt.Errorf("failed to write to right: %w", err)
			}
		}
	}

	rightToLeft := func() error {
		for {
			p, err := right.ReadPacket()
			if err != nil {
				return fmt.Errorf("failed to read from right: %w", err)
			}
			if err := left.WritePacket(p); err != nil {
				return fmt.Errorf("failed to write to left: %w", err)
			}
		}
	}

	return task.Run(ctx, leftToRight, rightToLeft)
}

// read from leftReader, write to rightWriter, read from rightReader, write to leftWriter.
// When any direction met an error, returns the error.
// When both directions return nil, the function returns nil.
func RelayConn(ctx context.Context, left, right net.Conn) error {
	leftToRight := func() error {
		_, err := io.Copy(right, left)
		if err != nil {
			return fmt.Errorf("failed to copy from left to right: %w", err)
		}
		if cw, ok := right.(buf.CloseWriter); ok {
			return cw.CloseWrite()
		}
		return nil
	}

	rightToLeft := func() error {
		_, err := io.Copy(left, right)
		if err != nil {
			return fmt.Errorf("failed to copy from right to left: %w", err)
		}
		if cw, ok := left.(buf.CloseWriter); ok {
			return cw.CloseWrite()
		}
		return nil
	}

	return task.Run(ctx, leftToRight, rightToLeft)
}
