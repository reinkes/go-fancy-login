package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}

	// Test default values are set
	if cfg.DefaultRegion == "" {
		t.Error("DefaultRegion should have a default value")
	}

	if cfg.BinDir == "" {
		t.Error("BinDir should have a default value")
	}

	if cfg.AWSDir == "" {
		t.Error("AWSDir should have a default value")
	}

	if cfg.KubeDir == "" {
		t.Error("KubeDir should have a default value")
	}

	// Test platform-specific paths
	homeDir, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		expectedBinDir := filepath.Join(homeDir, "AppData", "Local", "fancy-login")
		if cfg.BinDir != expectedBinDir {
			t.Errorf("Windows BinDir = %v, expected %v", cfg.BinDir, expectedBinDir)
		}
		if !strings.HasSuffix(cfg.AWSProfileTemp, "aws_profile.ps1") {
			t.Errorf("Windows AWSProfileTemp should end with aws_profile.ps1, got %v", cfg.AWSProfileTemp)
		}
	} else {
		expectedBinDir := filepath.Join(homeDir, ".local", "bin")
		if cfg.BinDir != expectedBinDir {
			t.Errorf("Unix BinDir = %v, expected %v", cfg.BinDir, expectedBinDir)
		}
		if cfg.AWSProfileTemp != "/tmp/aws_profile.sh" {
			t.Errorf("Unix AWSProfileTemp = %v, expected /tmp/aws_profile.sh", cfg.AWSProfileTemp)
		}
	}
}

func TestNewConfigEnvironmentVariables(t *testing.T) {
	// Save original environment variables
	originalRegion := os.Getenv("FANCY_DEFAULT_REGION")
	originalBinDir := os.Getenv("FANCY_BIN_DIR")
	originalVerbose := os.Getenv("FANCY_VERBOSE")
	originalDebug := os.Getenv("FANCY_DEBUG")

	// Restore environment after test
	defer func() {
		os.Setenv("FANCY_DEFAULT_REGION", originalRegion)
		os.Setenv("FANCY_BIN_DIR", originalBinDir)
		os.Setenv("FANCY_VERBOSE", originalVerbose)
		os.Setenv("FANCY_DEBUG", originalDebug)
	}()

	// Set test environment variables
	os.Setenv("FANCY_DEFAULT_REGION", "us-west-2")
	os.Setenv("FANCY_BIN_DIR", "/custom/bin")
	os.Setenv("FANCY_VERBOSE", "true")
	os.Setenv("FANCY_DEBUG", "true")

	cfg := NewConfig()

	if cfg.DefaultRegion != "us-west-2" {
		t.Errorf("DefaultRegion = %v, expected us-west-2", cfg.DefaultRegion)
	}

	if cfg.BinDir != "/custom/bin" {
		t.Errorf("BinDir = %v, expected /custom/bin", cfg.BinDir)
	}

	if !cfg.FancyVerbose {
		t.Error("FancyVerbose should be true when FANCY_VERBOSE=true")
	}

	if !cfg.FancyDebug {
		t.Error("FancyDebug should be true when FANCY_DEBUG=true")
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are properly defined
	colors := map[string]string{
		"Green":  Green,
		"Yellow": Yellow,
		"Cyan":   Cyan,
		"Red":    Red,
		"Reset":  Reset,
		"Bold":   Bold,
	}

	for name, color := range colors {
		if color == "" {
			t.Errorf("Color constant %s should not be empty", name)
		}
		if !strings.HasPrefix(color, "\033[") {
			t.Errorf("Color constant %s should start with ANSI escape sequence, got: %s", name, color)
		}
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		AWSProfileTemp: "/test/aws_profile.sh",
		DefaultRegion:  "eu-central-1",
		FancyVerbose:   true,
		ForceAWSLogin:  true,
		UseK9S:         true,
		FancyDebug:     false,
		BinDir:         "/test/bin",
		AWSDir:         "/test/.aws",
		KubeDir:        "/test/.kube",
	}

	// Test that all fields are accessible and properly set
	if cfg.DefaultRegion != "eu-central-1" {
		t.Errorf("DefaultRegion = %v, expected eu-central-1", cfg.DefaultRegion)
	}

	if !cfg.FancyVerbose {
		t.Error("FancyVerbose should be true")
	}

	if !cfg.ForceAWSLogin {
		t.Error("ForceAWSLogin should be true")
	}

	if !cfg.UseK9S {
		t.Error("UseK9S should be true")
	}

	if cfg.FancyDebug {
		t.Error("FancyDebug should be false")
	}
}

func TestHomeDirectoryHandling(t *testing.T) {
	// Test that NewConfig handles home directory properly
	cfg := NewConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory, skipping test")
	}

	// AWS and Kube directories should be under home directory
	if !strings.HasPrefix(cfg.AWSDir, homeDir) {
		t.Errorf("AWSDir should be under home directory. Got: %s, Home: %s", cfg.AWSDir, homeDir)
	}

	if !strings.HasPrefix(cfg.KubeDir, homeDir) {
		t.Errorf("KubeDir should be under home directory. Got: %s, Home: %s", cfg.KubeDir, homeDir)
	}
}

// Benchmark the config creation
func BenchmarkNewConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewConfig()
	}
}

func TestEnvironmentVariableParsing(t *testing.T) {
	tests := []struct {
		name          string
		envVar        string
		envValue      string
		expectedBool  bool
		testFieldName string
	}{
		{"FANCY_VERBOSE true", "FANCY_VERBOSE", "true", true, "FancyVerbose"},
		{"FANCY_VERBOSE 1", "FANCY_VERBOSE", "1", true, "FancyVerbose"},
		{"FANCY_VERBOSE yes", "FANCY_VERBOSE", "yes", false, "FancyVerbose"},
		{"FANCY_VERBOSE false", "FANCY_VERBOSE", "false", false, "FancyVerbose"},
		{"FANCY_VERBOSE empty", "FANCY_VERBOSE", "", false, "FancyVerbose"},
		{"FANCY_DEBUG true", "FANCY_DEBUG", "true", true, "FancyDebug"},
		{"FANCY_DEBUG false", "FANCY_DEBUG", "false", false, "FancyDebug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv(tt.envVar)
			defer os.Setenv(tt.envVar, original)

			// Set test value
			os.Setenv(tt.envVar, tt.envValue)

			cfg := NewConfig()

			var actualBool bool
			switch tt.testFieldName {
			case "FancyVerbose":
				actualBool = cfg.FancyVerbose
			case "FancyDebug":
				actualBool = cfg.FancyDebug
			default:
				t.Fatalf("Unknown test field: %s", tt.testFieldName)
			}

			if actualBool != tt.expectedBool {
				t.Errorf("%s with value %q: got %v, expected %v",
					tt.envVar, tt.envValue, actualBool, tt.expectedBool)
			}
		})
	}
}
