package echo

// Fatal logs a message at FATAL level and exits the application.
func (l *Logger) Fatal(v ...any) {
	if l.logLevel >= 0 {
		l.fatalLogger.Fatal(v...)
	}
}

// Fatalf formats and logs a message at FATAL level before exiting the application.
func (l *Logger) Fatalf(format string, v ...any) {
	if l.logLevel >= 0 {
		l.fatalLogger.Fatalf(format, v...)
	}
}

// Fatalln logs a message at FATAL level with a newline and exits the application.
func (l *Logger) Fatalln(v ...any) {
	if l.logLevel >= 0 {
		l.fatalLogger.Fatalln(v...)
	}
}

// Error logs a message at ERROR level if the current log level allows it.
func (l *Logger) Error(v ...any) {
	if l.logLevel >= 1 {
		l.errorLogger.Print(v...)
	}
}

// Errorf formats and logs an ERROR-level message when enabled.
func (l *Logger) Errorf(format string, v ...any) {
	if l.logLevel >= 1 {
		l.errorLogger.Printf(format, v...)
	}
}

// Errorln logs an ERROR-level message with a newline when enabled.
func (l *Logger) Errorln(v ...any) {
	if l.logLevel >= 1 {
		l.errorLogger.Println(v...)
	}
}

// Warn logs a message at WARN level if the current log level allows it.
func (l *Logger) Warn(v ...any) {
	if l.logLevel >= 2 {
		l.warnLogger.Print(v...)
	}
}

// Warnf formats and logs a WARN-level message when enabled.
func (l *Logger) Warnf(format string, v ...any) {
	if l.logLevel >= 2 {
		l.warnLogger.Printf(format, v...)
	}
}

// Warnln logs a WARN-level message with a newline when enabled.
func (l *Logger) Warnln(v ...any) {
	if l.logLevel >= 2 {
		l.warnLogger.Println(v...)
	}
}

// Info logs a message at INFO level if the current log level allows it.
func (l *Logger) Info(v ...any) {
	if l.logLevel >= 3 {
		l.infoLogger.Print(v...)
	}
}

// Infof formats and logs an INFO-level message when enabled.
func (l *Logger) Infof(format string, v ...any) {
	if l.logLevel >= 3 {
		l.infoLogger.Printf(format, v...)
	}
}

// Infoln logs an INFO-level message with a newline when enabled.
func (l *Logger) Infoln(v ...any) {
	if l.logLevel >= 3 {
		l.infoLogger.Println(v...)
	}
}

// Print writes a message through the embedded standard logger at INFO level.
func (l *Logger) Print(v ...any) {
	if l.logLevel >= 3 {
		l.Logger.Print(v...)
	}
}

// Printf formats and writes a message through the embedded standard logger at INFO level.
func (l *Logger) Printf(format string, v ...any) {
	if l.logLevel >= 3 {
		l.Logger.Printf(format, v...)
	}
}

// Println writes a message through the embedded standard logger with a newline at INFO level.
func (l *Logger) Println(v ...any) {
	if l.logLevel >= 3 {
		l.Logger.Println(v...)
	}
}

// Debug logs a message at DEBUG level if the current log level allows it.
func (l *Logger) Debug(v ...any) {
	if l.logLevel >= 4 {
		l.debugLogger.Print(v...)
	}
}

// Debugf formats and logs a DEBUG-level message when enabled.
func (l *Logger) Debugf(format string, v ...any) {
	if l.logLevel >= 4 {
		l.debugLogger.Printf(format, v...)
	}
}

// Debugln logs a DEBUG-level message with a newline when enabled.
func (l *Logger) Debugln(v ...any) {
	if l.logLevel >= 4 {
		l.debugLogger.Println(v...)
	}
}

// Trace logs a message at TRACE level if the current log level allows it.
func (l *Logger) Trace(v ...any) {
	if l.logLevel >= 5 {
		l.traceLogger.Print(v...)
	}
}

// Tracef formats and logs a TRACE-level message when enabled.
func (l *Logger) Tracef(format string, v ...any) {
	if l.logLevel >= 5 {
		l.traceLogger.Printf(format, v...)
	}
}

// Traceln logs a TRACE-level message with a newline when enabled.
func (l *Logger) Traceln(v ...any) {
	if l.logLevel >= 5 {
		l.traceLogger.Println(v...)
	}
}
