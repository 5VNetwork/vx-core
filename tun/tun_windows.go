/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2025 WireGuard LLC. All Rights Reserved.
 */

package tun

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	sync "sync"
	"sync/atomic"
	"time"
	_ "unsafe"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/tun/internal/wintun"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

// Contains code copied from golang.zx2c4.com/wireguard/tun/tun_windows.go

const (
	rateMeasurementGranularity = uint64((time.Second / 2) / time.Nanosecond)
	spinloopRateThreshold      = 800000000 / 8                                   // 800mbps
	spinloopDuration           = uint64(time.Millisecond / 80 / time.Nanosecond) // ~1gbit/s
)

var (
	WintunTunnelType          = "WireGuard"
	WintunStaticRequestedGUID *windows.GUID
)

type Event int

const (
	EventUp = 1 << iota
	EventDown
	EventMTUUpdate
)

type NativeTun struct {
	wt        *wintun.Adapter
	name      string
	handle    windows.Handle
	rate      rateJuggler
	session   wintun.Session
	readWait  windows.Handle
	events    chan Event
	running   sync.WaitGroup
	closeOnce sync.Once
	close     atomic.Bool
	forcedMTU int
	mtu       int32
	// outSizes  []int
	ip4        netip.Addr
	ip6        netip.Addr
	dnsServers []netip.Addr
}

func NewTun(config *TunOption) (TunDeviceWithInfo, error) {
	var err error
	/* create tun device */
	path := config.Path
	switch runtime.GOARCH {
	case "amd64":
		path = filepath.Join(path, "amd64")
	case "386":
		path = filepath.Join(path, "x86")
	case "arm":
		path = filepath.Join(path, "arm")
	case "arm64":
		path = filepath.Join(path, "arm64")
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
	}
	log.Debug().Msgf("win dll path: %s", path)
	pathUint16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}
	_, err = windows.AddDllDirectory(pathUint16)
	if err != nil {
		return nil, fmt.Errorf("failed to add DLL directory: %w", err)
	}
	wt, err := wintun.CreateAdapter(config.Name, WintunTunnelType, WintunStaticRequestedGUID)
	if err != nil {
		return nil, fmt.Errorf("createAdapter failed: %w", err)
	}

	tun := &NativeTun{
		wt:        wt,
		name:      config.Name,
		handle:    windows.InvalidHandle,
		events:    make(chan Event, 10),
		forcedMTU: int(config.Mtu),
		mtu:       int32(config.Mtu),
	}

	/* configure the TUN device */
	tunLuid := winipcfg.LUID(tun.LUID())
	if config.Ip4.IsValid() {
		err = tunLuid.AddIPAddress(config.Ip4)
		if err != nil {
			return nil, fmt.Errorf("failed to set IPv4 address for interface: %v", err)
		}
		a := config.Ip4.Addr()
		tun.ip4 = a
		if len(config.Dns4) > 0 {
			tun.dnsServers = config.Dns4
			err = tunLuid.SetDNS(windows.AF_INET, config.Dns4, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to set ipv4 DNS for interface: %v", err)
			}
		}
		for _, route := range config.Route4 {
			err = tunLuid.AddRoute(route, netip.IPv4Unspecified(), 0)
			if err != nil {
				return nil, fmt.Errorf("failed to add route for interface: %v", err)
			}
		}
		mibIpIfRow, err := tunLuid.IPInterface(windows.AF_INET)
		if err != nil {
			return nil, fmt.Errorf("failed to get MIB_IPINTERFACE_ROW for interface: %v", err)
		}
		mibIpIfRow.Metric = config.Metric
		mibIpIfRow.UseAutomaticMetric = false
		mibIpIfRow.NLMTU = config.Mtu
		err = mibIpIfRow.Set()
		if err != nil {
			return nil, fmt.Errorf("failed to set metric for interface (ipv4): %v", err)
		}
		// delete multicast route entries
		err = tunLuid.DeleteRoute(netip.MustParsePrefix("224.0.0.0/4"), netip.IPv4Unspecified())
		if err != nil {
			return nil, fmt.Errorf("failed to delete route entry:224.0.0.0/4: %v", err)
		}
		err = tunLuid.DeleteRoute(netip.MustParsePrefix("255.255.255.255/32"), netip.IPv4Unspecified())
		if err != nil {
			return nil, fmt.Errorf("failed to delete route entry:255.255.255.255/32: %v", err)
		}
	} else {
		err = tunLuid.FlushIPAddresses(windows.AF_INET)
		if err != nil {
			return nil, fmt.Errorf("failed to flush ipv4 unicast addresses: %v", err)
		}
	}

	/* ipv6 */
	if config.Ip6.IsValid() {
		err = tunLuid.AddIPAddress(config.Ip6)
		if err != nil {
			return nil, fmt.Errorf("failed to set IPv6 address for interface: %v", err)
		}
		a := config.Ip6.Addr()
		tun.ip6 = a
		// set dns server
		if len(config.Dns6) > 0 {
			tun.dnsServers = append(tun.dnsServers, config.Dns6...)
			err = tunLuid.SetDNS(windows.AF_INET6, config.Dns6, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to set ipv6 DNS for interface: %v", err)
			}
		}
		for _, route := range config.Route6 {
			err = tunLuid.AddRoute(route, netip.IPv6Unspecified(), 0)
			if err != nil {
				return nil, fmt.Errorf("failed to add route: %v", err)
			}
		}
		mibIpIfRow, err := tunLuid.IPInterface(windows.AF_INET6)
		if err != nil {
			return nil, fmt.Errorf("failed to get MIB_IPINTERFACE_ROW for interface: %v", err)
		}
		mibIpIfRow.Metric = config.Metric
		mibIpIfRow.UseAutomaticMetric = false
		mibIpIfRow.NLMTU = config.Mtu
		err = mibIpIfRow.Set()
		if err != nil {
			return nil, fmt.Errorf("failed to set metric for interface (ipv6): %v", err)
		}
		err = tunLuid.DeleteRoute(netip.MustParsePrefix("ff00::/8"), netip.IPv6Unspecified())
		if err != nil {
			return nil, fmt.Errorf("failed to delete route entry:ff00::/8: %v", err)
		}
	} else {
		err = tunLuid.FlushIPAddresses(windows.AF_INET6)
		if err != nil {
			return nil, fmt.Errorf("failed to flush ipv6 unicast addresses: %v", err)
		}
	}

	// tun.session, err = wt.StartSession(0x800000) // Ring capacity, 8 MiB
	// if err != nil {
	// 	tun.wt.Close()
	// 	close(tun.events)
	// 	return nil, fmt.Errorf("error starting session: %w", err)
	// }
	// tun.readWait = tun.session.ReadWaitEvent()

	// i, err := net.InterfaceByName(config.Name)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get interface %s, %w", config.Name, err)
	// }
	// maddrs, err := i.MulticastAddrs()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get multicast addresses for interface %s, %w", config.Name, err)
	// }
	// for _, maddr := range maddrs {
	// 	log.Println("multicast address:", maddr.String())
	// }

	return tun, nil
}

func (tun *NativeTun) IP4() netip.Addr {
	return tun.ip4
}

func (tun *NativeTun) IP6() netip.Addr {
	return tun.ip6
}

func (tun *NativeTun) DnsServers() []netip.Addr {
	return tun.dnsServers
}

func (tun *NativeTun) Start() error {
	var err error
	tun.session, err = tun.wt.StartSession(0x800000) // Ring capacity, 8 MiB
	if err != nil {
		tun.wt.Close()
		close(tun.events)
		return fmt.Errorf("error starting session: %w", err)
	}
	tun.readWait = tun.session.ReadWaitEvent()
	return nil
}

// func DisableDNSRegistration(luid winipcfg.LUID) error {
// 	guid, err := luid.GUID()
// 	if err != nil {
// 		return err
// 	}

// 	dnsInterfaceSettings := &winipcfg.DnsInterfaceSettings{
// 		Version:             winipcfg.DnsInterfaceSettingsVersion1,
// 		Flags:               winipcfg.DnsInterfaceSettingsFlagRegistrationEnabled,
// 		RegistrationEnabled: 0,
// 	}

// 	// For >= Windows 10 1809
// 	err = winipcfg.SetInterfaceDnsSettings(*guid, dnsInterfaceSettings)
// 	if err == nil || !errors.Is(err, windows.ERROR_PROC_NOT_FOUND) {
// 		return err
// 	}

// 	// For < Windows 10 1809
// 	// return luid.fallbackDisableDNSRegistration()
// 	return nil
// }

type rateJuggler struct {
	current       atomic.Uint64
	nextByteCount atomic.Uint64
	nextStartTime atomic.Int64
	changing      atomic.Bool
}

//go:linkname procyield runtime.procyield
func procyield(cycles uint32)

//go:linkname nanotime runtime.nanotime
func nanotime() int64

func (tun *NativeTun) Name() string {
	return tun.name
}

func (tun *NativeTun) File() *os.File {
	return nil
}

func (tun *NativeTun) Events() <-chan Event {
	return tun.events
}

func (tun *NativeTun) Close() error {
	var err error
	tun.closeOnce.Do(func() {
		tun.close.Store(true)
		windows.SetEvent(tun.readWait)
		tun.running.Wait()
		tun.session.End()
		if tun.wt != nil {
			tun.wt.Close()
		}
		close(tun.events)
	})
	return err
}

func (tun *NativeTun) MTU() (int, error) {
	return tun.forcedMTU, nil
}

// TODO: This is a temporary hack. We really need to be monitoring the interface in real time and adapting to MTU changes.
func (tun *NativeTun) ForceMTU(mtu int) {
	if tun.close.Load() {
		return
	}
	update := tun.forcedMTU != mtu
	tun.forcedMTU = mtu
	if update {
		tun.events <- EventMTUUpdate
	}
}

func (tun *NativeTun) BatchSize() int {
	// TODO: implement batching with wintun
	return 1
}

// Note: Read() and Write() assume the caller comes only from a single thread; there's no locking.

func (tun *NativeTun) Read(bufs [][]byte, sizes []int, offset int) (int, error) {
	tun.running.Add(1)
	defer tun.running.Done()
retry:
	if tun.close.Load() {
		return 0, os.ErrClosed
	}
	start := nanotime()
	shouldSpin := tun.rate.current.Load() >= spinloopRateThreshold && uint64(start-tun.rate.nextStartTime.Load()) <= rateMeasurementGranularity*2
	for {
		if tun.close.Load() {
			return 0, os.ErrClosed
		}
		packet, err := tun.session.ReceivePacket()
		switch err {
		case nil:
			// TODO: no copy
			n := copy(bufs[0][offset:], packet)
			sizes[0] = n
			tun.session.ReleaseReceivePacket(packet)
			tun.rate.update(uint64(n))
			return 1, nil
		case windows.ERROR_NO_MORE_ITEMS:
			if !shouldSpin || uint64(nanotime()-start) >= spinloopDuration {
				windows.WaitForSingleObject(tun.readWait, windows.INFINITE)
				goto retry
			}
			procyield(1)
			continue
		case windows.ERROR_HANDLE_EOF:
			return 0, os.ErrClosed
		case windows.ERROR_INVALID_DATA:
			return 0, errors.New("send ring corrupt")
		}
		return 0, fmt.Errorf("Read failed: %w", err)
	}
}

func (tun *NativeTun) Write(bufs [][]byte, offset int) (int, error) {
	tun.running.Add(1)
	defer tun.running.Done()
	if tun.close.Load() {
		return 0, os.ErrClosed
	}

	for i, buf := range bufs {
		packetSize := len(buf) - offset
		tun.rate.update(uint64(packetSize))

		packet, err := tun.session.AllocateSendPacket(packetSize)
		switch err {
		case nil:
			// TODO: Explore options to eliminate this copy.
			copy(packet, buf[offset:])
			tun.session.SendPacket(packet)
			continue
		case windows.ERROR_HANDLE_EOF:
			return i, os.ErrClosed
		case windows.ERROR_BUFFER_OVERFLOW:
			continue // Dropping when ring is full.
		default:
			return i, fmt.Errorf("Write failed: %w", err)
		}
	}
	return len(bufs), nil
}

// LUID returns Windows interface instance ID.
func (tun *NativeTun) LUID() uint64 {
	tun.running.Add(1)
	defer tun.running.Done()
	if tun.close.Load() {
		return 0
	}
	return tun.wt.LUID()
}

// RunningVersion returns the running version of the Wintun driver.
func (tun *NativeTun) RunningVersion() (version uint32, err error) {
	return wintun.RunningVersion()
}

func (rate *rateJuggler) update(packetLen uint64) {
	now := nanotime()
	total := rate.nextByteCount.Add(packetLen)
	period := uint64(now - rate.nextStartTime.Load())
	if period >= rateMeasurementGranularity {
		if !rate.changing.CompareAndSwap(false, true) {
			return
		}
		rate.nextStartTime.Store(now)
		rate.current.Store(total * uint64(time.Second/time.Nanosecond) / period)
		rate.nextByteCount.Store(0)
		rate.changing.Store(false)
	}
}

func (t *NativeTun) ReadPacket() (*buf.Buffer, error) {
	b := buf.NewWithSize(t.mtu)
	var sizeArray [1]int
	_, err := t.Read([][]byte{b.BytesTo(b.Cap())}, sizeArray[:], 0)
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Extend(int32(sizeArray[0]))
	return b, nil
}

func (t *NativeTun) WritePacket(b *buf.Buffer) error {
	defer b.Release()
	_, err := t.Write([][]byte{b.Bytes()}, 0)
	return err
}

func (t *NativeTun) WritePackets(b buf.MultiBuffer) error {
	defer buf.ReleaseMulti(b)
	bufs := make([][]byte, len(b))
	for i, buf := range b {
		bufs[i] = buf.Bytes()
	}
	_, err := t.Write(bufs, 0)
	return err
}

func (t *NativeTun) ReadPackets() (buf.MultiBuffer, error) {
	b := buf.NewWithSize(t.mtu)
	var sizeArray [1]int
	_, err := t.Read([][]byte{b.BytesTo(b.Cap())}, sizeArray[:], 0)
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Extend(int32(sizeArray[0]))
	return buf.MultiBuffer{b}, nil
}

type ChangeCallbackUnregister interface {
	Unregister() error
}

type TunManager struct {
	tun    *NativeTun
	option *TunOption
}

func NewTunManager(option *TunOption, tun *NativeTun) (*TunManager, error) {
	return &TunManager{
		tun:    tun,
		option: option,
	}, nil
}

func (t *TunManager) SetTunSupport6(support6 bool) error {
	tunLuid := winipcfg.LUID(t.tun.LUID())

	if support6 {
		// err := tunLuid.AddIPAddress(t.option.Ip6)
		// if err != nil {
		// 	return fmt.Errorf("failed to set IPv6 address for interface: %v", err)
		// }
		// a := t.option.Ip6.Addr()
		// t.tun.ip6 = a
		// // set dns server
		// if len(t.option.Dns6) > 0 {
		// 	t.tun.dnsServers = append(t.tun.dnsServers, t.option.Dns6...)
		// 	err = tunLuid.SetDNS(windows.AF_INET6, t.option.Dns6, nil)
		// 	if err != nil {
		// 		return fmt.Errorf("failed to set ipv6 DNS for interface: %v", err)
		// 	}
		// }
		for _, route := range t.option.Route6 {
			err := tunLuid.AddRoute(route, netip.IPv6Unspecified(), 0)
			if err != nil {
				return fmt.Errorf("failed to add route: %v", err)
			}
		}
		mibIpIfRow, err := tunLuid.IPInterface(windows.AF_INET6)
		if err != nil {
			return fmt.Errorf("failed to get MIB_IPINTERFACE_ROW for interface: %v", err)
		}
		mibIpIfRow.Metric = t.option.Metric
		mibIpIfRow.UseAutomaticMetric = false
		mibIpIfRow.NLMTU = t.option.Mtu
		err = mibIpIfRow.Set()
		if err != nil {
			return fmt.Errorf("failed to set metric for interface (ipv6): %v", err)
		}
		err = tunLuid.DeleteRoute(netip.MustParsePrefix("ff00::/8"), netip.IPv6Unspecified())
		if err != nil {
			log.Debug().Err(err).Msg("delete route entry: ff00::/8")
		}
	} else {
		// err := tunLuid.FlushDNS(windows.AF_INET6)
		// if err != nil {
		// 	return fmt.Errorf("failed to flush ipv6 DNS for interface: %v", err)
		// }
		// err = tunLuid.FlushIPAddresses(windows.AF_INET6)
		// if err != nil {
		// 	return fmt.Errorf("failed to flush ipv6 unicast addresses: %v", err)
		// }
		err := tunLuid.FlushRoutes(windows.AF_INET6)
		if err != nil {
			return fmt.Errorf("failed to flush route: %v", err)
		}
	}
	return nil
}
