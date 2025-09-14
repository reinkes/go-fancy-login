package utils

import (
	"fmt"
	"os"
	"time"

	"fancy-login/internal/config"
)

// Logger provides logging functionality
type Logger struct {
	verbose bool
}

// NewLogger creates a new logger instance
func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// FancyLog prints debug messages when verbose mode is enabled
func (l *Logger) FancyLog(message string) {
	if l.verbose {
		fmt.Printf("[fancy-login] %s\n", message)
	}
}

// LogInfo prints informational messages
func (l *Logger) LogInfo(message string) {
	fmt.Printf("%s🔹 %s%s\n", config.Cyan, message, config.Reset)
}

// LogSuccess prints success messages (only in verbose mode)
func (l *Logger) LogSuccess(message string) {
	if l.verbose {
		fmt.Printf("%s✅ %s%s\n", config.Green, message, config.Reset)
	}
}

// LogWarning prints warning messages
func (l *Logger) LogWarning(message string) {
	fmt.Printf("%s⚠️ %s%s\n", config.Yellow, message, config.Reset)
}

// LogError prints error messages
func (l *Logger) LogError(message string) {
	fmt.Printf("%s❌ %s%s\n", config.Red, message, config.Reset)
}

// LogCompletion prints completion messages (only in verbose mode)
func (l *Logger) LogCompletion(message string) {
	if l.verbose {
		fmt.Printf("\n%s🎉 %s%s\n", config.Cyan, message, config.Reset)
	}
}

// Die prints error and exits
func (l *Logger) Die(message string) {
	l.LogError(message)
	os.Exit(1)
}

// Spinner represents a loading spinner
type Spinner struct {
	message string
	chars   []rune
	index   int
	running bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		chars:   []rune{'|', '/', '-', '\\'},
		index:   0,
		running: false,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.running = true
	go func() {
		for s.running {
			fmt.Printf("\r%s%s %c %s", config.Cyan, s.message, s.chars[s.index], config.Reset)
			s.index = (s.index + 1) % len(s.chars)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.running = false
	fmt.Printf("\r%60s\r", "") // Clear the line
}
