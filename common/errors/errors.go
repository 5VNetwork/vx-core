package errors

import (
	"context"
	"errors"

	"github.com/5vnetwork/vx-core/common/strings"

	"github.com/rs/zerolog/log"
)

var ErrIdle = errors.New("Idle")
var ErrClosed = errors.New("closed")

type errorWrap struct {
	error
}

func New(values ...interface{}) *errorWrap {
	return &errorWrap{
		error: errors.New(strings.Concat(values...)),
	}
}

func (e *errorWrap) Base(err error) *errorWrap {
	return &errorWrap{
		error: errors.Join(e.error, err),
	}
}

func (e *errorWrap) Unwrap() error {
	return e.error
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func (e *errorWrap) WriteToLog(ctx context.Context) {
	log.Ctx(ctx).Warn().Err(e.error).Send()
}

func Join(errs ...error) error {
	var err error
	for _, e := range errs {
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

type AuthError struct {
	error
}

type LeftToRightError struct {
	error
}

func NewLeftToRightError(err error) LeftToRightError {
	return LeftToRightError{err}
}

func (e LeftToRightError) Is(target error) bool {
	if _, ok := target.(*LeftToRightError); ok {
		return true
	}
	return false
}

func (e LeftToRightError) Error() string {
	return "relay leftToRight: " + e.error.Error()
}

func (e LeftToRightError) Unwrap() error {
	return e.error
}

type RightToLeftError struct {
	error
}

func NewRightToLeftError(err error) RightToLeftError {
	return RightToLeftError{err}
}

func (e RightToLeftError) Error() string {
	return "relay rightToLeft: " + e.error.Error()
}

func (e RightToLeftError) Unwrap() error {
	return e.error
}
