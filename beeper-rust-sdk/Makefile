# Makefile for Beeper Desktop API Rust SDK

.PHONY: build test lint fmt check clean demo run-example help

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the project"
	@echo "  test         - Run all tests with coverage"
	@echo "  lint         - Run clippy linter"
	@echo "  fmt          - Format code"
	@echo "  check        - Run fmt + clippy + test"
	@echo "  clean        - Clean build artifacts"
	@echo "  demo         - Run the test application demo"
	@echo "  run-example  - Run basic example"
	@echo "  run-advanced - Run advanced example"
	@echo "  docs         - Generate documentation"
	@echo "  publish-dry  - Dry run of cargo publish"

# Build the project
build:
	cargo build

# Build in release mode
build-release:
	cargo build --release

# Run tests with coverage
test:
	cargo test --all-features

# Run tests with coverage report
test-coverage:
	cargo tarpaulin --out html --output-dir coverage

# Run clippy linter
lint:
	cargo clippy --all-targets --all-features -- -D warnings

# Format code
fmt:
	cargo fmt

# Check if code is formatted
fmt-check:
	cargo fmt -- --check

# Run all checks (format, lint, test)
check: fmt-check lint test

# Clean build artifacts
clean:
	cargo clean

# Run the basic example
run-example:
	cargo run --example basic

# Run the advanced example
run-advanced:
	cargo run --example advanced

# Run the test application
demo:
	cargo run --bin testapp

# Run the archive chats tool
run-archive:
	cargo run --bin archive-chats

# Generate documentation
docs:
	cargo doc --all-features --no-deps

# Open documentation in browser
docs-open:
	cargo doc --all-features --no-deps --open

# Check for security vulnerabilities
audit:
	cargo audit

# Update dependencies
update:
	cargo update

# Dry run publish to crates.io
publish-dry:
	cargo publish --dry-run

# Publish to crates.io
publish:
	cargo publish

# Install development dependencies
install-dev-deps:
	cargo install cargo-tarpaulin
	cargo install cargo-audit
	cargo install cargo-expand

# Run benchmarks (if any)
bench:
	cargo bench

# Check for unused dependencies
unused-deps:
	cargo +nightly udeps

# Profile the build
profile-build:
	cargo build --timings

# Run specific test
test-unit TEST:
	cargo test ${TEST}

# Run integration tests only
test-integration:
	cargo test --test '*'

# Build and run with specific features
run-with-features FEATURES:
	cargo run --features ${FEATURES} --example basic

# Cross-compile for different targets
cross-compile TARGET:
	cargo build --target ${TARGET}