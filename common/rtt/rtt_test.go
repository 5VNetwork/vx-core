package rtt

import (
	"testing"
	"time"
)

func TestICMPRTT(t *testing.T) {
	// Test with a reliable host (Google DNS)
	rtt, err := PingRTT("8.8.8.8", 5*time.Second)
	if err != nil {
		t.Logf("ICMP ping failed (may require elevated privileges): %v", err)
		t.Skip("Skipping ICMP test - may require elevated privileges")
		return
	}

	if rtt <= 0 {
		t.Errorf("Expected positive RTT, got %v", rtt)
	}

	if rtt > 5*time.Second {
		t.Errorf("RTT %v exceeded timeout", rtt)
	}

	t.Logf("ICMP RTT to 8.8.8.8: %v", rtt)
}

func TestICMPRTTWithDomain(t *testing.T) {
	// Test with domain name resolution
	rtt, err := PingRTT("google.com", 5*time.Second)
	if err != nil {
		t.Logf("ICMP ping to domain failed (may require elevated privileges): %v", err)
		t.Skip("Skipping ICMP domain test - may require elevated privileges")
		return
	}

	if rtt <= 0 {
		t.Errorf("Expected positive RTT, got %v", rtt)
	}

	t.Logf("ICMP RTT to google.com: %v", rtt)
}