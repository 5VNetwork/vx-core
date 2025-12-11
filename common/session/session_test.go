package session

import (
	"log"
	"testing"
)

func TestSessionIDToString(t *testing.T) {
	sid := NewID()
	log.Print(sid)
}
