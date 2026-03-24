# AGENTS.md — providers/

> Provider 层入口。详细设计请跟随链接。

## 职责

定义统一 `Provider` 接口 + 实现 Kiro / Codex / GeminiCLI / iFlow 适配器。

## 文件速查

| 文件 | 关键类型 |
|------|---------|
| `interface.go` | `Provider`, `Model`, `CredentialManager`, `UsageLimiter` |
| `request.go` | `Request`, `GenerationConfig`, `Tool` |
| `response.go` | `Response`, `Choice`, `Usage` |
| `message.go` | `Message`, `ContentPart`（多模态） |
| `tool_call.go` | `ToolCall`, `FunctionDefinitionParam` |
| `accumulator.go` | `ResponseAccumulator` |
| `http_errors.go` | `HTTPError`, `ErrorType` |
| `register.go` | `Register()`, `GetProvider()` |

## 关键约束

1. `GenerateContent` **必须**由 `GenerateContentStream` + `ResponseAccumulator` 聚合
2. 所有 HTTP 错误**必须**归一化为 5 种 `HTTPError` 类型
3. `req.Credential` 类型断言失败会 panic — 调用者负责匹配

## 深入阅读

- [新增 Provider 指南](../docs/design-docs/add-provider.md)
- [错误处理规范](../docs/design-docs/error-handling.md)
- [流式响应机制](../docs/design-docs/streaming.md)
- [平台差异对照](../docs/references/platform-comparison.md)
