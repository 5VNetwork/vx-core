package dlhelper

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog/log"
)

type controller func(network, address string, fd uintptr) error

type DefaultListener struct {
	controllers []controller
}

type combinedListener struct {
	net.Listener
	locker *FileLocker // for unix domain socket
}

func (l *combinedListener) Close() error {
	if l.locker != nil {
		l.locker.Release()
		l.locker = nil
	}
	return l.Listener.Close()
}

func getControlFunc(ctx context.Context, sockopt *SocketSetting, controllers []controller) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		var controlErr error
		err := c.Control(func(fd uintptr) {
			if sockopt != nil {
				if err := applyInboundSocketOptions(ctx, network, address, fd, sockopt); err != nil {
					controlErr = fmt.Errorf("failed to apply socket options: %w", err)
					return
				}
			}

			setReusePort(fd) // nolint: staticcheck

			for _, controller := range controllers {
				if err := controller(network, address, fd); err != nil {
					controlErr = fmt.Errorf("failed to apply external controller: %w", err)
					return
				}
			}
		})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to apply Control")
			return err
		}
		if controlErr != nil {
			log.Ctx(ctx).Error().Err(controlErr).Msg("controlErr")
			return controlErr
		}
		return nil
	}
}

func (dl *DefaultListener) Listen(ctx context.Context, addr net.Addr, sockopt *SocketSetting) (net.Listener, error) {
	var lc net.ListenConfig
	var network, address string
	// callback is called after the Listen function returns
	// this is used to wrap the listener and do some post processing
	callback := func(l net.Listener, err error) (net.Listener, error) {
		return l, err
	}
	switch addr := addr.(type) {
	case *net.TCPAddr:
		network = addr.Network()
		address = addr.String()
		lc.Control = getControlFunc(ctx, sockopt, dl.controllers)
		if sockopt != nil && (sockopt.TcpKeepAliveInterval != 0 || sockopt.TcpKeepAliveIdle != 0) {
			lc.KeepAlive = time.Duration(-1)
		}
	case *net.UnixAddr:
		lc.Control = nil
		network = addr.Network()
		address = addr.Name
		if (runtime.GOOS == "linux" || runtime.GOOS == "android") && address[0] == '@' {
			// linux abstract unix domain socket is lockfree
			if len(address) > 1 && address[1] == '@' {
				// but may need padding to work with haproxy
				fullAddr := make([]byte, len(syscall.RawSockaddrUnix{}.Path))
				copy(fullAddr, address[1:])
				address = string(fullAddr)
			}
		} else {
			// normal unix domain socket
			var fileMode *os.FileMode
			// parse file mode from address
			if s := strings.Split(address, ","); len(s) == 2 {
				fMode, err := strconv.ParseUint(s[1], 8, 32)
				if err != nil {
					return nil, fmt.Errorf("failed to parse file mode, %w", err)
				}
				address = s[0]
				fm := os.FileMode(fMode)
				fileMode = &fm
			}
			// normal unix domain socket needs lock
			locker := &FileLocker{
				path: address + ".lock",
			}
			if err := locker.Acquire(); err != nil {
				return nil, err
			}
			// set file mode for unix domain socket when it is created
			callback = func(l net.Listener, err error) (net.Listener, error) {
				if err != nil {
					locker.Release()
					return nil, err
				}
				l = &combinedListener{Listener: l, locker: locker}
				if fileMode == nil {
					return l, err
				}
				if cerr := os.Chmod(address, *fileMode); cerr != nil {
					// failed to set file mode, close the listener
					l.Close()
					return nil, fmt.Errorf("failed to set file mode for file:%v, %w", address, cerr)
				}
				return l, err
			}
		}
	}

	l, err := lc.Listen(ctx, network, address)
	l, err = callback(l, err)
	if err == nil && sockopt != nil && sockopt.AcceptProxyProtocol {
		policyFunc := func(upstream net.Addr) (proxyproto.Policy, error) {
			return proxyproto.REQUIRE, nil
		}
		l = &proxyproto.Listener{Listener: l, Policy: policyFunc}
	}
	// TODO: stats
	if sockopt != nil && sockopt.StatsReadCounter != nil && sockopt.StatsWriteCounter != nil {
		l = mynet.NewStatsListener(l, sockopt.StatsReadCounter, sockopt.StatsWriteCounter)
	}
	return l, err
}

func (dl *DefaultListener) ListenPacket(ctx context.Context, network, address string, sockopt *SocketSetting) (net.PacketConn, error) {
	var lc net.ListenConfig

	if address == "" {
		if sockopt.GetLocalAddr6() != "" {
			address = sockopt.LocalAddr6
		} else if sockopt.GetLocalAddr4() != "" {
			address = sockopt.LocalAddr4
		}
	}

	lc.Control = getControlFunc(ctx, sockopt, dl.controllers)
	pc, err := lc.ListenPacket(ctx, network, address)
	if err != nil {
		return nil, err
	}
	if sockopt != nil && sockopt.StatsReadCounter != nil && sockopt.StatsWriteCounter != nil {
		pc = mynet.NewStatsPacketConn(pc, sockopt.StatsReadCounter, sockopt.StatsWriteCounter)
	}
	return pc, nil
}

// RegisterListenerController adds a controller to the effective system listener.
// The controller can be used to operate on file descriptors before they are put into use.
//
// v2ray:api:beta
func RegisterListenerController(controller func(network, address string, fd uintptr) error) error {
	if controller == nil {
		return errors.New("nil listener controller")
	}

	effectiveListener.controllers = append(effectiveListener.controllers, controller)
	return nil
}
