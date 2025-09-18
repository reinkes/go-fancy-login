package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Test verbose logger
	verboseLogger := NewLogger(true)
	if verboseLogger == nil {
		t.Fatal("NewLogger(true) returned nil")
	}
	if !verboseLogger.verbose {
		t.Error("Logger created with verbose=true should have verbose=true")
	}

	// Test non-verbose logger
	quietLogger := NewLogger(false)
	if quietLogger == nil {
		t.Fatal("NewLogger(false) returned nil")
	}
	if quietLogger.verbose {
		t.Error("Logger created with verbose=false should have verbose=false")
	}
}

// Helper function to capture stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return ""
	}
	return buf.String()
}

func TestFancyLogVerbose(t *testing.T) {
	logger := NewLogger(true)
	testMessage := "Test debug message"

	output := captureOutput(func() {
		logger.FancyLog(testMessage)
	})

	expectedPrefix := "[fancy-login]"
	if !strings.Contains(output, expectedPrefix) {
		t.Errorf("FancyLog output should contain '%s', got: %s", expectedPrefix, output)
	}

	if !strings.Contains(output, testMessage) {
		t.Errorf("FancyLog output should contain test message '%s', got: %s", testMessage, output)
	}
}

func TestFancyLogQuiet(t *testing.T) {
	logger := NewLogger(false)
	testMessage := "Test debug message"

	output := captureOutput(func() {
		logger.FancyLog(testMessage)
	})

	if output != "" {
		t.Errorf("FancyLog in quiet mode should produce no output, got: %s", output)
	}
}

func TestLogInfo(t *testing.T) {
	logger := NewLogger(false) // verbose setting shouldn't matter for LogInfo
	testMessage := "Test info message"

	output := captureOutput(func() {
		logger.LogInfo(testMessage)
	})

	if !strings.Contains(output, testMessage) {
		t.Errorf("LogInfo output should contain test message '%s', got: %s", testMessage, output)
	}

	// Check for info emoji
	if !strings.Contains(output, "🔹") {
		t.Errorf("LogInfo output should contain info emoji 🔹, got: %s", output)
	}
}

func TestLogSuccessVerbose(t *testing.T) {
	logger := NewLogger(true)
	testMessage := "Test success message"

	output := captureOutput(func() {
		logger.LogSuccess(testMessage)
	})

	if !strings.Contains(output, testMessage) {
		t.Errorf("LogSuccess output should contain test message '%s', got: %s", testMessage, output)
	}

	// Check for success emoji
	if !strings.Contains(output, "✅") {
		t.Errorf("LogSuccess output should contain success emoji ✅, got: %s", output)
	}
}

func TestLogSuccessQuiet(t *testing.T) {
	logger := NewLogger(false)
	testMessage := "Test success message"

	output := captureOutput(func() {
		logger.LogSuccess(testMessage)
	})

	if output != "" {
		t.Errorf("LogSuccess in quiet mode should produce no output, got: %s", output)
	}
}

func TestLogWarning(t *testing.T) {
	logger := NewLogger(false) // verbose setting shouldn't matter for LogWarning
	testMessage := "Test warning message"

	output := captureOutput(func() {
		logger.LogWarning(testMessage)
	})

	if !strings.Contains(output, testMessage) {
		t.Errorf("LogWarning output should contain test message '%s', got: %s", testMessage, output)
	}

	// Check for warning emoji
	if !strings.Contains(output, "⚠️") {
		t.Errorf("LogWarning output should contain warning emoji ⚠️, got: %s", output)
	}
}

func TestLogError(t *testing.T) {
	logger := NewLogger(false) // verbose setting shouldn't matter for LogError
	testMessage := "Test error message"

	output := captureOutput(func() {
		logger.LogError(testMessage)
	})

	if !strings.Contains(output, testMessage) {
		t.Errorf("LogError output should contain test message '%s', got: %s", testMessage, output)
	}

	// Check for error emoji
	if !strings.Contains(output, "❌") {
		t.Errorf("LogError output should contain error emoji ❌, got: %s", output)
	}
}

func TestDie(t *testing.T) {
	// We can't easily test Die() since it calls os.Exit()
	// But we can test that the method exists and is accessible
	logger := NewLogger(false)

	// Test that Die method exists (this will compile if it exists)
	dieFunc := logger.Die
	// Function pointers are never nil, so we just verify it exists
	_ = dieFunc
}

func TestLoggerVerbosityToggle(t *testing.T) {
	// Test that the same logger behaves differently with different verbosity settings
	testMessage := "Test message"

	// Test verbose behavior
	verboseLogger := NewLogger(true)
	verboseOutput := captureOutput(func() {
		verboseLogger.FancyLog(testMessage)
		verboseLogger.LogSuccess(testMessage)
	})

	// Test quiet behavior
	quietLogger := NewLogger(false)
	quietOutput := captureOutput(func() {
		quietLogger.FancyLog(testMessage)
		quietLogger.LogSuccess(testMessage)
	})

	// Verbose should produce output
	if len(verboseOutput) == 0 {
		t.Error("Verbose logger should produce output")
	}

	// Quiet should produce less output
	if len(quietOutput) >= len(verboseOutput) {
		t.Error("Quiet logger should produce less output than verbose logger")
	}
}

func TestMultipleLogCalls(t *testing.T) {
	logger := NewLogger(true)

	output := captureOutput(func() {
		logger.LogInfo("Info 1")
		logger.LogWarning("Warning 1")
		logger.LogError("Error 1")
		logger.FancyLog("Debug 1")
		logger.LogSuccess("Success 1")
	})

	// Check that all messages appear in output
	expectedMessages := []string{"Info 1", "Warning 1", "Error 1", "Debug 1", "Success 1"}
	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Output should contain '%s', got: %s", msg, output)
		}
	}

	// Check that different emojis appear
	expectedEmojis := []string{"🔹", "⚠️", "❌", "✅"}
	for _, emoji := range expectedEmojis {
		if !strings.Contains(output, emoji) {
			t.Errorf("Output should contain emoji '%s', got: %s", emoji, output)
		}
	}
}

// Benchmark logger operations
func BenchmarkNewLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewLogger(true)
	}
}

func BenchmarkLogInfo(b *testing.B) {
	logger := NewLogger(false)
	message := "Benchmark test message"

	// Redirect output to discard for benchmarking
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogInfo(message)
	}
}

func BenchmarkFancyLogVerbose(b *testing.B) {
	logger := NewLogger(true)
	message := "Benchmark test message"

	// Redirect output to discard for benchmarking
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.FancyLog(message)
	}
}

func BenchmarkFancyLogQuiet(b *testing.B) {
	logger := NewLogger(false)
	message := "Benchmark test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.FancyLog(message)
	}
}
