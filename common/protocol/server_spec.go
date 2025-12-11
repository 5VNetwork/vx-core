package protocol

import (
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
)

type ValidationStrategy interface {
	IsValid() bool
	Invalidate()
}

type alwaysValidStrategy struct{}

func AlwaysValid() ValidationStrategy {
	return alwaysValidStrategy{}
}

func (alwaysValidStrategy) IsValid() bool {
	return true
}

func (alwaysValidStrategy) Invalidate() {}

type timeoutValidStrategy struct {
	until time.Time
}

func BeforeTime(t time.Time) ValidationStrategy {
	return &timeoutValidStrategy{
		until: t,
	}
}

func (s *timeoutValidStrategy) IsValid() bool {
	return s.until.After(time.Now())
}

func (s *timeoutValidStrategy) Invalidate() {
	s.until = time.Time{}
}

type ServerSpec struct {
	sync.RWMutex
	dest            net.Destination
	protocolSetting interface{}
	valid           ValidationStrategy
}

func NewServerSpec(dest net.Destination, valid ValidationStrategy, ps interface{}) *ServerSpec {
	return &ServerSpec{
		dest:            dest,
		protocolSetting: ps,
		valid:           valid,
	}
}

func (s *ServerSpec) Destination() net.Destination {
	return s.dest
}

// func (s *ServerSpec) HasUser(user *MemoryUser) bool {
// 	s.RLock()
// 	defer s.RUnlock()

// 	for _, u := range s.users {
// 		if u.Account.Equals(user.Account) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (s *ServerSpec) AddUser(user *MemoryUser) {
// 	if s.HasUser(user) {
// 		return
// 	}

// 	s.Lock()
// 	defer s.Unlock()

// 	s.users = append(s.users, user)
// }

func (s *ServerSpec) GetProtocolSetting() interface{} {
	return s.protocolSetting
}

func (s *ServerSpec) IsValid() bool {
	return s.valid.IsValid()
}

func (s *ServerSpec) Invalidate() {
	s.valid.Invalidate()
}
