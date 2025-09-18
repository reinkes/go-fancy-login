package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConfigWizard handles the interactive configuration setup
type ConfigWizard struct {
	config      *FancyConfig
	awsProfiles []AWSProfile
	k8sContexts []KubernetesContext
	reader      *bufio.Reader
	addNewOnly  bool // If true, only configure new profiles
}

// NewConfigWizard creates a new configuration wizard
func NewConfigWizard() *ConfigWizard {
	return &ConfigWizard{
		config: DefaultFancyConfig(),
		reader: bufio.NewReader(os.Stdin),
	}
}

// NewConfigWizardWithMode creates a new configuration wizard with specific mode
func NewConfigWizardWithMode(addNewOnly bool) *ConfigWizard {
	wizard := NewConfigWizard()
	wizard.addNewOnly = addNewOnly
	return wizard
}

// Run executes the configuration wizard
func (w *ConfigWizard) Run() error {
	fmt.Printf("%s🎯 Fancy Login Configuration Wizard%s\n", Yellow+Bold, Reset)
	fmt.Printf("%s========================================%s\n\n", Yellow, Reset)

	// Try to load existing configuration
	existingConfig, err := LoadFancyConfig()
	if err == nil && len(existingConfig.ProfileConfigs) > 0 {
		fmt.Printf("%s📋 Found existing configuration with %d profiles%s\n", Cyan, len(existingConfig.ProfileConfigs), Reset)
		fmt.Printf("Configuration mode:\n")
		fmt.Printf("  1. Override all (reconfigure all profiles)\n")
		fmt.Printf("  2. Add new profiles only (keep existing, add new ones)\n")
		fmt.Printf("Choice [2]: ")

		choice := w.readInput()
		if choice == "1" {
			fmt.Printf("%s⚠️  This will replace your existing configuration!%s\n", Yellow, Reset)
			fmt.Printf("Are you sure? [y/N]: ")
			confirm := w.readInput()
			if confirm == "" || strings.ToLower(confirm)[0] != 'y' {
				w.addNewOnly = true
				w.config = existingConfig
			}
		} else {
			w.addNewOnly = true
			w.config = existingConfig
		}
		fmt.Println()
	}

	// Load existing configurations
	if err := w.discoverConfigurations(); err != nil {
		return fmt.Errorf("failed to discover configurations: %w", err)
	}

	// Show discovered configurations
	w.showDiscoveredConfigurations()

	// Configure profiles
	if err := w.configureProfiles(); err != nil {
		return fmt.Errorf("failed to configure profiles: %w", err)
	}

	// Configure global settings
	w.configureGlobalSettings()

	// Save configuration
	if err := w.saveConfiguration(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\n%s✅ Configuration wizard completed successfully!%s\n", Green+Bold, Reset)
	fmt.Printf("%sConfiguration saved to: %s%s\n", Green, GetFancyConfigPath(), Reset)

	return nil
}

// discoverConfigurations discovers existing AWS and Kubernetes configurations
func (w *ConfigWizard) discoverConfigurations() error {
	fmt.Printf("%s🔍 Discovering existing configurations...%s\n\n", Cyan, Reset)

	// Discover AWS profiles
	awsConfigPath := GetAWSConfigPath()
	fmt.Printf("Looking for AWS config at: %s\n", awsConfigPath)

	profiles, err := ParseAWSProfiles(awsConfigPath)
	if err != nil {
		fmt.Printf("%s⚠️  Warning: Could not parse AWS config: %v%s\n", Yellow, err, Reset)
		w.awsProfiles = []AWSProfile{}
	} else {
		w.awsProfiles = profiles
		fmt.Printf("%s✅ Found %d AWS profiles%s\n", Green, len(profiles), Reset)
	}

	// Discover Kubernetes contexts
	kubeConfigPath := GetKubeConfigPath()
	fmt.Printf("Looking for Kubernetes config at: %s\n", kubeConfigPath)

	contexts, err := ParseKubernetesContexts(kubeConfigPath)
	if err != nil {
		fmt.Printf("%s⚠️  Warning: Could not parse Kubernetes config: %v%s\n", Yellow, err, Reset)
		w.k8sContexts = []KubernetesContext{}
	} else {
		w.k8sContexts = contexts
		fmt.Printf("%s✅ Found %d Kubernetes contexts%s\n", Green, len(contexts), Reset)
	}

	return nil
}

// showDiscoveredConfigurations displays what was found
func (w *ConfigWizard) showDiscoveredConfigurations() {
	fmt.Printf("\n%s📋 Discovered Configurations:%s\n", Cyan+Bold, Reset)
	fmt.Printf("%s================================%s\n\n", Cyan, Reset)

	// Show AWS profiles
	if len(w.awsProfiles) > 0 {
		fmt.Printf("%sAWS Profiles:%s\n", Yellow+Bold, Reset)
		for i, profile := range w.awsProfiles {
			status := "Standard"
			if profile.IsSSO {
				status = "SSO"
			}
			accountInfo := "Unknown Account"
			if profile.AccountID != "" {
				accountInfo = fmt.Sprintf("Account: %s", profile.AccountID)
			}

			// Show if profile is already configured
			configStatus := ""
			if _, exists := w.config.ProfileConfigs[profile.Name]; exists {
				configStatus = fmt.Sprintf(" %s[Configured]%s", Green, Reset)
			}

			fmt.Printf("  %d. %s (%s, %s)%s\n", i+1, profile.Name, status, accountInfo, configStatus)
		}
		fmt.Println()
	}

	// Show Kubernetes contexts
	if len(w.k8sContexts) > 0 {
		fmt.Printf("%sKubernetes Contexts:%s\n", Yellow+Bold, Reset)
		for i, ctx := range w.k8sContexts {
			namespace := "default"
			if ctx.Namespace != "" {
				namespace = ctx.Namespace
			}
			fmt.Printf("  %d. %s (Cluster: %s, Namespace: %s)\n", i+1, ctx.Name, ctx.Cluster, namespace)
		}
		fmt.Println()
	}
}

// configureProfiles configures each AWS profile individually
func (w *ConfigWizard) configureProfiles() error {
	fmt.Printf("%s🔗 Configuring AWS Profiles%s\n", Cyan+Bold, Reset)
	fmt.Printf("%s========================%s\n\n", Cyan, Reset)

	if len(w.awsProfiles) == 0 {
		fmt.Printf("%s⚠️  No AWS profiles found. You can configure profiles manually later.%s\n\n", Yellow, Reset)
		return nil
	}

	// Filter profiles if we're only adding new ones
	profilesToConfigure := w.awsProfiles
	if w.addNewOnly {
		var newProfiles []AWSProfile
		var existingCount int
		for _, profile := range w.awsProfiles {
			if _, exists := w.config.ProfileConfigs[profile.Name]; !exists {
				newProfiles = append(newProfiles, profile)
			} else {
				existingCount++
			}
		}
		profilesToConfigure = newProfiles

		if existingCount > 0 {
			fmt.Printf("%s📋 Skipping %d existing profiles%s\n", Cyan, existingCount, Reset)
		}
		if len(newProfiles) == 0 {
			fmt.Printf("%s✅ No new profiles found. All profiles are already configured.%s\n\n", Green, Reset)
			return nil
		}
		fmt.Printf("%s🆕 Found %d new profiles to configure%s\n\n", Green, len(newProfiles), Reset)
	}

	fmt.Printf("Let's configure %s profiles. This determines:\n",
		func() string {
			if w.addNewOnly {
				return "new"
			}
			return "each"
		}())
	fmt.Printf("  • Whether to auto-login to ECR\n")
	fmt.Printf("  • Which Kubernetes context to use\n")
	fmt.Printf("  • Whether to auto-launch K9s\n\n")

	for i, profile := range profilesToConfigure {
		fmt.Printf("%s📝 Configuring Profile %d/%d: %s%s%s%s\n",
			Bold, i+1, len(profilesToConfigure), Yellow, profile.Name, Reset, Bold)
		fmt.Printf("%s%s\n", strings.Repeat("─", 50), Reset)

		if profile.AccountID != "" {
			fmt.Printf("Account ID: %s%s%s\n", Cyan, profile.AccountID, Reset)
		}
		if profile.Region != "" {
			fmt.Printf("Region: %s%s%s\n", Cyan, profile.Region, Reset)
		}
		if profile.IsSSO {
			fmt.Printf("Type: %sSSO Profile%s\n", Green, Reset)
		}
		fmt.Println()

		// Ask if user wants to configure this profile
		fmt.Printf("Configure this profile? [Y/n]: ")
		configure := w.readInput()
		if configure != "" && strings.ToLower(configure)[0] == 'n' {
			fmt.Println("Skipping profile.")
			continue
		}

		// Get profile configuration
		profileConfig, err := w.getProfileConfiguration(profile)
		if err != nil {
			return err
		}

		// Store profile configuration directly
		w.config.ProfileConfigs[profile.Name] = ProfileConfig{
			Name:          profile.Name,
			AccountID:     profile.AccountID,
			ECRLogin:      profileConfig.ECRLogin,
			ECRRegion:     profileConfig.ECRRegion,
			K8sContext:    profileConfig.K8sContext,
			K9sAutoLaunch: profileConfig.K9sAutoLaunch,
		}

		fmt.Printf("%s✅ Profile %s configured%s\n\n", Green, profile.Name, Reset)
	}

	return nil
}

// ProfileConfiguration holds temporary configuration for a profile during wizard
type ProfileConfiguration struct {
	Name          string
	ECRLogin      bool
	ECRRegion     string
	K8sContext    string
	K9sAutoLaunch bool
	Namespace     string
}

// getProfileConfiguration gets configuration for a specific profile
func (w *ConfigWizard) getProfileConfiguration(profile AWSProfile) (*ProfileConfiguration, error) {
	config := &ProfileConfiguration{
		Name: profile.Name,
	}

	// ECR login
	fmt.Printf("Enable ECR login for profile %s? [Y/n]: ", profile.Name)
	ecrInput := w.readInput()
	config.ECRLogin = ecrInput == "" || strings.ToLower(ecrInput)[0] == 'y'

	// ECR region
	if config.ECRLogin {
		defaultRegion := "eu-central-1"
		if profile.Region != "" {
			defaultRegion = profile.Region
		}
		fmt.Printf("ECR region for %s [%s]: ", profile.Name, defaultRegion)
		region := w.readInput()
		if region == "" {
			region = defaultRegion
		}
		config.ECRRegion = region
	}

	// Kubernetes context
	if len(w.k8sContexts) > 0 {
		fmt.Printf("Select Kubernetes context for profile %s:\n", profile.Name)
		for i, ctx := range w.k8sContexts {
			fmt.Printf("  %d. %s\n", i+1, ctx.Name)
		}
		fmt.Printf("  0. None\n")
		fmt.Printf("Choice [0]: ")

		choice := w.readInput()
		if choice != "" && choice != "0" {
			if idx, err := strconv.Atoi(choice); err == nil && idx > 0 && idx <= len(w.k8sContexts) {
				config.K8sContext = w.k8sContexts[idx-1].Name
			}
		}
	}

	// K9s auto-launch
	if config.K8sContext != "" {
		fmt.Printf("Auto-launch K9s for profile %s? [y/N]: ", profile.Name)
		k9sInput := w.readInput()
		config.K9sAutoLaunch = k9sInput != "" && strings.ToLower(k9sInput)[0] == 'y'

		// Kubernetes namespace (optional)
		if config.K9sAutoLaunch {
			fmt.Printf("Kubernetes namespace for K9s (optional) [default]: ")
			namespaceInput := w.readInput()
			if namespaceInput != "" && namespaceInput != "default" {
				config.Namespace = namespaceInput
			}
		}
	}

	return config, nil
}

// configureGlobalSettings configures global settings
func (w *ConfigWizard) configureGlobalSettings() {
	fmt.Printf("%s⚙️  Global Settings%s\n", Cyan+Bold, Reset)
	fmt.Printf("%s================%s\n\n", Cyan, Reset)

	// Default region
	fmt.Printf("Default AWS region [%s]: ", w.config.Settings.DefaultRegion)
	region := w.readInput()
	if region != "" {
		w.config.Settings.DefaultRegion = region
	}

	// Mark wizard as completed
	w.config.Settings.ConfigWizardRun = true
}

// saveConfiguration saves the configuration
func (w *ConfigWizard) saveConfiguration() error {
	fmt.Printf("%s💾 Saving Configuration%s\n", Cyan+Bold, Reset)
	fmt.Printf("%s===================%s\n\n", Cyan, Reset)

	configPath := GetFancyConfigPath()
	fmt.Printf("Save configuration to: %s\n", configPath)
	fmt.Printf("Proceed? [Y/n]: ")

	confirm := w.readInput()
	if confirm != "" && strings.ToLower(confirm)[0] == 'n' {
		return fmt.Errorf("configuration save cancelled")
	}

	return w.config.SaveFancyConfig()
}

// readInput reads a line of input from the user
func (w *ConfigWizard) readInput() string {
	input, _ := w.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// RunConfigWizardIfNeeded runs the config wizard if configuration doesn't exist or hasn't been run
func RunConfigWizardIfNeeded() error {
	config, err := LoadFancyConfig()
	if err != nil {
		return err
	}

	// If config exists and wizard has been run, skip
	if config.Settings.ConfigWizardRun {
		return nil
	}

	// Check if config file exists
	configPath := GetFancyConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		// Config exists but wizard hasn't been marked as run
		fmt.Printf("%s⚠️  Configuration file exists but wizard hasn't been completed.%s\n", Yellow, Reset)
		fmt.Printf("Run configuration wizard to update settings? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input))[0] != 'y' {
			return nil
		}
	}

	// Run the wizard
	wizard := NewConfigWizard()
	return wizard.Run()
}

// RunConfigWizard explicitly runs the configuration wizard
func RunConfigWizard() error {
	wizard := NewConfigWizard()
	return wizard.Run()
}
