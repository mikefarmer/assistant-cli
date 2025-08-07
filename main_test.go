package main

import (
	"testing"

	"github.com/mikefarmer/assistant-cli/cmd"
)

func TestVersion(t *testing.T) {
	// Test that version variable exists and can be set
	if version == "" {
		t.Error("version should not be empty")
	}

	// Test that version is properly set to default value
	if version != "dev" {
		t.Errorf("expected version to be 'dev', got '%s'", version)
	}
}

func TestSetVersion(t *testing.T) {
	_ = t // Test parameter not needed for this simple test

	// Test that we can set version through cmd package
	originalVersion := version
	testVersion := "test-version"

	// Reset version to original value after test
	defer func() {
		version = originalVersion
		cmd.SetVersion(originalVersion)
	}()

	// Set test version
	version = testVersion
	cmd.SetVersion(testVersion)

	// This test mainly ensures the SetVersion function doesn't panic
	// The actual verification would require accessing the cmd package's internal state
}
