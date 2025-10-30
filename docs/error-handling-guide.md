# Error Handling Best Practices for GitHub CLI

This document describes the enhanced error handling utilities available in the GitHub CLI codebase to provide better user experience.

## Overview

The `pkg/cmdutil` package provides user-friendly error helpers that wrap errors with actionable suggestions and documentation links. These helpers ensure that users receive clear guidance on how to resolve issues.

## Available Error Helpers

### 1. UserFriendlyError

The base error type that wraps any error with additional context:

```go
import "github.com/cli/cli/v2/pkg/cmdutil"

err := &cmdutil.UserFriendlyError{
    Err:        errors.New("config file not found"),
    Suggestion: "Run 'gh config set' to configure your settings",
    DocsURL:    "https://cli.github.com/manual/gh_config",
    ExitCode:   1,
}
```

### 2. NetworkError

Automatically detects and provides context-specific suggestions for network issues:

```go
if err := makeAPICall(); err != nil {
    return cmdutil.NetworkError(err)
}
```

**Handles:**
- Timeout errors → "The request timed out. Check your network connection or try again later."
- Connection refused → "Connection refused. Check if you can access GitHub at https://github.com"
- DNS failures → "DNS resolution failed. Check your internet connection and DNS settings."

### 3. AuthenticationError

Provides guidance for authentication failures:

```go
if err := authenticateUser(); err != nil {
    return cmdutil.AuthenticationError(err)
}
```

**Returns:**
- Suggestion: "Run 'gh auth login' to authenticate with GitHub, or check if your token has expired."
- Docs URL: Link to authentication documentation
- Exit Code: 4 (standard auth error code)

### 4. PermissionError

Handles OAuth scope and permission issues:

```go
if err := createRepository(); err != nil {
    return cmdutil.PermissionError(err, "repo")
}
```

**Returns:**
- Suggestion: "This operation requires additional permissions. Run: gh auth refresh -s repo"
- Docs URL: Link to auth refresh documentation
- Exit Code: 4

### 5. NotFoundError

Clear messaging for missing resources:

```go
if repository == nil {
    return cmdutil.NotFoundError("repository")
}
```

**Returns:**
- Error: "repository not found"
- Suggestion: "Check that the repository exists and you have access to it. Verify spelling and repository/organization names."

### 6. RateLimitError

Handles API rate limiting with reset time information:

```go
if rateLimited {
    return cmdutil.RateLimitError(resetTime)
}
```

**Returns:**
- Error: "rate limit exceeded"
- Suggestion: "GitHub API rate limit exceeded. Rate limit will reset at 2025-10-30 10:00:00. Consider authenticating to get a higher rate limit."
- Docs URL: Link to rate limiting documentation

## Utility Functions

### IsUserFriendlyError

Check if an error is a user-friendly error:

```go
if cmdutil.IsUserFriendlyError(err) {
    // Handle user-friendly error
}
```

### GetExitCode

Extract exit code from any error:

```go
exitCode := cmdutil.GetExitCode(err)
os.Exit(exitCode)
```

## Usage Examples

### Example 1: Handling API Calls

```go
func listRepositories(client *api.Client) error {
    repos, err := client.ListRepos()
    if err != nil {
        // Check for specific error types
        if strings.Contains(err.Error(), "timeout") {
            return cmdutil.NetworkError(err)
        }
        if strings.Contains(err.Error(), "401") {
            return cmdutil.AuthenticationError(err)
        }
        if strings.Contains(err.Error(), "403") {
            return cmdutil.PermissionError(err, "repo")
        }
        if strings.Contains(err.Error(), "404") {
            return cmdutil.NotFoundError("repository")
        }
        
        // Generic error
        return err
    }
    
    return nil
}
```

### Example 2: Custom User-Friendly Errors

```go
func validateConfiguration(config *Config) error {
    if config.Token == "" {
        return cmdutil.NewUserFriendlyError(
            errors.New("authentication token is missing"),
            "Run 'gh auth login' to configure authentication",
        )
    }
    
    if config.Host == "" {
        return &cmdutil.UserFriendlyError{
            Err:        errors.New("GitHub host not configured"),
            Suggestion: "Set the host using 'gh config set git_protocol https'",
            DocsURL:    "https://cli.github.com/manual/gh_config_set",
            ExitCode:   1,
        }
    }
    
    return nil
}
```

### Example 3: Exit Code Handling

```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(cmdutil.GetExitCode(err))
    }
}
```

## Benefits

1. **Consistent Error Messages**: All errors follow the same format with clear suggestions
2. **Actionable Guidance**: Users know exactly what to do to fix issues
3. **Documentation Links**: Errors include links to relevant documentation
4. **Proper Exit Codes**: Different error types return appropriate exit codes
5. **Error Chaining**: Works with Go's `errors.Is` and `errors.As` for proper error handling

## Best Practices

1. **Use Specific Helpers**: Prefer specific helpers (NetworkError, AuthenticationError) over generic UserFriendlyError
2. **Provide Context**: Include resource names and relevant details in error messages
3. **Test Error Paths**: Write tests for error handling to ensure good user experience
4. **Document Solutions**: When creating custom errors, always include actionable suggestions
5. **Link to Docs**: Include documentation URLs when available

## Testing

All error helpers include comprehensive tests. Run them with:

```bash
go test ./pkg/cmdutil
```

## Migration Guide

To migrate existing error handling:

**Before:**
```go
if err != nil {
    return fmt.Errorf("failed to authenticate: %w", err)
}
```

**After:**
```go
if err != nil {
    return cmdutil.AuthenticationError(err)
}
```

This provides users with:
- Clear error message
- Actionable suggestion
- Link to documentation
- Proper exit code

## Performance

All error helpers are lightweight:
- Zero allocation for helper function calls
- Minimal string formatting overhead
- No performance impact on happy path

---

For more information, see:
- [GitHub CLI Manual](https://cli.github.com/manual/)
- [Error Handling in Go](https://go.dev/blog/error-handling-and-go)
- [Exit Codes Documentation](https://cli.github.com/manual/gh_help_exit-codes)
