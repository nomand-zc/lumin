package converter

import (
	"encoding/json"
	"fmt"

	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/providers/codex/types"
)

// BuildRequest 将统一 Request 转换为 Codex Responses API 请求格式
// 对齐 codex-rs/codex-api/src/common.rs 中的 ResponsesApiRequest
func BuildRequest(req *providers.Request) (*types.ResponsesAPIRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	apiReq := &types.ResponsesAPIRequest{
		Model:             req.Model,
		ToolChoice:        "auto",
		ParallelToolCalls: true,
		Store:             false,
		Stream:            true,
	}

	// 转换消息为 Responses API input items
	input, instructions := convertMessages(req.Messages)
	apiReq.Input = input
	apiReq.Instructions = instructions

	// 转换 tools
	if len(req.Tools) > 0 {
		apiReq.Tools = convertTools(req.Tools)
	}

	// 处理 tool_choice
	if req.Metadata != nil {
		if tc, ok := req.Metadata["tool_choice"].(string); ok && tc != "" {
			apiReq.ToolChoice = tc
		}
		if pt, ok := req.Metadata["parallel_tool_calls"].(bool); ok {
			apiReq.ParallelToolCalls = pt
		}
		if store, ok := req.Metadata["store"].(bool); ok {
			apiReq.Store = store
		}
		if serviceTier, ok := req.Metadata["service_tier"].(string); ok {
			apiReq.ServiceTier = serviceTier
		}
	}

	// 转换 reasoning 配置
	apiReq.Reasoning = convertReasoning(&req.GenerationConfig, req.Metadata)

	// 对齐 codex: 仅在有 reasoning 时才包含 "reasoning.encrypted_content"
	if apiReq.Reasoning != nil {
		apiReq.Include = []string{"reasoning.encrypted_content"}
	}

	// 支持从 metadata 设置 prompt_cache_key
	if req.Metadata != nil {
		if pck, ok := req.Metadata["prompt_cache_key"].(string); ok && pck != "" {
			apiReq.PromptCacheKey = pck
		}
	}

	return apiReq, nil
}

// convertMessages 将统一 Messages 转换为 Responses API 的 input items
// 返回 input items 和 instructions (来自 system 消息)
func convertMessages(messages []providers.Message) ([]types.ResponseItem, string) {
	var input []types.ResponseItem
	var instructions string

	for _, msg := range messages {
		switch msg.Role {
		case providers.RoleSystem:
			// system 消息映射为 instructions
			instructions = msg.Content
		case providers.RoleUser:
			item := convertUserMessage(&msg)
			input = append(input, item)
		case providers.RoleAssistant:
			items := convertAssistantMessage(&msg)
			input = append(input, items...)
		case providers.RoleTool:
			item := convertToolMessage(&msg)
			input = append(input, item)
		}
	}

	return input, instructions
}

// convertUserMessage 将 user 消息转换为 Responses API message item
func convertUserMessage(msg *providers.Message) types.ResponseItem {
	item := types.ResponseItem{
		Type: "message",
		Role: "user",
	}

	var content []types.ContentItem
	if msg.Content != "" {
		content = append(content, types.ContentItem{
			Type: "input_text",
			Text: msg.Content,
		})
	}
	for _, cp := range msg.ContentParts {
		switch cp.Type {
		case providers.ContentTypeText:
			if cp.Text != nil {
				content = append(content, types.ContentItem{
					Type: "input_text",
					Text: *cp.Text,
				})
			}
		case providers.ContentTypeImage:
			if cp.Image != nil && cp.Image.URL != "" {
				content = append(content, types.ContentItem{
					Type:     "input_image",
					ImageURL: cp.Image.URL,
				})
			}
		}
	}
	item.Content = content
	return item
}

// convertAssistantMessage 将 assistant 消息转换为 Responses API items
// assistant 消息可能包含文本内容和 tool_calls，需要拆分为多个 item
func convertAssistantMessage(msg *providers.Message) []types.ResponseItem {
	var items []types.ResponseItem

	// 处理 reasoning 内容
	if msg.ReasoningContent != "" {
		items = append(items, types.ResponseItem{
			Type: "reasoning",
			Summary: []types.ReasoningSummaryItem{
				{
					Type: "summary_text",
					Text: msg.ReasoningContent,
				},
			},
		})
	}

	// 处理文本内容
	if msg.Content != "" {
		items = append(items, types.ResponseItem{
			Type: "message",
			Role: "assistant",
			Content: []types.ContentItem{
				{
					Type: "output_text",
					Text: msg.Content,
				},
			},
		})
	}

	// 处理 tool calls —— 每个 tool call 转换为一个 function_call item
	for _, tc := range msg.ToolCalls {
		items = append(items, types.ResponseItem{
			Type:      "function_call",
			Name:      tc.Function.Name,
			Arguments: string(tc.Function.Arguments),
			CallID:    tc.ID,
		})
	}

	return items
}

// convertToolMessage 将 tool 消息转换为 function_call_output item
func convertToolMessage(msg *providers.Message) types.ResponseItem {
	return types.ResponseItem{
		Type:   "function_call_output",
		CallID: msg.ToolID,
		Output: msg.Content,
	}
}

// convertTools 将统一 Tool 列表转换为 Responses API 工具定义
func convertTools(tools []providers.Tool) []json.RawMessage {
	var result []json.RawMessage
	for _, t := range tools {
		tool := types.ToolFunction{
			Type: "function",
			Function: &types.ToolFunctionSpec{
				Name:        t.Name,
				Description: t.Description,
				Strict:      false,
				Parameters:  t.Parameters,
			},
		}
		data, err := json.Marshal(tool)
		if err != nil {
			continue
		}
		result = append(result, data)
	}
	return result
}

// convertReasoning 将 GenerationConfig 中的推理相关配置转换为 Reasoning
// 对齐 codex-rs/core/src/client.rs 中 build_responses_request 的 reasoning 构建逻辑
func convertReasoning(cfg *providers.GenerationConfig, metadata map[string]any) *types.Reasoning {
	if cfg == nil {
		return nil
	}

	var reasoning *types.Reasoning

	if cfg.ReasoningEffort != nil && *cfg.ReasoningEffort != "" {
		if reasoning == nil {
			reasoning = &types.Reasoning{}
		}
		reasoning.Effort = *cfg.ReasoningEffort
	}

	// 对齐 codex: summary 默认为 "auto"，支持通过 metadata 覆盖
	if reasoning != nil {
		reasoning.Summary = "auto"
		if metadata != nil {
			if rs, ok := metadata["reasoning_summary"].(string); ok && rs != "" {
				reasoning.Summary = rs
			}
		}
	}

	return reasoning
}
