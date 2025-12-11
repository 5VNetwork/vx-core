package dispatcher

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/i"

	"github.com/5vnetwork/vx-core/common/buf"

	"github.com/rs/zerolog/log"
)

type TimeoutReaderWriter struct {
	buf.ReaderWriter
	timeout i.TimeoutSetting
	idle    *signal.ActivityChecker
	upOnly  bool
}

func (w *TimeoutReaderWriter) CloseWrite() error {
	w.idle.SetTimeout(w.timeout.UpLinkOnlyTimeout())
	return w.ReaderWriter.CloseWrite()
}

func (w *TimeoutReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if !w.upOnly {
		w.idle.Update()
	}
	return w.ReaderWriter.WriteMultiBuffer(mb)
}

func (w *TimeoutReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	w.idle.Update()
	m, err := w.ReaderWriter.ReadMultiBuffer()
	if err != nil {
		if errors.Is(err, io.EOF) {
			w.idle.SetTimeout(w.timeout.DownLinkOnlyTimeout())
		}
	}
	return m, err
}

func (w *TimeoutReaderWriter) OkayToUnwrapReader() int {
	return 1
}
func (w *TimeoutReaderWriter) UnwrapReader() any {
	return w.ReaderWriter
}
func (w *TimeoutReaderWriter) OkayToUnwrapWriter() int {
	return 1
}
func (w *TimeoutReaderWriter) UnwrapWriter() any {
	return w.ReaderWriter
}

type TimeoutDeadlineRW struct {
	i.DeadlineRW
	timeout i.TimeoutSetting
	idle    *signal.ActivityChecker
	upOnly  bool
}

func (w *TimeoutDeadlineRW) CloseWrite() error {
	w.idle.SetTimeout(w.timeout.UpLinkOnlyTimeout())
	return w.DeadlineRW.CloseWrite()
}

func (w *TimeoutDeadlineRW) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if !w.upOnly {
		w.idle.Update()
	}
	return w.DeadlineRW.WriteMultiBuffer(mb)
}

func (w *TimeoutDeadlineRW) ReadMultiBuffer() (buf.MultiBuffer, error) {
	w.idle.Update()
	m, err := w.DeadlineRW.ReadMultiBuffer()
	if err != nil {
		if errors.Is(err, io.EOF) {
			w.idle.SetTimeout(w.timeout.DownLinkOnlyTimeout())
		}
	}
	return m, err
}

func (w *TimeoutDeadlineRW) OkayToUnwrapReader() int {
	return 1
}
func (w *TimeoutDeadlineRW) UnwrapReader() any {
	return w.DeadlineRW
}
func (w *TimeoutDeadlineRW) OkayToUnwrapWriter() int {
	return 1
}
func (w *TimeoutDeadlineRW) UnwrapWriter() any {
	return w.DeadlineRW
}

type linkStatsAdder interface {
	AddThroughput(uint64)
	AddPing(uint64)
}

type linkStats struct {
	ctx                    context.Context
	ohStats                linkStatsAdder
	initialWriteTime       time.Time
	prevWriteTime          time.Time
	writeCounter           atomic.Uint64
	hasDoneCalculatingRate bool
	initialReadTime        time.Time
	hadAddedPing           bool
}

func (w *linkStats) UpTraffic(n uint64) {
	if w.initialReadTime.IsZero() {
		w.initialReadTime = time.Now()
	}
}

func (w *linkStats) DownTraffic(n uint64) {
	if !w.hadAddedPing {
		w.ohStats.AddPing(uint64(time.Since(w.initialReadTime).Milliseconds()))
		w.hadAddedPing = true
	}
	if !w.hasDoneCalculatingRate && w.ohStats != nil {
		if w.initialWriteTime.IsZero() {
			w.initialWriteTime = time.Now()
			// w.prevWriteTime = time.Now()
		}
		if !w.prevWriteTime.IsZero() && time.Since(w.prevWriteTime).Seconds() > 1 {
			w.hasDoneCalculatingRate = true
		} else {
			w.prevWriteTime = time.Now()
			w.writeCounter.Add(n)
			if w.writeCounter.Load() >= common.OneKB*10 {
				elapsed := time.Since(w.initialWriteTime).Seconds()
				rate := float64(w.writeCounter.Swap(0)) / elapsed
				if rate > 1024*1024*100 {
					log.Ctx(w.ctx).Warn().Float64("elapsed", elapsed).Uint64("rate", uint64(rate)).Msg("throughput is too high")
					w.hasDoneCalculatingRate = true
				} else {
					log.Ctx(w.ctx).Debug().Float64("rate(MBps)", rate/1000/1000).Msg("throughput")
					w.ohStats.AddThroughput(uint64(rate))
				}
				w.initialWriteTime = time.Now()
			}
		}
	}
}

type atomicCounter struct {
	counter *atomic.Uint64
}

func (c atomicCounter) UpTraffic(n uint64) {
	c.counter.Add(n)
}
func (c atomicCounter) DownTraffic(n uint64) {
	c.counter.Add(n)
}

// // calculate throughput of this connection
// type LinkStatsReaderWriter struct {
// 	buf.ReaderWriter

// 	ctx                    context.Context
// 	ohStats                linkStatsAdder
// 	initialWriteTime       time.Time
// 	prevWriteTime          time.Time
// 	writeCounter           atomic.Uint64
// 	hasDoneCalculatingRate bool
// 	initialReadTime        time.Time
// 	hadAddedPing           bool
// }

// func (w *LinkStatsReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
// 	if w.initialReadTime.IsZero() {
// 		w.initialReadTime = time.Now()
// 	}
// 	return w.ReaderWriter.ReadMultiBuffer()
// }

// // downlink, right to left
// func (w *LinkStatsReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
// 	if !w.hadAddedPing {
// 		w.ohStats.AddPing(uint64(time.Since(w.initialReadTime).Milliseconds()))
// 		w.hadAddedPing = true
// 	}
// 	if !w.hasDoneCalculatingRate && w.ohStats != nil {
// 		if w.initialWriteTime.IsZero() {
// 			w.initialWriteTime = time.Now()
// 			// w.prevWriteTime = time.Now()
// 		}
// 		if !w.prevWriteTime.IsZero() && time.Since(w.prevWriteTime).Seconds() > 1 {
// 			w.hasDoneCalculatingRate = true
// 		} else {
// 			w.prevWriteTime = time.Now()
// 			w.writeCounter.Add(uint64(mb.Len()))
// 			if w.writeCounter.Load() >= common.OneMB {
// 				elapsed := time.Since(w.initialWriteTime).Seconds()
// 				rate := float64(w.writeCounter.Swap(0)) / elapsed
// 				if rate > 1024*1024*100 {
// 					log.Ctx(w.ctx).Warn().Float64("elapsed", elapsed).Uint64("rate", uint64(rate)).Msg("throughput is too high")
// 					w.hasDoneCalculatingRate = true
// 				} else {
// 					log.Ctx(w.ctx).Debug().Float64("rate(MBps)", rate/1000/1000).Msg("throughput")
// 					w.ohStats.AddThroughput(uint64(rate))
// 				}
// 				w.initialWriteTime = time.Now()
// 			}
// 		}
// 	}
// 	return w.ReaderWriter.WriteMultiBuffer(mb)
// }

// func (w *LinkStatsReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
// 	if !w.hasDoneCalculatingRate && w.ohStats != nil {
// 		if w.initialWriteTime.IsZero() {
// 			w.initialWriteTime = time.Now()
// 		}
// 		if !w.prevWriteTime.IsZero() && time.Since(w.prevWriteTime).Seconds() > 0.5 {
// 			// if the two consecutive writes are more than 0.5s apart, stop calculating the rate using this connection
// 			w.hasDoneCalculatingRate = true
// 			writeCounter := w.writeCounter.Load()
// 			if writeCounter >= common.OneMB*5 {
// 				elapsed := time.Since(w.initialWriteTime).Seconds()
// 				rate := float64(writeCounter) / elapsed
// 				w.ohStats.AddThroughput(uint64(rate))
// 			}
// 		} else {
// 			w.prevWriteTime = time.Now()
// 			w.writeCounter.Add(uint64(mb.Len()))
// 			if w.writeCounter.Load() >= common.OneMB*5 {
// 				w.hasDoneCalculatingRate = true
// 				elapsed := time.Since(w.initialWriteTime).Seconds()
// 				rate := float64(w.writeCounter.Swap(0)) / elapsed
// 				w.ohStats.AddThroughput(uint64(rate))
// 			}
// 		}
// 	}
// 	return w.ReaderWriter.WriteMultiBuffer(mb)
// }

type LinkStatsPacketConn struct {
	ctx context.Context
	udp.PacketReaderWriter
	ohStats                linkStatsAdder
	initialWriteTime       time.Time
	prevWriteTime          time.Time
	writeCounter           atomic.Uint64
	hasDoneCalculatingRate bool
}

// downlink, right to left
func (w *LinkStatsPacketConn) WritePacket(packet *udp.Packet) error {
	if !w.hasDoneCalculatingRate && w.ohStats != nil {
		if w.initialWriteTime.IsZero() {
			w.initialWriteTime = time.Now()
		}
		if !w.prevWriteTime.IsZero() && time.Since(w.prevWriteTime).Seconds() > 1 {
			// if the two consecutive writes are more than 0.5s apart, stop calculating the rate using this connection
			w.hasDoneCalculatingRate = true
		} else {
			w.prevWriteTime = time.Now()
			w.writeCounter.Add(uint64(packet.Payload.Len()))
			if w.writeCounter.Load() >= common.OneMB {
				elapsed := time.Since(w.initialWriteTime).Seconds()
				rate := float64(w.writeCounter.Swap(0)) / elapsed
				if rate > 1024*1024*100 {
					log.Ctx(w.ctx).Warn().Uint64("rate", uint64(rate)).Msg("throughput is too high")
					w.hasDoneCalculatingRate = true
				} else {
					log.Ctx(w.ctx).Debug().Uint64("rate(MBps)", uint64(rate/1000/1000)).Msg("throughput")
					w.ohStats.AddThroughput(uint64(rate))
				}
				w.initialWriteTime = time.Now()
			}
		}
	}
	return w.PacketReaderWriter.WritePacket(packet)
}

type TimeoutPacketConn struct {
	idle *signal.ActivityChecker
	udp.PacketReaderWriter
}

func (p *TimeoutPacketConn) ReadPacket() (*udp.Packet, error) {
	p.idle.Update()
	return p.PacketReaderWriter.ReadPacket()
}

func (p *TimeoutPacketConn) WritePacket(packet *udp.Packet) error {
	return p.PacketReaderWriter.WritePacket(packet)
}

// change fake ip to real ip
type RealIpPacketConn struct {
	udp.PacketReaderWriter
	m              map[net.Address]net.Address // fake ip to real ip, real ip to real ip
	realIpToFakeIp sync.Map                    // real ip (net.Address) to fake ip (net.Address)
	domainToRealIp sync.Map                    // domain (net.Address) to real ip (net.Address)

	fakeDns i.FakeDnsPool
	dns     i.IPResolver
	ctx     context.Context
}

// assume to be called sequentially
func (p *RealIpPacketConn) ReadPacket() (*udp.Packet, error) {
	packet, err := p.PacketReaderWriter.ReadPacket()
	if err != nil {
		return nil, err
	}
	originalTarget := packet.Target.Address
	if originalTarget.Family().IsIP() {
		if v, ok := p.m[originalTarget]; ok {
			packet.Target.Address = v
			return packet, nil
		}
		if p.fakeDns.IsIPInIPPool(originalTarget) {
			if d := p.fakeDns.GetDomainFromFakeDNS(originalTarget); d != "" {
				ips, err := p.dns.LookupIP(p.ctx, d)
				if err != nil {
					return nil, err
				}
				if len(ips) == 0 {
					return nil, errors.New("failed to find ip for a domain")
				}
				newTarget := net.IPAddress(ips[0])
				packet.Target.Address = newTarget
				p.m[originalTarget] = newTarget
				p.realIpToFakeIp.Store(newTarget, originalTarget)
			} else {
				return nil, errors.New("failed to find domain for a fake ip")
			}
		} else {
			p.m[originalTarget] = originalTarget
		}
	}
	return packet, nil
}

func (p *RealIpPacketConn) WritePacket(packet *udp.Packet) error {
	if v, ok := p.realIpToFakeIp.Load(packet.Source.Address); ok {
		packet.Source.Address = v.(net.Address)
	} else if packet.Source.Address.Family().IsDomain() {
		if v, ok := p.domainToRealIp.Load(packet.Source.Address); ok {
			packet.Source.Address = v.(net.Address)
		} else {
			ips, err := p.dns.LookupIP(p.ctx, packet.Source.Address.Domain())
			if err != nil && len(ips) == 0 {
				packet.Release()
				log.Warn().Ctx(p.ctx).Err(err).
					Str("domain", packet.Source.Address.Domain()).
					Msg("failed to lookup ip for a domain")
				return nil
			}
			newSource := net.IPAddress(ips[0])
			p.domainToRealIp.Store(packet.Source.Address, newSource)
			packet.Source.Address = newSource
		}
	}
	return p.PacketReaderWriter.WritePacket(packet)
}

type StatsReaderWriter struct {
	buf.ReaderWriter
	// might be nil
	upCounter session.UpCounters
	// might be nil
	downCounter session.DownCounters
}

func (w *StatsReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.downCounter.DownTraffic(uint64(mb.Len()))
	return w.ReaderWriter.WriteMultiBuffer(mb)
}

func (w *StatsReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := w.ReaderWriter.ReadMultiBuffer()
	w.upCounter.UpTraffic(uint64(mb.Len()))
	return mb, err
}

type StatsDeadlineRW struct {
	i.DeadlineRW
	// might be nil
	upCounter session.UpCounters
	// might be nil
	downCounter session.DownCounters
}

func (w *StatsDeadlineRW) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.downCounter.DownTraffic(uint64(mb.Len()))
	return w.DeadlineRW.WriteMultiBuffer(mb)
}

func (w *StatsDeadlineRW) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := w.DeadlineRW.ReadMultiBuffer()
	w.upCounter.UpTraffic(uint64(mb.Len()))
	return mb, err
}

type StatsPacketConn struct {
	udp.PacketReaderWriter
	// might be nil
	upCounter session.UpCounters
	// might be nil
	downCounter session.DownCounters
}

func (p *StatsPacketConn) ReadPacket() (*udp.Packet, error) {
	packet, err := p.PacketReaderWriter.ReadPacket()
	if err != nil {
		return nil, err
	}
	p.upCounter.UpTraffic(uint64(packet.Payload.Len()))
	return packet, nil
}

func (p *StatsPacketConn) WritePacket(packet *udp.Packet) error {
	p.downCounter.DownTraffic(uint64(packet.Payload.Len()))
	return p.PacketReaderWriter.WritePacket(packet)
}
