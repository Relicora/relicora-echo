package echo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupLoggerWithFiles(t *testing.T, logLevel string) (*Logger, string, Config) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "echo_logger_test")
	if err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		LogLevel:        logLevel,
		OutputPath:      filepath.Join(tempDir, "output.log"),
		FatalOutputPath: filepath.Join(tempDir, "fatal.log"),
		ErrorOutputPath: filepath.Join(tempDir, "error.log"),
		WarnOutputPath:  filepath.Join(tempDir, "warn.log"),
		InfoOutputPath:  filepath.Join(tempDir, "info.log"),
		DebugOutputPath: filepath.Join(tempDir, "debug.log"),
		TraceOutputPath: filepath.Join(tempDir, "trace.log"),
	}

	logger := New(cfg)
	return logger, tempDir, cfg
}

func readFileContents(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestNewLogger(t *testing.T) {
	logger, tempDir, _ := setupLoggerWithFiles(t, "INFO")
	defer os.RemoveAll(tempDir)
	defer logger.Close()

	if logger == nil {
		t.Fatal("Logger is nil")
	}
	if logger.logLevel != 3 {
		t.Fatalf("Expected logLevel 3, got %d", logger.logLevel)
	}
}

func TestAllMethodsWithDifferentLevelsAndFiles(t *testing.T) {
	methodCases := []struct {
		name      string
		call      func(*Logger, string)
		minLevel  int
		outputKey string
	}{
		{"Print", func(l *Logger, msg string) { l.Print(msg) }, 3, "output"},
		{"Printf", func(l *Logger, msg string) { l.Printf("%s", msg) }, 3, "output"},
		{"Println", func(l *Logger, msg string) { l.Println(msg) }, 3, "output"},

		{"Error", func(l *Logger, msg string) { l.Error(msg) }, 1, "error"},
		{"Errorf", func(l *Logger, msg string) { l.Errorf("%s", msg) }, 1, "error"},
		{"Errorln", func(l *Logger, msg string) { l.Errorln(msg) }, 1, "error"},

		{"Warn", func(l *Logger, msg string) { l.Warn(msg) }, 2, "warn"},
		{"Warnf", func(l *Logger, msg string) { l.Warnf("%s", msg) }, 2, "warn"},
		{"Warnln", func(l *Logger, msg string) { l.Warnln(msg) }, 2, "warn"},

		{"Info", func(l *Logger, msg string) { l.Info(msg) }, 3, "info"},
		{"Infof", func(l *Logger, msg string) { l.Infof("%s", msg) }, 3, "info"},
		{"Infoln", func(l *Logger, msg string) { l.Infoln(msg) }, 3, "info"},

		{"Debug", func(l *Logger, msg string) { l.Debug(msg) }, 4, "debug"},
		{"Debugf", func(l *Logger, msg string) { l.Debugf("%s", msg) }, 4, "debug"},
		{"Debugln", func(l *Logger, msg string) { l.Debugln(msg) }, 4, "debug"},

		{"Trace", func(l *Logger, msg string) { l.Trace(msg) }, 5, "trace"},
		{"Tracef", func(l *Logger, msg string) { l.Tracef("%s", msg) }, 5, "trace"},
		{"Traceln", func(l *Logger, msg string) { l.Traceln(msg) }, 5, "trace"},
	}

	levels := []struct {
		name  string
		value int
	}{
		{"FATAL", 0},
		{"ERROR", 1},
		{"WARN", 2},
		{"INFO", 3},
		{"DEBUG", 4},
		{"TRACE", 5},
	}

	for _, level := range levels {
		level := level
		t.Run(level.name, func(t *testing.T) {
			logger, tempDir, cfg := setupLoggerWithFiles(t, level.name)
			defer os.RemoveAll(tempDir)
			defer logger.Close()

			for _, mc := range methodCases {
				mc := mc
				t.Run(mc.name, func(t *testing.T) {
					message := fmt.Sprintf("%s-%s", mc.name, level.name)
					mc.call(logger, message)

					var outputPath string
					switch mc.outputKey {
					case "output":
						outputPath = cfg.OutputPath
					case "fatal":
						outputPath = cfg.FatalOutputPath
					case "error":
						outputPath = cfg.ErrorOutputPath
					case "warn":
						outputPath = cfg.WarnOutputPath
					case "info":
						outputPath = cfg.InfoOutputPath
					case "debug":
						outputPath = cfg.DebugOutputPath
					case "trace":
						outputPath = cfg.TraceOutputPath
					default:
						t.Fatal("unknown output key")
					}

					content := readFileContents(t, outputPath)
					messageFound := strings.Contains(content, message)
					if level.value >= mc.minLevel {
						if !messageFound {
							t.Errorf("wanted %s to log %q to %s", mc.name, message, filepath.Base(outputPath))
						}
					} else {
						if messageFound {
							t.Errorf("expected %s not to log %q at %s level", mc.name, message, level.name)
						}
					}
				})
			}
		})
	}
}

func TestLoggerWithStandardLogCompatibility(t *testing.T) {
	logger, tempDir, _ := setupLoggerWithFiles(t, "INFO")
	defer os.RemoveAll(tempDir)
	defer logger.Close()

	var stdLogger *log.Logger = logger.Logger
	buf := bytes.NewBuffer(nil)
	stdLogger.SetOutput(buf)
	stdLogger.SetPrefix("STD: ")
	stdLogger.SetFlags(0)

	stdLogger.Print("hello")
	stdLogger.Printf("%s", "world")
	stdLogger.Println("test")
	if err := stdLogger.Output(2, "output"); err != nil {
		t.Fatalf("Output failed: %v", err)
	}

	if !strings.Contains(buf.String(), "STD: hello") {
		t.Error("expected stdLogger.Print output")
	}
	if !strings.Contains(buf.String(), "STD: world") {
		t.Error("expected stdLogger.Printf output")
	}
	if !strings.Contains(buf.String(), "STD: test") {
		t.Error("expected stdLogger.Println output")
	}
	if !strings.Contains(buf.String(), "STD: output") {
		t.Error("expected stdLogger.Output output")
	}
}

func TestLogLevelSwitch(t *testing.T) {
	levels := []struct {
		input         string
		expectedLevel int
	}{
		{"FATAL", 0},
		{"ERROR", 1},
		{"WARN", 2},
		{"INFO", 3},
		{"DEBUG", 4},
		{"TRACE", 5},
		{"UNKNOWN", 3},
	}

	for _, test := range levels {
		t.Run(test.input, func(t *testing.T) {
			logger, tempDir, cfg := setupLoggerWithFiles(t, test.input)
			defer os.RemoveAll(tempDir)
			defer logger.Close()

			if logger.logLevel != test.expectedLevel {
				t.Fatalf("For level %s expected %d, got %d", test.input, test.expectedLevel, logger.logLevel)
			}

			if cfg.LogLevel != test.input {
				t.Fatalf("Config should preserve log level value")
			}
		})
	}
}

func TestNextRotationTimeUsesSystemTimezone(t *testing.T) {
	now := time.Date(2026, 4, 24, 1, 30, 0, 0, time.Local)
	next := nextRotationTime(now, 2, 0)
	expected := time.Date(2026, 4, 24, 2, 0, 0, 0, time.Local)
	if !next.Equal(expected) {
		t.Fatalf("expected next rotation %v, got %v", expected, next)
	}

	now = time.Date(2026, 4, 24, 2, 0, 0, 0, time.Local)
	next = nextRotationTime(now, 2, 0)
	expected = time.Date(2026, 4, 25, 2, 0, 0, 0, time.Local)
	if !next.Equal(expected) {
		t.Fatalf("expected next rotation %v, got %v", expected, next)
	}
}

func TestRotatingFileRotate(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal(err)
	}

	rf := &rotatingFile{path: path, file: f}
	if _, err := rf.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}

	rotateAt := time.Date(2026, 4, 24, 2, 0, 0, 0, time.Local)
	if err := rf.Rotate(rotateAt); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected new file to exist: %v", err)
	}

	rotatedPath := rotateFilePath(path, rotateAt)
	if _, err := os.Stat(rotatedPath); err != nil {
		t.Fatalf("expected rotated file %q: %v", rotatedPath, err)
	}

	content, err := os.ReadFile(rotatedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "hello") {
		t.Fatalf("rotated file missing expected content")
	}
}
