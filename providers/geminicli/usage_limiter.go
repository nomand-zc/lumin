package geminicli

import (
	"context"

	"github.com/nomand-zc/lumin-client/credentials"
	"github.com/nomand-zc/lumin-client/usagerule"
)

// Models 返回默认支持的模型列表
func (p *geminicliProvider) Models(_ context.Context) ([]string, error) {
	models := make([]string, len(ModelList))
	copy(models, ModelList)
	return models, nil
}

// ListModels 获取当前凭证支持的模型列表
// GeminiCLI 不提供动态模型列表接口，直接返回默认模型列表
func (p *geminicliProvider) ListModels(ctx context.Context, _ credentials.Credential) ([]string, error) {
	return p.Models(ctx)
}

// DefaultUsageRules 获取供应商默认的用量规则列表
// GeminiCLI 使用 Google Cloud 项目配额，无固定的用量规则
func (p *geminicliProvider) DefaultUsageRules(_ context.Context) ([]*usagerule.UsageRule, error) {
	return []*usagerule.UsageRule{}, nil
}

// GetUsageRules 获取当前凭证的用量规则列表
// GeminiCLI 的配额由 Google Cloud 项目管理，SDK 层面无法获取具体规则
func (p *geminicliProvider) GetUsageRules(_ context.Context, _ credentials.Credential) ([]*usagerule.UsageRule, error) {
	return []*usagerule.UsageRule{}, nil
}

// GetUsageStats 获取当前凭证的用量统计信息
// GeminiCLI 的用量统计由 Google Cloud Console 管理，SDK 层面暂不支持查询
func (p *geminicliProvider) GetUsageStats(_ context.Context, _ credentials.Credential) ([]*usagerule.UsageStats, error) {
	return []*usagerule.UsageStats{}, nil
}
