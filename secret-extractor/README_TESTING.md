# Testing Guide for Supabase Secret Extractor

This document describes the testing strategy and how to run tests for the secret extractor.

## Test Structure

```
secret-extractor/
├── pkg/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go           # Unit tests for configuration
│   ├── extractor/
│   │   ├── extractor.go
│   │   ├── extractor_test.go        # Unit tests for extractor logic
│   │   ├── kong.go
│   │   ├── kong_test.go             # Unit tests for Kong parsing
│   │   ├── validator.go
│   │   └── validator_test.go        # Unit tests for validation
│   └── k8s/
│       ├── secrets.go
│       └── secrets_test.go          # Unit tests for K8s operations
└── test_integration_test.go         # Integration tests (require K8s)
```

## Running Tests

### Unit Tests Only

Run all unit tests (no Kubernetes cluster required):

```bash
cd secret-extractor
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

### Specific Package Tests

Test a specific package:

```bash
go test ./pkg/config
go test ./pkg/extractor
go test ./pkg/k8s
```

### Integration Tests

Integration tests require a running Kubernetes cluster with Supabase deployed.

Run integration tests:

```bash
go test -tags=integration -v
```

Skip integration tests:

```bash
go test -short ./...
```

## Test Coverage

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

View coverage in terminal:

```bash
go test -cover ./...
```

Expected coverage:
- `pkg/config`: ~90%
- `pkg/extractor`: ~70% (excludes K8s integration)
- `pkg/k8s`: ~40% (requires mocking)

## Test Categories

### 1. Configuration Tests (`config_test.go`)

Tests for configuration management:
- Default configuration values
- Custom configuration
- Configuration validation

```bash
go test ./pkg/config -v
```

### 2. Validator Tests (`validator_test.go`)

Tests for secret validation:
- JWT format validation
- Secret presence validation
- Minimum length validation
- URL validation

```bash
go test ./pkg/extractor -run TestValidate -v
```

### 3. Kong Parser Tests (`kong_test.go`)

Tests for Kong configuration parsing:
- YAML parsing
- Consumer extraction
- Credential extraction
- Error handling

```bash
go test ./pkg/extractor -run TestKong -v
```

### 4. Extractor Structure Tests (`extractor_test.go`)

Tests for data structures and mappings:
- Secret structure validation
- Data type correctness
- Field accessibility

```bash
go test ./pkg/extractor -run TestExtractor -v
```

### 5. Kubernetes Tests (`secrets_test.go`)

Tests for Kubernetes resource structure:
- Secret labels
- Secret type
- Metadata structure

```bash
go test ./pkg/k8s -v
```

### 6. Integration Tests (`test_integration_test.go`)

End-to-end tests with real Kubernetes:
- Full extraction workflow
- Secret persistence
- Validation with real data

```bash
# Requires Kubernetes cluster with Supabase
go test -tags=integration -v
```

## Test Data

### Sample Kong Configuration

```yaml
_format_version: "3.0"
consumers:
  - username: anon
    keyauth_credentials:
      - key: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.anon.key
  - username: service_role
    keyauth_credentials:
      - key: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.service.key
```

### Sample JWT Token

Valid JWT format (3 parts separated by dots):
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiJ9.ZopqoUt20nEV9cklpv9e3yw3PVyZLmKs5qLD6nGL1SI
```

## Mocking Strategy

Since we can't interact with a real Kubernetes cluster in unit tests:

1. **Structure tests**: Verify data structures and types
2. **Logic tests**: Test validation and parsing logic
3. **Integration tests**: Use build tags to separate K8s-dependent tests

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run unit tests
        run: cd secret-extractor && go test -v ./...
      - name: Generate coverage
        run: cd secret-extractor && go test -coverprofile=coverage.out ./...
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./secret-extractor/coverage.out
```

## Benchmarking

Run benchmark tests:

```bash
go test -bench=. ./...
```

Example benchmark test:

```go
func BenchmarkValidateJWT(b *testing.B) {
    token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        isJWT(token)
    }
}
```

## Common Test Commands

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run specific test
go test ./pkg/extractor -run TestValidate

# Run tests matching pattern
go test ./... -run JWT

# List all tests
go test ./... -list .

# Run tests in parallel
go test -parallel 4 ./...

# Verbose output
go test -v ./...

# Short mode (skip slow tests)
go test -short ./...
```

## Debugging Tests

Enable verbose logging in tests:

```go
func TestSomething(t *testing.T) {
    t.Logf("Debug: value = %v", someValue)
    // ...
}
```

Run single test with logs:

```bash
go test ./pkg/extractor -run TestValidate -v
```

## Writing New Tests

### Test Naming Convention

- File: `*_test.go`
- Function: `TestXxx` or `Test_xxx`
- Example: `TestValidateJWTTokens`

### Table-Driven Tests

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Test Best Practices

1. **Isolation**: Each test should be independent
2. **Clear names**: Test names should describe what they test
3. **One assertion**: Test one thing at a time when possible
4. **Table-driven**: Use table-driven tests for multiple cases
5. **Cleanup**: Use `defer` for cleanup operations
6. **Error messages**: Provide clear failure messages
7. **Fast tests**: Keep unit tests fast (< 1 second)
8. **Build tags**: Use build tags to separate integration tests

## Troubleshooting

### Tests Fail: "no such file or directory"

Make sure you're in the correct directory:
```bash
cd secret-extractor
go test ./...
```

### Tests Fail: "cannot find package"

Run go mod download:
```bash
go mod download
go mod tidy
```

### Integration Tests Fail

Integration tests require:
- Running Kubernetes cluster
- Supabase deployed in `supabase-test` namespace
- Proper RBAC permissions

Skip integration tests:
```bash
go test -short ./...
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing Techniques](https://go.dev/blog/table-driven-tests)
