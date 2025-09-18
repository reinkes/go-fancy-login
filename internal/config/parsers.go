package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// AWSProfile represents an AWS profile from ~/.aws/config
type AWSProfile struct {
	Name        string
	AccountID   string
	Region      string
	SSOStartURL string
	SSORegion   string
	SSORole     string
	IsSSO       bool
}

// KubernetesContext represents a Kubernetes context from ~/.kube/config
type KubernetesContext struct {
	Name      string
	Cluster   string
	Namespace string
	User      string
}

// KubeConfig represents the structure of ~/.kube/config
type KubeConfig struct {
	APIVersion     string `yaml:"apiVersion"`
	Kind           string `yaml:"kind"`
	CurrentContext string `yaml:"current-context"`
	Contexts       []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster   string `yaml:"cluster"`
			User      string `yaml:"user"`
			Namespace string `yaml:"namespace,omitempty"`
		} `yaml:"context"`
	} `yaml:"contexts"`
	Clusters []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
}

// ParseAWSProfiles parses AWS profiles from ~/.aws/config
func ParseAWSProfiles(awsConfigPath string) ([]AWSProfile, error) {
	if awsConfigPath == "" {
		homeDir, _ := os.UserHomeDir()
		awsConfigPath = filepath.Join(homeDir, ".aws", "config")
	}

	file, err := os.Open(awsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open AWS config file %s: %w", awsConfigPath, err)
	}
	defer file.Close()

	var profiles []AWSProfile
	var currentProfile *AWSProfile
	profileRegex := regexp.MustCompile(`^\[profile\s+(.+)\]$`)
	defaultRegex := regexp.MustCompile(`^\[default\]$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for profile section
		if matches := profileRegex.FindStringSubmatch(line); matches != nil {
			// Save previous profile if exists
			if currentProfile != nil {
				profiles = append(profiles, *currentProfile)
			}
			// Start new profile
			currentProfile = &AWSProfile{
				Name: matches[1],
			}
		} else if defaultRegex.MatchString(line) {
			// Save previous profile if exists
			if currentProfile != nil {
				profiles = append(profiles, *currentProfile)
			}
			// Start default profile
			currentProfile = &AWSProfile{
				Name: "default",
			}
		} else if currentProfile != nil {
			// Parse profile properties
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "sso_account_id":
					currentProfile.AccountID = value
					currentProfile.IsSSO = true
				case "region":
					currentProfile.Region = value
				case "sso_start_url":
					currentProfile.SSOStartURL = value
					currentProfile.IsSSO = true
				case "sso_region":
					currentProfile.SSORegion = value
				case "sso_role_name":
					currentProfile.SSORole = value
				}
			}
		}
	}

	// Don't forget the last profile
	if currentProfile != nil {
		profiles = append(profiles, *currentProfile)
	}

	return profiles, scanner.Err()
}

// ParseKubernetesContexts parses Kubernetes contexts from ~/.kube/config
func ParseKubernetesContexts(kubeConfigPath string) ([]KubernetesContext, error) {
	if kubeConfigPath == "" {
		homeDir, _ := os.UserHomeDir()
		kubeConfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	data, err := os.ReadFile(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kubernetes config file %s: %w", kubeConfigPath, err)
	}

	var kubeConfig KubeConfig
	if err := yaml.Unmarshal(data, &kubeConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Kubernetes config file %s: %w", kubeConfigPath, err)
	}

	var contexts []KubernetesContext
	for _, ctx := range kubeConfig.Contexts {
		contexts = append(contexts, KubernetesContext{
			Name:      ctx.Name,
			Cluster:   ctx.Context.Cluster,
			User:      ctx.Context.User,
			Namespace: ctx.Context.Namespace,
		})
	}

	return contexts, nil
}

// FindAccountIDForProfile attempts to find the AWS account ID for a profile
// This could be extended to actually call AWS CLI if needed
func FindAccountIDForProfile(profile string) (string, error) {
	// For now, try to parse from the profile name if it follows common patterns
	// This could be enhanced to actually call `aws sts get-caller-identity`

	// Try to extract from common naming patterns
	patterns := []string{
		`(\d{12})`,   // Direct account ID
		`-(\d{12})-`, // Account ID in middle
		`_(\d{12})_`, // Account ID with underscores
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(profile); len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not determine account ID for profile %s", profile)
}

// GetAWSConfigPath returns the path to AWS config file
func GetAWSConfigPath() string {
	if path := os.Getenv("AWS_CONFIG_FILE"); path != "" {
		return path
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".aws", "config")
}

// GetKubeConfigPath returns the path to Kubernetes config file
func GetKubeConfigPath() string {
	if path := os.Getenv("KUBECONFIG"); path != "" {
		return path
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".kube", "config")
}
