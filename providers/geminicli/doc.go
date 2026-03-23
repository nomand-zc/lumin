// Package geminicli 实现了 Gemini CLI (Cloud Code) 的 Provider。
//
// 协议对齐说明：
// 本包的请求/响应格式对齐 gemini-cli (https://github.com/google-gemini/gemini-cli)
// 的 converter.ts 和 server.ts 实现，以及 CLIProxyAPIPlus 中的转换逻辑。
//
// 包结构：
//   - types/     : Gemini CLI API 的请求和响应 DTO 定义
//   - converter/ : 统一 Request/Response 与 Gemini CLI 格式之间的转换逻辑
//   - sse/       : SSE 流协议解析与业务处理
package geminicli
