package tun

import (
	"io"
	"net/netip"
	"os"
	sync "sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"gvisor.dev/gvisor/pkg/tcpip/header"

	"golang.org/x/sys/unix"
)

type Event int

const (
	EventUp = 1 << iota
	EventDown
	EventMTUUpdate
)

type ReadWriteCloserTun struct {
	Rw         io.ReadWriteCloser
	Offset     int32
	once       sync.Once
	name       string
	ip4        netip.Addr
	ip6        netip.Addr
	dnsServers []netip.Addr
}

func NewTun(config *TunOption) (TunDeviceWithInfo, error) {
	dupFd, err := unix.Dup(int(config.FD))
	if err != nil {
		return nil, err
	}
	err = unix.SetNonblock(dupFd, true)
	if err != nil {
		unix.Close(dupFd)
		return nil, err
	}

	file := os.NewFile(uintptr(dupFd), "/dev/tun")

	dnsServers := make([]netip.Addr, 0, len(config.Dns4)+len(config.Dns6))
	dnsServers = append(dnsServers, config.Dns4...)
	dnsServers = append(dnsServers, config.Dns6...)

	return &ReadWriteCloserTun{
		Rw:         file,
		Offset:     config.Offset,
		ip4:        config.Ip4.Addr(),
		ip6:        config.Ip6.Addr(),
		dnsServers: dnsServers,
		name:       config.Name,
	}, nil
}

func (u *ReadWriteCloserTun) Start() error {
	return nil
}

func (u *ReadWriteCloserTun) Close() error {
	var err error
	u.once.Do(func() {
		err = u.Rw.Close()
	})
	return err
}

func (t *ReadWriteCloserTun) Name() string {
	return t.name
}

func (t *ReadWriteCloserTun) DnsServers() []netip.Addr {
	return t.dnsServers
}

func (t *ReadWriteCloserTun) IP4() netip.Addr {
	return t.ip4
}

func (t *ReadWriteCloserTun) IP6() netip.Addr {
	return t.ip6
}

func (u *ReadWriteCloserTun) ReadPacket() (*buf.Buffer, error) {
	b := buf.New()
	n, err := u.Rw.Read(b.BytesTo(b.Cap()))
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Resize(4, int32(n))
	return b, nil
}

var ipv4FourBytes = []byte{0, 0, 0, 2}
var ipv6FourBytes = []byte{0, 0, 0, 30}

func (u *ReadWriteCloserTun) WritePacket(p *buf.Buffer) error {
	defer p.Release()
	if u.Offset > 0 {
		ipv4 := header.IPVersion(p.Bytes()) == header.IPv4Version
		p.RetreatStart(u.Offset)
		if ipv4 {
			p.Write(ipv4FourBytes)
		} else {
			p.Write(ipv6FourBytes)
		}
	}
	_, err := u.Rw.Write(p.Bytes())
	return err
}

// type NativeTun struct {
// 	name        string
// 	tunFile     *os.File
// 	events      chan Event
// 	errors      chan error
// 	routeSocket int
// 	closeOnce   sync.Once

// 	ip4        netip.Addr
// 	ip6        netip.Addr
// 	dnsServers []netip.Addr
// }

// func NewTun(config *TunOption) (Tun, error) {
// 	dupFd, err := unix.Dup(int(config.FD))
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = unix.SetNonblock(dupFd, true)
// 	if err != nil {
// 		unix.Close(dupFd)
// 		return nil, err
// 	}

// 	file := os.NewFile(uintptr(dupFd), "/dev/tun")

// 	dnsServers := make([]netip.Addr, 0, len(config.Dns4)+len(config.Dns6))
// 	dnsServers = append(dnsServers, config.Dns4...)
// 	dnsServers = append(dnsServers, config.Dns6...)

// 	t := &NativeTun{
// 		name:       config.Name,
// 		tunFile:    file,
// 		events:     make(chan Event, 10),
// 		errors:     make(chan error, 5),
// 		ip4:        config.Ip4.Addr(),
// 		ip6:        config.Ip6.Addr(),
// 		dnsServers: dnsServers,
// 	}

// 	// tunIfindex, err := func() (int, error) {
// 	// 	iface, err := net.InterfaceByName(t.name)
// 	// 	if err != nil {
// 	// 		return -1, err
// 	// 	}
// 	// 	return iface.Index, nil
// 	// }()
// 	// if err != nil {
// 	// 	t.tunFile.Close()
// 	// 	return nil, err
// 	// }

// 	// t.routeSocket, err = socketCloexec(unix.AF_ROUTE, unix.SOCK_RAW, unix.AF_UNSPEC)
// 	// if err != nil {
// 	// 	t.tunFile.Close()
// 	// 	return nil, err
// 	// }

// 	// go t.routineRouteListener(tunIfindex)

// 	return t, nil
// }

// func (t *NativeTun) Start() error {
// 	return nil
// }

// func (t *NativeTun) Name() string {
// 	return t.name
// }

// func (t *NativeTun) DnsServers() []netip.Addr {
// 	return t.dnsServers
// }

// func (t *NativeTun) IP4() netip.Addr {
// 	return t.ip4
// }

// func (t *NativeTun) IP6() netip.Addr {
// 	return t.ip6
// }

// func (t *NativeTun) ReadPacket() (*buf.Buffer, error) {
// 	b := buf.New()
// 	var sizeArray [1]int
// 	_, err := t.Read([][]byte{b.BytesTo(b.Cap())}, sizeArray[:], 0)
// 	if err != nil {
// 		b.Release()
// 		return nil, err
// 	}
// 	b.Resize(4, int32(sizeArray[0]))
// 	return b, nil
// }

// func (t *NativeTun) WritePacket(b *buf.Buffer) error {
// 	b.RetreatStart(4)
// 	defer b.Release()
// 	_, err := t.Write([][]byte{b.Bytes()}, 4)
// 	return err
// }

// func (t *NativeTun) ReadPackets() (buf.MultiBuffer, error) {
// 	b := buf.New()
// 	var sizeArray [1]int
// 	_, err := t.Read([][]byte{b.BytesTo(b.Cap())}, sizeArray[:], 0)
// 	if err != nil {
// 		b.Release()
// 		return nil, err
// 	}
// 	b.Resize(4, int32(sizeArray[0]))
// 	return buf.MultiBuffer{b}, nil
// }
// func (t *NativeTun) WritePackets(mb buf.MultiBuffer) error {
// 	for _, b := range mb {
// 		b.RetreatStart(4)
// 	}
// 	defer buf.ReleaseMulti(mb)

// 	bufs := make([][]byte, len(mb))
// 	for i, buf := range mb {
// 		bufs[i] = buf.Bytes()
// 	}
// 	_, err := t.Write(bufs, 4)
// 	return err
// }

// const utunControlName = "com.apple.net.utun_control"

// func retryInterfaceByIndex(index int) (iface *net.Interface, err error) {
// 	for i := 0; i < 20; i++ {
// 		iface, err = net.InterfaceByIndex(index)
// 		if err != nil && errors.Is(err, unix.ENOMEM) {
// 			time.Sleep(time.Duration(i) * time.Second / 3)
// 			continue
// 		}
// 		return iface, err
// 	}
// 	return nil, err
// }

// // func (tun *NativeTun) routineRouteListener(tunIfindex int) {
// // 	var (
// // 		statusUp  bool
// // 		statusMTU int
// // 	)

// // 	defer close(tun.events)

// // 	data := make([]byte, os.Getpagesize())
// // 	for {
// // 	retry:
// // 		n, err := unix.Read(tun.routeSocket, data)
// // 		if err != nil {
// // 			if errno, ok := err.(unix.Errno); ok && errno == unix.EINTR {
// // 				goto retry
// // 			}
// // 			tun.errors <- err
// // 			return
// // 		}

// // 		if n < 14 {
// // 			continue
// // 		}

// // 		if data[3 /* type */] != unix.RTM_IFINFO {
// // 			continue
// // 		}
// // 		ifindex := int(*(*uint16)(unsafe.Pointer(&data[12 /* ifindex */])))
// // 		if ifindex != tunIfindex {
// // 			continue
// // 		}

// // 		iface, err := retryInterfaceByIndex(ifindex)
// // 		if err != nil {
// // 			tun.errors <- err
// // 			return
// // 		}

// // 		// Up / Down event
// // 		up := (iface.Flags & net.FlagUp) != 0
// // 		if up != statusUp && up {
// // 			tun.events <- EventUp
// // 		}
// // 		if up != statusUp && !up {
// // 			tun.events <- EventDown
// // 		}
// // 		statusUp = up

// // 		// MTU changes
// // 		if iface.MTU != statusMTU {
// // 			tun.events <- EventMTUUpdate
// // 		}
// // 		statusMTU = iface.MTU
// // 	}
// // }

// func (tun *NativeTun) File() *os.File {
// 	return tun.tunFile
// }

// func (tun *NativeTun) Events() <-chan Event {
// 	return tun.events
// }

// func (tun *NativeTun) Read(bufs [][]byte, sizes []int, offset int) (int, error) {
// 	// TODO: the BSDs look very similar in Read() and Write(). They should be
// 	// collapsed, with platform-specific files containing the varying parts of
// 	// their implementations.
// 	select {
// 	case err := <-tun.errors:
// 		return 0, err
// 	default:
// 		buf := bufs[0]
// 		n, err := tun.tunFile.Read(buf[:])
// 		if n < 4 {
// 			return 0, err
// 		}
// 		sizes[0] = n
// 		return 1, err
// 	}
// }

// func (tun *NativeTun) Write(bufs [][]byte, offset int) (int, error) {
// 	if offset < 4 {
// 		return 0, io.ErrShortBuffer
// 	}
// 	for i, buf := range bufs {
// 		buf = buf[offset-4:]
// 		buf[0] = 0x00
// 		buf[1] = 0x00
// 		buf[2] = 0x00
// 		switch buf[4] >> 4 {
// 		case 4:
// 			buf[3] = unix.AF_INET
// 		case 6:
// 			buf[3] = unix.AF_INET6
// 		default:
// 			return i, unix.EAFNOSUPPORT
// 		}
// 		if _, err := tun.tunFile.Write(buf); err != nil {
// 			return i, err
// 		}
// 	}
// 	return len(bufs), nil
// }

// func (tun *NativeTun) Close() error {
// 	var err1, err2 error
// 	tun.closeOnce.Do(func() {
// 		err1 = tun.tunFile.Close()
// 		if tun.routeSocket != -1 {
// 			unix.Shutdown(tun.routeSocket, unix.SHUT_RDWR)
// 			err2 = unix.Close(tun.routeSocket)
// 		} else if tun.events != nil {
// 			close(tun.events)
// 		}
// 	})
// 	if err1 != nil {
// 		return err1
// 	}
// 	return err2
// }

// func (tun *NativeTun) setMTU(n int) error {
// 	fd, err := socketCloexec(
// 		unix.AF_INET,
// 		unix.SOCK_DGRAM,
// 		0,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	defer unix.Close(fd)

// 	var ifr unix.IfreqMTU
// 	copy(ifr.Name[:], tun.name)
// 	ifr.MTU = int32(n)
// 	err = unix.IoctlSetIfreqMTU(fd, &ifr)
// 	if err != nil {
// 		return fmt.Errorf("failed to set MTU on %s: %w", tun.name, err)
// 	}

// 	return nil
// }

// func (tun *NativeTun) MTU() (int, error) {
// 	fd, err := socketCloexec(
// 		unix.AF_INET,
// 		unix.SOCK_DGRAM,
// 		0,
// 	)
// 	if err != nil {
// 		return 0, err
// 	}

// 	defer unix.Close(fd)

// 	ifr, err := unix.IoctlGetIfreqMTU(fd, tun.name)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to get MTU on %s: %w", tun.name, err)
// 	}

// 	return int(ifr.MTU), nil
// }

// func (tun *NativeTun) BatchSize() int {
// 	return 1
// }

// func socketCloexec(family, sotype, proto int) (fd int, err error) {
// 	// See go/src/net/sys_cloexec.go for background.
// 	syscall.ForkLock.RLock()
// 	defer syscall.ForkLock.RUnlock()

// 	fd, err = unix.Socket(family, sotype, proto)
// 	if err == nil {
// 		unix.CloseOnExec(fd)
// 	}
// 	return
// }
