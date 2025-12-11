package user

import (
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/uuid"
)

type NullUser struct {
	counter atomic.Uint64
}

func (u *NullUser) Uid() uuid.UUID {
	return uuid.UUID{}
}

func (u *NullUser) Level() uint32 {
	return 0
}

func (u *NullUser) Secret() uuid.UUID {
	return uuid.UUID{}
}

func (u *NullUser) Counter() *atomic.Uint64 {
	return &u.counter
}
