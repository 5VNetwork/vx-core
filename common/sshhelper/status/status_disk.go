package status

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// DiskInfo represents information about a disk partition
type DiskInfo struct {
	Filesystem    string
	SizeTotal     int64 // in bytes
	SizeUsed      int64 // in bytes
	SizeAvailable int64 // in bytes
	UsagePercent  float64
	MountedOn     string
}

// convertToBytes converts human readable size to bytes
func convertToBytes(size string) (int64, error) {
	size = strings.TrimSpace(size)
	if size == "0" {
		return 0, nil
	}

	// Extract the numeric part and unit
	var num float64
	var unit string
	_, err := fmt.Sscanf(size, "%f%s", &num, &unit)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	// Convert to bytes based on unit
	multiplier := int64(1)
	switch strings.ToUpper(unit) {
	case "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(num * float64(multiplier)), nil
}

/*
Filesystem     1K-blocks    Used Available Use% Mounted on
udev              484896       0    484896   0% /dev
tmpfs              99328     472     98856   1% /run
/dev/sda1       10089736 3077584   6478036  33% /
tmpfs             496632       0    496632   0% /dev/shm
tmpfs               5120       0      5120   0% /run/lock
/dev/sda15        126678   11840    114838  10% /boot/efi
tmpfs              99324       0     99324   0% /run/user/1000
*/
func ParseDfOutput(output string) ([]DiskInfo, error) {
	var disks []DiskInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// Skip the header line
	scanner.Scan()

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 6 {
			continue
		}

		// Parse numeric values
		total, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing total size: %v", err)
		}

		used, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing used size: %v", err)
		}

		avail, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing available size: %v", err)
		}

		// Parse percentage (remove % sign)
		usageStr := strings.TrimRight(fields[4], "%")
		usage, err := strconv.ParseFloat(usageStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing usage percentage: %v", err)
		}

		disk := DiskInfo{
			Filesystem:    fields[0],
			SizeTotal:     total,
			SizeUsed:      used,
			SizeAvailable: avail,
			UsagePercent:  usage,
			MountedOn:     fields[len(fields)-1],
		}
		disks = append(disks, disk)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning output: %v", err)
	}

	return disks, nil
}

// formatSize converts bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
