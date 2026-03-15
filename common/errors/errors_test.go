package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorIs(t *testing.T) {
	err := NewLeftToRightError(errors.New("test"))
	assert.True(t, errors.Is(err, LeftToRightError{}))

	err1 := NewRightToLeftError(errors.New("test"))
	assert.True(t, errors.Is(err1, RightToLeftError{}))
}
