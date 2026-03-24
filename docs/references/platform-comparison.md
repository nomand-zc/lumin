# 平台差异对照

## 概览

| 维度 | Kiro | Codex | GeminiCLI | iFlow |
|------|------|-------|-----------|-------|
| **流协议** | AWS EventStream | SSE (text/event-stream) | SSE | — |
| **协议参考** | Kiro 原生 | codex-rs/codex-api Responses API | gemini-cli converter.ts | — |
| **事件解析** | 注册式事件派发 | SSE 事件类型映射 | SSE 数据解析 | — |

## Converter 差异

### ReqBuilder

所有平台的 ReqBuilder 拆解为相同的子构建器结构：
- SystemPromptBuilder
- HistoryBuilder
- CurrentMessageBuilder
- ToolsBuilder
- PreprocessBuilder

但各自的 DTO 结构不同，需要分别实现。

### RespParser

- **Kiro**: 注册式事件解析器，根据事件类型路由到对应 Parser
- **Codex**: SSE 事件类型直接映射到统一 Response
- **GeminiCLI**: SSE 数据字段解析

## 非标准错误映射

见 → [error-handling.md](../design-docs/error-handling.md)

## 透传字段

### Request.Header

所有平台均支持动态 Header 透传：
```go
for key, value := range req.Header {
    request.Header.Set(key, value)
}
```

### Request.Metadata

用于传递平台扩展字段（tool_choice、response_format 等），各平台按需取用。

### ToolCall.ExtraFields

平台特定透传字段（如 Gemini 3 的 thought_signature），不参与核心逻辑但需完整保留。
