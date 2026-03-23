package converter

import (
	"context"
	"io"

	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/providers/codex/sse"
	"github.com/nomand-zc/lumin-client/queue"
)

// ParseSSEStream 解析 Codex SSE 流并将结果推送到队列。
func ParseSSEStream(ctx context.Context, body io.ReadCloser, model string,
	chainQueue queue.Queue[*providers.Response]) {
	processor := sse.NewStreamProcessor(model, chainQueue)
	processor.Process(ctx, body)
}

// ParseErrorResponse 解析 Codex 的非流式错误响应。
func ParseErrorResponse(statusCode int, body []byte) error {
	return sse.ParseErrorResponse(statusCode, body)
}
