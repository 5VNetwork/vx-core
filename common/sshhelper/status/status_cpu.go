package status

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type CPUStat struct {
	CPU       string
	User      uint64
	Nice      uint64
	System    uint64
	Idle      uint64
	Iowait    uint64
	Irq       uint64
	Softirq   uint64
	Steal     uint64
	Guest     uint64
	GuestNice uint64
}

// GetCPUStats reads /proc/stat and returns CPU statistics
// expected input:
// cpu  1001370 17361 706987 1369338765 328267 0 128294 21095 0 0
// cpu0 503074 8389 381024 684596825 174006 0 63274 10996 0 0
// cpu1 498296 8972 325962 684741940 154261 0 65019 10098 0 0
func GetCPUStats(output string) ([]CPUStat, error) {
	var err error

	var stats []CPUStat
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		stat := CPUStat{CPU: fields[0]}
		numbers := make([]uint64, 10)

		// Convert string fields to numbers
		for i := 0; i < 10; i++ {
			numbers[i], err = strconv.ParseUint(fields[i+1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing field %d: %v", i, err)
			}
		}

		stat.User = numbers[0]
		stat.Nice = numbers[1]
		stat.System = numbers[2]
		stat.Idle = numbers[3]
		stat.Iowait = numbers[4]
		stat.Irq = numbers[5]
		stat.Softirq = numbers[6]
		stat.Steal = numbers[7]
		stat.Guest = numbers[8]
		stat.GuestNice = numbers[9]

		stats = append(stats, stat)
	}

	return stats, nil
}

// Calculate CPU usage percentage between two measurements
func calculateCPUUsage(old, new CPUStat) float64 {
	oldTotal := old.User + old.Nice + old.System + old.Idle + old.Iowait +
		old.Irq + old.Softirq + old.Steal + old.Guest + old.GuestNice

	newTotal := new.User + new.Nice + new.System + new.Idle + new.Iowait +
		new.Irq + new.Softirq + new.Steal + new.Guest + new.GuestNice

	totalDelta := newTotal - oldTotal
	idleDelta := (new.Idle + new.Iowait) - (old.Idle + old.Iowait)

	if totalDelta == 0 {
		return 0.0
	}

	return 100.0 * (float64(totalDelta-idleDelta) / float64(totalDelta))
}

// CPU usage: 6.97% user, 13.22% sys, 79.80% idle
func ParseMacCPUUsage(output string) (user, system, idle float64, err error) {
	// Example output: "CPU usage: 7.98% user, 20.44% sys, 71.57% idle"
	re := regexp.MustCompile(`CPU usage: (\d+\.\d+)% user, (\d+\.\d+)% sys, (\d+\.\d+)% idle`)
	matches := re.FindStringSubmatch(output)

	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("unexpected format in CPU usage output: %s", output)
	}

	user, err = strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse user percentage: %w", err)
	}

	system, err = strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse system percentage: %w", err)
	}

	idle, err = strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse idle percentage: %w", err)
	}

	return user, system, idle, nil
}
