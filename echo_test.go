package echo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	cfg := Config{
		LogLevel: "INFO",
	}

	logger := New(cfg)
	if logger == nil {
		t.Fatal("Logger is nil")
	}

	if logger.logLevel != 3 {
		t.Errorf("Expected logLevel 3, got %d", logger.logLevel)
	}

	logger.Close()
}

func TestLogLevels(t *testing.T) {
	// Create temp dir for output
	tempDir, err := os.MkdirTemp("", "echo_logger_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "output.log")

	cfg := Config{
		LogLevel:        "INFO",
		OutputPath:      outputPath,
		ErrorOutputPath: filepath.Join(tempDir, "error.log"),
	}

	logger := New(cfg)
	defer logger.Close()

	// Test INFO level
	logger.Info("Test info message")
	logger.Error("Test error message")
	logger.Debug("Test debug message") // Should not log

	// Read output file
	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	outputStr := string(outputContent)

	if !strings.Contains(outputStr, "Test info message") {
		t.Error("INFO message not found in output")
	}

	if !strings.Contains(outputStr, "Test error message") {
		t.Error("ERROR message not found in output")
	}

	if strings.Contains(outputStr, "Test debug message") {
		t.Error("DEBUG message should not be logged at INFO level")
	}

	// Check error file
	errorContent, err := os.ReadFile(cfg.ErrorOutputPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(errorContent), "Test error message") {
		t.Error("ERROR message not found in error file")
	}
}

func TestFileCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "echo_logger_file_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := Config{
		LogLevel:        "DEBUG",
		OutputPath:      filepath.Join(tempDir, "output.log"),
		FatalOutputPath: filepath.Join(tempDir, "fatal.log"),
		ErrorOutputPath: filepath.Join(tempDir, "error.log"),
		WarnOutputPath:  filepath.Join(tempDir, "warn.log"),
		InfoOutputPath:  filepath.Join(tempDir, "info.log"),
		DebugOutputPath: filepath.Join(tempDir, "debug.log"),
		TraceOutputPath: filepath.Join(tempDir, "trace.log"),
	}

	logger := New(cfg)
	defer logger.Close()

	logger.Info("Info message")
	logger.Debug("Debug message")
	logger.Trace("Trace message")

	// Check if files exist
	files := []string{cfg.OutputPath, cfg.InfoOutputPath, cfg.DebugOutputPath, cfg.TraceOutputPath}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("File %s was not created", file)
		}
	}
}

func TestClose(t *testing.T) {
	cfg := Config{
		LogLevel:   "INFO",
		OutputPath: "/tmp/test_close.log", // Use a path that will create file
	}

	logger := New(cfg)

	// Log something to create file
	logger.Info("Test")

	// Close
	err := logger.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Try to close again (should not error)
	err = logger.Close()
	if err != nil {
		t.Errorf("Second close returned error: %v", err)
	}
}

func TestEmbeddedLogger(t *testing.T) {
	// Since we can't easily inject writer, test that embedded Logger exists
	cfg := Config{
		LogLevel: "INFO",
	}

	logger := New(cfg)
	defer logger.Close()

	if logger.Logger == nil {
		t.Error("Embedded Logger is nil")
	}

	// Test that we can use it as *log.Logger
	// But since it's embedded, logger.Printf should work
	logger.Printf("Test printf")

	// To check output, we need to capture it, but since it's to stdout, hard.
	// Just check that no panic
}

func TestLogLevelSwitch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "echo_logger_level_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "output.log")

	tests := []struct {
		level         string
		expectedLevel int
	}{
		{"FATAL", 0},
		{"ERROR", 1},
		{"WARN", 2},
		{"INFO", 3},
		{"DEBUG", 4},
		{"TRACE", 5},
		{"UNKNOWN", 3}, // default to INFO
	}

	for _, test := range tests {
		cfg := Config{
			LogLevel:   test.level,
			OutputPath: outputPath,
		}

		logger := New(cfg)
		if logger.logLevel != test.expectedLevel {
			t.Errorf("For level %s, expected %d, got %d", test.level, test.expectedLevel, logger.logLevel)
		}
		logger.Close()
	}
}
