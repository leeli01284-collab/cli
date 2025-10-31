# Performance Optimizations and Best Practices

This document outlines the performance optimizations and best practices implemented in the GitHub CLI.

## Caching Strategy

### HTTP Response Caching

The CLI implements HTTP response caching to reduce latency and API calls:

- **Location**: `api/http_client.go`
- **Implementation**: Uses `HTTPClientOptions.EnableCache` and `CacheTTL`
- **Benefits**: 
  - Reduces redundant API calls
  - Improves response times for repeated requests
  - Respects GitHub API rate limits

#### Usage Example

```go
opts := api.HTTPClientOptions{
    AppVersion:  "1.0.0",
    EnableCache: true,
    CacheTTL:    time.Hour,
    Config:      config,
}
client, err := api.NewHTTPClient(opts)
```

### Update Check Caching

The CLI caches update checks to avoid excessive API requests:

- **Location**: `internal/update/update.go`
- **Strategy**: Checks for updates at most once every 24 hours
- **Storage**: Stores last check timestamp in `state.yml`
- **Benefits**:
  - Reduces unnecessary GitHub API calls
  - Improves CLI startup performance
  - Respects user bandwidth

## Error Handling Best Practices

### Context Propagation

All API calls should use proper context propagation for:
- Request cancellation
- Timeout management
- Request tracing

```go
ctx := context.Background()
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### User-Friendly Error Messages

Error messages should be:
- Clear and actionable
- Include relevant context
- Suggest next steps when possible

## CI/CD Environment Detection

The CLI automatically detects CI/CD environments to:
- Skip interactive prompts
- Disable update notifications
- Optimize for non-interactive execution

Detected environments:
- GitHub Actions
- Travis CI
- CircleCI
- Jenkins
- TeamCity
- And more (see `IsCI()` in `internal/update/update.go`)

## Terminal Detection

The CLI detects terminal capabilities to:
- Enable/disable color output appropriately
- Show/hide progress indicators
- Optimize display for TTY vs pipe

## Resource Optimization

### Memory Management

- Request bodies are properly closed with `defer`
- Response bodies are drained to `io.Discard` before closing
- JSON decoding uses streaming where possible

### Connection Reuse

The HTTP client reuses connections through the standard `http.Transport` pooling mechanism.

## Performance Monitoring

To enable verbose HTTP logging for debugging:

```bash
export GH_DEBUG=api
gh <command>
```

## Best Practices for Contributors

When adding new features:

1. **Use caching where appropriate** - Leverage existing cache infrastructure
2. **Propagate context** - Use `context.Context` for all API calls
3. **Close resources** - Always defer close for readers and connections
4. **Drain response bodies** - Copy to `io.Discard` before closing
5. **Add timeouts** - Set reasonable timeouts for network operations
6. **Check environment** - Use `IsCI()` and `IsTerminal()` to adjust behavior

## Migration Guide

### For Users

No action required - all optimizations are backward compatible.

### For Extension Developers

If your extension makes HTTP requests, consider:

1. Using the CLI's HTTP client for automatic caching:
   ```go
   import "github.com/cli/go-gh/v2/pkg/api"
   ```

2. Respecting the same environment variables:
   - `GH_NO_UPDATE_NOTIFIER`
   - `GH_NO_EXTENSION_UPDATE_NOTIFIER`
   - `CI`

## Benchmarking

To benchmark CLI performance:

```bash
time gh <command>
```

For more detailed profiling:

```bash
go test -bench=. -benchmem ./...
```

## Future Optimizations

Potential areas for future optimization:
- GraphQL query batching
- Parallel API requests where safe
- Local caching of repository metadata
- Compression support for large payloads
