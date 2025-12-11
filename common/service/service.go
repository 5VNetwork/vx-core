//go:build windows

package service

import (
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

func ConnectRemote(host string, access uint32) (*mgr.Mgr, error) {
	var s *uint16
	if host != "" {
		var err error
		s, err = syscall.UTF16PtrFromString(host)
		if err != nil {
			return nil, err
		}
	}
	h, err := windows.OpenSCManager(s, nil, access)
	if err != nil {
		return nil, err
	}
	return &mgr.Mgr{Handle: h}, nil
}

// OpenServiceWithQueryStatus retrieves access to service name, so it can
// be queried.
func OpenService(m *mgr.Mgr, name string, access uint32) (*mgr.Service, error) {
	namePointer, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	h, err := windows.OpenService(m.Handle, namePointer, access)
	if err != nil {
		return nil, err
	}
	return &mgr.Service{Name: name, Handle: h}, nil
}
