package codex

import (
	"maps"
)

const (
	// Codex API 默认端点
	DefaultEndpoint = "https://api.openai.com/v1"
)

var defaultOptions = Options{
	endpoint: DefaultEndpoint,
	headers: map[string]string{
		"Content-Type": "application/json",
	},
}

// Options 配置选项
type Options struct {
	endpoint string
	headers  map[string]string
}

// Option 配置选项函数
type Option func(*Options)

// WithEndpoint 设置 API 端点
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.endpoint = endpoint
	}
}

// WithHeader 设置单个 Header
func WithHeader(key, value string) Option {
	return func(o *Options) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

// WithHeaders 合并多个 Header
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		maps.Copy(o.headers, headers)
	}
}

// ResponsesURL 返回 Responses API 的完整 URL
func (o *Options) ResponsesURL() string {
	return o.endpoint + "/responses"
}
