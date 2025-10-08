# Contributing to Beeper Desktop API Rust SDK

Thank you for your interest in contributing to the Beeper Desktop API Rust SDK! We welcome contributions from everyone.

## Getting Started

### Prerequisites

- Rust 1.70.0 or later
- Git

### Setting up the development environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/beeper-go-sdk.git
   cd beeper-go-sdk/beeper-rust-sdk
   ```

3. Install development dependencies:
   ```bash
   make install-dev-deps
   ```

4. Run the test suite to make sure everything works:
   ```bash
   make test
   ```

## Development Workflow

### Making Changes

1. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. Make your changes, following the coding guidelines below.

3. Add tests for your changes:
   - Unit tests in the same file as your code (in a `tests` module)
   - Integration tests in the `tests/` directory
   - Examples in the `examples/` directory if applicable

4. Run the test suite:
   ```bash
   make check
   ```

5. Commit your changes:
   ```bash
   git add .
   git commit -m "Add my new feature"
   ```

6. Push to your fork and submit a pull request.

### Coding Guidelines

#### Code Style

- Use `cargo fmt` to format your code. Run `make fmt` to format all code.
- Use `cargo clippy` to check for common issues. Run `make lint` to check all code.
- Follow Rust naming conventions:
  - `snake_case` for variables, functions, modules
  - `PascalCase` for types, traits, enums
  - `SCREAMING_SNAKE_CASE` for constants

#### Documentation

- All public functions and types must have documentation comments (`///`)
- Include examples in doc comments where helpful
- Use `cargo doc` to generate documentation locally

#### Error Handling

- Use the `Result` type for fallible operations
- Create specific error types for different failure modes
- Include context in error messages

#### Testing

- Write unit tests for individual functions
- Write integration tests for complete workflows
- Use the `wiremock` crate for mocking HTTP responses in tests
- Test both success and error cases
- Aim for high test coverage

### Project Structure

```
src/
├── lib.rs              # Main library entry point
├── client.rs           # HTTP client implementation
├── config.rs           # Configuration builder
├── error.rs            # Error types
├── utils.rs            # Utility functions
├── version.rs          # Version constant
├── resources/          # API resource modules
│   ├── mod.rs
│   ├── shared.rs       # Shared types
│   ├── accounts.rs     # Accounts API
│   ├── chats.rs        # Chats API
│   ├── messages.rs     # Messages API
│   ├── contacts.rs     # Contacts API
│   ├── token.rs        # Token API
│   └── app.rs          # App API
└── bin/                # Binary applications
    ├── testapp.rs      # Interactive test application
    └── archive_chats.rs # Chat archiving tool
```

### Adding New API Endpoints

When adding support for new API endpoints:

1. Add the request/response types to the appropriate module in `src/resources/`
2. Add the method to the appropriate resource struct
3. Add comprehensive tests (both unit and integration)
4. Update the examples if the new endpoint is commonly used
5. Update the documentation

### Testing Guidelines

#### Unit Tests

Place unit tests in a `tests` module within the same file:

```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_my_function() {
        // Test implementation
    }
}
```

#### Integration Tests

Place integration tests in the `tests/` directory. Use `wiremock` to mock API responses:

```rust
#[tokio::test]
async fn test_api_workflow() {
    let (mock_server, client) = setup_mock_client().await;
    
    Mock::given(method("GET"))
        .and(path("/v0/endpoint"))
        .respond_with(ResponseTemplate::new(200).set_body_json(expected_response))
        .mount(&mock_server)
        .await;
    
    let result = client.resource().method().await.unwrap();
    // Assertions
}
```

#### Running Tests

```bash
# Run all tests
make test

# Run specific tests
cargo test test_name

# Run tests with coverage
make test-coverage

# Run integration tests only
make test-integration
```

## Pull Request Process

1. Ensure your code follows the style guidelines
2. Include tests for new functionality
3. Update documentation as needed
4. Make sure all CI checks pass
5. Write a clear pull request description explaining:
   - What changes you made
   - Why you made them
   - How to test the changes

## Release Process

The release process is handled by maintainers:

1. Update version in `Cargo.toml` and `src/version.rs`
2. Update `CHANGELOG.md`
3. Create a git tag
4. Publish to crates.io using `cargo publish`

## Getting Help

- Open an issue on GitHub for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues and discussions first

## Code of Conduct

This project follows the Rust Code of Conduct. Please be respectful and constructive in all interactions.

Thank you for contributing!