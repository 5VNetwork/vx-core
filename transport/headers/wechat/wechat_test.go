package wechat_test

import (
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	. "github.com/5vnetwork/vx-core/transport/headers/wechat"
)

func TestUTPWrite(t *testing.T) {
	videoRaw, err := NewVideoChat(nil, &VideoConfig{})
	common.Must(err)

	video := videoRaw.(*VideoChat)

	payload := buf.New()
	video.Serialize(payload.Extend(video.Size()))

	if payload.Len() != video.Size() {
		t.Error("expected payload size ", video.Size(), " but got ", payload.Len())
	}
}
