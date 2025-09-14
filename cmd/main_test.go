package main

import (
	"strings"
	"testing"
)

func TestContainsDev(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		expected bool
	}{
		{
			name:     "Profile with _DEV_ should return true",
			profile:  "PROJECT_DEV_ENVIRONMENT",
			expected: true,
		},
		{
			name:     "Profile without _DEV_ should return false",
			profile:  "PROJECT_PROD_ENVIRONMENT",
			expected: false,
		},
		{
			name:     "Empty profile should return false",
			profile:  "",
			expected: false,
		},
		{
			name:     "Profile with dev (lowercase) should return false",
			profile:  "PROJECT_dev_ENVIRONMENT",
			expected: false,
		},
		{
			name:     "Profile with DEV but not _DEV_ should return false",
			profile:  "PROJECTDEVENVIRONMENT",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDev(tt.profile)
			if result != tt.expected {
				t.Errorf("containsDev(%q) = %v, expected %v", tt.profile, result, tt.expected)
			}
		})
	}
}

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

// Benchmark the containsDev function
func BenchmarkContainsDev(b *testing.B) {
	testProfile := "OV_DEV_ENVIRONMENT"
	for i := 0; i < b.N; i++ {
		containsDev(testProfile)
	}
}

// Test edge cases for profile names
func TestContainsDevEdgeCases(t *testing.T) {
	edgeCases := []struct {
		profile  string
		expected bool
		desc     string
	}{
		{"_DEV_", true, "minimal _DEV_ pattern"},
		{"PREFIX_DEV_SUFFIX", true, "standard pattern"},
		{"MULTI_DEV_DEV_TEST", true, "multiple _DEV_ occurrences"},
		{"_DEVELOPMENT_", false, "similar but not exact pattern"},
		{"DEV", false, "without underscores"},
		{"_DEV", false, "missing trailing underscore"},
		{"DEV_", false, "missing leading underscore"},
		{strings.Repeat("A", 100) + "_DEV_" + strings.Repeat("B", 100), true, "very long profile name"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := containsDev(tc.profile)
			if result != tc.expected {
				t.Errorf("containsDev(%q) = %v, expected %v (%s)", tc.profile, result, tc.expected, tc.desc)
			}
		})
	}
}
