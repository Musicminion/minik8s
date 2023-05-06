package proxy

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSaveIPTables(t *testing.T) {
	path := "test-save-iptables"

	// Ensure file does not exist
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	// Save iptables
	err := SaveIPTables(path)
	if err != nil {
		t.Errorf("failed to save iptables: %v", err)
	}

	// Ensure file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file was not created: %v", err)
	}

	// Cleanup
	// os.Remove(path)
}

func TestRestoreIPTables(t *testing.T) {
	path := "test-restore-iptables"

	// Ensure file does not exist
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	// Create test iptables file
	err := ioutil.WriteFile(path, []byte("*filter\n:INPUT ACCEPT [0:0]\n:FORWARD ACCEPT [0:0]\n:OUTPUT ACCEPT [0:0]\nCOMMIT\n"), 0644)
	if err != nil {
		t.Errorf("failed to create test iptables file: %v", err)
	}

	// Restore iptables
	err = RestoreIPTables(path)
	if err != nil {
		t.Errorf("failed to restore iptables: %v", err)
	}

	// Cleanup
	os.Remove(path)
}
