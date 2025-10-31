package errors

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestUserFriendlyError_Error(t *testing.T) {
	tests := []struct {
		name       string
		err        *UserFriendlyError
		wantString string
	}{
		{
			name: "带建议的错误",
			err: &UserFriendlyError{
				Err:        errors.New("原始错误"),
				Message:    "操作失败",
				Suggestion: "请重试",
			},
			wantString: "操作失败\n\n💡 提示：请重试",
		},
		{
			name: "无建议的错误",
			err: &UserFriendlyError{
				Err:     errors.New("原始错误"),
				Message: "操作失败",
			},
			wantString: "操作失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantString {
				t.Errorf("Error() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestWrapWithSuggestion(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		message    string
		suggestion string
		wantNil    bool
	}{
		{
			name:       "包装非空错误",
			err:        errors.New("原始错误"),
			message:    "操作失败",
			suggestion: "请重试",
			wantNil:    false,
		},
		{
			name:       "包装空错误",
			err:        nil,
			message:    "操作失败",
			suggestion: "请重试",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapWithSuggestion(tt.err, tt.message, tt.suggestion)
			if (got == nil) != tt.wantNil {
				t.Errorf("WrapWithSuggestion() = %v, wantNil %v", got, tt.wantNil)
			}
			if !tt.wantNil {
				if !strings.Contains(got.Error(), tt.message) {
					t.Errorf("错误消息应包含 %q", tt.message)
				}
				if !strings.Contains(got.Error(), tt.suggestion) {
					t.Errorf("错误消息应包含建议 %q", tt.suggestion)
				}
			}
		})
	}
}

func TestHTTPErrorWithSuggestion(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		statusCode     int
		wantNil        bool
		wantContains   []string
	}{
		{
			name:         "401 未授权",
			err:          errors.New("unauthorized"),
			statusCode:   http.StatusUnauthorized,
			wantNil:      false,
			wantContains: []string{"身份验证", "gh auth"},
		},
		{
			name:         "403 禁止访问",
			err:          errors.New("forbidden"),
			statusCode:   http.StatusForbidden,
			wantNil:      false,
			wantContains: []string{"权限", "访问权限"},
		},
		{
			name:         "404 未找到",
			err:          errors.New("not found"),
			statusCode:   http.StatusNotFound,
			wantNil:      false,
			wantContains: []string{"未找到", "检查"},
		},
		{
			name:         "429 速率限制",
			err:          errors.New("rate limit"),
			statusCode:   http.StatusTooManyRequests,
			wantNil:      false,
			wantContains: []string{"速率限制", "rate_limit"},
		},
		{
			name:         "500 服务器错误",
			err:          errors.New("server error"),
			statusCode:   http.StatusInternalServerError,
			wantNil:      false,
			wantContains: []string{"服务器错误", "500"},
		},
		{
			name:       "空错误",
			err:        nil,
			statusCode: http.StatusOK,
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTTPErrorWithSuggestion(tt.err, tt.statusCode)
			if (got == nil) != tt.wantNil {
				t.Errorf("HTTPErrorWithSuggestion() = %v, wantNil %v", got, tt.wantNil)
			}
			if !tt.wantNil {
				gotStr := got.Error()
				for _, want := range tt.wantContains {
					if !strings.Contains(gotStr, want) {
						t.Errorf("错误消息应包含 %q，实际为：%s", want, gotStr)
					}
				}
			}
		})
	}
}

func TestIsUserFriendlyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "是 UserFriendlyError",
			err: &UserFriendlyError{
				Err:     errors.New("test"),
				Message: "test",
			},
			want: true,
		},
		{
			name: "不是 UserFriendlyError",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "包装的 UserFriendlyError",
			err: WrapWithSuggestion(errors.New("test"), "message", "suggestion"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUserFriendlyError(tt.err); got != tt.want {
				t.Errorf("IsUserFriendlyError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSuggestion(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "有建议",
			err:  WrapWithSuggestion(errors.New("test"), "message", "这是建议"),
			want: "这是建议",
		},
		{
			name: "无建议的 UserFriendlyError",
			err: &UserFriendlyError{
				Err:     errors.New("test"),
				Message: "message",
			},
			want: "",
		},
		{
			name: "标准错误",
			err:  errors.New("standard error"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSuggestion(tt.err); got != tt.want {
				t.Errorf("GetSuggestion() = %v, want %v", got, tt.want)
			}
		})
	}
}
