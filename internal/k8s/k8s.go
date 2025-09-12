package k8s

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"fancy-login/internal/config"
	"fancy-login/internal/utils"
)

// K8sManager handles Kubernetes operations
type K8sManager struct {
	config *config.Config
	logger *utils.Logger
}

// NewK8sManager creates a new Kubernetes manager
func NewK8sManager(cfg *config.Config, logger *utils.Logger) *K8sManager {
	return &K8sManager{
		config: cfg,
		logger: logger,
	}
}

// SelectKubernetesContext selects and switches Kubernetes context
func (k8s *K8sManager) SelectKubernetesContext(awsProfile string) (string, error) {
	k8s.logger.FancyLog("Entered select_kubernetes_context")
	
	if k8s.shouldSkipK8sContext(awsProfile) {
		return k8s.handleDEVProfile(awsProfile)
	}
	
	// Load context mappings
	mappings, err := config.LoadContextMappings()
	if err != nil {
		k8s.logger.FancyLog(fmt.Sprintf("Failed to load context mappings: %v", err))
		mappings = []config.ContextMapping{}
	}
	
	// Check for mapped context
	for _, mapping := range mappings {
		if config.MatchesPattern(awsProfile, mapping.Pattern) {
			k8s.logger.FancyLog(fmt.Sprintf("Matched pattern: %s, using context: %s", mapping.Pattern, mapping.Context))
			
			if err := k8s.switchK8sContext(mapping.Context); err != nil {
				k8s.logger.LogWarning(fmt.Sprintf("Failed to switch to context %s: %v", mapping.Context, err))
			}
			
			return k8s.formatContextSummary(mapping.Context, awsProfile), nil
		}
	}
	
	// No mapping found, use fzf to select
	context, err := k8s.selectContextWithFzf()
	if err != nil {
		k8s.logger.FancyLog("No context selected or error occurred")
		// Return current context or fallback
		return k8s.getCurrentContextSummary(awsProfile)
	}
	
	if err := k8s.switchK8sContext(context); err != nil {
		k8s.logger.LogWarning(fmt.Sprintf("Failed to switch to context %s: %v", context, err))
	}
	
	return k8s.formatContextSummary(context, awsProfile), nil
}

// HandleK9sLaunch handles launching k9s for DEVENG profiles
func (k8s *K8sManager) HandleK9sLaunch(awsProfile string) error {
	if !strings.HasSuffix(awsProfile, "DEVENG") {
		return nil
	}
	
	if k8s.config.UseK9S {
		return k8s.launchK9sWithNamespace(awsProfile)
	}
	
	fmt.Printf("\n%sDo you want to open k9s in the derived namespace? (y/n): %s", config.Cyan, config.Reset)
	var response string
	fmt.Scanln(&response)
	
	if response == "y" {
		return k8s.launchK9sWithNamespace(awsProfile)
	}
	
	return nil
}

// shouldSkipK8sContext determines if context selection should be skipped for DEV profiles
func (k8s *K8sManager) shouldSkipK8sContext(awsProfile string) bool {
	skip := strings.Contains(awsProfile, "_DEV_")
	k8s.logger.FancyLog(fmt.Sprintf("should_skip_k8s_context: %s matches *_DEV_* = %t", awsProfile, skip))
	return skip
}

// handleDEVProfile handles context selection for DEV profiles
func (k8s *K8sManager) handleDEVProfile(awsProfile string) (string, error) {
	mappings, err := config.LoadContextMappings()
	if err != nil {
		k8s.logger.FancyLog(fmt.Sprintf("Failed to load context mappings: %v", err))
		mappings = []config.ContextMapping{}
	}
	
	var mappedContext string
	for _, mapping := range mappings {
		if config.MatchesPattern(awsProfile, mapping.Pattern) {
			mappedContext = mapping.Context
			break
		}
	}
	
	// Load namespace mappings
	namespaceMappings, err := config.LoadNamespaceMappings()
	if err != nil {
		k8s.logger.FancyLog(fmt.Sprintf("Failed to load namespace mappings: %v", err))
		namespaceMappings = make(map[string]string)
	}
	
	// Try to get namespace from profile
	namespace, err := config.GetNamespaceFromProfile(awsProfile, namespaceMappings)
	if err == nil {
		k8s.setITerm2Namespace(namespace)
		if mappedContext != "" {
			return fmt.Sprintf("%s🌱 Kubernetes Context:%s %s%s%s %s(ns: %s)%s",
				config.Green, config.Reset, config.Bold, mappedContext, config.Reset,
				config.Cyan, namespace, config.Reset), nil
		}
		return fmt.Sprintf("%s🌱 Kubernetes Context:%s %s(ns: %s)%s",
			config.Green, config.Reset, config.Cyan, namespace, config.Reset), nil
	}
	
	if mappedContext != "" {
		return fmt.Sprintf("%s🌱 Kubernetes Context:%s %s%s%s",
			config.Green, config.Reset, config.Bold, mappedContext, config.Reset), nil
	}
	
	return fmt.Sprintf("%s🌱 Kubernetes Context:%s (skipped for DEV profile)", 
		config.Green, config.Reset), nil
}

// selectContextWithFzf uses fzf to select a Kubernetes context
func (k8s *K8sManager) selectContextWithFzf() (string, error) {
	k8s.logger.FancyLog("Selecting Kubernetes Context...")
	
	// Get available contexts
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get contexts: %w", err)
	}
	
	contexts := strings.TrimSpace(string(output))
	if contexts == "" {
		return "", fmt.Errorf("no contexts available")
	}
	
	// Use fzf to select
	fzfCmd := exec.Command("fzf", "--prompt=Select Kubernetes Context: ")
	fzfCmd.Stdin = strings.NewReader(contexts)
	
	result, err := fzfCmd.Output()
	if err != nil {
		return "", err
	}
	
	context := strings.TrimSpace(string(result))
	k8s.logger.FancyLog(fmt.Sprintf("K8s context selected: %s", context))
	
	return context, nil
}

// switchK8sContext switches to the specified Kubernetes context
func (k8s *K8sManager) switchK8sContext(context string) error {
	if k8s.config.FancyVerbose {
		k8s.logger.LogInfo(fmt.Sprintf("Switching to Kubernetes context: %s", context))
		cmd := exec.Command("kubectl", "config", "use-context", context)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	
	cmd := exec.Command("kubectl", "config", "use-context", context)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// getCurrentContextSummary returns the current context summary
func (k8s *K8sManager) getCurrentContextSummary(awsProfile string) (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%s🌱 Kubernetes Context:%s (none selected)", 
			config.Green, config.Reset), nil
	}
	
	currentContext := strings.TrimSpace(string(output))
	return k8s.formatContextSummary(currentContext, awsProfile), nil
}

// formatContextSummary formats the context summary with namespace if available
func (k8s *K8sManager) formatContextSummary(context, awsProfile string) string {
	namespaceMappings, err := config.LoadNamespaceMappings()
	if err != nil {
		namespaceMappings = make(map[string]string)
	}
	
	namespace, err := config.GetNamespaceFromProfile(awsProfile, namespaceMappings)
	if err == nil {
		k8s.setITerm2Namespace(namespace)
		return fmt.Sprintf("%s🌱 Kubernetes Context:%s %s%s%s %s(ns: %s)%s",
			config.Green, config.Reset, config.Bold, context, config.Reset,
			config.Cyan, namespace, config.Reset)
	}
	
	return fmt.Sprintf("%s🌱 Kubernetes Context:%s %s%s%s",
		config.Green, config.Reset, config.Bold, context, config.Reset)
}

// setITerm2Namespace sets the terminal tab title and badge (cross-platform)
func (k8s *K8sManager) setITerm2Namespace(namespace string) {
	if namespace == "" {
		return
	}
	
	switch runtime.GOOS {
	case "darwin":
		// macOS iTerm2
		if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
			// Set tab title
			fmt.Printf("\033]1;ns:%s\007", namespace)
			
			// Set badge
			badge := fmt.Sprintf("🟢 ns:%s", namespace)
			encoded := base64.StdEncoding.EncodeToString([]byte(badge))
			fmt.Printf("\033]1337;SetBadgeFormat=%s\a", encoded)
		}
	case "windows":
		// Windows Terminal
		if os.Getenv("WT_SESSION") != "" {
			// Set tab title for Windows Terminal
			fmt.Printf("\033]0;ns:%s\007", namespace)
		}
	default:
		// Linux terminals (most support standard title escape sequence)
		fmt.Printf("\033]0;ns:%s\007", namespace)
	}
}

// launchK9sWithNamespace launches k9s with the derived namespace
func (k8s *K8sManager) launchK9sWithNamespace(awsProfile string) error {
	namespaceMappings, err := config.LoadNamespaceMappings()
	if err != nil {
		return fmt.Errorf("failed to load namespace mappings: %w", err)
	}
	
	namespace, err := config.GetNamespaceFromProfile(awsProfile, namespaceMappings)
	if err != nil {
		k8s.logger.LogError(fmt.Sprintf("Unable to derive namespace from profile: %s", awsProfile))
		return err
	}
	
	k8s.logger.FancyLog(fmt.Sprintf("Launching k9s in %s.", namespace))
	
	cmd := exec.Command("k9s", "-n", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Inherit current environment and set AWS_PROFILE
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_PROFILE=%s", awsProfile))
	
	return cmd.Run()
}