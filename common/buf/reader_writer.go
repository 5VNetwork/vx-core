package buf

import "time"

type RW struct {
	Reader
	Writer
}

func NewRW(r Reader, w Writer) *RW {
	return &RW{
		Reader: r,
		Writer: w,
	}
}

type RWD struct {
	Reader
	Writer
	SetDeadline
}

func NewRWD(r Reader, w Writer, d SetDeadline) *RWD {
	return &RWD{
		Reader:      r,
		Writer:      w,
		SetDeadline: d,
	}
}

type SetDeadline interface {
	SetReadDeadline(time.Time) error
}

func (r *RWD) OkayToUnwrapReader() int {
	return 1
}

func (r *RWD) UnwrapReader() any {
	return r.Reader
}

func (r *RWD) OkayToUnwrapWriter() int {
	return 1
}

func (r *RWD) UnwrapWriter() any {
	return r.Writer
}
