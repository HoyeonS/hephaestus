# Contributing to Hephaestus

We love your input! We want to make contributing to Hephaestus as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code lints
6. Issue that pull request!

## Pull Request Process

1. Update the README.md with details of changes to the interface, if applicable
2. Update the docs/ with any necessary documentation
3. The PR will be merged once you have the sign-off of two other developers
4. If you haven't already, complete the Contributor License Agreement ("CLA")

## Any contributions you make will be under the MIT Software License

In short, when you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.

## Report bugs using GitHub's [issue tracker](https://github.com/HoyeonS/hephaestus/issues)

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/HoyeonS/hephaestus/issues/new); it's that easy!

## Write bug reports with detail, background, and sample code

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Code Style Guidelines

### Go Code

Follow the official Go style guide and common practices:

1. Use `gofmt` to format your code
2. Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
3. Document all exported functions, types, and packages
4. Write meaningful commit messages
5. Include tests for new functionality

### Protocol Buffers

1. Use consistent naming:
   - Service names: PascalCase
   - Method names: PascalCase
   - Message names: PascalCase
   - Field names: snake_case

2. Include comments for all messages and fields

### Project Structure

```
hephaestus/
├── cmd/                    # Command-line applications
├── internal/              # Internal packages
├── pkg/                   # Public packages
├── proto/                 # Protocol Buffers
├── docs/                  # Documentation
└── examples/              # Example code
```

## Testing Guidelines

1. **Unit Tests**
   - Test all exported functions
   - Use table-driven tests
   - Mock external dependencies
   - Aim for >80% coverage

2. **Integration Tests**
   - Test component interactions
   - Use docker-compose for dependencies
   - Clean up test resources

3. **Performance Tests**
   - Benchmark critical paths
   - Test with realistic data volumes
   - Monitor resource usage

## Documentation Guidelines

1. **Code Documentation**
   - Document all exported symbols
   - Include examples where appropriate
   - Explain complex algorithms
   - Document assumptions

2. **API Documentation**
   - Document all API endpoints
   - Include request/response examples
   - Document error conditions
   - Keep OpenAPI spec updated

3. **Architecture Documentation**
   - Update HLD/LLD for major changes
   - Include diagrams where helpful
   - Document design decisions

## Commit Message Guidelines

Format:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation only changes
- style: Changes that do not affect the meaning of the code
- refactor: Code change that neither fixes a bug nor adds a feature
- perf: Code change that improves performance
- test: Adding missing tests
- chore: Changes to the build process or auxiliary tools

Example:
```
feat(log-processor): add support for structured logging

- Add JSON log format support
- Include timestamp and correlation ID
- Update documentation

Closes #123
```

## Branch Naming Convention

- Feature branches: `feature/description`
- Bug fixes: `fix/description`
- Documentation: `docs/description`
- Performance improvements: `perf/description`

## Release Process

1. **Version Bump**
   - Update version in code
   - Update CHANGELOG.md
   - Create release notes

2. **Testing**
   - Run full test suite
   - Perform integration tests
   - Check documentation

3. **Release**
   - Tag release in git
   - Create GitHub release
   - Update documentation

## Getting Help

- Join our [Discord server](https://discord.gg/hephaestus)
- Check the [FAQ](docs/FAQ.md)
- Ask in GitHub Issues
- Contact the maintainers

## Code of Conduct

### Our Pledge

We pledge to make participation in our project and our community a harassment-free experience for everyone.

### Our Standards

Examples of behavior that contributes to creating a positive environment include:

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

### Our Responsibilities

Project maintainers are responsible for clarifying the standards of acceptable behavior and are expected to take appropriate and fair corrective action in response to any instances of unacceptable behavior.

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team. All complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances.

## License

By contributing, you agree that your contributions will be licensed under its MIT License. 