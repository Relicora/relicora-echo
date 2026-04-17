package echo

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Logger struct {
	baseLogger  *log.Logger
	fatalLogger *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
	traceLogger *log.Logger

	files    []*os.File
	logLevel int
	config   Config
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
}

func New(cfg Config) *Logger {
	l := &Logger{
		config: cfg,
	}

	// helper
	openFile := func(path string) *os.File {
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
	l.baseLogger = log.New(baseWriter, "LOG:   ", log.Ldate|log.Ltime)
	l.fatalLogger = log.New(buildWriter(cfg.FatalOutputPath), "FATAL: ", log.Ldate|log.Ltime)
	l.errorLogger = log.New(buildWriter(cfg.ErrorOutputPath), "ERROR: ", log.Ldate|log.Ltime)
	l.warnLogger = log.New(buildWriter(cfg.WarnOutputPath), "WARN:  ", log.Ldate|log.Ltime)
	l.infoLogger = log.New(buildWriter(cfg.InfoOutputPath), "INFO:  ", log.Ldate|log.Ltime)
	l.debugLogger = log.New(buildWriter(cfg.DebugOutputPath), "DEBUG: ", log.Ldate|log.Ltime)
	l.traceLogger = log.New(buildWriter(cfg.TraceOutputPath), "TRACE: ", log.Ldate|log.Ltime)

	// log level
	switch strings.ToUpper(cfg.LogLevel) {
	case "FATAL":
		l.logLevel = 0
	case "ERROR":
		l.logLevel = 1
	case "WARN":
		l.logLevel = 2
	case "INFO":
		l.logLevel = 3
	case "DEBUG":
		l.logLevel = 4
	case "TRACE":
		l.logLevel = 5
	default:
		l.logLevel = 3
	}

	l.infoLogger.Printf("Logger started with log_level='%s'", cfg.LogLevel)

	return l
}

func (l *Logger) Close() error {
	for _, f := range l.files {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
