package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FancyConfig represents the main configuration structure
type FancyConfig struct {
	ProfileConfigs map[string]ProfileConfig `yaml:"profile_configs"`
	Settings       GlobalSettings           `yaml:"settings"`
}

// ProfileConfig holds configuration for a specific AWS profile
type ProfileConfig struct {
	Name            string `yaml:"name"`
	AccountID       string `yaml:"account_id,omitempty"`
	ECRLogin        bool   `yaml:"ecr_login"`
	ECRRegion       string `yaml:"ecr_region"`
	K8sContext      string `yaml:"k8s_context"`
	K9sAutoLaunch   bool   `yaml:"k9s_auto_launch"`
	NamespacePrefix string `yaml:"namespace_prefix,omitempty"`
}

// GlobalSettings contains global configuration options
type GlobalSettings struct {
	DefaultRegion      string `yaml:"default_region"`
	ConfigWizardRun    bool   `yaml:"config_wizard_run"`
	PreferLocalConfigs bool   `yaml:"prefer_local_configs"`
}

// DefaultFancyConfig returns a default configuration
func DefaultFancyConfig() *FancyConfig {
	return &FancyConfig{
		ProfileConfigs: make(map[string]ProfileConfig),
		Settings: GlobalSettings{
			DefaultRegion:      "eu-central-1",
			ConfigWizardRun:    false,
			PreferLocalConfigs: true,
		},
	}
}

// LoadFancyConfig loads the fancy configuration from file
func LoadFancyConfig() (*FancyConfig, error) {
	configPath := GetFancyConfigPath()

	// If config doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultFancyConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config FancyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Ensure maps are initialized
	if config.ProfileConfigs == nil {
		config.ProfileConfigs = make(map[string]ProfileConfig)
	}

	return &config, nil
}

// SaveFancyConfig saves the fancy configuration to file
func (fc *FancyConfig) SaveFancyConfig() error {
	configPath := GetFancyConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(fc)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return nil
}

// GetFancyConfigPath returns the path to the fancy config file
func GetFancyConfigPath() string {
	// Check for local config first (for development)
	localConfig := ".fancy-config.yaml"
	if _, err := os.Stat(localConfig); err == nil {
		abs, _ := filepath.Abs(localConfig)
		return abs
	}

	// Default to home directory
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".fancy-config.yaml")
}

// GetProfileConfig returns the profile config for a given AWS profile
func (fc *FancyConfig) GetProfileConfig(profile string) (*ProfileConfig, error) {
	if config, exists := fc.ProfileConfigs[profile]; exists {
		return &config, nil
	}
	return nil, fmt.Errorf("no configuration found for profile: %s", profile)
}

// ShouldPerformECRLogin determines if ECR login should be performed for a profile
func (fc *FancyConfig) ShouldPerformECRLogin(profile string) bool {
	config, err := fc.GetProfileConfig(profile)
	if err != nil {
		return false
	}
	return config.ECRLogin
}

// ShouldAutoLaunchK9s determines if K9s should be auto-launched for a profile
func (fc *FancyConfig) ShouldAutoLaunchK9s(profile string) bool {
	config, err := fc.GetProfileConfig(profile)
	if err != nil {
		return false
	}
	return config.K9sAutoLaunch
}

// GetK8sContextForProfile returns the Kubernetes context for a profile
func (fc *FancyConfig) GetK8sContextForProfile(profile string) string {
	config, err := fc.GetProfileConfig(profile)
	if err != nil {
		return ""
	}
	return config.K8sContext
}

// GetECRRegionForProfile returns the ECR region for a profile
func (fc *FancyConfig) GetECRRegionForProfile(profile string) string {
	config, err := fc.GetProfileConfig(profile)
	if err != nil {
		return fc.Settings.DefaultRegion
	}
	if config.ECRRegion == "" {
		return fc.Settings.DefaultRegion
	}
	return config.ECRRegion
}
