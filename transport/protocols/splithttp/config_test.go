package splithttp_test

import (
	"testing"

	. "github.com/5vnetwork/vx-core/transport/protocols/splithttp"
)

func Test_GetNormalizedPath(t *testing.T) {
	c := SplitHttpConfig{
		Path: "/?world",
	}

	path := c.GetNormalizedPath()
	if path != "/" {
		t.Error("Unexpected: ", path)
	}
}
