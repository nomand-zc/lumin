package sse

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// Event 代表一个 SSE 事件
type Event struct {
	Data []byte
}

// Parse 解析 SSE 流，通过回调函数逐个返回事件。
// 支持多行 data: 拼接，空行表示事件结束。
func Parse(ctx context.Context, body io.ReadCloser, onEvent func(Event) error) error {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	// 设置较大的缓冲区以处理大的 SSE 事件
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var bufferedLines []string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// 处理 data: 开头的行，累加到缓冲区
		if strings.HasPrefix(line, "data: ") {
			bufferedLines = append(bufferedLines, strings.TrimSpace(line[6:]))
			continue
		}

		// 空行表示一个 SSE 事件结束
		if line == "" {
			if len(bufferedLines) == 0 {
				continue
			}

			// 拼接多行 data
			chunk := strings.Join(bufferedLines, "\n")
			bufferedLines = nil

			if strings.TrimSpace(chunk) == "" {
				continue
			}

			if err := onEvent(Event{Data: []byte(chunk)}); err != nil {
				return err
			}
			continue
		}
		// 忽略其他行（如 event:、id:、注释等）
	}

	return scanner.Err()
}
