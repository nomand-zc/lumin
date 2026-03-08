package parser

import (
	"context"
	"encoding/json"

	"github.com/nomand-zc/lumin/log"
	"github.com/nomand-zc/lumin/providers"
	"github.com/nomand-zc/lumin/utils"
)

// errorParser 处理 error 类型消息
type errorParser struct{}

func init() {
	Register(&errorParser{})
}

func (p *errorParser) MessageType() string { return MessageTypeError }
func (p *errorParser) EventType() string   { return "" }

func (p *errorParser) Parse(ctx context.Context, msg *StreamMessage, opts ...OptionFunc) (*providers.Response, error) {
	var errorData struct {
		Type    string `json:"__type"`
		Message string `json:"message"`
	}

	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &errorData); err != nil {
			log.Warnf("解析错误消息载荷失败: %v", err)
			errorData.Message = utils.Bytes2Str(msg.Payload)
		}
	}

	return providers.NewResponse(ctx,
		providers.WithResponseError(&providers.ResponseError{
			Message: errorData.Message,
			Type:    "error",
			Code:    utils.ToPtr(errorData.Type),
		}),
	), nil
}
