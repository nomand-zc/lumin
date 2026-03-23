package converter

import (
	"github.com/nomand-zc/lumin-client/providers/geminicli/types"
)

// ConvertToolChoice 将统一 tool_choice 转换为 GeminiToolConfig
func ConvertToolChoice(toolChoice any) *types.GeminiToolConfig {
	config := &types.GeminiToolConfig{
		FunctionCallingConfig: &types.GeminiFunctionCallingConfig{},
	}

	switch v := toolChoice.(type) {
	case string:
		switch v {
		case "auto":
			config.FunctionCallingConfig.Mode = "AUTO"
		case "none":
			config.FunctionCallingConfig.Mode = "NONE"
		case "any", "required":
			config.FunctionCallingConfig.Mode = "ANY"
		default:
			return nil
		}
	case map[string]any:
		tcType, _ := v["type"].(string)
		tcName, _ := v["name"].(string)
		switch tcType {
		case "auto":
			config.FunctionCallingConfig.Mode = "AUTO"
		case "none":
			config.FunctionCallingConfig.Mode = "NONE"
		case "any":
			config.FunctionCallingConfig.Mode = "ANY"
		case "tool", "function":
			config.FunctionCallingConfig.Mode = "ANY"
			if tcName != "" {
				config.FunctionCallingConfig.AllowedFunctionNames = []string{tcName}
			}
		default:
			return nil
		}
	default:
		return nil
	}

	return config
}
