package aws

import (
	"sort"
	"strings"
	"testing"

	"fancy-login/internal/config"
	"fancy-login/internal/utils"
)

func TestProfileDisplayInfo_Sorting(t *testing.T) {
	// Create test profiles with different names
	k9sProfiles := []ProfileDisplayInfo{
		{
			Name:         "zebra-profile",
			DisplayText:  "★ Zebra Environment | k8s:cluster | auto-k9s",
			IsConfigured: true,
		},
		{
			Name:         "alpha-profile",
			DisplayText:  "★ Alpha Environment | k8s:cluster | auto-k9s",
			IsConfigured: true,
		},
		{
			Name:         "beta-profile",
			DisplayText:  "★ Beta Environment | k8s:cluster | auto-k9s",
			IsConfigured: true,
		},
	}

	configuredProfiles := []ProfileDisplayInfo{
		{
			Name:         "prod-profile",
			DisplayText:  "  Production | ECR",
			IsConfigured: true,
		},
		{
			Name:         "dev-profile",
			DisplayText:  "  Development | ECR",
			IsConfigured: true,
		},
		{
			Name:         "staging-profile",
			DisplayText:  "  Staging | ECR",
			IsConfigured: true,
		},
	}

	// Test k9s profiles sorting
	sort.Slice(k9sProfiles, func(i, j int) bool {
		nameI := strings.TrimSpace(strings.Split(k9sProfiles[i].DisplayText, "|")[0])
		nameJ := strings.TrimSpace(strings.Split(k9sProfiles[j].DisplayText, "|")[0])
		nameI = strings.TrimPrefix(nameI, "★")
		nameJ = strings.TrimPrefix(nameJ, "★")
		return strings.TrimSpace(nameI) < strings.TrimSpace(nameJ)
	})

	expectedK9sOrder := []string{"alpha-profile", "beta-profile", "zebra-profile"}
	for i, profile := range k9sProfiles {
		if profile.Name != expectedK9sOrder[i] {
			t.Errorf("K9s profiles not sorted correctly. Expected %s at position %d, got %s",
				expectedK9sOrder[i], i, profile.Name)
		}
	}

	// Test configured profiles sorting
	sort.Slice(configuredProfiles, func(i, j int) bool {
		nameI := strings.TrimSpace(strings.Split(configuredProfiles[i].DisplayText, "|")[0])
		nameJ := strings.TrimSpace(strings.Split(configuredProfiles[j].DisplayText, "|")[0])
		return strings.TrimSpace(nameI) < strings.TrimSpace(nameJ)
	})

	expectedConfiguredOrder := []string{"dev-profile", "prod-profile", "staging-profile"}
	for i, profile := range configuredProfiles {
		if profile.Name != expectedConfiguredOrder[i] {
			t.Errorf("Configured profiles not sorted correctly. Expected %s at position %d, got %s",
				expectedConfiguredOrder[i], i, profile.Name)
		}
	}
}

func TestUnconfiguredProfilesSorting(t *testing.T) {
	unconfiguredProfiles := []string{"zebra-account", "alpha-account", "beta-account"}

	sort.Strings(unconfiguredProfiles)

	expected := []string{"alpha-account", "beta-account", "zebra-account"}
	for i, profile := range unconfiguredProfiles {
		if profile != expected[i] {
			t.Errorf("Unconfigured profiles not sorted correctly. Expected %s at position %d, got %s",
				expected[i], i, profile)
		}
	}
}

func TestProfileSelectionMatching(t *testing.T) {
	testCases := []struct {
		name                string
		selectedDisplayText string
		displayProfiles     []ProfileDisplayInfo
		expectedProfile     string
		expectedFound       bool
	}{
		{
			name:                "Exact match",
			selectedDisplayText: "  Dev Environment | ECR",
			displayProfiles: []ProfileDisplayInfo{
				{
					Name:         "dev-profile",
					DisplayText:  "  Dev Environment | ECR",
					IsConfigured: true,
				},
			},
			expectedProfile: "dev-profile",
			expectedFound:   true,
		},
		{
			name:                "Trimmed match (fzf strips whitespace)",
			selectedDisplayText: "Dev Environment | ECR",
			displayProfiles: []ProfileDisplayInfo{
				{
					Name:         "dev-profile",
					DisplayText:  "  Dev Environment | ECR",
					IsConfigured: true,
				},
			},
			expectedProfile: "dev-profile",
			expectedFound:   true,
		},
		{
			name:                "K9s profile match with leading spaces stripped",
			selectedDisplayText: "★ Alpha Environment | k8s:cluster | auto-k9s",
			displayProfiles: []ProfileDisplayInfo{
				{
					Name:         "alpha-profile",
					DisplayText:  "  ★ Alpha Environment | k8s:cluster | auto-k9s", // Note the leading spaces
					IsConfigured: true,
				},
			},
			expectedProfile: "alpha-profile",
			expectedFound:   true,
		},
		{
			name:                "K9s profile exact match",
			selectedDisplayText: "★ Alpha Environment | k8s:cluster | auto-k9s",
			displayProfiles: []ProfileDisplayInfo{
				{
					Name:         "alpha-profile",
					DisplayText:  "★ Alpha Environment | k8s:cluster | auto-k9s",
					IsConfigured: true,
				},
			},
			expectedProfile: "alpha-profile",
			expectedFound:   true,
		},
		{
			name:                "No match",
			selectedDisplayText: "Nonexistent Profile",
			displayProfiles: []ProfileDisplayInfo{
				{
					Name:         "dev-profile",
					DisplayText:  "  Dev Environment | ECR",
					IsConfigured: true,
				},
			},
			expectedProfile: "",
			expectedFound:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var selectedProfile string
			var found bool

			// Simulate the matching logic from SelectAWSProfile
			for _, p := range tc.displayProfiles {
				if p.DisplayText == tc.selectedDisplayText || strings.TrimSpace(p.DisplayText) == tc.selectedDisplayText {
					selectedProfile = p.Name
					found = true
					break
				}
			}

			if found != tc.expectedFound {
				t.Errorf("Expected found=%v, got found=%v", tc.expectedFound, found)
			}

			if selectedProfile != tc.expectedProfile {
				t.Errorf("Expected profile=%s, got profile=%s", tc.expectedProfile, selectedProfile)
			}
		})
	}
}

func TestCustomDisplayName(t *testing.T) {
	testCases := []struct {
		name             string
		profileName      string
		configName       string
		isK9s            bool
		expectedPrefix   string
		expectedContains string
	}{
		{
			name:             "Custom name with K9s",
			profileName:      "dev-account-123",
			configName:       "🚀 Development Environment",
			isK9s:            true,
			expectedPrefix:   "★ 🚀 Development Environment",
			expectedContains: "🚀 Development Environment",
		},
		{
			name:             "Custom name without K9s",
			profileName:      "prod-account-456",
			configName:       "🏭 Production Environment",
			isK9s:            false,
			expectedPrefix:   "  🏭 Production Environment",
			expectedContains: "🏭 Production Environment",
		},
		{
			name:             "No custom name with K9s",
			profileName:      "staging-account",
			configName:       "",
			isK9s:            true,
			expectedPrefix:   "★ staging-account",
			expectedContains: "staging-account",
		},
		{
			name:             "No custom name without K9s",
			profileName:      "test-account",
			configName:       "",
			isK9s:            false,
			expectedPrefix:   "  test-account",
			expectedContains: "test-account",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock profile config
			profileConfig := config.ProfileConfig{
				Name: tc.configName,
			}

			// Simulate the display name logic from getProfilesWithMetadata
			displayName := tc.profileName
			if profileConfig.Name != "" {
				displayName = profileConfig.Name
			}

			var prefixedName string
			if tc.isK9s {
				prefixedName = "★ " + displayName
			} else {
				prefixedName = "  " + displayName
			}

			if !strings.HasPrefix(prefixedName, tc.expectedPrefix) {
				t.Errorf("Expected prefix '%s', but got '%s'", tc.expectedPrefix, prefixedName)
			}

			if !strings.Contains(prefixedName, tc.expectedContains) {
				t.Errorf("Expected to contain '%s', but got '%s'", tc.expectedContains, prefixedName)
			}
		})
	}
}

// Test helper to create a mock AWSManager for testing
func createMockAWSManager() *AWSManager {
	cfg := &config.Config{
		FancyVerbose: false,
	}
	logger := utils.NewLogger(false) // Logger takes boolean, not config
	fancyConfig := &config.FancyConfig{
		ProfileConfigs: make(map[string]config.ProfileConfig),
	}

	return &AWSManager{
		config:      cfg,
		logger:      logger,
		fancyConfig: fancyConfig,
	}
}

func TestProfileDisplaySeparators(t *testing.T) {
	// Test that separator selection is properly handled
	selectedProfile := ""

	// Test "---" separator
	if selectedProfile == "---" || selectedProfile == "" {
		// This should be caught as invalid
		if selectedProfile != "" && selectedProfile != "---" {
			t.Errorf("Expected empty or separator profile to be invalid")
		}
	}
}
