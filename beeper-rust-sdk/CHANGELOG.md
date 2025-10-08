# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of the Beeper Desktop API Rust SDK
- Full feature parity with the Go SDK
- Support for all API endpoints:
  - Accounts management
  - Chat operations (create, search, archive, reminders)
  - Message operations (search, send)
  - Contact search
  - App operations (search, open, download assets)
  - Token information
- Comprehensive error handling with typed errors
- Pagination support for all paginated endpoints
- Retry logic with exponential backoff
- Async/await support with Tokio
- Builder pattern for configuration
- Environment variable configuration
- Request/response logging with tracing
- Comprehensive test suite with mock server testing
- Integration tests covering complete workflows
- Command-line tools:
  - Interactive test application (`testapp`)
  - Chat archive tool (`archive-chats`)
- Examples demonstrating basic and advanced usage
- Full documentation with examples
- CI/CD pipeline with GitHub Actions
- Security auditing with cargo-audit
- Code coverage reporting
- Support for both rustls and native TLS backends

### Documentation

- Complete API documentation with examples
- README with usage examples
- Contributing guidelines
- Comprehensive integration tests as documentation

### Tools

- Makefile for common development tasks
- GitHub Actions CI pipeline
- Code formatting with rustfmt
- Linting with clippy
- Security auditing
- Coverage reporting

## [0.1.0] - 2024-01-XX

### Added

- Initial public release