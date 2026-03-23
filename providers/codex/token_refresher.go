package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	jujuerrors "github.com/juju/errors"
	"github.com/nomand-zc/lumin-client/credentials"
	codexcreds "github.com/nomand-zc/lumin-client/credentials/codex"
	"github.com/nomand-zc/lumin-client/httpclient"
	"github.com/nomand-zc/lumin-client/providers"
)

// Refresh 刷新 Codex 的 OAuth2 令牌，直接修改入参 creds 中的字段。
// 使用 auth.openai.com 的 OAuth2 refresh_token 流程获取新的 access_token。
// 对齐 codex-rs/core/src/codex.rs 中的 token 刷新逻辑和
// providers/codex/refresh_token.sh 中描述的请求格式。
func (p *codexProvider) Refresh(ctx context.Context, creds credentials.Credential) error {
	codexCreds, ok := creds.(*codexcreds.Credential)
	if !ok {
		return fmt.Errorf("invalid credentials type, expected *codexcreds.Credential")
	}

	return p.refreshOAuth2Token(ctx, codexCreds)
}

// tokenRefreshResp Codex OAuth2 token 刷新响应
// 对齐 refresh_token.sh 中的响应格式
type tokenRefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // Token 有效期（秒）
	Scope        string `json:"scope,omitempty"`

	// 错误字段
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// refreshOAuth2Token 使用 refresh_token 通过 Codex OAuth2 端点刷新 access_token
// 对齐 refresh_token.sh 中的请求格式：
//
//	POST https://auth.openai.com/oauth/token
//	Content-Type: application/x-www-form-urlencoded
//	grant_type=refresh_token&client_id=app_EMoamEEZ73f0CkXaXp7hrann&refresh_token=<token>
func (p *codexProvider) refreshOAuth2Token(ctx context.Context, creds *codexcreds.Credential) error {
	tokenURI := creds.TokenURI
	if tokenURI == "" {
		tokenURI = codexcreds.DefaultTokenURI
	}

	clientID := creds.ClientID
	if clientID == "" {
		clientID = codexcreds.DefaultClientID
	}

	// 构建 application/x-www-form-urlencoded 请求体
	// 对齐 refresh_token.sh 中的请求格式，codex 不需要 client_secret
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {creds.RefreshToken},
	}

	req, err := http.NewRequestWithContext(httpclient.EnablePrintRespBody(ctx),
		http.MethodPost, tokenURI, strings.NewReader(formData.Encode()))
	if err != nil {
		return jujuerrors.Annotate(err, "create token refresh request failed")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CodexOAuth/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return jujuerrors.Annotatef(err, "token refresh request failed, url=%s", tokenURI)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return jujuerrors.Annotate(err, "read token refresh response failed")
	}

	// 处理 HTTP 错误状态码
	if resp.StatusCode != http.StatusOK {
		var errResp tokenRefreshResp
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil && errResp.Error != "" {
			if errResp.Error == "invalid_grant" {
				return providers.ErrInvalidGrant
			}
			return newHTTPError(resp.StatusCode,
				fmt.Sprintf("token refresh failed: %s - %s", errResp.Error, errResp.ErrorDescription),
				respBody)
		}
		return newHTTPErrorf(resp.StatusCode, respBody,
			"token refresh failed, status=%d", resp.StatusCode)
	}

	// 解析成功响应
	var result tokenRefreshResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return jujuerrors.Annotatef(err, "parse token refresh response failed, status=%d, body=%s",
			resp.StatusCode, string(respBody))
	}

	if result.Error != "" {
		if result.Error == "invalid_grant" {
			return providers.ErrInvalidGrant
		}
		return fmt.Errorf("token refresh error: %s - %s", result.Error, result.ErrorDescription)
	}

	// 更新凭证字段
	creds.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		creds.RefreshToken = result.RefreshToken
	}
	if result.IDToken != "" {
		creds.IDToken = result.IDToken
	}
	if result.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
		creds.ExpiresAt = &expiresAt
	}
	// 重置 raw 缓存，使 ToMap() 返回最新数据
	creds.ResetRaw()

	return nil
}
