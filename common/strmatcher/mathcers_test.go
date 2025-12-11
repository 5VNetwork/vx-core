package strmatcher

import (
	"fmt"
	"testing"
)

func TestToDomain(t *testing.T) {
	s, err := ToDomain("")
	fmt.Println(s, err)
}
