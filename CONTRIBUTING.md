# Contributing to Hephaestus

We love your input! We want to make contributing to Hephaestus as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code follows the Go style guide.
6. Issue that pull request!

## Code Style

### Go Guidelines

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Document all exported types, functions, and methods
- Keep functions focused and small
- Use meaningful variable names
- Handle errors explicitly

### Project-Specific Guidelines

1. **File Organization**
   - Place internal code in the `internal/` directory
   - Place reusable packages in the `pkg/` directory
   - Keep `cmd/` directory clean and minimal

2. **Naming Conventions**
   - Use descriptive package names
   - Avoid package name collisions
   - Use consistent naming across similar types/functions

3. **Error Handling**
   - Use custom error types when needed
   - Wrap errors with context
   - Don't panic in library code

4. **Testing**
   - Write table-driven tests
   - Use meaningful test names
   - Mock external dependencies
   - Aim for high test coverage

## Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Example:
```
Add error pattern matching for panic detection

- Implement regex-based pattern matching
- Add test cases for common panic scenarios
- Update documentation with new patterns

Fixes #123
```

## Pull Request Process

1. Update the README.md with details of changes to the interface
2. Update the documentation with any new configuration or dependencies
3. The PR will be merged once you have the sign-off of two maintainers

## Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run specific test
go test ./internal/collector -run TestCollector

# Run benchmarks
go test -bench=. ./...
```

## Setting Up Development Environment

1. Install Go 1.21 or later
2. Install development tools:
```bash
make tools
```

3. Set up pre-commit hooks:
```bash
make init
```

4. Create a development configuration:
```bash
cp config/config.example.yaml config/config.yaml
```

## Documentation

- Document all exported types and functions
- Include examples in documentation
- Update API documentation when making changes
- Keep README.md up to date

## Issue and Feature Request Process

1. Check existing issues and pull requests
2. Use the issue template
3. Provide detailed reproduction steps
4. Include relevant logs and error messages
5. Specify your environment details

## Code Review Process

1. All code changes require review
2. Reviewers should focus on:
   - Code correctness
   - Test coverage
   - Documentation
   - Performance implications
   - Security implications

## Community

- Join our [Discord server](https://discord.gg/hephaestus)
- Follow our [Twitter](https://twitter.com/hephaestus)
- Read our [blog](https://hephaestus.dev/blog)

## License

By contributing, you agree that your contributions will be licensed under its MIT License. 