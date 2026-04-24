# relicora-echo

`relicora-echo` is a lightweight Go logging library with per-level file output, standard logger compatibility, and optional daily log rotation using the system local timezone.

## Features

- Hierarchical log levels: `FATAL`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`
- Separate file targets for each log level
- Fallback to a shared output file for levels without their own path
- Compatible with Go's standard `log.Logger`
- Optional daily rotation of log files at a configured local time
- Safe close semantics for file handles and rotation goroutines

## Installation

```bash
go get github.com/Relicora/relicora-echo@v0.3.0
```

Import the package in your code:

```go
import "github.com/Relicora/relicora-echo"
```

## Usage

Create a logger with the desired configuration and call methods for the desired log level.

```go
package main

import (
    "github.com/Relicora/relicora-echo"
)

func main() {
    cfg := echo.Config{
        LogLevel:        "INFO",
        OutputPath:      "logs/output.log",
        ErrorOutputPath: "logs/error.log",
        WarnOutputPath:  "logs/warn.log",
        InfoOutputPath:  "logs/info.log",
        DebugOutputPath: "logs/debug.log",
        TraceOutputPath: "logs/trace.log",
    }

    logger := echo.New(cfg)
    defer logger.Close()

    logger.Info("Application started")
    logger.Warn("This is a warning")
    logger.Error("Something went wrong")
}
```

### Using daily rotation

Enable daily rotation by setting `RotateDaily` and `RotationTime` in the configuration.
The rotation time uses the system local timezone.

```go
cfg := echo.Config{
    LogLevel:      "INFO",
    OutputPath:    "logs/output.log",
    RotateDaily:   true,
    RotationTime:  "02:00", // rotate at 02:00 local time each day
}

logger := echo.New(cfg)
defer logger.Close()
```

When rotation occurs, the existing file is renamed to include a timestamp suffix, for example: `output-2026-04-24_020000.log`.

## Configuration

`Config` fields:

- `LogLevel string` - minimum log level to output. Allowed values: `FATAL`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`. Defaults to `INFO` if missing or unknown.
- `OutputPath string` - shared base log file path used when specific per-level path is not provided.
- `FatalOutputPath string` - optional path for fatal logs.
- `ErrorOutputPath string` - optional path for error logs.
- `WarnOutputPath string` - optional path for warning logs.
- `InfoOutputPath string` - optional path for info logs.
- `DebugOutputPath string` - optional path for debug logs.
- `TraceOutputPath string` - optional path for trace logs.
- `RotateDaily bool` - enable daily rotation for file outputs.
- `RotationTime string` - local rotation time in `HH:MM` format, for example `23:30`.

## Logger methods

The logger supports standard methods for each level:

- `Fatal`, `Fatalf`, `Fatalln`
- `Error`, `Errorf`, `Errorln`
- `Warn`, `Warnf`, `Warnln`
- `Info`, `Infof`, `Infoln`
- `Print`, `Printf`, `Println`
- `Debug`, `Debugf`, `Debugln`
- `Trace`, `Tracef`, `Traceln`

The `Print*` methods are treated as `INFO`-level output.

## Standard logger compatibility

The embedded `*log.Logger` is available at `logger.Logger`, so existing code that expects a standard Go logger may use it directly.

```go
stdLogger := logger.Logger
stdLogger.Print("standard log")
```

## Shutdown

Always call `logger.Close()` before exiting your application to flush and close file handles cleanly.

```go
defer logger.Close()
```

## Testing

Run the package tests with:

```bash
go test ./...
```

The library includes tests for:

- log level parsing and behavior
- per-level file output
- standard logger compatibility
- daily rotation scheduling
- rotating file rename behavior

## License

This project is licensed under the terms of the [MIT License](LICENSE).
