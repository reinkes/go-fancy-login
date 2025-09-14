package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
	NamespaceConfig string
	AWSProfileTemp  string
	DefaultRegion   string
	FancyVerbose    bool
	ForceAWSLogin   bool
	UseK9S          bool
	FancyDebug      bool
	BinDir          string
	AWSDir          string
	KubeDir         string
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
		NamespaceConfig: getEnvWithDefault("FANCY_NAMESPACE_CONFIG", filepath.Join(binDir, ".fancy-namespaces.conf")),
		AWSProfileTemp:  getEnvWithDefault("FANCY_PROFILE_TEMP", awsProfileTemp),
		DefaultRegion:   getEnvWithDefault("FANCY_DEFAULT_REGION", "eu-central-1"),
		FancyVerbose:    getEnvBool("FANCY_VERBOSE"),
		FancyDebug:      getEnvBool("FANCY_DEBUG"),
		BinDir:          getEnvWithDefault("FANCY_BIN_DIR", binDir),
		AWSDir:          getEnvWithDefault("FANCY_AWS_DIR", filepath.Join(homeDir, ".aws")),
		KubeDir:         getEnvWithDefault("FANCY_KUBE_DIR", filepath.Join(homeDir, ".kube")),
	}
}

// ContextMapping represents a mapping from AWS profile pattern to k8s context
type ContextMapping struct {
	Pattern string
	Context string
}

// LoadContextMappings loads context mappings from .fancy-contexts.conf
func LoadContextMappings() ([]ContextMapping, error) {
	cfg := NewConfig()
	contextConf := filepath.Join(cfg.BinDir, ".fancy-contexts.conf")

	// Check for local config file first
	if _, err := os.Stat(".fancy-contexts.conf"); err == nil {
		contextConf = ".fancy-contexts.conf"
	}

	return parseContextFile(contextConf)
}

// LoadNamespaceMappings loads namespace mappings from .fancy-namespaces.conf
func LoadNamespaceMappings() (map[string]string, error) {
	config := NewConfig()
	return parseNamespaceFile(config.NamespaceConfig)
}

// parseContextFile parses the context configuration file
func parseContextFile(filename string) ([]ContextMapping, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open context config %s: %w", filename, err)
	}
	defer file.Close()

	var mappings []ContextMapping
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			pattern := strings.TrimSpace(parts[0])
			context := strings.TrimSpace(parts[1])
			mappings = append(mappings, ContextMapping{
				Pattern: pattern,
				Context: context,
			})
		}
	}

	return mappings, scanner.Err()
}

// parseNamespaceFile parses the namespace configuration file
func parseNamespaceFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open namespace config %s: %w", filename, err)
	}
	defer file.Close()

	mappings := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			code := strings.TrimSpace(parts[0])
			project := strings.TrimSpace(parts[1])
			mappings[code] = project
		}
	}

	return mappings, scanner.Err()
}

// GetNamespaceFromProfile derives the namespace from AWS profile name
func GetNamespaceFromProfile(profile string, namespaceMappings map[string]string) (string, error) {
	// Match pattern like XXX_YYY_DEVENG
	re := regexp.MustCompile(`^([A-Z]+)_([A-Z]+)_DEVENG$`)
	matches := re.FindStringSubmatch(profile)

	if len(matches) != 3 {
		return "", fmt.Errorf("profile %s does not match DEVENG pattern", profile)
	}

	projectCode := matches[1]
	environment := strings.ToLower(matches[2])

	projectName, exists := namespaceMappings[projectCode]
	if !exists {
		return "", fmt.Errorf("project code %s not found in namespace config", projectCode)
	}

	return fmt.Sprintf("%s-%s", environment, projectName), nil
}

// MatchesPattern checks if a profile matches a wildcard pattern
func MatchesPattern(profile, pattern string) bool {
	// Convert shell-style wildcards to regex
	regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(pattern, "*", ".*"), "?", ".") + "$"
	matched, _ := regexp.MatchString(regexPattern, profile)
	return matched
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
