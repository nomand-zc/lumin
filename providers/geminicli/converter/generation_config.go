package converter

import (
	"strings"

	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/providers/geminicli/types"
)

// ConvertGenerationConfig 将统一 GenerationConfig 转换为 Gemini 格式
func ConvertGenerationConfig(cfg *providers.GenerationConfig) *types.GeminiGenerationConfig {
	if cfg == nil {
		return nil
	}

	genCfg := &types.GeminiGenerationConfig{}
	hasConfig := false

	if cfg.Temperature != nil {
		genCfg.Temperature = cfg.Temperature
		hasConfig = true
	}
	if cfg.TopP != nil {
		genCfg.TopP = cfg.TopP
		hasConfig = true
	}
	if cfg.PresencePenalty != nil {
		genCfg.PresencePenalty = cfg.PresencePenalty
		hasConfig = true
	}
	if cfg.FrequencyPenalty != nil {
		genCfg.FrequencyPenalty = cfg.FrequencyPenalty
		hasConfig = true
	}
	if cfg.MaxTokens != nil {
		genCfg.MaxOutputTokens = cfg.MaxTokens
		hasConfig = true
	}
	if len(cfg.Stop) > 0 {
		genCfg.StopSequences = cfg.Stop
		hasConfig = true
	}

	// thinking 配置
	if cfg.ThinkingEnabled != nil {
		include := *cfg.ThinkingEnabled
		thinkingCfg := &types.GeminiThinkingConfig{
			IncludeThoughts: &include,
		}
		if cfg.ThinkingTokens != nil {
			budget := *cfg.ThinkingTokens
			thinkingCfg.ThinkingBudget = &budget
		}
		genCfg.ThinkingConfig = thinkingCfg
		hasConfig = true
	}

	// reasoning effort 映射为 thinkingLevel
	if cfg.ReasoningEffort != nil && *cfg.ReasoningEffort != "" {
		effort := strings.ToLower(*cfg.ReasoningEffort)
		if genCfg.ThinkingConfig == nil {
			genCfg.ThinkingConfig = &types.GeminiThinkingConfig{}
		}
		genCfg.ThinkingConfig.ThinkingLevel = effort
		include := effort != "none"
		genCfg.ThinkingConfig.IncludeThoughts = &include
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}
	return genCfg
}
