package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fancy-login/internal/aws"
	"fancy-login/internal/config"
	"fancy-login/internal/k8s"
	"fancy-login/internal/utils"
)

var (
	// Build-time variables (set via -ldflags)
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"

	// Command-line flags
	verbose       = flag.Bool("v", false, "Enable verbose output")
	k9sFlag       = flag.Bool("k", false, "Auto-launch k9s without prompting (for DEVENG profiles)")
	forceAWSLogin = flag.Bool("force-aws-login", false, "Force AWS SSO login even if a valid session exists")
	helpFlag      = flag.Bool("h", false, "Show help message")
	versionFlag   = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.BoolVar(verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(k9sFlag, "k9s", false, "Auto-launch k9s without prompting (for DEVENG profiles)")
	flag.BoolVar(helpFlag, "help", false, "Show help message")
	flag.Parse()

	if *versionFlag {
		showVersion()
		return
	}

	if *helpFlag {
		showHelp()
		return
	}

	// Initialize configuration
	cfg := config.NewConfig()
	cfg.FancyVerbose = *verbose
	cfg.ForceAWSLogin = *forceAWSLogin
	cfg.UseK9S = *k9sFlag

	// Set debug mode
	if cfg.FancyDebug {
		fmt.Println("Debug mode enabled")
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.FancyVerbose)

	// Initialize managers
	awsManager := aws.NewAWSManager(cfg, logger)
	k8sManager := k8s.NewK8sManager(cfg, logger)

	// Variables to aggregate results
	var k8sContextResult string
	var ecrResult string
	var ecrAttempted bool
	var accountIDSummary string

	// Select AWS profile
	awsProfile, err := awsManager.SelectAWSProfile()
	if err != nil {
		logger.Die(fmt.Sprintf("Failed to select AWS profile: %v", err))
	}

	// Set AWS_PROFILE environment variable for this process
	os.Setenv("AWS_PROFILE", awsProfile)

	// Handle AWS SSO login
	if err := awsManager.HandleAWSLogin(awsProfile, cfg.ForceAWSLogin); err != nil {
		logger.Die(fmt.Sprintf("AWS login failed: %v", err))
	}

	// Select Kubernetes context and get summary string
	k8sContextResult, err = k8sManager.SelectKubernetesContext(awsProfile)
	if err != nil {
		logger.LogWarning(fmt.Sprintf("Kubernetes context selection failed: %v", err))
		k8sContextResult = fmt.Sprintf("%s🌱 Kubernetes Context:%s (failed to select)", config.Green, config.Reset)
	}

	// Always get AWS account ID for summary
	if accountID, err := awsManager.GetAccountID(awsProfile); err == nil {
		accountIDSummary = accountID
	}

	// Handle ECR login only for _DEV_ profiles
	if err := awsManager.HandleECRLogin(awsProfile); err != nil {
		ecrResult = fmt.Sprintf("%s🐳 ECR login: failed%s", config.Red, config.Reset)
		ecrAttempted = true
		logger.FancyLog(fmt.Sprintf("ECR login failed: %v", err))
	} else if awsProfile != "" && containsDev(awsProfile) {
		ecrResult = fmt.Sprintf("%s🐳 ECR login: successful%s", config.Green, config.Reset)
		ecrAttempted = true
	}

	// Show summary before k9s prompt (unless verbose)
	if !cfg.FancyVerbose {
		fmt.Println()
		fmt.Printf("%s🦄  %sFancy Login Summary%s\n", config.Yellow, config.Bold, config.Reset)
		fmt.Printf("%s───────────────────────────────────────────────%s\n", config.Yellow, config.Reset)
		fmt.Printf("%s🔑 AWS Profile:%s %s%s%s\n", config.Yellow, config.Reset, config.Bold, awsProfile, config.Reset)
		if k8sContextResult != "" {
			fmt.Println(k8sContextResult)
		}
		if ecrAttempted {
			fmt.Println(ecrResult)
		}
		if accountIDSummary != "" {
			fmt.Printf("%s☁️  AWS Account ID:%s %s%s%s\n", config.Cyan, config.Reset, config.Bold, accountIDSummary, config.Reset)
		}
		fmt.Printf("%s───────────────────────────────────────────────%s\n", config.Yellow, config.Reset)
		fmt.Println()
	}

	// Handle k9s launch
	if err := k8sManager.HandleK9sLaunch(awsProfile); err != nil {
		logger.LogError(fmt.Sprintf("Failed to launch k9s: %v", err))
	}

	logger.LogCompletion("Script execution completed.")
}

func showHelp() {
	fmt.Printf(`Usage: %s [OPTIONS]

OPTIONS:
  -k, --k9s           Auto-launch k9s without prompting (for DEVENG profiles)
  -v, --verbose       Enable verbose output
  --force-aws-login   Force AWS SSO login even if a valid session exists
  -h, --help          Show this help message
  --version           Show version information

Description:
  Interactive tool for AWS SSO login and Kubernetes context selection.
  Automatically handles ECR login for development profiles and provides
  namespace-aware k9s integration.
  
  When using -k flag with DEVENG profiles, k9s will launch automatically
  in the derived namespace without user confirmation.
  Use -v to enable verbose output for debugging.
  Use --force-aws-login to always perform AWS SSO login, even if a valid session exists.

Version: %s
Build Time: %s
Git Commit: %s
`, os.Args[0], version, buildTime, gitCommit)
}

func showVersion() {
	fmt.Printf("fancy-login-go version %s\n", version)
	fmt.Printf("Build time: %s\n", buildTime)
	fmt.Printf("Git commit: %s\n", gitCommit)
}

func containsDev(profile string) bool {
	return strings.Contains(profile, "_DEV_")
}
