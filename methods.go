package echo

// FATAL
func (l *Logger) Fatal(v ...any) {
	if l.logLevel >= 0 {
		l.fatalLogger.Fatal(v...)
	}
}

func (l *Logger) Fatalf(format string, v ...any) {
	if l.logLevel >= 0 {
		l.fatalLogger.Fatalf(format, v...)
	}
}

func (l *Logger) Fatalln(v ...any) {
	if l.logLevel >= 0 {
		l.fatalLogger.Fatalln(v...)
	}
}

// ERROR
func (l *Logger) Error(v ...any) {
	if l.logLevel >= 1 {
		l.errorLogger.Print(v...)
	}
}

func (l *Logger) Errorf(format string, v ...any) {
	if l.logLevel >= 1 {
		l.errorLogger.Printf(format, v...)
	}
}

func (l *Logger) Errorln(v ...any) {
	if l.logLevel >= 1 {
		l.errorLogger.Println(v...)
	}
}

// WARN
func (l *Logger) Warn(v ...any) {
	if l.logLevel >= 2 {
		l.warnLogger.Print(v...)
	}
}

func (l *Logger) Warnf(format string, v ...any) {
	if l.logLevel >= 2 {
		l.warnLogger.Printf(format, v...)
	}
}

func (l *Logger) Warnln(v ...any) {
	if l.logLevel >= 2 {
		l.warnLogger.Println(v...)
	}
}

// INFO
func (l *Logger) Info(v ...any) {
	if l.logLevel >= 3 {
		l.infoLogger.Print(v...)
	}
}

func (l *Logger) Infof(format string, v ...any) {
	if l.logLevel >= 3 {
		l.infoLogger.Printf(format, v...)
	}
}

func (l *Logger) Infoln(v ...any) {
	if l.logLevel >= 3 {
		l.infoLogger.Println(v...)
	}
}

// LOG
func (l *Logger) Print(v ...any) {
	if l.logLevel >= 3 {
		l.Logger.Print(v...)
	}
}

func (l *Logger) Printf(format string, v ...any) {
	if l.logLevel >= 3 {
		l.Logger.Printf(format, v...)
	}
}

func (l *Logger) Println(v ...any) {
	if l.logLevel >= 3 {
		l.Logger.Println(v...)
	}
}

// DEBUG
func (l *Logger) Debug(v ...any) {
	if l.logLevel >= 4 {
		l.debugLogger.Print(v...)
	}
}

func (l *Logger) Debugf(format string, v ...any) {
	if l.logLevel >= 4 {
		l.debugLogger.Printf(format, v...)
	}
}

func (l *Logger) Debugln(v ...any) {
	if l.logLevel >= 4 {
		l.debugLogger.Println(v...)
	}
}

// TRACE
func (l *Logger) Trace(v ...any) {
	if l.logLevel >= 5 {
		l.traceLogger.Print(v...)
	}
}

func (l *Logger) Tracef(format string, v ...any) {
	if l.logLevel >= 5 {
		l.traceLogger.Printf(format, v...)
	}
}

func (l *Logger) Traceln(v ...any) {
	if l.logLevel >= 5 {
		l.traceLogger.Println(v...)
	}
}
