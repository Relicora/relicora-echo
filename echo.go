package echo

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	*log.Logger
	fatalLogger *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
	traceLogger *log.Logger

	files        []io.Closer
	rotationStop chan struct{}
	rotationDone chan struct{}
	logLevel     int
	config       Config
}

type Config struct {
	LogLevel        string
	OutputPath      string
	FatalOutputPath string
	ErrorOutputPath string
	WarnOutputPath  string
	InfoOutputPath  string
	DebugOutputPath string
	TraceOutputPath string

	RotateDaily  bool
	RotationTime string // format "15:04", local system timezone

	RotationMode  string // "none", "time", "size", "time-and-size"
	MaxSizeBytes  int64
	RetentionDays int
}

const (
	RotationModeNone        = "none"
	RotationModeTime        = "time"
	RotationModeSize        = "size"
	RotationModeTimeAndSize = "time-and-size"
)

type rotatingFile struct {
	sync.Mutex
	path          string
	file          *os.File
	maxSize       int64
	retentionDays int
	size          int64
}

var nowFunc = time.Now
var newTimer = time.NewTimer

func (r *rotatingFile) Write(p []byte) (int, error) {
	r.Lock()
	defer r.Unlock()
	if r.file == nil {
		return 0, fmt.Errorf("rotating file is closed")
	}

	if r.maxSize > 0 && r.size > 0 && r.size+int64(len(p)) > r.maxSize {
		if err := r.rotateLocked(nowFunc(), 0); err != nil {
			return 0, err
		}
	}

	n, err := r.file.Write(p)
	r.size += int64(n)
	return n, err
}

func (r *rotatingFile) Close() error {
	r.Lock()
	defer r.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

func (r *rotatingFile) Rotate(at time.Time, sequence int) error {
	r.Lock()
	defer r.Unlock()
	return r.rotateLocked(at, sequence)
}

func (r *rotatingFile) rotateByTime(at time.Time) error {
	return r.rotateLocked(at, 0)
}

func (r *rotatingFile) rotateBySize(at time.Time) error {
	return r.rotateLocked(at, 0)
}

func (r *rotatingFile) rotateLocked(at time.Time, sequence int) error {
	if r.file == nil {
		return nil
	}

	if err := r.file.Close(); err != nil {
		return err
	}

	archivePath, err := rotateFilePath(r.path, at, sequence)
	if err != nil {
		return err
	}
	if err := os.Rename(r.path, archivePath); err != nil && !os.IsNotExist(err) {
		f, openErr := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if openErr == nil {
			r.file = f
		}
		return err
	}

	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	r.file = f
	r.size = 0
	if r.retentionDays > 0 {
		if err := r.cleanupRotatedFiles(at); err != nil {
			return err
		}
	}
	return nil
}

func (r *rotatingFile) cleanupRotatedFiles(at time.Time) error {
	if r.retentionDays <= 0 {
		return nil
	}

	cutoff := at.AddDate(0, 0, -r.retentionDays)
	ext := filepath.Ext(r.path)
	base := strings.TrimSuffix(filepath.Base(r.path), ext)
	dir := filepath.Dir(r.path)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, base+"-") || !strings.HasSuffix(name, ext) {
			continue
		}
		stamp, ok := parseArchiveDate(name, ext)
		if !ok {
			continue
		}
		if !stamp.Before(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func parseArchiveDate(name, ext string) (time.Time, bool) {
	stem := strings.TrimSuffix(name, ext)
	match := regexp.MustCompile(`-(\d{4})-(\d{2})-(\d{2})(?:-\d+)?$`).FindStringSubmatch(stem)
	if len(match) != 4 {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", match[1]+"-"+match[2]+"-"+match[3])
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func nextRotationTime(now time.Time, hour, minute int) time.Time {
	local := now.In(time.Local)
	next := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, time.Local)
	if !next.After(local) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func resolveRotationMode(cfg Config) string {
	mode := strings.ToLower(strings.TrimSpace(cfg.RotationMode))
	switch mode {
	case RotationModeNone, RotationModeTime, RotationModeSize, RotationModeTimeAndSize:
		return mode
	default:
	}

	if cfg.RotateDaily && cfg.MaxSizeBytes > 0 {
		return RotationModeTimeAndSize
	}
	if cfg.RotateDaily {
		return RotationModeTime
	}
	if cfg.MaxSizeBytes > 0 {
		return RotationModeSize
	}
	return RotationModeNone
}

func rotationArchiveTime(at time.Time, hour, minute int) time.Time {
	if hour == 0 && minute == 0 {
		return at.AddDate(0, 0, -1)
	}
	return at
}

func nextArchiveSequence(path string, when time.Time) int {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)
	prefix := fmt.Sprintf("%s-%s-", base, when.Format("2006-01-02"))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}

	highest := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ext) {
			continue
		}
		suffix := strings.TrimPrefix(name, prefix)
		suffix = strings.TrimSuffix(suffix, ext)
		seq, err := strconv.Atoi(suffix)
		if err == nil && seq > highest {
			highest = seq
		}
	}
	return highest + 1
}

func rotateFilePath(path string, when time.Time, sequence int) (string, error) {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)
	archiveDate := rotationArchiveTime(when, when.Hour(), when.Minute())
	stamp := archiveDate.Format("2006-01-02")
	if sequence <= 0 {
		sequence = nextArchiveSequence(path, archiveDate)
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s-%04d%s", base, stamp, sequence, ext)), nil
}

// New creates a Logger configured with the provided settings.
//
// If cfg.RotateDaily is true and cfg.RotationTime is set, the logger starts a daily
// rotation goroutine that renames current log files at the configured local time.
func New(cfg Config) *Logger {
	l := &Logger{
		config: cfg,
	}

	rotationMode := resolveRotationMode(cfg)
	timeRotationEnabled := rotationMode == RotationModeTime || rotationMode == RotationModeTimeAndSize || rotationMode == RotationModeSize
	rotationEnabled := rotationMode != RotationModeNone || cfg.RetentionDays > 0
	rotationHour := 0
	rotationMinute := 0
	if cfg.RotationTime != "" && (rotationMode == RotationModeTime || rotationMode == RotationModeTimeAndSize) {
		parsed, err := time.Parse("15:04", cfg.RotationTime)
		if err != nil {
			log.Fatal(err)
		}
		rotationHour = parsed.Hour()
		rotationMinute = parsed.Minute()
	}

	// helper
	openFile := func(path string) io.Writer {
		if path == "" {
			return nil
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}

		if rotationEnabled {
			rf := &rotatingFile{path: path, file: f, maxSize: cfg.MaxSizeBytes, retentionDays: cfg.RetentionDays}
			l.files = append(l.files, rf)
			return rf
		}

		l.files = append(l.files, f)
		return f
	}

	// base writer
	var baseWriter io.Writer = os.Stdout
	baseFile := openFile(cfg.OutputPath)
	if baseFile != nil {
		baseWriter = io.MultiWriter(os.Stdout, baseFile)
	}

	// helper for select writer
	buildWriter := func(levelPath string) io.Writer {
		if levelPath != "" {
			f := openFile(levelPath)
			return io.MultiWriter(os.Stdout, f)
		}

		// fallback
		return baseWriter
	}

	// create loggers
	l.Logger = log.New(baseWriter, "LOG:   ", log.Ldate|log.Ltime)
	l.fatalLogger = log.New(buildWriter(cfg.FatalOutputPath), "FATAL: ", log.Ldate|log.Ltime)
	l.errorLogger = log.New(buildWriter(cfg.ErrorOutputPath), "ERROR: ", log.Ldate|log.Ltime)
	l.warnLogger = log.New(buildWriter(cfg.WarnOutputPath), "WARN:  ", log.Ldate|log.Ltime)
	l.infoLogger = log.New(buildWriter(cfg.InfoOutputPath), "INFO:  ", log.Ldate|log.Ltime)
	l.debugLogger = log.New(buildWriter(cfg.DebugOutputPath), "DEBUG: ", log.Ldate|log.Ltime)
	l.traceLogger = log.New(buildWriter(cfg.TraceOutputPath), "TRACE: ", log.Ldate|log.Ltime)

	// log level
	switch strings.ToUpper(cfg.LogLevel) {
	case FATAL:
		l.logLevel = 0
	case ERROR:
		l.logLevel = 1
	case WARN:
		l.logLevel = 2
	case INFO:
		l.logLevel = 3
	case DEBUG:
		l.logLevel = 4
	case TRACE:
		l.logLevel = 5
	default:
		l.logLevel = 3
	}

	if timeRotationEnabled {
		l.startRotation(rotationHour, rotationMinute)
	}

	l.infoLogger.Printf("Logger started with log_level='%s'", cfg.LogLevel)

	return l
}

func (l *Logger) startRotation(hour, minute int) {
	l.rotationStop = make(chan struct{})
	l.rotationDone = make(chan struct{})

	go func() {
		defer close(l.rotationDone)
		for {
			now := nowFunc()
			next := nextRotationTime(now, hour, minute)
			timer := newTimer(next.Sub(now))
			select {
			case <-timer.C:
				l.rotateAll(next)
			case <-l.rotationStop:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			}
		}
	}()
}

func (l *Logger) rotateAll(at time.Time) {
	for _, closer := range l.files {
		if rf, ok := closer.(*rotatingFile); ok {
			rf.Lock()
			err := rf.rotateByTime(at)
			rf.Unlock()
			if err != nil {
				log.Printf("log rotation failed for %s: %v", rf.path, err)
			}
		}
	}
}

// Close stops any active rotation goroutine and closes all open log files.
//
// It returns the first error encountered while closing files, if any.
func (l *Logger) Close() error {
	if l.rotationStop != nil {
		close(l.rotationStop)
		<-l.rotationDone
	}

	var firstErr error
	for i, closer := range l.files {
		if closer != nil {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
			l.files[i] = nil
		}
	}
	return firstErr
}
