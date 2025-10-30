package cmdutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFriendlyError(t *testing.T) {
	t.Run("basic error with suggestion", func(t *testing.T) {
		err := &UserFriendlyError{
			Err:        errors.New("something went wrong"),
			Suggestion: "try this instead",
		}
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "something went wrong")
		assert.Contains(t, errMsg, "Suggestion: try this instead")
	})
	
	t.Run("error with docs URL", func(t *testing.T) {
		err := &UserFriendlyError{
			Err:        errors.New("config error"),
			Suggestion: "check your config",
			DocsURL:    "https://example.com/docs",
		}
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "Learn more: https://example.com/docs")
	})
	
	t.Run("unwrap", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := &UserFriendlyError{
			Err: originalErr,
		}
		
		assert.Equal(t, originalErr, errors.Unwrap(err))
	})
}

func TestNewUserFriendlyError(t *testing.T) {
	err := NewUserFriendlyError(errors.New("test error"), "test suggestion")
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
	assert.Contains(t, err.Error(), "test suggestion")
}

func TestNetworkError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := NetworkError(nil)
		assert.NoError(t, err)
	})
	
	t.Run("timeout error", func(t *testing.T) {
		err := NetworkError(errors.New("connection timeout"))
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "timeout")
		assert.Contains(t, strings.ToLower(errMsg), "network")
	})
	
	t.Run("connection refused", func(t *testing.T) {
		err := NetworkError(errors.New("connection refused"))
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "connection refused")
		assert.Contains(t, strings.ToLower(errMsg), "github.com")
	})
	
	t.Run("dns error", func(t *testing.T) {
		err := NetworkError(errors.New("no such host"))
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, strings.ToLower(errMsg), "dns")
	})
	
	t.Run("includes docs URL", func(t *testing.T) {
		err := NetworkError(errors.New("network error"))
		require.Error(t, err)
		
		var ufe *UserFriendlyError
		require.True(t, errors.As(err, &ufe))
		assert.NotEmpty(t, ufe.DocsURL)
	})
}

func TestAuthenticationError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := AuthenticationError(nil)
		assert.NoError(t, err)
	})
	
	t.Run("auth error with suggestion", func(t *testing.T) {
		err := AuthenticationError(errors.New("authentication failed"))
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "gh auth login")
		assert.Contains(t, errMsg, "authentication failed")
	})
	
	t.Run("correct exit code", func(t *testing.T) {
		err := AuthenticationError(errors.New("auth failed"))
		require.Error(t, err)
		
		var ufe *UserFriendlyError
		require.True(t, errors.As(err, &ufe))
		assert.Equal(t, 4, ufe.ExitCode)
	})
}

func TestPermissionError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := PermissionError(nil, "repo")
		assert.NoError(t, err)
	})
	
	t.Run("permission error with scope", func(t *testing.T) {
		err := PermissionError(errors.New("forbidden"), "repo")
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "gh auth refresh")
		assert.Contains(t, errMsg, "repo")
	})
	
	t.Run("correct exit code", func(t *testing.T) {
		err := PermissionError(errors.New("forbidden"), "repo")
		require.Error(t, err)
		
		var ufe *UserFriendlyError
		require.True(t, errors.As(err, &ufe))
		assert.Equal(t, 4, ufe.ExitCode)
	})
}

func TestNotFoundError(t *testing.T) {
	t.Run("repository not found", func(t *testing.T) {
		err := NotFoundError("repository")
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "repository not found")
		assert.Contains(t, errMsg, "exists")
		assert.Contains(t, errMsg, "access")
	})
	
	t.Run("issue not found", func(t *testing.T) {
		err := NotFoundError("issue")
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "issue not found")
	})
}

func TestRateLimitError(t *testing.T) {
	t.Run("without reset time", func(t *testing.T) {
		err := RateLimitError("")
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, strings.ToLower(errMsg), "rate limit")
		assert.Contains(t, errMsg, "authenticating")
	})
	
	t.Run("with reset time", func(t *testing.T) {
		err := RateLimitError("2025-10-30 10:00:00")
		require.Error(t, err)
		
		errMsg := err.Error()
		assert.Contains(t, errMsg, "2025-10-30 10:00:00")
		assert.Contains(t, errMsg, "reset at")
	})
	
	t.Run("includes docs URL", func(t *testing.T) {
		err := RateLimitError("")
		require.Error(t, err)
		
		var ufe *UserFriendlyError
		require.True(t, errors.As(err, &ufe))
		assert.Contains(t, ufe.DocsURL, "rate-limiting")
	})
}

func TestIsUserFriendlyError(t *testing.T) {
	t.Run("user friendly error", func(t *testing.T) {
		err := &UserFriendlyError{
			Err: errors.New("test"),
		}
		assert.True(t, IsUserFriendlyError(err))
	})
	
	t.Run("wrapped user friendly error", func(t *testing.T) {
		err := NewUserFriendlyError(errors.New("test"), "suggestion")
		assert.True(t, IsUserFriendlyError(err))
	})
	
	t.Run("regular error", func(t *testing.T) {
		err := errors.New("regular error")
		assert.False(t, IsUserFriendlyError(err))
	})
	
	t.Run("nil error", func(t *testing.T) {
		assert.False(t, IsUserFriendlyError(nil))
	})
}

func TestGetExitCode(t *testing.T) {
	t.Run("nil error returns 0", func(t *testing.T) {
		assert.Equal(t, 0, GetExitCode(nil))
	})
	
	t.Run("user friendly error with custom exit code", func(t *testing.T) {
		err := &UserFriendlyError{
			Err:      errors.New("test"),
			ExitCode: 42,
		}
		assert.Equal(t, 42, GetExitCode(err))
	})
	
	t.Run("authentication error returns 4", func(t *testing.T) {
		err := AuthenticationError(errors.New("auth failed"))
		assert.Equal(t, 4, GetExitCode(err))
	})
	
	t.Run("regular error returns 1", func(t *testing.T) {
		err := errors.New("regular error")
		assert.Equal(t, 1, GetExitCode(err))
	})
	
	t.Run("wrapped user friendly error", func(t *testing.T) {
		err := NewUserFriendlyError(errors.New("test"), "suggestion")
		assert.Equal(t, 1, GetExitCode(err))
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("errors.Is works with wrapped errors", func(t *testing.T) {
		originalErr := errors.New("original")
		wrappedErr := &UserFriendlyError{
			Err:        originalErr,
			Suggestion: "try this",
		}
		
		assert.True(t, errors.Is(wrappedErr, originalErr))
	})
	
	t.Run("errors.As works with nested errors", func(t *testing.T) {
		err := AuthenticationError(errors.New("auth failed"))
		
		var ufe *UserFriendlyError
		require.True(t, errors.As(err, &ufe))
		assert.Equal(t, 4, ufe.ExitCode)
	})
}
