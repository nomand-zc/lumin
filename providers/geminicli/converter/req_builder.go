package converter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/providers/geminicli/types"
)

// 默认安全设置
var defaultSafetySettings = []map[string]string{
	{"category": "HARM_CATEGORY_HARASSMENT", "threshold": "OFF"},
	{"category": "HARM_CATEGORY_HATE_SPEECH", "threshold": "OFF"},
	{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT", "threshold": "OFF"},
	{"category": "HARM_CATEGORY_DANGEROUS_CONTENT", "threshold": "OFF"},
	{"category": "HARM_CATEGORY_CIVIC_INTEGRITY", "threshold": "BLOCK_NONE"},
}

// thoughtSignatureSkip 用于跳过 thoughtSignature 验证的默认值
const thoughtSignatureSkip = "skip_thought_signature_validator"

// BuildRequest 将统一 Request 转换为 GeminiCLI 请求格式
func BuildRequest(req *providers.Request, projectID string) (*types.GeminiCLIRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	inner := &types.GeminiCLIInner{
		Contents:       make([]types.GeminiContent, 0),
		SafetySettings: defaultSafetySettings,
	}

	// 转换消息
	for _, msg := range req.Messages {
		switch msg.Role {
		case providers.RoleSystem:
			inner.SystemInstruction = convertSystemMessage(&msg)
		case providers.RoleUser:
			inner.Contents = append(inner.Contents, convertUserMessage(&msg))
		case providers.RoleAssistant:
			inner.Contents = append(inner.Contents, convertAssistantMessage(&msg))
		case providers.RoleTool:
			inner.Contents = append(inner.Contents, convertToolMessage(&msg))
		}
	}

	// 将分散的 functionResponse 消息合并到对应的 functionCall 之后
	inner.Contents = fixToolResponseGrouping(inner.Contents)

	// 转换 tools 为 functionDeclarations
	if len(req.Tools) > 0 {
		inner.Tools = convertTools(req.Tools)
	}

	// 处理 tool_choice
	if req.Metadata != nil {
		if toolChoice, ok := req.Metadata["tool_choice"]; ok {
			inner.ToolConfig = ConvertToolChoice(toolChoice)
		}
	}

	// 映射 GenerationConfig
	if genConfig := ConvertGenerationConfig(&req.GenerationConfig); genConfig != nil {
		inner.GenerationConfig = genConfig
	}

	// 从 Metadata 中提取可选字段
	var userPromptID string
	var enabledCreditTypes []string
	if req.Metadata != nil {
		if v, ok := req.Metadata["user_prompt_id"].(string); ok {
			userPromptID = v
		}
		if v, ok := req.Metadata["session_id"].(string); ok {
			inner.SessionID = v
		}
		if v, ok := req.Metadata["enabled_credit_types"].([]string); ok {
			enabledCreditTypes = v
		}
		if v, ok := req.Metadata["labels"].(map[string]string); ok {
			inner.Labels = v
		}
		if v, ok := req.Metadata["cached_content"].(string); ok {
			inner.CachedContent = v
		}
	}

	return &types.GeminiCLIRequest{
		Model:              req.Model,
		Project:            projectID,
		UserPromptID:       userPromptID,
		Request:            inner,
		EnabledCreditTypes: enabledCreditTypes,
	}, nil
}

// convertSystemMessage 将 system 消息提取为 systemInstruction
func convertSystemMessage(msg *providers.Message) *types.GeminiContent {
	systemParts := make([]types.GeminiPart, 0)
	if msg.Content != "" {
		systemParts = append(systemParts, types.GeminiPart{Text: msg.Content})
	}
	for _, cp := range msg.ContentParts {
		if cp.Type == providers.ContentTypeText && cp.Text != nil {
			systemParts = append(systemParts, types.GeminiPart{Text: *cp.Text})
		}
	}
	if len(systemParts) == 0 {
		return nil
	}
	return &types.GeminiContent{
		Role:  "user",
		Parts: systemParts,
	}
}

// convertUserMessage 将 user 消息转换为 GeminiContent
func convertUserMessage(msg *providers.Message) types.GeminiContent {
	content := types.GeminiContent{
		Role:  "user",
		Parts: make([]types.GeminiPart, 0),
	}

	if msg.Content != "" {
		content.Parts = append(content.Parts, types.GeminiPart{Text: msg.Content})
	}

	for _, cp := range msg.ContentParts {
		switch cp.Type {
		case providers.ContentTypeText:
			if cp.Text != nil {
				content.Parts = append(content.Parts, types.GeminiPart{Text: *cp.Text})
			}
		case providers.ContentTypeImage:
			if cp.Image != nil {
				if part := convertImagePart(cp.Image); part != nil {
					content.Parts = append(content.Parts, *part)
				}
			}
		}
	}

	return content
}

// convertAssistantMessage 将 assistant 消息转换为 GeminiContent（role=model）
func convertAssistantMessage(msg *providers.Message) types.GeminiContent {
	content := types.GeminiContent{
		Role:  "model",
		Parts: make([]types.GeminiPart, 0),
	}

	// thinking/reasoning 内容
	if msg.ReasoningContent != "" {
		isThought := true
		content.Parts = append(content.Parts, types.GeminiPart{
			Text:    msg.ReasoningContent,
			Thought: &isThought,
		})
	}

	// 文本内容
	if msg.Content != "" {
		content.Parts = append(content.Parts, types.GeminiPart{Text: msg.Content})
	}

	// tool calls 转换为 functionCall
	for _, tc := range msg.ToolCalls {
		fc := &types.GeminiFunctionCall{
			Name: tc.Function.Name,
		}
		if len(tc.Function.Arguments) > 0 {
			var args map[string]any
			if err := json.Unmarshal(tc.Function.Arguments, &args); err == nil {
				fc.Args = args
			}
		}
		part := types.GeminiPart{
			ThoughtSignature: thoughtSignatureSkip,
			FunctionCall:     fc,
		}
		if tc.ExtraFields != nil {
			if sig, ok := tc.ExtraFields["thoughtSignature"].(string); ok && sig != "" {
				part.ThoughtSignature = sig
			}
		}
		content.Parts = append(content.Parts, part)
	}

	return content
}

// convertToolMessage 将 tool 消息转换为 GeminiContent（role=function）
func convertToolMessage(msg *providers.Message) types.GeminiContent {
	funcName := msg.ToolName
	if funcName == "" && msg.ToolID != "" {
		parts := strings.Split(msg.ToolID, "-")
		if len(parts) > 1 {
			funcName = strings.Join(parts[0:len(parts)-1], "-")
		} else {
			funcName = msg.ToolID
		}
	}

	return types.GeminiContent{
		Role: "function",
		Parts: []types.GeminiPart{
			{
				FunctionResponse: &types.GeminiFuncResponse{
					Name: funcName,
					Response: types.GeminiFuncRespValue{
						Result: msg.Content,
					},
				},
			},
		},
	}
}

// convertImagePart 将图片数据转换为 GeminiPart
func convertImagePart(img *providers.Image) *types.GeminiPart {
	if img == nil || len(img.Data) == 0 {
		return nil
	}

	mimeType := "image/png"
	switch img.Format {
	case "jpg", "jpeg":
		mimeType = "image/jpeg"
	case "webp":
		mimeType = "image/webp"
	case "gif":
		mimeType = "image/gif"
	}

	return &types.GeminiPart{
		InlineData: &types.GeminiInlineData{
			MimeType: mimeType,
			Data:     base64.StdEncoding.EncodeToString(img.Data),
		},
	}
}

// convertTools 将统一 Tool 列表转换为 Gemini 工具定义
func convertTools(tools []providers.Tool) []types.GeminiTool {
	decls := make([]types.GeminiFunctionDecl, 0, len(tools))
	for _, t := range tools {
		decl := types.GeminiFunctionDecl{
			Name:        t.Name,
			Description: t.Description,
		}
		if t.Parameters.Type != "" {
			decl.ParametersJsonSchema = t.Parameters
		}
		decls = append(decls, decl)
	}
	return []types.GeminiTool{{FunctionDeclarations: decls}}
}

// fixToolResponseGrouping 将分散的 function role 消息正确地分组到对应的 model（含 functionCall）消息之后
func fixToolResponseGrouping(contents []types.GeminiContent) []types.GeminiContent {
	type functionCallGroup struct {
		responsesNeeded int
	}

	result := make([]types.GeminiContent, 0, len(contents))
	var pendingGroups []*functionCallGroup
	var collectedResponses []types.GeminiPart

	for _, content := range contents {
		if content.Role == "function" {
			for _, part := range content.Parts {
				if part.FunctionResponse != nil {
					collectedResponses = append(collectedResponses, part)
				}
			}
			for i := len(pendingGroups) - 1; i >= 0; i-- {
				group := pendingGroups[i]
				if len(collectedResponses) >= group.responsesNeeded {
					groupResponses := collectedResponses[:group.responsesNeeded]
					collectedResponses = collectedResponses[group.responsesNeeded:]
					result = append(result, types.GeminiContent{
						Role:  "function",
						Parts: groupResponses,
					})
					pendingGroups = append(pendingGroups[:i], pendingGroups[i+1:]...)
					break
				}
			}
			continue
		}

		if content.Role == "model" {
			functionCallCount := 0
			for _, part := range content.Parts {
				if part.FunctionCall != nil {
					functionCallCount++
				}
			}
			result = append(result, content)
			if functionCallCount > 0 {
				pendingGroups = append(pendingGroups, &functionCallGroup{
					responsesNeeded: functionCallCount,
				})
			}
			continue
		}

		result = append(result, content)
	}

	// 处理剩余的 pending groups
	for _, group := range pendingGroups {
		if len(collectedResponses) >= group.responsesNeeded {
			groupResponses := collectedResponses[:group.responsesNeeded]
			collectedResponses = collectedResponses[group.responsesNeeded:]
			result = append(result, types.GeminiContent{
				Role:  "function",
				Parts: groupResponses,
			})
		}
	}

	return result
}
