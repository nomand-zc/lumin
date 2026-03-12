package auth

import (
	"context"
	"errors"

	"github.com/nomand-zc/lumin-client/credentials"
	"github.com/nomand-zc/lumin-client/providers"
)

// CheckAvailability 校验凭证的可用性，返回凭证当前的状态。
// 判断逻辑：
//  1. 先做本地校验（格式、过期时间等）
//  2. 尝试刷新 Token，根据返回的错误类型判断状态
//  3. 校验用量规则是否触发限制
func CheckAvailability(ctx context.Context, provider providers.Provider, creds credentials.Credential) (credentials.CredentialStatus, error) {
	// 基本校验
	if err := creds.Validate(); err != nil {
		return credentials.StatusInvalidated, nil
	}

	// 检查是否过期
	if creds.IsExpired() {
		// 尝试刷新，判断是临时过期还是永久失效
		if err := provider.Refresh(ctx, creds); err != nil {
			// 根据错误类型判断凭证状态
			if errors.Is(err, providers.ErrInvalidGrant) {
				return credentials.StatusInvalidated, nil
			}
			var httpErr *providers.HTTPError
			if errors.As(err, &httpErr) {
				switch httpErr.ErrorType {
				case providers.ErrorTypeForbidden:
					return credentials.StatusBanned, nil
				case providers.ErrorTypeRateLimit:
					return credentials.StatusUsageLimited, nil
				}
			}
			// 其他错误（如网络问题），返回过期状态和错误信息
			return credentials.StatusExpired, err
		}
	}

	// 校验用量规则
	stats, err := provider.GetUsageStats(ctx, creds)
	if err == nil {
		for _, s := range stats {
			if s.IsTriggered() {
				return credentials.StatusUsageLimited, nil
			}
		}
	}

	return credentials.StatusAvailable, nil
}
