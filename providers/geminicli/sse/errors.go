package sse

import (
	"encoding/json"
	"fmt"

	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/providers/geminicli/types"
)

// ParseErrorResponse 解析 Gemini CLI 的非流式错误响应
func ParseErrorResponse(statusCode int, body []byte) error {
	var errResp types.GeminiCLIErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		msg := fmt.Sprintf("Gemini CLI API error [%d]: %s", errResp.Error.Code, errResp.Error.Message)
		return newHTTPError(statusCode, msg, body)
	}

	return newHTTPError(statusCode,
		fmt.Sprintf("Gemini CLI API error, status=%d", statusCode), body)
}

// classifyHTTPStatus 根据 HTTP 状态码分类错误类型和错误码
func classifyHTTPStatus(code int) (providers.ErrorType, int) {
	switch code {
	case 429:
		return providers.ErrorTypeRateLimit, providers.ErrorCodeRateLimit
	case 401:
		return providers.ErrorTypeUnauthorized, providers.ErrorCodeUnauthorized
	case 403:
		return providers.ErrorTypeForbidden, providers.ErrorCodeForbidden
	case 400:
		return providers.ErrorTypeBadRequest, providers.ErrorCodeBadRequest
	default:
		return providers.ErrorTypeServerError, providers.ErrorCodeServerError
	}
}

// newHTTPError 根据状态码创建标准化的 HTTPError
func newHTTPError(statusCode int, message string, body []byte) *providers.HTTPError {
	errType, errCode := classifyHTTPStatus(statusCode)
	return &providers.HTTPError{
		ErrorType:     errType,
		ErrorCode:     errCode,
		Message:       message,
		RawStatusCode: statusCode,
		RawBody:       body,
	}
}
