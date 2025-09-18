package main

import (
	"testing"
)

func TestVersionVariables(t *testing.T) {
	// Test that version variables are initialized with default values
	if version == "" {
		t.Error("version variable should not be empty")
	}
	if buildTime == "" {
		t.Error("buildTime variable should not be empty")
	}
	if gitCommit == "" {
		t.Error("gitCommit variable should not be empty")
	}
}

func TestShowVersionOutput(t *testing.T) {
	// Capture output by redirecting stdout (basic test)
	// Note: This is a simple test - in a real scenario you might use testify or similar
	// to capture and test output more robustly

	// Just verify the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("showVersion() panicked: %v", r)
		}
	}()

	// This will output to stdout but won't fail the test
	// In a more sophisticated test, we'd capture and verify the output
	showVersion()
}

func TestFlagVariables(t *testing.T) {
	// Test that flag variables are properly initialized
	if verbose == nil {
		t.Error("verbose flag should be initialized")
	}
	if k9sFlag == nil {
		t.Error("k9sFlag should be initialized")
	}
	if forceAWSLogin == nil {
		t.Error("forceAWSLogin flag should be initialized")
	}
	if helpFlag == nil {
		t.Error("helpFlag should be initialized")
	}
	if versionFlag == nil {
		t.Error("versionFlag should be initialized")
	}
}
