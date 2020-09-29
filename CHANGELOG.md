# Changelog
All notable changes to this project will be documented in this file.

**ATTN**: This project uses [semantic versioning](http://semver.org/).

## [Unreleased]

### Added
- Added authentication failed test.

## [v1.2.0] - 2020-07-10
### Added
- Added options to Dial. It is possible to set timeout and deadline settings.

### Fixed
- Change `SERVERDATA_AUTH_ID` and `SERVERDATA_EXECCOMMAND_ID` from 42 to 0. Conan Exiles has a bug because of which it 
always responds 42 regardless of the value of the request ID. This is no longer relevant, so the values have been 
changed.

### Changed
- Renamed `DefaultTimeout` const to `DefaultDeadline`
- Changed default timeouts from 10 seconds to 5 seconds

## [v1.1.2] - 2020-05-13
### Added
- Added go modules (go 1.13).
- Added golangci.yml linter config. To run linter use `golangci-lint run` command.
- Added CHANGELOG.md.
- Added more tests.

## v1.0.0 - 2019-07-27
### Added
- Initial implementation.

[Unreleased]: https://github.com/gorcon/rcon/compare/v1.2.0...HEAD
[v1.2.0]: https://github.com/gorcon/rcon/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/gorcon/rcon/compare/v1.0.0...v1.1.2
