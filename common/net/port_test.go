package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortFromBytes(t *testing.T) {
	bytes := []byte{39, 16}
	port := PortFromBytes(bytes)
	assert.Equal(t, int(port), 10000)
}
