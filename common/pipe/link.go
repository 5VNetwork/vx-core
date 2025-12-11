package pipe

import (
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
)

type Link struct {
	Reader *Pipe
	Writer *Pipe
}

func (l *Link) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return l.Reader.ReadMultiBuffer()
}

func (l *Link) WriteMultiBuffer(mb buf.MultiBuffer) error {
	return l.Writer.WriteMultiBuffer(mb)
}

func (p *Link) SetDeadline(t time.Time) error {
	p.Reader.readDeadline.set(t)
	p.Writer.writeDeadline.set(t)
	return nil
}

func (p *Link) SetReadDeadline(t time.Time) error {
	p.Reader.readDeadline.set(t)
	return nil
}

func (p *Link) SetWriteDeadline(t time.Time) error {
	p.Writer.writeDeadline.set(t)
	return nil
}

func (l *Link) Close() error {
	l.CloseRead()
	l.CloseWrite()
	return nil
}

func (l *Link) ReadMultiBufferTimeout(d time.Duration) (buf.MultiBuffer, error) {
	return l.Reader.ReadMultiBufferTimeout(d)
}

func (l *Link) CloseWrite() error {
	l.Writer.Close()
	return nil
}
func (l *Link) CloseRead() error {
	l.Reader.Close()
	return nil
}

func (l *Link) Interrupt(err error) {
	l.Writer.Interrupt(err)
	l.Reader.Interrupt(err)
}

func NewLinks(limit int32, discard bool) (*Link, *Link) {
	pipeA := NewPipe(limit, discard)
	pipeB := NewPipe(limit, discard)

	return &Link{
			Reader: pipeA,
			Writer: pipeB,
		}, &Link{
			Reader: pipeB,
			Writer: pipeA,
		}
}
