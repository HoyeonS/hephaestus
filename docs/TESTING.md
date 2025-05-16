# Testing Guide

## Overview
This document describes the testing strategy and procedures for the Hephaestus project. The testing suite is organized into three main categories:
- Unit Tests
- Integration Tests
- End-to-End Tests

## Test Structure
```
tests/
├── unit/
│   ├── collector/
│   │   ├── collector_test.go
│   │   ├── parser_test.go
│   │   └── detector_test.go
│   ├── analyzer/
│   │   └── analyzer_test.go
│   ├── generator/
│   │   └── generator_test.go
│   ├── deployment/
│   │   └── deployment_test.go
│   └── knowledge/
│       └── knowledge_test.go
├── integration/
│   ├── collector_analyzer_test.go
│   ├── analyzer_generator_test.go
│   └── generator_deployment_test.go
└── e2e/
    └── full_pipeline_test.go
```

## Running Tests

### Prerequisites
- Go 1.21 or higher
- Docker (for integration tests)
- Access to test configuration files

### Commands

1. Run all tests:
```bash
go test ./...
```

2. Run specific test category:
```bash
# Unit tests only
go test ./tests/unit/...

# Integration tests only
go test ./tests/integration/...

# E2E tests only
go test ./tests/e2e/...
```

3. Run with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

4. Run tests in verbose mode:
```bash
go test -v ./...
```

## Test Categories

### Unit Tests
Unit tests focus on testing individual components in isolation. Each component has its own test file with comprehensive test cases.

#### Coverage Requirements
- Minimum 80% code coverage for each component
- All public methods must have test cases
- Error cases must be tested
- Edge cases must be covered

### Integration Tests
Integration tests verify the interaction between components. They ensure that data flows correctly through the system.

#### Test Scenarios
- Collector to Analyzer data flow
- Analyzer to Generator data flow
- Generator to Deployment data flow
- Knowledge Base integration

### End-to-End Tests
E2E tests verify the entire system works together. They simulate real-world scenarios.

#### Test Scenarios
- Complete error detection and fix pipeline
- Configuration changes
- Error pattern updates
- System recovery scenarios

## Writing Tests

### Test File Structure
```go
package mypackage_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Test file structure
func TestComponent(t *testing.T) {
    // Setup
    // Test cases
    // Teardown
}
```

### Best Practices
1. Use table-driven tests for multiple test cases
2. Mock external dependencies
3. Clean up resources in teardown
4. Use meaningful test names
5. Add comments explaining complex test scenarios

### Example Test Pattern
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Mocking

### Mock Generation
We use `mockery` to generate mocks:
```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks
mockery --all --dir internal --output tests/mocks
```

### Using Mocks
```go
func TestWithMocks(t *testing.T) {
    mockService := &mocks.Service{}
    mockService.On("Method", mock.Anything).Return("result", nil)
    
    // Use mock in test
}
```

## CI/CD Integration

### GitHub Actions
Tests are automatically run in CI:
- On pull requests
- On merge to main
- Nightly for long-running tests

### Test Reports
- Coverage reports are generated and uploaded
- Test results are published to GitHub
- Failed tests block merges

## Troubleshooting

### Common Issues
1. **Tests Hanging**: Check for unclosed channels or goroutines
2. **Flaky Tests**: Look for race conditions or timing issues
3. **Resource Leaks**: Ensure cleanup in teardown

### Debug Tools
- Use `-race` flag for race detection
- Use `-v` flag for verbose output
- Use logging for complex scenarios

## Adding New Tests

### Checklist
1. [ ] Create test file in appropriate directory
2. [ ] Add both positive and negative test cases
3. [ ] Mock external dependencies
4. [ ] Add comments explaining test scenarios
5. [ ] Verify coverage meets requirements
6. [ ] Add any necessary test fixtures
7. [ ] Update test documentation if needed

### Review Process
- Tests are reviewed as part of PR process
- Coverage reports must be included
- Test plan must be documented
- Edge cases must be considered 