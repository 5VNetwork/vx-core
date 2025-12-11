package status

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/common/sshhelper"

	"github.com/rs/zerolog/log"
)

type Status struct {
	CpuUsage    uint32 // percentage
	DiskUsed    uint32 // bytes
	DiskAll     uint32 // bytes
	MemAvail    uint64 // bytes
	MemAll      uint64 // bytes
	NetInUsage  uint64 // bytes
	NetOutUsage uint64 // bytes
	NetInSpeed  uint32 // bytes/s
	NetOutSpeed uint32 // bytes/s
}

const (
	separator     = "xvxvxvxv"
	cpuLinux      = "cat /proc/stat | grep cpu"
	diskLinux     = "df"
	memLinux      = "cat /proc/meminfo | grep -E 'Mem|Swap'"
	netStatsLinux = "cat /proc/net/dev"
	cpuBSD        = "top -l 1 | grep 'CPU usage'"
	diskBSD       = "df -k"
	memBSD        = "top -l 1 | grep PhysMem"
	netStatsBSD   = "netstat -ibn"
)

func GetStatusStream(ctx context.Context, client *sshhelper.Client, interval time.Duration) (chan *Status, error) {
	isLinux := true
	// linux or bsd
	isDarwin, _ := client.Output("uname -a 2>&1 | grep 'Darwin'", false)
	if isDarwin != "" {
		isLinux = false
	} else {
		isBsd, _ := client.Output("uname -a 2>&1 | grep 'BSD'", false)
		if isBsd != "" {
			isLinux = false
		}
	}

	echoCmd := "\necho " + separator + "\n"
	setLangCmd := "export LANG=en_US.UTF-8"

	// build commands to run
	commandsToRun := []string{setLangCmd}
	if isLinux {
		commandsToRun = append(commandsToRun, cpuLinux, diskLinux, memLinux, netStatsLinux)
	} else {
		commandsToRun = append(commandsToRun, cpuBSD, diskBSD, memBSD, netStatsBSD)
	}

	ch := make(chan *Status, 1)
	var parser parser
	if isLinux {
		parser = &linuxParser{
			interval: interval,
		}
	} else {
		parser = &bsdParser{}
	}

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				output, err := client.CombinedOutput(strings.Join(commandsToRun, echoCmd), false)
				if err != nil {
					log.Printf("failed to RunCommandCombined, result: %s, error: %s", output, err)
					return
				}
				status, err := parser.parse(output)
				if err != nil {
					log.Printf("failed to parse status: %s", err)
					return
				}
				ch <- status
				time.Sleep(interval)
			}
		}
	}()

	return ch, nil
}

type parser interface {
	parse(output string) (*Status, error)
}

type linuxParser struct {
	lastCpuStat    *CPUStat
	lastNetReceive uint64
	lastNetSend    uint64
	interval       time.Duration
}

func (p *linuxParser) parse(output string) (*Status, error) {
	lines := strings.Split(output, separator+"\n")
	ss := &Status{}
	// cpu
	cpuPart := lines[1]
	cpuStats, err := GetCPUStats(cpuPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu stats: %w", err)
	}
	if len(cpuStats) == 0 {
		return nil, fmt.Errorf("no cpu stats")
	}
	if p.lastCpuStat == nil {
		p.lastCpuStat = &cpuStats[0]
	} else {
		ss.CpuUsage = uint32(calculateCPUUsage(*p.lastCpuStat, cpuStats[0]))
		p.lastCpuStat = &cpuStats[0]
	}

	// disk
	diskPart := lines[2]
	diskStats, err := ParseDfOutput(diskPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	usedDiskSize := uint32(0)
	allDiskSize := uint32(0)
	calculatedDisks := make([]string, 0, len(diskStats))
	for _, disk := range diskStats {
		if shouldCalculateDisk(disk.Filesystem, disk.MountedOn) &&
			!slices.Contains(calculatedDisks, disk.Filesystem) {
			calculatedDisks = append(calculatedDisks, disk.Filesystem)
			usedDiskSize += uint32(disk.SizeUsed)
			allDiskSize += uint32(disk.SizeTotal)
		}
	}
	ss.DiskUsed = usedDiskSize
	ss.DiskAll = allDiskSize
	// mem
	memPart := lines[3]
	memInfo, err := GetMemInfo(memPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get mem info: %w", err)
	}
	ss.MemAvail = memInfo.MemAvailable
	ss.MemAll = memInfo.MemTotal
	// net
	netPart := lines[4]
	netStats, err := ParseNetDev(netPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get net stats: %w", err)
	}
	receivedTotal := uint64(0)
	sendTotal := uint64(0)
	for _, netStat := range netStats {
		if shouldCalculateInterface(netStat.Interface) {
			receivedTotal += netStat.Receive.Bytes
			sendTotal += netStat.Transmit.Bytes
		}
	}
	ss.NetInUsage = receivedTotal
	ss.NetOutUsage = sendTotal
	if p.lastNetReceive != 0 {
		ss.NetInSpeed = uint32(receivedTotal-p.lastNetReceive) / uint32(p.interval.Seconds())
		ss.NetOutSpeed = uint32(sendTotal-p.lastNetSend) / uint32(p.interval.Seconds())
	}
	p.lastNetReceive = receivedTotal
	p.lastNetSend = sendTotal
	return ss, nil
}

func shouldCalculateDisk(fs, mount string) bool {
	if strings.HasPrefix(fs, "/dev") ||
		strings.HasPrefix(fs, "//") ||
		strings.HasPrefix(mount, "/mnt") {
		return true
	}
	return false
}

func shouldCalculateInterface(iface string) bool {
	if strings.HasPrefix(iface, "lo") ||
		strings.HasPrefix(iface, "utun") ||
		strings.HasPrefix(iface, "tun") {
		return false
	}
	return true
}

type bsdParser struct {
	lastNetReceive uint64
	lastNetSend    uint64
	timeInterface  time.Duration
}

func (p *bsdParser) parse(output string) (*Status, error) {
	lines := strings.Split(output, separator+"\n")
	status := &Status{}
	// cpu
	cpuPart := lines[1]
	_, _, idle, err := ParseMacCPUUsage(cpuPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu stats: %w", err)
	}
	status.CpuUsage = 100 - uint32(idle)
	// disk
	diskPart := lines[2]
	diskStats, err := ParseDfOutput(diskPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	usedDiskSize := uint32(0)
	allDiskSize := uint32(0)
	calculatedDisks := make([]string, 0, len(diskStats))
	for _, disk := range diskStats {
		if shouldCalculateDisk(disk.Filesystem, disk.MountedOn) &&
			!slices.Contains(calculatedDisks, disk.Filesystem) {
			calculatedDisks = append(calculatedDisks, disk.Filesystem)
			usedDiskSize += uint32(disk.SizeUsed)
			allDiskSize += uint32(disk.SizeTotal)
		}
	}
	status.DiskUsed = usedDiskSize
	status.DiskAll = allDiskSize
	// mem
	memPart := lines[3]
	memInfo, err := ParsePhysMemInfoMac(memPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get mem info: %w", err)
	}
	status.MemAvail = uint64(memInfo.UnusedMB)
	status.MemAll = uint64(memInfo.UsedTotal + memInfo.UsedTotal)
	// net
	netPart := lines[4]
	netStats, err := ParseNetstatIbn(netPart)
	if err != nil {
		return nil, fmt.Errorf("failed to get net stats: %w", err)
	}
	receivedTotal := uint64(0)
	sendTotal := uint64(0)
	calculatedInterfaces := make([]string, 0, len(netStats))
	for _, netStat := range netStats {
		if shouldCalculateInterface(netStat.Name) &&
			!slices.Contains(calculatedInterfaces, netStat.Name) {
			calculatedInterfaces = append(calculatedInterfaces, netStat.Name)
			receivedTotal += uint64(netStat.RxBytes)
			sendTotal += uint64(netStat.TxBytes)
		}
	}
	if p.lastNetReceive != 0 {
		status.NetInSpeed = uint32(receivedTotal-p.lastNetReceive) / uint32(p.timeInterface.Seconds())
		status.NetOutSpeed = uint32(sendTotal-p.lastNetSend) / uint32(p.timeInterface.Seconds())
	}
	p.lastNetReceive = receivedTotal
	p.lastNetSend = sendTotal
	status.NetInUsage = receivedTotal
	status.NetOutUsage = sendTotal
	return status, nil
}
