package echo

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
}

type rotatingFile struct {
	sync.Mutex
	path string
	file *os.File
}

var nowFunc = time.Now
var newTimer = time.NewTimer

func (r *rotatingFile) Write(p []byte) (int, error) {
	r.Lock()
	defer r.Unlock()
	if r.file == nil {
		return 0, fmt.Errorf("rotating file is closed")
	}
	return r.file.Write(p)
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

func (r *rotatingFile) Rotate(at time.Time) error {
	r.Lock()
	defer r.Unlock()

	if r.file == nil {
		return nil
	}

	if err := r.file.Close(); err != nil {
		return err
	}

	rotatedPath := rotateFilePath(r.path, at)
	if err := os.Rename(r.path, rotatedPath); err != nil && !os.IsNotExist(err) {
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
	return nil
}

func nextRotationTime(now time.Time, hour, minute int) time.Time {
	local := now.In(time.Local)
	next := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, time.Local)
	if !next.After(local) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func rotateFilePath(path string, when time.Time) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)
	suffix := when.Format("2006-01-02_150405")
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", base, suffix, ext))
}

// New creates a Logger configured with the provided settings.
//
// If cfg.RotateDaily is true and cfg.RotationTime is set, the logger starts a daily
// rotation goroutine that renames current log files at the configured local time.
func New(cfg Config) *Logger {
	l := &Logger{
		config: cfg,
	}

	rotationEnabled := cfg.RotateDaily && cfg.RotationTime != ""
	rotationHour := 0
	rotationMinute := 0
	if rotationEnabled {
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
			rf := &rotatingFile{path: path, file: f}
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

	if rotationEnabled {
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
			if err := rf.Rotate(at); err != nil {
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
