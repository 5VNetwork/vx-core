package sshhelper_test

import (
	"testing"

	"github.com/5vnetwork/vx-core/test"
)

// TestDetectFirewall demonstrates firewall detection
// This is a manual test that requires SSH credentials
func TestDetectFirewall(t *testing.T) {
	t.Skip("Manual test - requires SSH server")
	test.LoadEnvVariables("../../.env")

	client, err := test.GetTestSshClientLocal()
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer client.Close()

	// Test firewall detection
	firewallType, err := client.DetectFirewall()
	if err != nil {
		t.Fatalf("failed to detect firewall: %v", err)
	}

	t.Logf("Detected firewall type: %s", firewallType)
}

// TestOpenPort demonstrates opening a port
func TestOpenPort(t *testing.T) {
	t.Skip("Manual test - requires SSH server")
	test.LoadEnvVariables("../../.env")

	client, err := test.GetTestSshClientLocal()
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer client.Close()

	// Test opening a port
	err = client.OpenPort(8080, "tcp")
	if err != nil {
		t.Fatalf("failed to open port: %v", err)
	}

	t.Log("Successfully opened port 8080/tcp")

	// Test idempotency - opening again should not error
	err = client.OpenPort(8080, "tcp")
	if err != nil {
		t.Fatalf("failed to open port (idempotent check): %v", err)
	}

	t.Log("Idempotent check passed")
}

// TestDeletePortRule demonstrates deleting a port rule
func TestDeletePortRule(t *testing.T) {
	t.Skip("Manual test - requires SSH server")

	client, err := test.GetTestSshClientLocal()
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer client.Close()

	// First open a port
	err = client.OpenPort(8080, "tcp")
	if err != nil {
		t.Fatalf("failed to open port: %v", err)
	}

	// Then delete the rule
	err = client.DeletePortRule(8080, "tcp")
	if err != nil {
		t.Fatalf("failed to delete port rule: %v", err)
	}

	t.Log("Successfully deleted port rule for 8080/tcp")

	// Test idempotency - deleting again should not error
	err = client.DeletePortRule(8080, "tcp")
	if err != nil {
		t.Fatalf("failed to delete port rule (idempotent check): %v", err)
	}

	t.Log("Idempotent delete check passed")
}

// TestBulkOperations demonstrates opening multiple ports
func TestBulkOperations(t *testing.T) {
	t.Skip("Manual test - requires SSH server")

	client, err := test.GetTestSshClientLocal()
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer client.Close()

	// Test opening multiple ports
	ports := []uint32{8080, 8081, 8082}
	err = client.OpenPorts(ports, "tcp")
	if err != nil {
		t.Fatalf("failed to open ports: %v", err)
	}

	t.Logf("Successfully opened ports: %v", ports)

	// Test deleting multiple ports
	err = client.DeletePortRules(ports, "tcp")
	if err != nil {
		t.Fatalf("failed to delete port rules: %v", err)
	}

	t.Logf("Successfully deleted port rules: %v", ports)
}

// TestEnableDisableFirewall demonstrates firewall control
func TestEnableDisableFirewall(t *testing.T) {
	t.Skip("Manual test - requires SSH server")
	test.LoadEnvVariables("../../.env")
	client, err := test.GetTestSshClientLocal()
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer client.Close()

	// Test enabling firewall
	// err = client.EnableFirewall()
	// if err != nil {
	// 	t.Fatalf("failed to enable firewall: %v", err)
	// }

	t.Log("Successfully enabled firewall")

	// Test getting status
	status, err := client.GetFirewallStatus()
	if err != nil {
		t.Fatalf("failed to get firewall status: %v", err)
	}

	t.Logf("Firewall status:\n%s", status)
}
