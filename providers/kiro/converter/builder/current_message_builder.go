package builder

import (
	"strings"

	"github.com/nomand-zc/lumin/providers"
	"github.com/nomand-zc/lumin/providers/kiro/converter/builder/types"
)

// CurrentMessageBuilder 负责解析最后一条消息（currentMessage）：
//   - 若最后一条是 assistant 消息：将其加入 history，currentContent 设为 "Continue"
//   - 若最后一条是 user 消息（含 ContentParts）：分别提取文本和图片
//   - 若最后一条是 tool 消息：转换为 types.ToolResult
//   - currentContent 为空时的兜底逻辑
//
// 结果写入 BuildContext.CurrentContent、CurrentImages、CurrentToolResults
// 同时可能追加 BuildContext.History
type CurrentMessageBuilder struct{}

// Build 实现 MessageBuilder 接口
func (b *CurrentMessageBuilder) Build(ctx *BuildContext) error {
	messages := ctx.Messages
	totalMessages := len(messages)
	lastMsg := messages[totalMessages-1]

	var contentBuilder strings.Builder
	var currentToolResults []types.ToolResult
	var currentImages []types.Image

	if lastMsg.Role == providers.RoleAssistant {
		// 最后一条是 assistant 消息：将其加入 history，currentMessage 设为 "Continue"
		assistantMsg := BuildAssistantMessage(lastMsg)
		ctx.History = append(ctx.History, types.HistoryItem{AssistantResponseMessage: &assistantMsg})
		contentBuilder.WriteString("Continue")
	} else {
		// 最后一条是 user/tool 消息：确保 history 末尾是 assistantResponseMessage
		if len(ctx.History) > 0 {
			lastHistoryItem := ctx.History[len(ctx.History)-1]
			if lastHistoryItem.AssistantResponseMessage == nil {
				ctx.History = append(ctx.History, types.HistoryItem{
					AssistantResponseMessage: &types.AssistantResponseMessage{Content: "Continue"},
				})
			}
		}

		// 解析最后一条 user 消息的内容
		if len(lastMsg.ContentParts) > 0 {
			for _, part := range lastMsg.ContentParts {
				switch part.Type {
				case providers.ContentTypeText:
					if part.Text != nil {
						contentBuilder.WriteString(*part.Text)
					}
				case providers.ContentTypeImage:
					if part.Image != nil {
						img := ConvertImage(part.Image)
						if img != nil {
							currentImages = append(currentImages, *img)
						}
					}
				}
			}
		} else if lastMsg.Role == providers.RoleTool {
			// RoleTool 消息作为 toolResult
			currentToolResults = append(currentToolResults, types.ToolResult{
				ToolUseId: lastMsg.ToolID,
				Status:    "success",
				Content:   []types.ToolResultContent{{Text: lastMsg.Content}},
			})
		} else {
			contentBuilder.WriteString(lastMsg.Content)
		}

		// content 兆底
		if contentBuilder.Len() == 0 {
			if len(currentToolResults) > 0 {
				contentBuilder.WriteString("Tool results provided.")
			} else {
				contentBuilder.WriteString("Continue")
			}
		}
	}

	ctx.CurrentContent = contentBuilder.String()
	ctx.CurrentImages = currentImages
	ctx.CurrentToolResults = currentToolResults
	return nil
}