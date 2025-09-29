package aws

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"fancy-login/internal/config"
	"fancy-login/internal/utils"
)

// AWSManager handles AWS operations
type AWSManager struct {
	config      *config.Config
	logger      *utils.Logger
	fancyConfig *config.FancyConfig
}

// NewAWSManager creates a new AWS manager
func NewAWSManager(cfg *config.Config, logger *utils.Logger, fancyConfig *config.FancyConfig) *AWSManager {
	return &AWSManager{
		config:      cfg,
		logger:      logger,
		fancyConfig: fancyConfig,
	}
}

// SelectAWSProfile allows user to select an AWS profile using fzf
func (aws *AWSManager) SelectAWSProfile() (string, error) {
	displayProfiles, err := aws.getProfilesWithMetadata()
	if err != nil {
		return "", err
	}

	if len(displayProfiles) == 0 {
		aws.logger.Die("No AWS profiles found in ~/.aws/config")
	}

	configuredCount := aws.countConfiguredProfiles(displayProfiles)
	totalCount := aws.countRealProfiles(displayProfiles)

	aws.logger.FancyLog("☁️ AWS Profile Selection")
	aws.logger.FancyLog(fmt.Sprintf("Found %d configured profiles out of %d total AWS profiles",
		configuredCount, totalCount))

	// Create display text for fzf
	var displayTexts []string
	for _, p := range displayProfiles {
		displayTexts = append(displayTexts, p.DisplayText)
	}

	// Use fzf to select profile with proper TTY handling and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "fzf", "--prompt=Select AWS Profile: ")
	cmd.Stdin = strings.NewReader(strings.Join(displayTexts, "\n"))

	// fzf needs full terminal access - redirect both stderr and pass through TTY
	cmd.Stderr = os.Stderr

	// Try to open /dev/tty for fzf to use for input/output
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer tty.Close()
		// Let fzf use the TTY for its interface
		cmd.ExtraFiles = []*os.File{tty}
	}

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("profile selection timed out after 60 seconds")
		}
		return "", fmt.Errorf("profile selection failed: %w", err)
	}

	selectedDisplayText := strings.TrimSpace(string(output))
	if selectedDisplayText == "" {
		aws.logger.Die("No profile selected. Exiting.")
	}

	// Find the actual profile name from the selected display text
	var selectedProfile string
	var isConfigured bool
	for _, p := range displayProfiles {
		// Handle both exact match and trimmed match (fzf may strip leading whitespace)
		if p.DisplayText == selectedDisplayText || strings.TrimSpace(p.DisplayText) == selectedDisplayText {
			selectedProfile = p.Name
			isConfigured = p.IsConfigured
			break
		}
	}

	// Handle separator selection (shouldn't happen but be safe)
	if selectedProfile == "---" || selectedProfile == "" {
		return "", fmt.Errorf("invalid profile selection")
	}

	aws.logger.FancyLog(fmt.Sprintf("Profile selected: %s (configured: %v)", selectedProfile, isConfigured))

	// If profile is not configured, offer to run configuration
	if !isConfigured {
		aws.logger.LogWarning(fmt.Sprintf("Profile '%s' is not configured in fancy-config", selectedProfile))
		fmt.Printf("%sWould you like to configure this profile now? (y/N): %s", config.Cyan, config.Reset)

		// Use /dev/tty for proper terminal input handling
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			aws.logger.LogWarning("Failed to open /dev/tty for input, continuing with unconfigured profile")
		} else {
			defer tty.Close()
			var response string
			if _, err := fmt.Fscanln(tty, &response); err != nil {
				aws.logger.LogWarning("Failed to read user input, continuing with unconfigured profile")
			}

			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				aws.logger.LogInfo("Run 'fancy-login-go --config' to configure profiles")
				return "", fmt.Errorf("profile configuration needed")
			}
		}
		aws.logger.LogWarning("Continuing with unconfigured profile...")
	}

	// Export profile to temp file for shell integration
	if err := aws.exportProfileToTemp(selectedProfile); err != nil {
		aws.logger.LogWarning(fmt.Sprintf("Failed to export profile to temp file: %v", err))
	}

	aws.logger.LogSuccess(fmt.Sprintf("Selected AWS Profile: %s", selectedProfile))
	return selectedProfile, nil
}

// countConfiguredProfiles counts how many profiles are configured
func (aws *AWSManager) countConfiguredProfiles(profiles []ProfileDisplayInfo) int {
	count := 0
	for _, p := range profiles {
		if p.IsConfigured {
			count++
		}
	}
	return count
}

// countRealProfiles counts actual profiles (excludes separators and headers)
func (aws *AWSManager) countRealProfiles(profiles []ProfileDisplayInfo) int {
	count := 0
	for _, p := range profiles {
		if p.Name != "---" && !strings.HasPrefix(p.DisplayText, "===") && !strings.HasPrefix(p.DisplayText, "✓") {
			count++
		}
	}
	return count
}

// HandleAWSLogin checks and handles AWS SSO authentication
func (aws *AWSManager) HandleAWSLogin(profile string, forceLogin bool) error {
	aws.logger.FancyLog(fmt.Sprintf("Checking AWS SSO session for profile %s...", profile))

	if !forceLogin {
		if aws.isSessionValid(profile) {
			aws.logger.LogSuccess(fmt.Sprintf("AWS SSO session is still valid for %s.", profile))
			return nil
		}
	}

	isSSO, err := aws.isSSOMProfile(profile)
	if err != nil {
		return err
	}

	if isSSO {
		return aws.performSSOMLogin(profile)
	}

	aws.logger.LogWarning(fmt.Sprintf("Unable to authenticate with profile %s. This might not be an SSO profile.", profile))

	fmt.Printf("%sDo you want to continue anyway? (y/n): %s", config.Cyan, config.Reset)

	// Use /dev/tty for proper terminal input handling
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		aws.logger.LogError(fmt.Sprintf("Failed to open /dev/tty for input: %v", err))
		return err
	}
	defer tty.Close()

	var response string
	_, err = fmt.Fscanln(tty, &response)
	if err != nil {
		aws.logger.LogError(fmt.Sprintf("Error reading user input: %v", err))
		return err
	}

	if response != "y" {
		aws.logger.Die("User chose to exit due to authentication issues.")
	}

	aws.logger.LogWarning("Continuing with potentially invalid credentials...")
	return nil
}

// HandleECRLogin performs ECR login based on configuration
func (aws *AWSManager) HandleECRLogin(profile string) error {
	if !aws.fancyConfig.ShouldPerformECRLogin(profile) {
		return nil
	}

	aws.logger.FancyLog("ECR login based on configuration...")

	accountID, err := aws.getAccountID(profile)
	if err != nil {
		aws.logger.LogError("Failed to retrieve AWS account ID. Your session may have expired or is not authenticated.")
		return err
	}

	region := aws.fancyConfig.GetECRRegionForProfile(profile)
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			region = aws.config.DefaultRegion
		}
	}

	aws.logger.FancyLog(fmt.Sprintf("Account ID: %s, Region: %s", accountID, region))

	var spinner *utils.Spinner
	if !aws.config.FancyVerbose {
		spinner = utils.NewSpinner("🐳 Logging in to ECR...")
		spinner.Start()
	}

	// Get ECR login password and login to docker
	cmd1 := exec.Command("aws", "ecr", "get-login-password", "--region", region, "--profile", profile)
	cmd2 := exec.Command("docker", "login", "--username", "AWS", "--password-stdin",
		fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, region))

	cmd2.Stdin, _ = cmd1.StdoutPipe()

	if err := cmd1.Start(); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to start ECR login command: %w", err)
	}

	if err := cmd2.Start(); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to start docker login command: %w", err)
	}

	if err := cmd1.Wait(); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("ECR get-login-password failed: %w", err)
	}

	if err := cmd2.Wait(); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		aws.logger.LogError("ECR login failed.")
		return fmt.Errorf("docker login failed: %w", err)
	}

	if spinner != nil {
		spinner.Stop()
	}

	aws.logger.FancyLog("ECR login successful")
	if aws.config.FancyVerbose {
		aws.logger.LogSuccess("Docker: Login Succeeded")
	}

	return nil
}

// GetAccountID retrieves the AWS account ID for the current profile
func (aws *AWSManager) GetAccountID(profile string) (string, error) {
	return aws.getAccountID(profile)
}

// ProfileDisplayInfo holds information for displaying profiles in selection
type ProfileDisplayInfo struct {
	Name         string
	DisplayText  string
	IsConfigured bool
	Metadata     string
}

// getProfilesWithMetadata returns profiles with rich metadata for display
func (aws *AWSManager) getProfilesWithMetadata() ([]ProfileDisplayInfo, error) {
	// Get profiles from AWS config
	awsProfiles, err := aws.getAWSConfigProfiles()
	if err != nil {
		return nil, err
	}

	var displayProfiles []ProfileDisplayInfo

	// Separate profiles by type for better organization
	var k9sProfiles []ProfileDisplayInfo
	var configuredProfiles []ProfileDisplayInfo
	configuredCount := 0

	// First pass: collect all profiles and find the longest name for alignment
	type profileInfo struct {
		ProfileName string
		Config      config.ProfileConfig
		IsK9s       bool
	}
	var allConfiguredProfiles []profileInfo

	for profileName := range aws.fancyConfig.ProfileConfigs {
		// Check if this profile exists in AWS config
		found := false
		for _, awsProfile := range awsProfiles {
			if awsProfile == profileName {
				found = true
				break
			}
		}

		if found {
			profileConfig := aws.fancyConfig.ProfileConfigs[profileName]
			allConfiguredProfiles = append(allConfiguredProfiles, profileInfo{
				ProfileName: profileName,
				Config:      profileConfig,
				IsK9s:       profileConfig.K9sAutoLaunch,
			})
			configuredCount++
		}
	}

	// Calculate the maximum length for alignment
	maxNameLength := 0
	for _, profile := range allConfiguredProfiles {
		// Use the custom name from config if set, otherwise use the profile name
		displayName := profile.ProfileName
		if profile.Config.Name != "" {
			displayName = profile.Config.Name
		}

		var prefixedName string
		if profile.IsK9s {
			prefixedName = fmt.Sprintf("★ %s", displayName)
		} else {
			prefixedName = fmt.Sprintf("  %s", displayName)
		}

		if len(prefixedName) > maxNameLength {
			maxNameLength = len(prefixedName)
		}
	}

	// Second pass: format profiles with proper alignment
	for _, profile := range allConfiguredProfiles {
		metadata := aws.buildProfileMetadata(profile.Config)

		var displayText string
		var prefixedName string

		// Use the custom name from config if set, otherwise use the profile name
		displayName := profile.ProfileName
		if profile.Config.Name != "" {
			displayName = profile.Config.Name
		}

		if profile.IsK9s {
			prefixedName = fmt.Sprintf("★ %s", displayName)
		} else {
			prefixedName = fmt.Sprintf("  %s", displayName)
		}

		// Pad to align the pipe character
		padding := maxNameLength - len(prefixedName)
		if padding < 0 {
			padding = 0
		}

		if metadata != "" {
			displayText = fmt.Sprintf("%s%s %s", prefixedName, strings.Repeat(" ", padding), metadata)
		} else {
			displayText = prefixedName
		}

		profileInfo := ProfileDisplayInfo{
			Name:         profile.ProfileName,
			DisplayText:  displayText,
			IsConfigured: true,
			Metadata:     metadata,
		}

		if profile.IsK9s {
			k9sProfiles = append(k9sProfiles, profileInfo)
		} else {
			configuredProfiles = append(configuredProfiles, profileInfo)
		}
	}

	// Sort profiles by display name within each category
	sort.Slice(k9sProfiles, func(i, j int) bool {
		// Extract display name from DisplayText (remove prefix and metadata)
		nameI := strings.TrimSpace(strings.Split(k9sProfiles[i].DisplayText, "|")[0])
		nameJ := strings.TrimSpace(strings.Split(k9sProfiles[j].DisplayText, "|")[0])
		nameI = strings.TrimPrefix(nameI, "★")
		nameJ = strings.TrimPrefix(nameJ, "★")
		return strings.TrimSpace(nameI) < strings.TrimSpace(nameJ)
	})

	sort.Slice(configuredProfiles, func(i, j int) bool {
		// Extract display name from DisplayText (remove prefix and metadata)
		nameI := strings.TrimSpace(strings.Split(configuredProfiles[i].DisplayText, "|")[0])
		nameJ := strings.TrimSpace(strings.Split(configuredProfiles[j].DisplayText, "|")[0])
		return strings.TrimSpace(nameI) < strings.TrimSpace(nameJ)
	})

	// Add k9s profiles first (most important for daily use)
	if len(k9sProfiles) > 0 {
		displayProfiles = append(displayProfiles, ProfileDisplayInfo{
			Name:         "---",
			DisplayText:  "=== QUICK ACCESS (K9S AUTO-LAUNCH) ===",
			IsConfigured: false,
			Metadata:     "",
		})
		displayProfiles = append(displayProfiles, k9sProfiles...)
	}

	// Add other configured profiles
	if len(configuredProfiles) > 0 {
		if len(k9sProfiles) > 0 {
			displayProfiles = append(displayProfiles, ProfileDisplayInfo{
				Name:         "---",
				DisplayText:  "",
				IsConfigured: false,
				Metadata:     "",
			})
		}
		displayProfiles = append(displayProfiles, ProfileDisplayInfo{
			Name:         "---",
			DisplayText:  "=== OTHER CONFIGURED PROFILES ===",
			IsConfigured: false,
			Metadata:     "",
		})
		displayProfiles = append(displayProfiles, configuredProfiles...)
	}

	// Add separator if we have both configured and unconfigured profiles
	unconfiguredProfiles := []string{}
	for _, awsProfile := range awsProfiles {
		if _, exists := aws.fancyConfig.ProfileConfigs[awsProfile]; !exists {
			unconfiguredProfiles = append(unconfiguredProfiles, awsProfile)
		}
	}

	// Sort unconfigured profiles alphabetically
	sort.Strings(unconfiguredProfiles)

	if len(unconfiguredProfiles) > 0 {
		if configuredCount > 0 {
			displayProfiles = append(displayProfiles, ProfileDisplayInfo{
				Name:         "---",
				DisplayText:  "",
				IsConfigured: false,
				Metadata:     "",
			})
		}
		displayProfiles = append(displayProfiles, ProfileDisplayInfo{
			Name:         "---",
			DisplayText:  "=== UNCONFIGURED PROFILES ===",
			IsConfigured: false,
			Metadata:     "",
		})

		// Add unconfigured profiles
		for _, profileName := range unconfiguredProfiles {
			displayProfiles = append(displayProfiles, ProfileDisplayInfo{
				Name:         profileName,
				DisplayText:  fmt.Sprintf("           %s", profileName),
				IsConfigured: false,
				Metadata:     "",
			})
		}
	} else if configuredCount > 0 {
		// Add helpful hint when all profiles are configured
		displayProfiles = append(displayProfiles, ProfileDisplayInfo{
			Name:         "---",
			DisplayText:  "",
			IsConfigured: false,
			Metadata:     "",
		})
		displayProfiles = append(displayProfiles, ProfileDisplayInfo{
			Name:         "---",
			DisplayText:  "✓ All AWS profiles are configured! Run --config to modify settings.",
			IsConfigured: false,
			Metadata:     "",
		})
	}

	return displayProfiles, nil
}

// getAWSConfigProfiles reads AWS profiles from ~/.aws/config
func (aws *AWSManager) getAWSConfigProfiles() ([]string, error) {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".aws", "config")

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open AWS config: %w", err)
	}
	defer file.Close()

	var profiles []string
	re := regexp.MustCompile(`^\[profile\s+(.+)\]`)
	defaultRe := regexp.MustCompile(`^\[default\]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for named profiles
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			profiles = append(profiles, matches[1])
		}
		// Check for default profile
		if defaultRe.MatchString(line) {
			profiles = append(profiles, "default")
		}
	}

	return profiles, scanner.Err()
}

// buildProfileMetadata creates a display string with profile configuration info
func (aws *AWSManager) buildProfileMetadata(config config.ProfileConfig) string {
	var parts []string

	if config.ECRLogin {
		parts = append(parts, "ECR")
	}

	if config.K8sContext != "" {
		parts = append(parts, fmt.Sprintf("k8s:%s", config.K8sContext))
	}

	if config.K9sAutoLaunch {
		parts = append(parts, "auto-k9s")
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("| %s", strings.Join(parts, " | "))
}

// isSessionValid checks if the AWS session is valid for the given profile
func (aws *AWSManager) isSessionValid(profile string) bool {
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profile, "--query", "Account", "--output", "text")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// isSSOMProfile checks if the profile is an SSO profile
func (aws *AWSManager) isSSOMProfile(profile string) (bool, error) {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".aws", "config")

	file, err := os.Open(configPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inProfile := false
	profilePattern := fmt.Sprintf("[profile %s]", profile)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == profilePattern {
			inProfile = true
			continue
		}

		if strings.HasPrefix(line, "[") && inProfile {
			break
		}

		if inProfile && strings.Contains(line, "sso_") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// performSSOMLogin performs AWS SSO login
func (aws *AWSManager) performSSOMLogin(profile string) error {
	aws.logger.FancyLog(fmt.Sprintf("SSO profile detected. Session expired or not found for %s.", profile))
	aws.logger.FancyLog(fmt.Sprintf("Attempting SSO login for profile %s...", profile))

	var cmd *exec.Cmd
	if !aws.config.FancyVerbose {
		spinner := utils.NewSpinner("🔑 AWS SSO login...")
		spinner.Start()

		cmd = exec.Command("aws", "sso", "login", "--profile", profile)
		cmd.Stdout = nil
		cmd.Stderr = nil

		err := cmd.Run()
		spinner.Stop()

		if err != nil {
			aws.logger.Die(fmt.Sprintf("AWS SSO login failed for %s.", profile))
		}
	} else {
		cmd = exec.Command("aws", "sso", "login", "--profile", profile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			aws.logger.Die(fmt.Sprintf("AWS SSO login failed for %s.", profile))
		}
	}

	// Verify login
	if !aws.isSessionValid(profile) {
		aws.logger.Die(fmt.Sprintf("AWS SSO login verification failed for %s.", profile))
	}

	aws.logger.LogSuccess(fmt.Sprintf("AWS SSO login successful for %s.", profile))
	return nil
}

// getAccountID gets the AWS account ID for a profile
func (aws *AWSManager) getAccountID(profile string) (string, error) {
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profile, "--query", "Account", "--output", "text")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// exportProfileToTemp exports the AWS profile to a temp file for shell integration
func (aws *AWSManager) exportProfileToTemp(profile string) error {
	if runtime.GOOS == "windows" {
		// Create both PowerShell and batch files for Windows
		psContent := fmt.Sprintf("$env:AWS_PROFILE=\"%s\"\n", profile)
		if err := os.WriteFile(aws.config.AWSProfileTemp, []byte(psContent), 0644); err != nil {
			return err
		}

		// Also create a .bat file for Command Prompt users
		batFile := strings.Replace(aws.config.AWSProfileTemp, ".ps1", ".bat", 1)
		batContent := fmt.Sprintf("set AWS_PROFILE=%s\n", profile)
		return os.WriteFile(batFile, []byte(batContent), 0644)
	} else {
		// Unix shell script format
		content := fmt.Sprintf("export AWS_PROFILE=%s\n", profile)
		return os.WriteFile(aws.config.AWSProfileTemp, []byte(content), 0644)
	}
}
