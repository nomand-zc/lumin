package parser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nomand-zc/lumin/providers"
	"github.com/nomand-zc/lumin/utils"
)

// toolCallRequestParser 处理标准 tool_call_request 事件
type toolCallRequestParser struct{}

func init() {
	Register(&toolCallRequestParser{})
}

func (p *toolCallRequestParser) MessageType() string { return MessageTypeEvent }
func (p *toolCallRequestParser) EventType() string   { return EventTypeToolCallRequest }

func (p *toolCallRequestParser) Parse(ctx context.Context, msg *StreamMessage, opts ...OptionFunc) (*providers.Response, error) {
	var data struct {
		ToolCallID string          `json:"toolCallId"`
		ToolName   string          `json:"toolName"`
		Input      json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(msg.Payload, &data); err != nil {
		return nil, fmt.Errorf("解析 tool_call_request 事件载荷失败: %w", err)
	}

	// 解析 input：直接使用原始 JSON 字节，避免反序列化-再序列化的往返开销
	args := []byte("{}")
	if len(data.Input) > 0 && utils.Bytes2Str(data.Input) != "null" && utils.Bytes2Str(data.Input) != "{}" {
		args = data.Input
	}

	// 解析选项参数
	parseOpt := &ParseOption{}
	for _, opt := range opts {
		opt(parseOpt)
	}

	// 使用ParseOption中的ToolCallIndexManager获取工具调用索引
	var index int
	if parseOpt.ToolCallIndexManager != nil {
		index = parseOpt.ToolCallIndexManager.GetToolCallIndex(data.ToolCallID)
	}

	toolCall := providers.ToolCall{
		ID:    data.ToolCallID,
		Type:  "function",
		Index: &index, // 设置正确的索引
		Function: providers.FunctionDefinitionParam{
			Name:      data.ToolName,
			Arguments: args,
		},
	}

	return providers.NewResponse(ctx,
		providers.WithObject(providers.ObjectChatCompletion),
		providers.WithIsPartial(false),
		providers.WithChoices(providers.Choice{
			Index: 0,
			Delta: providers.Message{
				Role:      providers.RoleAssistant,
				ToolCalls: []providers.ToolCall{toolCall},
			},
		}),
	), nil
}
