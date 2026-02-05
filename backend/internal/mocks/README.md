# Mocks for Backend Interfaces

This directory contains auto-generated mocks for all internal interfaces using [Mockery](https://vektra.github.io/mockery/).

## Generated Interfaces

The following 24 mock interfaces are available for testing:

### Repository Mocks

- **OrderRepository** - Mock for order data access
- **ProductRepository** - Mock for product data access

### Service Mocks

- **OrderService** - Mock for order business logic
- **ProductService** - Mock for product business logic
- **TokenService** - Mock for token management
- **TokenRefreshService** - Mock for token refresh operations
- **CacheManager** - Mock for caching operations

### Platform API Mocks

- **APIClient** - Mock for platform API client
- **PlatformAPI** - Mock for platform API operations
- **PlatformSKUChecker** - Mock for SKU validation
- **PlatformWholesaleAPI** - Mock for wholesale API operations
- **TikTokClient** - Mock for TikTok platform client
- **shippingClient** - Mock for shipping operations

### Utility Mocks

- **ConfigManager** - Mock for configuration management
- **ClientFactory** - Mock for client creation
- **DBGetter** - Mock for database access utilities
- **FunctionHandler** - Mock for function execution
- **JobHandler** - Mock for job execution
- **OrderManager** - Mock for order management
- **SheetsClient** - Mock for Google Sheets operations
- **SheetsClientInterface** - Mock for Sheets client interface
- **SheetWriterClient** - Mock for Sheets writing operations

### Callback Mocks

- **ProgressCallback** - Mock for progress tracking callbacks
- **TokenRefreshFunc** - Mock for token refresh callbacks

## Usage

### Using Mocks in Tests

```go
package services_test

import (
    "testing"
    "github.com/omni/backend/internal/mocks"
    "github.com/stretchr/testify/require"
)

func TestOrderService(t *testing.T) {
    // Create a mock repository
    mockRepo := new(mocks.OrderRepository)

    // Set expectations
    mockRepo.On("GetOrder", mock.Anything, "order123").
        Return(&domain.Order{ID: "order123"}, nil)

    // Use the mock in your service
    service := services.NewOrderService(mockRepo)
    order, err := service.GetOrder(context.Background(), "order123")

    // Verify
    require.NoError(t, err)
    require.Equal(t, "order123", order.ID)
    mockRepo.AssertExpectations(t)
}
```

### Expected Mock Behavior

All generated mocks implement the original interface exactly, supporting:

- Method call expectations with `On()`
- Argument matchers with `mock.Anything`, `mock.MatchedBy()`, etc.
- Return value specification with `Return()`
- Call count verification with `AssertExpectations()` and `AssertCalled()`
- Run-time behavior customization with `Run()`

## Regenerating Mocks

When you add or modify interfaces in the `internal/` directory, regenerate all mocks:

```bash
# From the backend directory
mockery --all --dir ./internal --output internal/mocks --outpkg mocks

# Or use the Makefile target (if make is available)
make mocks
```

## Configuration

Mocks are configured in `.mockery.yaml`:

```yaml
output: internal/mocks
outpkg: mocks
all: true
```

This configuration:

- Outputs all mocks to `internal/mocks/`
- Uses the `mocks` package name
- Generates mocks for ALL interfaces found in the specified directory

## Dependencies

- **mockery** v2.53.5 - Interface mock generation tool
- **testify/mock** - Mock assertion library

## Troubleshooting

### Mocks not compiling

1. Check that all interface definitions are in `internal/` subdirectories
2. Run `go mod tidy` to ensure dependencies are correct
3. Verify interfaces don't use unsupported types (e.g., unexported types from other packages)

### Missing a mock?

1. Ensure the interface is in a file under `internal/`
2. Rerun the mock generation command
3. Check for any syntax errors in interface definitions

## Best Practices

1. **Use mocks for unit tests only** - For integration tests, prefer real implementations
2. **Set clear expectations** - Mock behaviors should match production expectations
3. **Verify calls** - Use `AssertExpectations()` to ensure mocks were used correctly
4. **Mock at boundaries** - Mock external dependencies, not internal logic
5. **Keep mocks updated** - Regenerate mocks whenever interfaces change

## Related

- Original interfaces: `internal/*/interfaces.go`
- Test examples: `tests/unit/`
- Test configuration: `backend/Makefile`
