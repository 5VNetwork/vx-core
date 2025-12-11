package status

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// NetDevStats represents statistics for a network interface
type NetDevStats struct {
	Interface string
	Receive   NetDevReceive
	Transmit  NetDevTransmit
}

// NetDevReceive holds receive statistics
type NetDevReceive struct {
	Bytes      uint64
	Packets    uint64
	Errors     uint64
	Drop       uint64
	FIFO       uint64
	Frame      uint64
	Compressed uint64
	Multicast  uint64
}

// NetDevTransmit holds transmit statistics
type NetDevTransmit struct {
	Bytes      uint64
	Packets    uint64
	Errors     uint64
	Drop       uint64
	FIFO       uint64
	Collisions uint64
	Carrier    uint64
	Compressed uint64
}

func ParseNetDev(output string) ([]NetDevStats, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip first two lines (headers)
	scanner.Scan() // Inter-|   Receive                                                |  Transmit
	scanner.Scan() // face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed

	var stats []NetDevStats

	for scanner.Scan() {
		line := scanner.Text()

		// Split interface name and stats
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		interfaceName := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])

		// Ensure we have all 16 fields
		if len(fields) != 16 {
			return nil, fmt.Errorf("invalid field count for interface %s: got %d, want 16",
				interfaceName, len(fields))
		}

		// Convert string fields to uint64
		values := make([]uint64, 16)
		for i, field := range fields {
			val, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing value for %s: %v", interfaceName, err)
			}
			values[i] = val
		}

		stat := NetDevStats{
			Interface: interfaceName,
			Receive: NetDevReceive{
				Bytes:      values[0],
				Packets:    values[1],
				Errors:     values[2],
				Drop:       values[3],
				FIFO:       values[4],
				Frame:      values[5],
				Compressed: values[6],
				Multicast:  values[7],
			},
			Transmit: NetDevTransmit{
				Bytes:      values[8],
				Packets:    values[9],
				Errors:     values[10],
				Drop:       values[11],
				FIFO:       values[12],
				Collisions: values[13],
				Carrier:    values[14],
				Compressed: values[15],
			},
		}

		stats = append(stats, stat)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
	}

	return stats, nil
}

type NetStatInfo struct {
	Name      string // Interface name
	MTU       int    // Maximum transmission unit
	Network   string // Network address/prefix
	Address   string // Interface address
	RxBytes   int64  // Bytes received
	RxPackets int64  // Packets received
	RxErrors  int64  // Receive errors
	TxBytes   int64  // Bytes transmitted
	TxPackets int64  // Packets transmitted
	TxErrors  int64  // Transmit errors
}

func ParseNetstatIbn(output string) ([]NetStatInfo, error) {
	var stats []NetStatInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip the first header line
	scanner.Scan()

	var currentStat *NetStatInfo

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		// Skip empty lines
		if len(fields) == 0 {
			continue
		}

		// If first field starts with a letter, it's a new interface
		if len(fields[0]) > 0 && unicode.IsLetter(rune(fields[0][0])) {
			// Save previous interface stats if exists
			if currentStat != nil {
				stats = append(stats, *currentStat)
			}

			// Start new interface stats
			currentStat = &NetStatInfo{
				Name: fields[0],
			}

			// Parse MTU if available
			if len(fields) >= 2 {
				mtu, err := strconv.Atoi(fields[1])
				if err == nil {
					currentStat.MTU = mtu
				}
			}

			continue
		}

		// Skip lines that don't have enough fields
		if len(fields) < 7 || currentStat == nil {
			continue
		}

		// Parse Address/Network line
		if strings.Contains(line, "<Link>") {
			// Parse statistics
			rxBytes, _ := strconv.ParseInt(fields[6], 10, 64)
			rxPackets, _ := strconv.ParseInt(fields[4], 10, 64)
			rxErrors, _ := strconv.ParseInt(fields[5], 10, 64)
			txBytes, _ := strconv.ParseInt(fields[9], 10, 64)
			txPackets, _ := strconv.ParseInt(fields[7], 10, 64)
			txErrors, _ := strconv.ParseInt(fields[8], 10, 64)

			currentStat.RxBytes = rxBytes
			currentStat.RxPackets = rxPackets
			currentStat.RxErrors = rxErrors
			currentStat.TxBytes = txBytes
			currentStat.TxPackets = txPackets
			currentStat.TxErrors = txErrors
		} else {
			// Store network/address info
			currentStat.Network = fields[2]
			if len(fields) >= 4 {
				currentStat.Address = fields[3]
			}
		}
	}

	// Add the last interface
	if currentStat != nil {
		stats = append(stats, *currentStat)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return stats, nil
}
