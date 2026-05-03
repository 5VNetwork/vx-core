// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package service

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// Windows API constants and functions
const (
	SDDL_REVISION_1           = 1
	DACL_SECURITY_INFORMATION = 0x00000004
)

var (
	advapi32 = windows.NewLazySystemDLL("advapi32.dll")

	procConvertStringSecurityDescriptorToSecurityDescriptor = advapi32.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	procSetServiceObjectSecurity                            = advapi32.NewProc("SetServiceObjectSecurity")
)

// ModifyServicePermissions modifies the service security descriptor to allow non-admin users to access it
func ModifyServicePermissions(handle windows.Handle) error {
	// SDDL string that allows:
	// - SYSTEM: Full control
	// - Built-in Administrators: Full control
	// - Interactive Users: Query status and read control
	// - Service Users: Query status and read control
	// - Interactive Users: Read permissions
	// - World: Full control (for service operations)
	sddl := "D:(A;;CCLCSWRPWPDTLOCRRC;;;SY)" +
		"(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;BA)" +
		"(A;;CCLCSWLOCRRC;;;IU)" +
		"(A;;CCLCSWLOCRRC;;;SU)" +
		"(A;;RPWPLC;;;IU)" +
		"S:(AU;FA;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;WD)"

	// Convert SDDL string to security descriptor
	var securityDescriptor unsafe.Pointer
	ret, _, err := procConvertStringSecurityDescriptorToSecurityDescriptor.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(sddl))),
		uintptr(SDDL_REVISION_1),
		uintptr(unsafe.Pointer(&securityDescriptor)),
		0,
	)
	if ret == 0 {
		return fmt.Errorf("failed to convert security descriptor: %v", err)
	}
	defer windows.LocalFree(windows.Handle(securityDescriptor))

	// Set the security descriptor on the service
	ret, _, err = procSetServiceObjectSecurity.Call(
		uintptr(handle),
		uintptr(DACL_SECURITY_INFORMATION),
		uintptr(securityDescriptor),
	)
	if ret == 0 {
		return fmt.Errorf("failed to set service security: %v", err)
	}

	return nil
}

// install service. If already installed, remove it then install
func InstallService(name, desc, path string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		err = RemoveService(name)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	s, err = m.CreateService(name, path, mgr.Config{
		DisplayName: desc,
		// ServiceStartName: "LocalSystem",
	})
	if err != nil {
		return err
	}
	defer s.Close()

	// Modify service permissions to allow non-admin access
	err = ModifyServicePermissions(s.Handle)
	if err != nil {
		s.Delete()
		return fmt.Errorf("failed to modify service permissions: %v", err)
	}

	// Event log source can outlive the SCM entry (e.g. sc delete, partial uninstall).
	// InstallAsEventCreate fails with "registry key already exists" if we skip this.
	err = eventlog.Remove(name)
	if err != nil {
		fmt.Printf("failed to remove event log source: %v", err)
	}

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}

	return nil
}

func RemoveService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}
