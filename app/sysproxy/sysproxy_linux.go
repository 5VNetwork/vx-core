package sysproxy

import (
	"sync"

	"github.com/5vnetwork/vx-core/i"
)

type SysProxy struct {
	sync.Mutex
	mon                 i.DefaultInterfaceChangeSubject
	info                i.DefaultInterfaceInfo
	proxySetting        *ProxySetting
	currentHardwarePort string
}

func NewSysProxy(ps *ProxySetting) *SysProxy {
	panic("not implemented")
}

func (s *SysProxy) WithMon(mon i.DefaultInterfaceChangeSubject) {
}

func (s *SysProxy) Close() error {

	return nil
}
