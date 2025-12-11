package pipe

import (
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/signal/done"
)

// discardOverflow is true when creating a pipe for udp traffic.
func NewPipe(limit int32, discardOverflow bool) *Pipe {
	p := &Pipe{
		readSignal:  signal.NewNotifier(),
		writeSignal: signal.NewNotifier(),
		done:        done.New(),
		//when limit is 0, and there is a write, if pipe.data is nil or empty,
		// write will succeed without error. The second write will block until that
		// data is removed
		limit:           limit,
		discardOverflow: discardOverflow,
		readDeadline:    makePipeDeadline(),
		writeDeadline:   makePipeDeadline(),
	}

	return p
}

type state byte

const (
	open state = iota
	closed
	errord
)

/*
1. 	if state is errord, no read/write happens
2. 	interrupt() marks state as errord"stm went wrong", close() mark it as closed,
	both will call done.Close(). Close() will cause any write return io.ErrClosedPipe, have no impacts on
	reads while there is still remaining data, and cause other reads return io.EOF
*/

type Pipe struct {
	sync.Mutex
	limit           int32
	discardOverflow bool
	data            buf.MultiBuffer
	readSignal      *signal.Notifier
	writeSignal     *signal.Notifier
	done            *done.Instance
	state           state
	err             error

	readDeadline  pipeDeadline
	writeDeadline pipeDeadline
}

var (
	errBufferFull      = errors.New("buffer full")
	errSlowDown        = errors.New("slow down")
	errPipeInterrupted = errors.New("pipe has been interrupted")
)

func (p *Pipe) getState(forRead bool) error {
	switch p.state {
	case open:
		if !forRead && p.limit >= 0 && p.data.Len() > p.limit {
			return errBufferFull
		}
		return nil
	case closed:
		if !forRead {
			return io.ErrClosedPipe
		}
		if !p.data.IsEmpty() {
			return nil
		}
		return io.EOF
	case errord:
		return errors.Join(errPipeInterrupted, p.err)
	default:
		panic("impossible case")
	}
}

func (p *Pipe) readMultiBufferInternal() (buf.MultiBuffer, error) {
	p.Lock()
	defer p.Unlock()

	if err := p.getState(true); err != nil {
		return nil, err
	}

	data := p.data
	p.data = nil
	return data, nil
}

func (p *Pipe) ReadMultiBuffer() (buf.MultiBuffer, error) {
	switch {
	case common.IsClosedChan(p.readDeadline.wait()):
		return nil, os.ErrDeadlineExceeded
	}

	for {
		data, err := p.readMultiBufferInternal()
		if err != nil {
			return nil, err
		}

		if data != nil {
			p.writeSignal.Signal() //okay to write
			return data, nil
		} else {
			select {
			case <-p.readSignal.Wait(): //wait for readSignal
			case <-p.done.Wait(): //wait for done
			case <-p.readDeadline.wait():
				return nil, os.ErrDeadlineExceeded
			}
		}

	}
}

func (p *Pipe) ReadMultiBufferTimeout(d time.Duration) (buf.MultiBuffer, error) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	for {
		data, err := p.readMultiBufferInternal()
		if data != nil || err != nil {
			p.writeSignal.Signal()
			return data, err
		}

		select {
		case <-p.readSignal.Wait():
		case <-p.done.Wait():
		case <-timer.C:
			return nil, buf.ErrReadTimeout
		}
	}
}

func (p *Pipe) writeMultiBufferInternal(mb buf.MultiBuffer) error {
	p.Lock()
	defer p.Unlock()

	if err := p.getState(false); err != nil {
		return err
	}

	if p.data == nil {
		p.data = mb
		return nil
	}

	p.data, _ = buf.MergeMulti(p.data, mb)
	return errSlowDown
}

//  1. when its full
//     a. discardOverflow -> discard the mb
//     b. not discardOverflow -> wait
//  2. when not full or empty, and return nil immedietlly
func (p *Pipe) WriteMultiBuffer(mb buf.MultiBuffer) error {
	switch {
	case common.IsClosedChan(p.writeDeadline.wait()):
		buf.ReleaseMulti(mb)
		return os.ErrDeadlineExceeded
	}

	if mb.IsEmpty() {
		return nil
	}

	for {
		err := p.writeMultiBufferInternal(mb)
		if err == nil {
			p.readSignal.Signal()
			return nil
		}

		if err == errSlowDown {
			p.readSignal.Signal()

			// Yield current goroutine. Hopefully the reading counterpart can pick up the payload.
			runtime.Gosched()
			return nil
		}

		if err == errBufferFull && p.discardOverflow {
			buf.ReleaseMulti(mb)
			return nil
		}

		if err != errBufferFull {
			buf.ReleaseMulti(mb)
			p.readSignal.Signal()
			return err
		}

		select {
		case <-p.writeSignal.Wait():
		case <-p.done.Wait():
			buf.ReleaseMulti(mb)
			return io.ErrClosedPipe
		case <-p.writeDeadline.wait():
			buf.ReleaseMulti(mb)
			return os.ErrDeadlineExceeded
		}
	}
}

func (p *Pipe) WriteMultiBufferTimeout(mb buf.MultiBuffer, d time.Duration) error {
	if mb.IsEmpty() {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	for {
		err := p.writeMultiBufferInternal(mb)
		if err == nil {
			p.readSignal.Signal()
			return nil
		}

		if err == errSlowDown {
			p.readSignal.Signal()

			// Yield current goroutine. Hopefully the reading counterpart can pick up the payload.
			runtime.Gosched()
			return nil
		}

		if err == errBufferFull && p.discardOverflow {
			buf.ReleaseMulti(mb)
			return nil
		}

		if err != errBufferFull {
			buf.ReleaseMulti(mb)
			p.readSignal.Signal()
			return err
		}

		select {
		case <-p.writeSignal.Wait():
		case <-p.done.Wait():
			buf.ReleaseMulti(mb)
			return io.ErrClosedPipe
		case <-timer.C:
			buf.ReleaseMulti(mb)
			return buf.ErrWriteTimeout
		}
	}
}

// CloseWrite
func (p *Pipe) Close() error {
	p.Lock()
	defer p.Unlock()
	if p.state == closed || p.state == errord {
		return nil
	}

	p.state = closed
	p.done.Close()
	return nil
}

// Interrupt implements common.Interruptible.
// release buffers, mark it a done
// can be used to tell the other end of the pipe that something went wrong.
func (p *Pipe) Interrupt(err error) {
	p.Lock()
	defer p.Unlock()

	if p.state == closed || p.state == errord {
		return
	}

	p.state = errord
	p.err = err

	if !p.data.IsEmpty() {
		buf.ReleaseMulti(p.data)
		p.data = nil
	}

	p.done.Close()
}

func (p *Pipe) CloseWrite() error {
	p.Close()
	return nil
}

// SetDeadline sets both read and write deadlines.
func (p *Pipe) SetDeadline(t time.Time) error {
	p.readDeadline.set(t)
	p.writeDeadline.set(t)
	return nil
}

// SetReadDeadline sets the read deadline.
func (p *Pipe) SetReadDeadline(t time.Time) error {
	p.readDeadline.set(t)
	return nil
}

// SetWriteDeadline sets the write deadline.
func (p *Pipe) SetWriteDeadline(t time.Time) error {
	p.writeDeadline.set(t)
	return nil
}

// pipeDeadline is an abstraction for handling timeouts.
type pipeDeadline struct {
	mu     sync.Mutex // Guards timer and cancel
	timer  *time.Timer
	cancel chan struct{} // Must be non-nil
}

func makePipeDeadline() pipeDeadline {
	return pipeDeadline{cancel: make(chan struct{})}
}

// set sets the point in time when the deadline will time out.
// A timeout event is signaled by closing the channel returned by waiter.
// Once a timeout has occurred, the deadline can be refreshed by specifying a
// t value in the future.
//
// A zero value for t prevents timeout.
func (d *pipeDeadline) set(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil && !d.timer.Stop() {
		<-d.cancel // Wait for the timer callback to finish and close cancel
	}
	d.timer = nil

	// Time is zero, then there is no deadline.
	closed := common.IsClosedChan(d.cancel)
	if t.IsZero() {
		if closed {
			d.cancel = make(chan struct{})
		}
		return
	}

	// Time in the future, setup a timer to cancel in the future.
	if dur := time.Until(t); dur > 0 {
		if closed {
			d.cancel = make(chan struct{})
		}
		d.timer = time.AfterFunc(dur, func() {
			close(d.cancel)
		})
		return
	}

	// Time in the past, so close immediately.
	if !closed {
		close(d.cancel)
	}
}

// wait returns a channel that is closed when the deadline is exceeded.
func (d *pipeDeadline) wait() chan struct{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.cancel
}
