package decode

import (
	"strconv"
	"strings"

	mynet "github.com/5vnetwork/vx-core/common/net"
)

// TryParsePorts parses a string of ports in format "123,5000-6000"
// Returns a non-empty slice of PortRange if ports is valid, otherwise returns nil
func TryParsePorts(ports string) []*mynet.PortRange {
	if ports == "" {
		return nil
	}

	var pr []*mynet.PortRange
	ranges := strings.Split(ports, ",")

	for _, r := range ranges {
		if strings.Contains(r, "-") {
			rangeParts := strings.Split(r, "-")
			if len(rangeParts) != 2 {
				return nil
			}

			from, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil
			}

			to, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil
			}

			pr = append(pr, &mynet.PortRange{From: uint32(from), To: uint32(to)})
		} else {
			port, err := strconv.Atoi(r)
			if err != nil {
				return nil
			}

			pr = append(pr, &mynet.PortRange{From: uint32(port), To: uint32(port)})
		}
	}

	if len(pr) == 0 {
		return nil
	}

	return pr
}
