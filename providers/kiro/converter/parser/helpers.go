package parser

import (
	"encoding/json"
	"strings"

	"github.com/nomand-zc/lumin-client/log"
	"github.com/nomand-zc/lumin-client/utils"
)

// isToolCallPayload 检查载荷是否包含工具调用信息
func isToolCallPayload(payload string) bool {
	return strings.Contains(payload, "\"toolUseId\":") ||
		strings.Contains(payload, "\"tool_use_id\":") ||
		(strings.Contains(payload, "\"name\":") && strings.Contains(payload, "\"input\":"))
}

// convertInputToArgs 将 any 类型的 input 转换为 JSON 字节
func convertInputToArgs(input any) []byte {
	if input == nil {
		return nil
	}
	if str, ok := input.(string); ok {
		return utils.Str2Bytes(str)
	}
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		log.Warnf("转换 input 为 JSON 失败: %v", err)
		return nil
	}
	return jsonBytes
}
