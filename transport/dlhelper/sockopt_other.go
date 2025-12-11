//go:build js || dragonfly || netbsd || openbsd || solaris
// +build js dragonfly netbsd openbsd solaris

package dlhelper

import "context"

func applyOutboundSocketOptions(ctx context.Context, network string, address string, fd uintptr, config *SocketSetting) error {
	return nil
}

func applyInboundSocketOptions(ctx context.Context, network string, fd uintptr, config *SocketSetting) error {
	return nil
}

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	return nil
}

func setReuseAddr(fd uintptr) error {
	return nil
}

func setReusePort(fd uintptr) error {
	return nil
}
