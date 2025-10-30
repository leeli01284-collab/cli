package cmdutil

import (
	"errors"
	"fmt"
	"strings"
)

// UserFriendlyError wraps an error with additional context and suggestions
// for users on how to resolve common issues.
type UserFriendlyError struct {
	Err         error
	Suggestion  string
	DocsURL     string
	ExitCode    int
}

func (e *UserFriendlyError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Err.Error())
	
	if e.Suggestion != "" {
		sb.WriteString("\n\nSuggestion: ")
		sb.WriteString(e.Suggestion)
	}
	
	if e.DocsURL != "" {
		sb.WriteString("\nLearn more: ")
		sb.WriteString(e.DocsURL)
	}
	
	return sb.String()
}

func (e *UserFriendlyError) Unwrap() error {
	return e.Err
}

// NewUserFriendlyError creates a new error with user-friendly guidance
func NewUserFriendlyError(err error, suggestion string) error {
	return &UserFriendlyError{
		Err:        err,
		Suggestion: suggestion,
		ExitCode:   1,
	}
}

// NetworkError creates a user-friendly error for network-related issues
func NetworkError(err error) error {
	if err == nil {
		return nil
	}
	
	suggestion := "Check your internet connection and try again. If behind a proxy, ensure it's configured correctly."
	
	// Check for specific network error types
	errMsg := err.Error()
	if strings.Contains(errMsg, "timeout") {
		suggestion = "The request timed out. Check your network connection or try again later."
	} else if strings.Contains(errMsg, "connection refused") {
		suggestion = "Connection refused. Check if you can access GitHub at https://github.com"
	} else if strings.Contains(errMsg, "no such host") {
		suggestion = "DNS resolution failed. Check your internet connection and DNS settings."
	}
	
	return &UserFriendlyError{
		Err:        err,
		Suggestion: suggestion,
		DocsURL:    "https://cli.github.com/manual/gh_help_environment",
		ExitCode:   1,
	}
}

// AuthenticationError creates a user-friendly error for auth issues
func AuthenticationError(err error) error {
	if err == nil {
		return nil
	}
	
	return &UserFriendlyError{
		Err:        err,
		Suggestion: "Run 'gh auth login' to authenticate with GitHub, or check if your token has expired.",
		DocsURL:    "https://cli.github.com/manual/gh_auth_login",
		ExitCode:   4,
	}
}

// PermissionError creates a user-friendly error for permission issues
func PermissionError(err error, requiredScope string) error {
	if err == nil {
		return nil
	}
	
	suggestion := fmt.Sprintf("This operation requires additional permissions. Run: gh auth refresh -s %s", requiredScope)
	
	return &UserFriendlyError{
		Err:        err,
		Suggestion: suggestion,
		DocsURL:    "https://cli.github.com/manual/gh_auth_refresh",
		ExitCode:   4,
	}
}

// NotFoundError creates a user-friendly error for resource not found issues
func NotFoundError(resource string) error {
	return &UserFriendlyError{
		Err:        fmt.Errorf("%s not found", resource),
		Suggestion: fmt.Sprintf("Check that the %s exists and you have access to it. Verify spelling and repository/organization names.", resource),
		ExitCode:   1,
	}
}

// RateLimitError creates a user-friendly error for rate limiting
func RateLimitError(resetTime string) error {
	suggestion := "GitHub API rate limit exceeded."
	if resetTime != "" {
		suggestion += fmt.Sprintf(" Rate limit will reset at %s.", resetTime)
	}
	suggestion += " Consider authenticating to get a higher rate limit."
	
	return &UserFriendlyError{
		Err:        errors.New("rate limit exceeded"),
		Suggestion: suggestion,
		DocsURL:    "https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting",
		ExitCode:   1,
	}
}

// IsUserFriendlyError checks if an error is a UserFriendlyError
func IsUserFriendlyError(err error) bool {
	var ufe *UserFriendlyError
	return errors.As(err, &ufe)
}

// GetExitCode extracts the exit code from an error, or returns 1 as default
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}
	
	var ufe *UserFriendlyError
	if errors.As(err, &ufe) {
		return ufe.ExitCode
	}
	
	// Check for other error types with exit codes
	type exitCoder interface {
		ExitCode() int
	}
	
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	
	return 1
}
