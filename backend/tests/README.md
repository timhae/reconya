# RecoNya Backend Test Suite

This directory contains the test suite for the RecoNya Go backend.

## Structure

```
tests/
├── README.md                    # This file
├── config/
│   └── test.env                # Test environment configuration
├── integration/                # Integration tests
│   ├── device_handlers_test.go # HTTP handler tests
│   └── device_service_test.go  # Service layer tests
└── testutils/                  # Test utilities and helpers
    ├── database.go             # Database setup utilities
    ├── fixtures.go             # Test data fixtures
    └── http.go                 # HTTP testing utilities
```

## Test Types

### Unit Tests
Located alongside the source code (e.g., `models/device_test.go`). These test individual functions and methods in isolation.

### Integration Tests
Located in `tests/integration/`. These test the interaction between different components, including:
- Service layer integration
- Database operations
- HTTP handlers

## Running Tests

### Using Make (Recommended)

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run tests with coverage report
make test-coverage

# Run tests with race detection
make test-race

# Watch for changes and run tests
make test-watch
```

### Using the Test Script

```bash
# Run all tests
./test.sh

# Run only unit tests
./test.sh unit

# Run only integration tests
./test.sh integration

# Run tests with coverage
./test.sh coverage

# Watch for changes
./test.sh watch
```

### Using Go Commands Directly

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Run specific package tests
go test ./models/...
go test ./tests/integration/...
```

## Test Configuration

Test configuration is handled through:
- `tests/config/test.env` - Test-specific environment variables
- `testutils.GetTestConfig()` - Programmatic test configuration

## Test Database

Tests use SQLite with in-memory databases by default for speed and isolation. Each test gets its own clean database instance.

## Test Utilities

### Database Utilities (`testutils/database.go`)
- `SetupTestDatabase()` - Creates a clean test database
- `SetupTestRepositoryFactory()` - Sets up repository factory with test DB
- `GetTestConfig()` - Returns test configuration

### Fixtures (`testutils/fixtures.go`)
- `CreateTestDevice()` - Creates test device objects
- `CreateTestNetwork()` - Creates test network objects
- `CreateTestEventLog()` - Creates test event log objects

### HTTP Utilities (`testutils/http.go`)
- `NewTestServer()` - Creates HTTP test server
- `AssertJSONResponse()` - Asserts JSON response format
- `AssertErrorResponse()` - Asserts error response format

## Writing Tests

### Unit Test Example

```go
func TestDevice_Creation(t *testing.T) {
    mac := "00:11:22:33:44:55"
    device := models.Device{
        ID:   uuid.New().String(),
        IPv4: "192.168.1.100",
        MAC:  &mac,
    }
    
    assert.NotEmpty(t, device.ID)
    assert.Equal(t, "192.168.1.100", device.IPv4)
}
```

### Integration Test Example

```go
func TestDeviceService_Integration(t *testing.T) {
    factory, cleanup := testutils.SetupTestRepositoryFactory(t)
    defer cleanup()
    
    deviceRepo := factory.NewDeviceRepository()
    // ... setup services
    
    testDevice := testutils.CreateTestDevice()
    err := deviceRepo.Save(testDevice)
    require.NoError(t, err)
    
    // ... assertions
}
```

## Coverage

Generate coverage reports with:

```bash
make test-coverage
```

This creates:
- `coverage.out` - Raw coverage data
- `coverage.html` - HTML coverage report

## CI/CD

Tests run automatically on:
- Push to main/master/develop branches
- Pull requests
- Backend code changes

See `.github/workflows/backend-tests.yml` for CI configuration.

## Best Practices

1. **Isolation**: Each test should be independent and not rely on other tests
2. **Cleanup**: Always clean up resources (use `defer cleanup()`)
3. **Descriptive Names**: Use descriptive test function names
4. **Table Tests**: Use table-driven tests for multiple similar test cases
5. **Test Data**: Use the fixtures from `testutils` for consistent test data
6. **Assertions**: Use `testify/assert` and `testify/require` for clear assertions
7. **Mocking**: Use interfaces and mocks for external dependencies when needed

## Dependencies

Test dependencies are managed in `go.mod`:
- `github.com/stretchr/testify` - Assertions and test utilities
- `github.com/DATA-DOG/go-sqlmock` - SQL mocking (for unit tests)

## Troubleshooting

### Database Lock Issues
If you see database lock errors, ensure tests are properly cleaning up database connections.

### Test Timeouts
Default timeout is 30 seconds. Increase with `-timeout` flag if needed.

### Missing Dependencies
Run `go mod download` to ensure all test dependencies are available.