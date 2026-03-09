package builder

import (
	"strings"

	"github.com/nomand-zc/lumin-client/providers"
)

// SystemPromptBuilder 负责从消息列表中提取 system prompt：
//   - 将所有 system 消息的内容用 "\n\n" 合并，写入 BuildContext.SystemPrompt
//   - 将非 system 消息写回 BuildContext.Messages
//   - 若过滤后消息列表为空，则设置 Done = true
type SystemPromptBuilder struct{}

// Build 实现 MessageBuilder 接口
func (b *SystemPromptBuilder) Build(ctx *BuildContext) error {
	var sb strings.Builder
	var nonSystemMessages []providers.Message

	for _, msg := range ctx.Messages {
		if msg.Role == providers.RoleSystem {
			if msg.Content != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n\n")
				}
				sb.WriteString(msg.Content)
			}
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}

	ctx.SystemPrompt = sb.String()
	ctx.Messages = nonSystemMessages

	if len(ctx.Messages) == 0 {
		ctx.Done = true
	}
	return nil
}
