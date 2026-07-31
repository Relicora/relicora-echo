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
	if err := rf.Rotate(rotateAt, 0); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected new file to exist: %v", err)
	}

	_, err = rotateFilePath(path, rotateAt, 0)
	if err != nil {
		t.Fatal(err)
	}
	entries, readErr := os.ReadDir(filepath.Dir(path))
	if readErr != nil {
		t.Fatal(readErr)
	}
	found := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "app-2026-04-24-") && strings.HasSuffix(entry.Name(), ".log") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a rotated file in %s", filepath.Dir(path))
	}

	var archivedContent string
	for _, entry := range entries {
		if entry.Name() != "app.log" {
			content, err := os.ReadFile(filepath.Join(filepath.Dir(path), entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			archivedContent = string(content)
			break
		}
	}
	if !strings.Contains(archivedContent, "hello") {
		t.Fatalf("rotated file missing expected content")
	}
}

func TestRotationArchiveTimeUsesPreviousDayForMidnight(t *testing.T) {
	at := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	got := rotationArchiveTime(at, 0, 0)
	expected := time.Date(2026, 7, 31, 0, 0, 0, 0, time.Local)
	if !got.Equal(expected) {
		t.Fatalf("expected midnight archive time %v, got %v", expected, got)
	}
}

func TestRotatingFileRotatesBySizeAndUsesSequenceNumber(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "app.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal(err)
	}

	rf := &rotatingFile{path: path, file: f, maxSize: 10}
	if _, err := rf.Write([]byte("0123456789")); err != nil {
		t.Fatal(err)
	}
	if _, err := rf.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 files, got %d", len(entries))
	}

	var rotatedName string
	for _, entry := range entries {
		if entry.Name() != "app.log" {
			rotatedName = entry.Name()
			break
		}
	}
	if rotatedName == "" {
		t.Fatal("expected a rotated file")
	}
	if !strings.Contains(rotatedName, "-0001") {
		t.Fatalf("expected sequence number in rotated file name, got %s", rotatedName)
	}
}

func TestRotatingFileRetentionDeletesOldRotatedFiles(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "app.log")
	if err := os.WriteFile(filepath.Join(tempDir, "app-2026-07-01.log"), []byte("old"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "app-2026-07-20.log"), []byte("old"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "app-2026-08-01.log"), []byte("new"), 0666); err != nil {
		t.Fatal(err)
	}

	rf := &rotatingFile{path: path, retentionDays: 7}
	if err := rf.cleanupRotatedFiles(time.Date(2026, 8, 8, 0, 0, 0, 0, time.Local)); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"app-2026-07-01.log", "app-2026-07-20.log"} {
		if _, err := os.Stat(filepath.Join(tempDir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed", name)
		}
	}
	if _, err := os.Stat(filepath.Join(tempDir, "app-2026-08-01.log")); err != nil {
		t.Fatalf("expected recent rotated file to remain: %v", err)
	}
}

func TestRotateBySizeUsesYesterdayDateAtMidnight(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "app.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal(err)
	}

	rf := &rotatingFile{path: path, file: f, maxSize: 1024}
	at := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	if err := rf.rotateBySize(at); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 files, got %d", len(entries))
	}

	var found bool
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "app-2026-07-31-") && strings.HasSuffix(entry.Name(), ".log") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected archived file with yesterday's date, got %v", entries)
	}
}
