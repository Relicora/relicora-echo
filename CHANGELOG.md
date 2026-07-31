# Changelog

All notable changes to this project will be documented in this file.

## [v0.3.1] - 2026-04-25
### Improved
- Added GoDoc comments for all public logger methods to improve editor hover help and documentation.

## [Unreleased]
- 

## [v0.4.0] - 2026-08-01
### Added
- Added explicit rotation modes: `none`, `time`, `size`, and `time-and-size`.
- Added size-based log rotation with configurable `MaxSizeBytes`.
- Added retention cleanup for rotated files through `RetentionDays`.
- Added midnight rollover support for size-based rotation, using the previous day in archive names and the next sequence number.

### Changed
- Time-based rotation now uses the explicit `RotationMode` selection model.
- Archive names now use the format `base-YYYY-MM-DD-000N.ext`.

## [v0.3.0] - 2026-04-24
### Added
- Daily log rotation support for file outputs.
- Rotation schedule configurable via `Config.RotateDaily` and `Config.RotationTime`.
- Rotation time is interpreted in the system local timezone.
- Safe rotation implementation with automatic file renaming and new file creation.
- Unit tests covering rotation scheduling and rotating file behavior.

## [v0.2.2] - 2026-04-20
### Added
- Comprehensive logger tests covering method, log level and file output compatibility.

## [v0.2.1] - 2026-04-20
### Fixed
- Recursive log method calls were corrected.
- `Logger.Close` became idempotent and safer for repeated close operations.
- Added regression tests for log shutdown behavior.

## [v0.2.0] - 2026-04-20
### Added
- Embedded `*log.Logger` compatibility for standard log usage.
- Added unit tests for standard logger compatibility with existing logger methods.

## [v0.1.1] - 2026-04-18
### Fixed
- `Close` now safely handles nil file entries.
- Log level switch was updated to consistently use named constants.

## [v0.1.0] - 2026-04-18
### Added
- Initial logger implementation with per-level file outputs and configurable log levels.
