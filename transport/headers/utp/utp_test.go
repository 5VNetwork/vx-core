package utp_test

import (
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	. "github.com/5vnetwork/vx-core/transport/headers/utp"
)

func TestUTPWrite(t *testing.T) {
	content := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	utpRaw, err := New(nil, &Config{})
	common.Must(err)

	utp := utpRaw.(*UTP)

	payload := buf.New()
	utp.Serialize(payload.Extend(utp.Size()))
	payload.Write(content)

	if payload.Len() != int32(len(content))+utp.Size() {
		t.Error("unexpected payload length: ", payload.Len())
	}
}
