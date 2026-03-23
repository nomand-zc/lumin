package sse

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// Event 代表一个 SSE 事件
type Event struct {
	// EventType SSE 事件名称（event: 行的值）
	EventType string
	// Data SSE 数据负载（data: 行的值）
	Data []byte
}

// Parse 解析 SSE 流，通过回调函数逐个返回事件。
// 支持多行 data: 拼接，空行表示事件结束。
// 对齐 codex-rs/codex-api/src/sse/responses.rs 中使用的 eventsource_stream 库的解析逻辑。
func Parse(ctx context.Context, body io.ReadCloser, onEvent func(Event) error) error {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	// 设置较大的缓冲区以处理大的 SSE 事件
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var eventType string
	var bufferedLines []string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// 处理 event: 行
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimSpace(line[7:])
			continue
		}

		// 处理 data: 行，累加到缓冲区
		if strings.HasPrefix(line, "data: ") {
			bufferedLines = append(bufferedLines, strings.TrimSpace(line[6:]))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			bufferedLines = append(bufferedLines, strings.TrimSpace(line[5:]))
			continue
		}

		// 空行表示一个 SSE 事件结束
		if line == "" {
			if len(bufferedLines) == 0 {
				eventType = ""
				continue
			}

			// 拼接多行 data
			chunk := strings.Join(bufferedLines, "\n")
			bufferedLines = nil

			if strings.TrimSpace(chunk) == "" {
				eventType = ""
				continue
			}

			if err := onEvent(Event{EventType: eventType, Data: []byte(chunk)}); err != nil {
				return err
			}
			eventType = ""
			continue
		}
		// 忽略其他行（如 id:、注释等）
	}

	return scanner.Err()
}
