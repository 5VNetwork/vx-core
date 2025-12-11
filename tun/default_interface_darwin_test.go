package tun

import (
	"testing"

	"golang.org/x/net/route"
)

func TestPrintRoute(t *testing.T) {
	rib, err := fetchRoutingTable()
	if err != nil {
		t.Fatalf("route.FetchRIB: %v", err)
	}
	msgs, err := parseRoutingTable(rib)
	if err != nil {
		t.Fatalf("route.ParseRIB: %v", err)
	}
	for _, m := range msgs {
		rm, ok := m.(*route.RouteMessage)
		if !ok {
			continue
		}
		printRoute(rm)
	}
}
