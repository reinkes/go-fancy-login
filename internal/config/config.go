package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Colors for terminal output
const (
	Green  = "\033[0;32m"
	Yellow = "\033[1;33m"
	Cyan   = "\033[1;36m"
	Red    = "\033[0;31m"
	Reset  = "\033[0m"
	Bold   = "\033[1m"
)

// Config holds all configuration for fancy-login
type Config struct {
	AWSProfileTemp string
	DefaultRegion  string
	FancyVerbose   bool
	ForceAWSLogin  bool
	UseK9S         bool
	FancyDebug     bool
	BinDir         string
	AWSDir         string
	KubeDir        string
}

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	// Platform-specific paths
	var binDir string
	var awsProfileTemp string

	if runtime.GOOS == "windows" {
		// Windows: Use AppData\Local for binaries, temp dir for profile scripts
		binDir = filepath.Join(homeDir, "AppData", "Local", "fancy-login")
		awsProfileTemp = filepath.Join(os.TempDir(), "aws_profile.ps1")
	} else {
		// Unix-like (Linux, macOS): Use .local/bin
		binDir = filepath.Join(homeDir, ".local", "bin")
		awsProfileTemp = "/tmp/aws_profile.sh"
	}

	return &Config{
		AWSProfileTemp: getEnvWithDefault("FANCY_PROFILE_TEMP", awsProfileTemp),
		DefaultRegion:  getEnvWithDefault("FANCY_DEFAULT_REGION", "eu-central-1"),
		FancyVerbose:   getEnvBool("FANCY_VERBOSE"),
		FancyDebug:     getEnvBool("FANCY_DEBUG"),
		BinDir:         getEnvWithDefault("FANCY_BIN_DIR", binDir),
		AWSDir:         getEnvWithDefault("FANCY_AWS_DIR", filepath.Join(homeDir, ".aws")),
		KubeDir:        getEnvWithDefault("FANCY_KUBE_DIR", filepath.Join(homeDir, ".kube")),
	}
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns environment variable as boolean
func getEnvBool(key string) bool {
	value := os.Getenv(key)
	return value == "1" || value == "true"
}
