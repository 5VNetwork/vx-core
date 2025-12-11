package status

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MemInfo represents system memory information
type MemInfo struct {
	// Memory
	MemTotal     uint64
	MemFree      uint64
	MemAvailable uint64
	// Swap
	SwapTotal uint64
	SwapFree  uint64
}

// parseKB converts a string like "1024 kB" to bytes
func parseKB(value string) (uint64, error) {
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return 0, fmt.Errorf("invalid format: %s", value)
	}

	// Parse the number
	kb, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse number: %v", err)
	}

	// Convert KB to bytes
	return kb * 1024, nil
}

func GetMemInfo(output string) (*MemInfo, error) {
	info := &MemInfo{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		bytes, err := parseKB(value)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", key, err)
		}

		switch key {
		case "MemTotal":
			info.MemTotal = bytes
		case "MemFree":
			info.MemFree = bytes
		case "MemAvailable":
			info.MemAvailable = bytes
		case "SwapTotal":
			info.SwapTotal = bytes
		case "SwapFree":
			info.SwapFree = bytes
		}
	}

	return info, nil
}

type PhysMemInfo struct {
	UsedTotal    int64 // Total used memory in MB
	WiredMB      int64 // Wired memory in MB
	CompressorMB int64 // Compressor memory in MB
	UnusedMB     int64 // Unused memory in MB
}

func ParsePhysMemInfoMac(output string) (*PhysMemInfo, error) {
	// Example: "PhysMem: 15G used (2824M wired, 6653M compressor), 80M unused."

	// First, let's extract the main components using regex
	re := regexp.MustCompile(`PhysMem: (\d+)([GM]) used \((\d+)M wired, (\d+)M compressor\), (\d+)M unused`)
	matches := re.FindStringSubmatch(output)

	if len(matches) != 6 {
		return nil, fmt.Errorf("unexpected format in PhysMem output: %s", output)
	}

	// Parse total used memory (handling both G and M units)
	usedValue, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total used memory: %w", err)
	}

	// Convert to MB if the unit is GB
	if matches[2] == "G" {
		usedValue *= 1024 // Convert GB to MB
	}

	// Parse wired memory
	wired, err := strconv.ParseInt(matches[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wired memory: %w", err)
	}

	// Parse compressor memory
	compressor, err := strconv.ParseInt(matches[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compressor memory: %w", err)
	}

	// Parse unused memory
	unused, err := strconv.ParseInt(matches[5], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unused memory: %w", err)
	}

	return &PhysMemInfo{
		UsedTotal:    usedValue,
		WiredMB:      wired,
		CompressorMB: compressor,
		UnusedMB:     unused,
	}, nil
}
