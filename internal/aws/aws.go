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
	aws.logger.FancyLog("Select an AWS Profile:")

	profiles, err := aws.getAWSProfiles()
	if err != nil {
		return "", err
	}

	if len(profiles) == 0 {
		aws.logger.Die("No AWS profiles found in ~/.aws/config")
	}

	aws.logger.FancyLog(fmt.Sprintf("Profiles found: %v", profiles))

	// Use fzf to select profile with proper TTY handling and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "fzf", "--prompt=Select AWS Profile: ")
	cmd.Stdin = strings.NewReader(strings.Join(profiles, "\n"))

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

	profile := strings.TrimSpace(string(output))
	if profile == "" {
		aws.logger.Die("No profile selected. Exiting.")
	}

	// Remove "profile " prefix if present
	profile = strings.TrimPrefix(profile, "profile ")

	aws.logger.FancyLog(fmt.Sprintf("Profile selected: %s", profile))

	// Export profile to temp file for shell integration
	if err := aws.exportProfileToTemp(profile); err != nil {
		aws.logger.LogWarning(fmt.Sprintf("Failed to export profile to temp file: %v", err))
	}

	aws.logger.LogSuccess(fmt.Sprintf("Selected AWS Profile: %s", profile))
	return profile, nil
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
	var response string
	_, err = fmt.Scanln(&response)
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

// getAWSProfiles reads AWS profiles from ~/.aws/config
func (aws *AWSManager) getAWSProfiles() ([]string, error) {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".aws", "config")

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open AWS config: %w", err)
	}
	defer file.Close()

	var profiles []string
	re := regexp.MustCompile(`^\[profile\s+(.+)\]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			profiles = append(profiles, matches[1])
		}
	}

	return profiles, scanner.Err()
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
