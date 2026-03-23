// Package codex 实现了 OpenAI Codex 的 Provider。
//
// 协议对齐说明：
// 本包的请求/响应格式对齐 codex (https://github.com/openai/codex)
// 的 codex-api SSE Responses API 实现。
//
// 包结构：
//   - types/     : Codex Responses API 的请求和响应 DTO 定义
//   - converter/ : 统一 Request/Response 与 Codex 格式之间的转换逻辑
//   - sse/       : SSE 流协议解析与业务处理
package codex
