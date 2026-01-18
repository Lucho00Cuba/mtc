# Contributing to MTC

**MTC (Merkle Tree Checksum)** — A deterministic directory checksum tool that generates directory checksums using Merkle Trees.

Thank you for your interest in contributing to MTC! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md) that all contributors are expected to follow. By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Bugs

Before reporting a bug, please:

1. Check if the issue already exists in the issue tracker
2. Verify you're using the latest version
3. Try to reproduce the issue with the latest code

When reporting a bug, please include:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Environment details (OS, architecture, Go version)
- Relevant logs or error messages

### Suggesting Features

Feature suggestions are welcome! Please:

1. Check if the feature has already been requested
2. Provide a clear description of the feature
3. Explain the use case and benefits
4. Consider implementation complexity

### Pull Requests

- Prefer small and focused pull requests
- Large refactors should be discussed in an issue first

1. **Fork the repository** and create a branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow the coding style and conventions
   - Write or update tests for your changes
   - Update documentation as needed
   - Ensure all tests pass locally

3. **Run quality checks**
   ```bash
   make lint
   make test
   make test-coverage
   ```

4. **Commit your changes**
   - Write clear, descriptive commit messages
   - Follow the [conventional commits](https://www.conventionalcommits.org/) format when possible
   - Keep commits focused and atomic

5. **Push and create a Pull Request**
   - Push your branch to your fork
   - Create a PR with a clear description
   - Reference any related issues
   - Ensure CI checks pass

## Development Setup

### Prerequisites

- Go 1.24 or later
- Make (optional, but recommended)

### Building

```bash
# Build for your platform
make build

# Build for all platforms
make build-all
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
go test -race ./...
```

### Linting

```bash
# Run linter
make lint

# Format code
make format
```

## Code Style

- Follow standard Go formatting (`gofmt`)
- Use `golangci-lint` for additional checks
- Write clear, self-documenting code
- Add comments for exported functions and types
- Keep functions focused and small
- Use meaningful variable and function names

## Testing Guidelines

- Write tests for new features and bug fixes
- Use table-driven tests when testing multiple scenarios
- Test error cases, not just happy paths
- Keep tests fast and independent
- Use `t.Helper()` in test helper functions

## Performance Considerations

MTC is optimized for large directory trees.
Please avoid changes that:

- Introduce extra filesystem traversals
- Increase memory allocations unnecessarily
- Add blocking IO inside directory loops

Benchmark large-tree performance when modifying hashing logic.

## Documentation

- Update README.md for user-facing changes
- Add code comments for complex logic
- Update CHANGELOG.md for significant changes
- Keep documentation clear and concise

## Commit Messages

Follow these guidelines for commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests when applicable
- Use conventional commit prefixes when possible:
  - `feat:` for new features
  - `fix:` for bug fixes
  - `docs:` for documentation changes
  - `test:` for test changes
  - `refactor:` for code refactoring
  - `chore:` for maintenance tasks

Example:
```
feat: add directory-level caching

Introduce persistent cache entries for directory nodes
to avoid recomputing unchanged Merkle branches.

Fixes #123
```

## Review Process

- All PRs require at least one approval
- Maintainers will review your PR and may request changes
- Address review comments promptly
- Keep discussions focused and constructive

## Questions?

If you have questions or need help, please:

- Open an issue with the `question` label
- Check existing documentation
- Review closed issues and PRs

Thank you for contributing to MTC! 🎉
