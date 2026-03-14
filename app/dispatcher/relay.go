// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"io"
	"runtime"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

func (d *Dispatcher) Relay(ctx context.Context, info *session.Info,
	left, right any, outbound i.Outbound) error {
	ups, downs := d.GetCounters.GetCounters(ctx, info, outbound)
	info.UpCounter = ups
	info.DownCounter = downs

	var activityChecker *signal.ActivityChecker
	if timeout := d.getTimeout(info); timeout != 0 {
		var cancelCause context.CancelCauseFunc
		ctx, cancelCause = context.WithCancelCause(ctx)
		activityChecker = signal.NewActivityChecker(func() {
			cancelCause(errors.ErrIdle)
			log.Ctx(ctx).Debug().Msg("idle timeout")
		}, timeout)
		info.ActivityChecker = activityChecker
	}

	leftToRight := func() error {
		err := d.copyReaderToWriter(ctx, info, left,
			right, true, activityChecker)
		if err == nil {
			if closeWriter, ok := right.(buf.CloseWriter); ok {
				closeWriter.CloseWrite()
			}
			if activityChecker != nil {
				activityChecker.SetTimeout(d.TimeoutSetting.DownLinkOnlyTimeout())
			}
		}
		if err != nil {
			err = errors.NewLeftToRightError(err)
		}
		return err
	}

	rightToLeft := func() error {
		err := d.copyReaderToWriter(ctx, info, right,
			left, false, activityChecker)
		if err == nil {
			if closeWriter, ok := left.(buf.CloseWriter); ok {
				closeWriter.CloseWrite()
			}
			if activityChecker != nil {
				activityChecker.SetTimeout(d.TimeoutSetting.UpLinkOnlyTimeout())
			}
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
func (d *Dispatcher) copyReaderToWriter(ctx context.Context, info *session.Info,
	reader, writer any, up bool, activityChecker *signal.ActivityChecker) error {

	// try splice copy on linux
	if runtime.GOOS == "linux" {
		// unwrap both reader and writer until they are not Unwrapper
		unwrapReader, unwrapWriter := true, true
		var innerMostReader any
		var innerMostWriter any
		// This loop is for finding the innermost reader and writer
		for {
			// try find the innermost reader
			if unwrapReader {
				var r any = reader
				for {
					if unwrapper, ok := r.(buf.UnwrapReader); ok {
						if unwrapper.OkayToUnwrapReader() == 1 {
							log.Debug().Type("reader", r).Msg("unwrap reader")
							r = unwrapper.UnwrapReader()
						} else if unwrapper.OkayToUnwrapReader() == -1 {
							unwrapReader = false
							innerMostReader = r
							break
						} else if unwrapper.OkayToUnwrapReader() == 0 {
							break
						}
					} else {
						unwrapReader = false
						innerMostReader = r
						break
					}
				}
			}
			mb, err := readFromReader(reader)
			if mb.Len() > 0 {
				// try find the innermost writer
				if unwrapWriter {
					var w any = writer
					for {
						if unwrapper, ok := w.(buf.UnwrapWriter); ok {
							if unwrapper.OkayToUnwrapWriter() == 1 {
								log.Debug().Type("writer", w).Msg("unwrap writer")
								w = unwrapper.UnwrapWriter()
								log.Debug().Type("writer", w).Msg("unwraped writer")
							} else if unwrapper.OkayToUnwrapWriter() == -1 {
								unwrapWriter = false
								innerMostWriter = w
								break
							} else if unwrapper.OkayToUnwrapWriter() == 0 {
								break
							}
						} else {
							unwrapWriter = false
							innerMostWriter = w
							break
						}
					}
				}
				if err := writeToWriter(writer, mb); err != nil {
					return err
				}
				if activityChecker != nil {
					activityChecker.Update()
				}
				if up {
					info.UpCounter.UpTraffic(uint64(mb.Len()))
				} else {
					info.DownCounter.DownTraffic(uint64(mb.Len()))
				}
			}
			if err != nil {
				return err
			}
			if !unwrapReader && !unwrapWriter {
				break
			}
		}
		if readFromer, ok := innerMostWriter.(io.ReaderFrom); ok {
			if ioReader, ok := innerMostReader.(io.Reader); ok {
				if activityChecker != nil {
					activityChecker.Cancel()
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

	if up {
		return buf.Copy(bufReader, bufWriter,
			&countersCopyHandler{info.UpCounter},
			&activityCopyHandler{activityChecker})
	} else {
		return buf.Copy(bufReader, bufWriter,
			&downCountersCopyHandler{info.DownCounter},
			&activityCopyHandler{activityChecker})
	}
}

type countersCopyHandler struct {
	counters session.UpCounters
}

func (h *countersCopyHandler) HandleData(mb buf.MultiBuffer) {
	h.counters.UpTraffic(uint64(mb.Len()))
}

type downCountersCopyHandler struct {
	counters session.DownCounters
}

func (h *downCountersCopyHandler) HandleData(mb buf.MultiBuffer) {
	h.counters.DownTraffic(uint64(mb.Len()))
}

type activityCopyHandler struct {
	activityChecker *signal.ActivityChecker
}

func (h *activityCopyHandler) HandleData(mb buf.MultiBuffer) {
	if h.activityChecker != nil {
		h.activityChecker.Update()
	}
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

func (p *Dispatcher) getTimeout(info *session.Info) time.Duration {
	if p.TimeoutSetting == nil {
		return 0
	}
	if info.Target.Port == 22 {
		return p.TimeoutSetting.SshIdleTimeout()
	}
	if info.Target.Port == 53 {
		return p.TimeoutSetting.DnsIdleTimeout()
	}
	if info.Target.Network == net.Network_TCP {
		return p.TimeoutSetting.TcpIdleTimeout()
	}
	return p.TimeoutSetting.UdpIdleTimeout()
}
