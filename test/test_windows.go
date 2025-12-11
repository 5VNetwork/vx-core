package test

import "golang.org/x/sys/windows"

func IsAdmin() bool {
	var sid *windows.SID
	sid, _ = windows.StringToSid("S-1-5-32-544")

	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return false
	}
	defer token.Close()

	member, _ := token.IsMember(sid)
	return member
}
