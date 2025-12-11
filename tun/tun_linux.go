//go:build linux && !android

package tun

import (
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/tun"
	gtun "gvisor.dev/gvisor/pkg/tcpip/link/tun"
)

type tunWrapper struct {
	device tun.Device
	name   string
}

func NewTun(name string) (TunDevice, error) {
	fd, err := gtun.Open(name)
	if err != nil {
		return nil, err
	}
	device, name, err := tun.CreateUnmonitoredTUNFromFD(fd)
	if err != nil {
		return nil, err
	}
	log.Info().Int("fd", fd).Str("name", name).Msg("fd")
	t := &tunWrapper{
		device: device,
		name:   name,
	}
	return t, nil
}

func (t *tunWrapper) Close() error {
	return t.device.Close()
}

func (t *tunWrapper) WritePacket(pkt *buf.Buffer) error {
	defer pkt.Release()
	_, err := t.device.Write([][]byte{pkt.Bytes()}, 0)
	if err != nil {
		return err
	}
	return nil
}

func (t *tunWrapper) ReadPacket() (*buf.Buffer, error) {
	b := buf.New()
	bufs := make([][]byte, 1)
	bufs[0] = b.BytesTo(b.Cap())
	sizes := []int{0}

	_, err := t.device.Read(bufs, sizes, 0)
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Extend(int32(sizes[0]))
	return b, nil
}

func (t *tunWrapper) Name() string {
	return t.name
}

func (t *tunWrapper) Start() error {
	return nil
}

// type tun0 struct {
// 	*TunOption
// 	fd   int
// 	file *os.File
// }

// func NewTun(config *TunOption) (Tun, error) {
// 	t := &tun0{
// 		TunOption: config,
// 	}
// 	fd, err := unix.Dup(int(config.FD))
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = unix.SetNonblock(fd, true)
// 	if err != nil {
// 		unix.Close(fd)
// 		return nil, err
// 	}
// 	t.fd = fd
// 	t.file = os.NewFile(uintptr(fd), "tun")
// 	return t, nil
// }

// func (t *tun0) Close() error {
// 	return t.file.Close()
// }

// func (t *tun0) WritePacket(pkt *buf.Buffer) error {
// 	defer pkt.Release()
// 	if t.Offset != 0 {
// 		pkt.RetreatStart(t.Offset)
// 		pkt.Zero(0, t.Offset)
// 	}
// 	_, err := t.file.Write(pkt.Bytes())
// 	return err
// }

// func (t *tun0) ReadPacket() (*buf.Buffer, error) {
// 	b := buf.New()
// 	n, err := t.file.Read(b.BytesTo(b.Cap()))
// 	if err != nil {
// 		b.Release()
// 		return nil, err
// 	}
// 	b.Resize(t.Offset, int32(n))
// 	return b, nil
// }

// func (t *tun0) ReadPackets() (buf.MultiBuffer, error) {
// 	b := buf.New()
// 	n, err := t.file.Read(b.BytesTo(b.Cap()))
// 	if err != nil {
// 		b.Release()
// 		return nil, err
// 	}
// 	b.Resize(t.Offset, int32(n))
// 	return buf.MultiBuffer{b}, nil
// }

// func (t *tun0) WritePackets(mb buf.MultiBuffer) error {
// 	defer buf.ReleaseMulti(mb)
// 	for i, b := range mb {
// 		if t.Offset > 0 {
// 			b.RetreatStart(t.Offset)
// 			b.Zero(0, t.Offset)
// 		}
// 		mb[i] = b
// 		_, err := t.file.Write(b.Bytes())
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (t *tun0) Name() string {
// 	return t.TunOption.Name
// }

// func (t *tun0) IP4() netip.Addr {
// 	return t.TunOption.Ip4.Addr()
// }

// func (t *tun0) IP6() netip.Addr {
// 	return t.TunOption.Ip6.Addr()
// }

// func (t *tun0) DnsServers() []netip.Addr {
// 	servers := make([]netip.Addr, 0, len(t.TunOption.Dns4)+len(t.TunOption.Dns6))
// 	servers = append(servers, t.TunOption.Dns4...)
// 	servers = append(servers, t.TunOption.Dns6...)
// 	return servers
// }

// func (t *tun0) Start() error {
// 	return nil
// }
