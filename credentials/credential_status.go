package credentials

// CredentialStatus 凭证可用性状态
type CredentialStatus int

const (
	// StatusAvailable 当前可用，凭证有效且可正常使用
	StatusAvailable CredentialStatus = iota
	// StatusExpired Token 已过期，可通过刷新（Refresh）恢复
	StatusExpired
	// StatusInvalidated 永久失效，无法恢复（如 refresh token 无效，对应 ErrInvalidGrant）
	StatusInvalidated
	// StatusBanned 被平台封禁（如 Kiro 账号被 AWS 临时封禁，body 含 TEMPORARILY_SUSPENDED）
	StatusBanned
	// StatusUsageLimited 触发了 usageRule 的用量限制
	StatusUsageLimited
	// StatusReauthRequired 需要用户重新走授权流程（如 GeminiCLI 的 VALIDATION_REQUIRED）
	StatusReauthRequired
)

// String 返回 CredentialStatus 的可读字符串表示
func (s CredentialStatus) String() string {
	switch s {
	case StatusAvailable:
		return "available"
	case StatusExpired:
		return "expired"
	case StatusInvalidated:
		return "invalidated"
	case StatusBanned:
		return "banned"
	case StatusUsageLimited:
		return "usage_limited"
	case StatusReauthRequired:
		return "reauth_required"
	default:
		return "unknown"
	}
}
