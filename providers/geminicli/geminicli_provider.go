package geminicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	geminicreds "github.com/nomand-zc/lumin-client/credentials/geminicli"
	"github.com/nomand-zc/lumin-client/httpclient"
	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/providers/geminicli/converter"
	"github.com/nomand-zc/lumin-client/queue"
)

func init() {
	providers.Register(NewProvider(providers.DefaultProviderName))
}

const (
	providerType     = "geminicli"
	defaultQueueSize = 100
)

type geminicliProvider struct {
	name       string
	httpClient httpclient.HTTPClient
	options    *Options
}

// NewProvider 创建一个新的 GeminiCLI provider
func NewProvider(name string, opts ...Option) *geminicliProvider {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &geminicliProvider{
		name:       name,
		options:    &options,
		httpClient: httpclient.New(httpclient.WithMiddleware(
			httpclient.LoggingMiddleware,
		)),
	}
}

// Name 返回 provider 名称
func (p *geminicliProvider) Name() string {
	return p.name
}

// Type 返回 provider 类型
func (p *geminicliProvider) Type() string {
	return providerType
}

// GenerateContent 生成内容（非流式）
// 通过流式接口聚合完整响应，对齐 kiro provider 的实现模式
func (p *geminicliProvider) GenerateContent(ctx context.Context, req *providers.Request) (*providers.Response, error) {
	reader, err := p.GenerateContentStream(ctx, req)
	if err != nil {
		return nil, err
	}

	acc := &providers.ResponseAccumulator{}
	if err := reader.Each(ctx, func(chunk *providers.Response) error {
		acc.AddChunk(chunk)
		return nil
	}); err != nil {
		return nil, err
	}

	resp := acc.Response()
	if resp == nil {
		return nil, fmt.Errorf("no response received from stream")
	}
	resp.Object = providers.ObjectChatCompletion
	resp.IsPartial = false
	resp.Done = true
	return resp, nil
}

// GenerateContentStream 流式生成内容
// 对齐 CLIProxyAPIPlus 中 GeminiCLIExecutor.ExecuteStream 的核心逻辑
func (p *geminicliProvider) GenerateContentStream(ctx context.Context, req *providers.Request) (queue.Consumer[*providers.Response], error) {
	// 1. 初始化调用上下文
	ctx, inv := providers.EnsureInvocationContext(ctx)
	inv.ID = uuid.NewString()

	// 2. 提取凭证
	geminiCreds, ok := req.Credential.(*geminicreds.Credential)
	if !ok {
		return nil, fmt.Errorf("invalid credentials type, expected *geminicreds.Credential")
	}

	// 3. 解析模型名（支持别名映射）
	model := req.Model
	if mapped, exists := ModelMap[model]; exists {
		model = mapped
	}
	inv.Model = model

	// 4. 使用 converter 构建 Gemini CLI 请求
	geminiReq, err := converter.BuildRequest(req, geminiCreds.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to build geminicli request: %w", err)
	}
	// 覆盖为映射后的模型名
	geminiReq.Model = model

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal geminicli request: %w", err)
	}

	// 5. 构建 HTTP 请求，对齐 CLIProxyAPIPlus 中的 URL 和 Headers
	url := p.options.StreamURL()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置 Headers，对齐 CLIProxyAPIPlus 中的 applyGeminiCLIHeaders
	for key, value := range p.options.headers {
		httpReq.Header.Set(key, value)
	}
	// 设置 Request 中调用者传递的动态 Header
	for key, value := range req.Header {
		httpReq.Header.Set(key, value)
	}
	httpReq.Header.Set("Authorization", "Bearer "+geminiCreds.AccessToken)
	httpReq.Header.Set("Accept", "text/event-stream")
	// 动态 User-Agent（包含 model 名），对齐 CLIProxyAPIPlus 中 applyGeminiCLIHeaders
	httpReq.Header.Set("User-Agent", GeminiCLIUserAgent(model))

	// 6. 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("geminicli HTTP request failed: %w", err)
	}

	// 7. 检查状态码，对齐 CLIProxyAPIPlus 中的错误处理
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, converter.ParseErrorResponse(resp.StatusCode, body)
	}

	// 8. 解析 SSE 流
	chainQueue := queue.New[*providers.Response](defaultQueueSize)
	go converter.ParseSSEStream(ctx, resp.Body, model, chainQueue)

	return chainQueue, nil
}
