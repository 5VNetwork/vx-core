package tun

import (
	"net/netip"

	"github.com/5vnetwork/vx-core/common/buf"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/tun"
)

type tunAdapter struct {
	*TunOption
	device tun.Device
}

func NewTun0(config *TunOption) (TunDeviceWithInfo, error) {
	fd, err := unix.Dup(int(config.FD))
	if err != nil {
		return nil, err
	}

	device, _, err := tun.CreateUnmonitoredTUNFromFD(fd)
	if err != nil {
		return nil, err
	}
	return &tunAdapter{TunOption: config, device: device}, nil
}

func (t *tunAdapter) Start() error {
	return nil
}

func (t *tunAdapter) Close() error {
	return t.device.Close()
}

func (t *tunAdapter) WritePacket(pkt *buf.Buffer) error {
	defer pkt.Release()
	_, err := t.device.Write([][]byte{pkt.Bytes()}, 0)
	return err
}

func (t *tunAdapter) ReadPacket() (*buf.Buffer, error) {
	b := buf.New()
	var sizeArray [1]int
	_, err := t.device.Read([][]byte{b.BytesTo(b.Cap())}, sizeArray[:], 0)
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Extend(int32(sizeArray[0]))
	return b, nil
}

func (t *tunAdapter) DnsServers() []netip.Addr {
	servers := make([]netip.Addr, 0, len(t.TunOption.Dns4)+len(t.TunOption.Dns6))
	servers = append(servers, t.TunOption.Dns4...)
	servers = append(servers, t.TunOption.Dns6...)
	return servers
}

func (t *tunAdapter) Name() string {
	return t.TunOption.Name
}

func (t *tunAdapter) IP4() netip.Addr {
	return t.TunOption.Ip4.Addr()
}

func (t *tunAdapter) IP6() netip.Addr {
	return t.TunOption.Ip6.Addr()
}
