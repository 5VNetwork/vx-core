package sysproxy

import (
	"github.com/5vnetwork/vx-core/i"
)

type SysProxy struct {
	// sync.Mutex
	// mon                 i.DefaultInterfaceChangeSubject
	// info                i.DefaultInterfaceInfo
	// proxySetting        *ProxySetting
	// currentHardwarePort string
}

func NewSysProxy(ps *ProxySetting) *SysProxy {
	setting := &SysProxy{
		// proxySetting: ps,
	}
	return setting
}

func (s *SysProxy) WithMon(mon i.DefaultInterfaceChangeSubject) {
	// s.mon = mon
	// s.info = mon.(i.DefaultInterfaceInfo)
}

func (s *SysProxy) Close() error {
	return nil
}
