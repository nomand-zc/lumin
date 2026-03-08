package parser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nomand-zc/lumin/providers"
)

// completionChunkParser 处理 completion_chunk 事件（流式增量）
type completionChunkParser struct{}

func init() {
	Register(&completionChunkParser{})
}

func (p *completionChunkParser) MessageType() string { return MessageTypeEvent }
func (p *completionChunkParser) EventType() string   { return EventTypeCompletionChunk }

func (p *completionChunkParser) Parse(ctx context.Context, msg *StreamMessage, opts ...OptionFunc) (*providers.Response, error) {
	var data struct {
		Content      string `json:"content"`
		Delta        string `json:"delta"`
		FinishReason string `json:"finish_reason"`
	}
	if err := json.Unmarshal(msg.Payload, &data); err != nil {
		return nil, fmt.Errorf("解析 completion_chunk 事件载荷失败: %w", err)
	}

	// 使用 delta 作为实际的文本增量，如果没有则使用 content
	textDelta := data.Delta
	if textDelta == "" {
		textDelta = data.Content
	}

	resp := providers.NewResponse(ctx,
		providers.WithObject(providers.ObjectChatCompletion),
		providers.WithChoices(providers.Choice{
			Index: 0,
			Delta: providers.Message{
				Role:    providers.RoleAssistant,
				Content: textDelta,
			},
		}),
	)

	// 如果有完成原因，标记为最终响应
	if data.FinishReason != "" {
		resp.Choices[0].FinishReason = &data.FinishReason
		resp.Done = true
		resp.IsPartial = false
	}

	return resp, nil
}
