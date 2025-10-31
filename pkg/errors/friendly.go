package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// UserFriendlyError 包装错误并添加用户友好的建议
type UserFriendlyError struct {
	Err        error
	Message    string
	Suggestion string
}

func (e *UserFriendlyError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s\n\n💡 提示：%s", e.Message, e.Suggestion)
	}
	return e.Message
}

func (e *UserFriendlyError) Unwrap() error {
	return e.Err
}

// WrapWithSuggestion 包装错误并添加建议
func WrapWithSuggestion(err error, message, suggestion string) error {
	if err == nil {
		return nil
	}
	return &UserFriendlyError{
		Err:        err,
		Message:    message,
		Suggestion: suggestion,
	}
}

// HTTPErrorWithSuggestion 为 HTTP 错误提供友好的建议
func HTTPErrorWithSuggestion(err error, statusCode int) error {
	if err == nil {
		return nil
	}

	var suggestion string
	var message string

	switch statusCode {
	case http.StatusUnauthorized:
		message = "身份验证失败"
		suggestion = "请检查您的 GitHub 令牌是否有效。运行 'gh auth status' 查看状态，或 'gh auth login' 重新登录"
	case http.StatusForbidden:
		message = "权限不足"
		suggestion = "您可能没有足够的权限执行此操作。请检查：\n  • 您是否有仓库的访问权限\n  • 您的令牌是否有所需的作用域\n  • 运行 'gh auth refresh -s <scope>' 添加所需权限"
	case http.StatusNotFound:
		message = "资源未找到"
		suggestion = "请检查：\n  • 仓库名称是否正确\n  • 资源（PR/Issue/Release 等）是否存在\n  • 您是否有访问权限"
	case http.StatusUnprocessableEntity:
		message = "请求无法处理"
		suggestion = "请检查输入的参数是否正确。某些操作可能有特定要求（例如合并 PR 需要所有检查通过）"
	case http.StatusTooManyRequests:
		message = "API 请求速率限制"
		suggestion = "您已达到 GitHub API 速率限制。请稍等片刻再试，或查看限制状态：'gh api rate_limit'"
	case http.StatusServiceUnavailable, http.StatusBadGateway:
		message = "GitHub 服务暂时不可用"
		suggestion = "这通常是暂时性问题。请稍后重试，或查看 GitHub 状态：https://www.githubstatus.com"
	default:
		if statusCode >= 500 {
			message = fmt.Sprintf("服务器错误 (HTTP %d)", statusCode)
			suggestion = "这是 GitHub 服务端错误。请稍后重试，或查看 GitHub 状态页面"
		} else {
			message = fmt.Sprintf("请求失败 (HTTP %d)", statusCode)
			suggestion = "请检查命令参数是否正确，或使用 --help 查看用法"
		}
	}

	return &UserFriendlyError{
		Err:        err,
		Message:    message,
		Suggestion: suggestion,
	}
}

// IsUserFriendlyError 检查错误是否是 UserFriendlyError
func IsUserFriendlyError(err error) bool {
	var ufe *UserFriendlyError
	return errors.As(err, &ufe)
}

// GetSuggestion 获取错误的建议（如果有）
func GetSuggestion(err error) string {
	var ufe *UserFriendlyError
	if errors.As(err, &ufe) {
		return ufe.Suggestion
	}
	return ""
}
