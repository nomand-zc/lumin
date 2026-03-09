package parser

import (
	"context"
	"encoding/json"

	"github.com/nomand-zc/lumin-client/log"
	"github.com/nomand-zc/lumin-client/providers"
	"github.com/nomand-zc/lumin-client/utils"
)

// exceptionParser 处理 exception 类型消息
type exceptionParser struct{}

func init() {
	Register(&exceptionParser{})
}

func (p *exceptionParser) MessageType() string { return MessageTypeException }
func (p *exceptionParser) EventType() string   { return "" }

func (p *exceptionParser) Parse(ctx context.Context, msg *StreamMessage, opts ...OptionFunc) (*providers.Response, error) {
	var exceptionData struct {
		Type    string `json:"__type"`
		Message string `json:"message"`
	}

	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &exceptionData); err != nil {
			log.Warnf("解析异常消息载荷失败: %v", err)
			exceptionData.Message = utils.Bytes2Str(msg.Payload)
		}
	}

	return providers.NewResponse(ctx,
		providers.WithResponseError(&providers.ResponseError{
			Message: exceptionData.Message,
			Type:    "exception",
			Code:    utils.ToPtr(exceptionData.Type),
		}),
	), nil
}
