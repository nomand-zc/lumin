package codex

import (
	"encoding/json"
	"time"

	"github.com/nomand-zc/lumin-client/credentials"
)

// Codex OAuth2 固定常量
const (
	DefaultClientID = "app_EMoamEEZ73f0CkXaXp7hrann"
	DefaultTokenURI = "https://auth.openai.com/oauth/token"
)

func init() {
	credentials.Register("codex", NewCredential[[]byte])
}

// Credential Codex 凭证结构体
type Credential struct {
	// AccessToken 当前访问令牌
	AccessToken string `json:"access_token"`
	// RefreshToken 刷新令牌
	RefreshToken string `json:"refresh_token,omitempty"`
	// IDToken ID 令牌
	IDToken string `json:"id_token,omitempty"`
	// ClientID OAuth2 客户端 ID
	ClientID string `json:"client_id,omitempty"`
	// TokenURI Token 端点 URL
	TokenURI string `json:"token_uri,omitempty"`
	// AccountID ChatGPT Account ID，用于 ChatGPT-Account-ID 请求头
	AccountID string `json:"account_id,omitempty"`
	// ExpiresAt 过期时间
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	raw map[string]any `json:"-"` // 原始凭证数据
}

// NewCredential 创建一个新的凭据实例
// 支持传入 JSON 字符串或 []byte，解析失败时返回 nil
func NewCredential[T string | []byte](raw T) credentials.Credential {
	var creds Credential
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil
	}

	// 填充默认值
	if creds.ClientID == "" {
		creds.ClientID = DefaultClientID
	}
	if creds.TokenURI == "" {
		creds.TokenURI = DefaultTokenURI
	}

	return &creds
}

// Clone 克隆凭据实例
func (c *Credential) Clone() credentials.Credential {
	clone := *c
	if c.ExpiresAt != nil {
		t := *c.ExpiresAt
		clone.ExpiresAt = &t
	}
	clone.raw = nil
	return &clone
}

// Validate 校验凭据的格式有效性（仅校验格式，不校验是否过期）
func (c *Credential) Validate() error {
	if c == nil {
		return credentials.ErrCredentialEmpty
	}
	if c.AccessToken == "" {
		return credentials.ErrAccessTokenEmpty
	}
	if c.RefreshToken == "" {
		return credentials.ErrRefreshTokenEmpty
	}
	return nil
}

// GetAccessToken 返回访问令牌
func (c *Credential) GetAccessToken() string {
	return c.AccessToken
}

// GetRefreshToken 返回刷新令牌
func (c *Credential) GetRefreshToken() string {
	return c.RefreshToken
}

// GetExpiresAt 返回过期时间
func (c *Credential) GetExpiresAt() *time.Time {
	return c.ExpiresAt
}

// GetUserInfo 返回用户信息
func (c *Credential) GetUserInfo() (credentials.UserInfo, error) {
	if c == nil {
		return credentials.UserInfo{}, nil
	}
	return credentials.UserInfo{}, nil
}

// IsExpired 检查凭据是否过期
func (c *Credential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(*c.ExpiresAt)
}

// ToMap 将凭据转换为 map 格式
func (c *Credential) ToMap() map[string]any {
	if c == nil {
		return nil
	}
	if c.raw == nil {
		c.raw = map[string]any{
			"access_token":  c.AccessToken,
			"refresh_token": c.RefreshToken,
			"id_token":      c.IDToken,
			"client_id":     c.ClientID,
			"token_uri":     c.TokenURI,
			"account_id":    c.AccountID,
			"expires_at":    c.ExpiresAt,
		}
	}
	return c.raw
}

// ResetRaw 重置 raw 缓存，使 ToMap() 下次调用时重新构建
func (c *Credential) ResetRaw() {
	c.raw = nil
}
