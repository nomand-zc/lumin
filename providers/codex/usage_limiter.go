package codex

import (
	"context"

	"github.com/nomand-zc/lumin-client/credentials"
	"github.com/nomand-zc/lumin-client/usagerule"
)

// Models 返回默认支持的模型列表
func (p *codexProvider) Models(_ context.Context) ([]string, error) {
	models := make([]string, len(ModelList))
	copy(models, ModelList)
	return models, nil
}

// ListModels 获取当前凭证支持的模型列表
// Codex 不提供动态模型列表接口，直接返回默认模型列表
func (p *codexProvider) ListModels(ctx context.Context, _ credentials.Credential) ([]string, error) {
	return p.Models(ctx)
}

// DefaultUsageRules 获取供应商默认的用量规则列表
// Codex 的配额由 OpenAI 账户管理，SDK 层面无固定用量规则
func (p *codexProvider) DefaultUsageRules(_ context.Context) ([]*usagerule.UsageRule, error) {
	return []*usagerule.UsageRule{}, nil
}

// GetUsageRules 获取当前凭证的用量规则列表
func (p *codexProvider) GetUsageRules(_ context.Context, _ credentials.Credential) ([]*usagerule.UsageRule, error) {
	return []*usagerule.UsageRule{}, nil
}

// GetUsageStats 获取当前凭证的用量统计信息
func (p *codexProvider) GetUsageStats(_ context.Context, _ credentials.Credential) ([]*usagerule.UsageStats, error) {
	return []*usagerule.UsageStats{}, nil
}
