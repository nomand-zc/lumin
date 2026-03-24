# 核心设计原则

## 1. 统一抽象，屏蔽差异

所有平台（Kiro / Codex / GeminiCLI / iFlow）对上层暴露完全一致的 `Provider` 接口。
调用者不感知底层协议差异（EventStream / SSE / 自定义协议）。

## 2. 流式优先

`GenerateContentStream()` 是一等公民。`GenerateContent()` 必须由流式接口 + `ResponseAccumulator` 聚合实现，不允许独立实现同步路径。

理由：避免同步/流式路径行为漂移，单一真相来源。

## 3. 错误语义化，不透传原始状态码

HTTP 错误必须归一化为 5 种语义类型（400/401/403/429/500）。
上层依赖语义类型做决策，`RawStatusCode` / `RawBody` 仅用于调试日志。

详见 → [error-handling.md](error-handling.md)

## 4. 凭证即值对象

凭证必须支持 `Clone()` 深拷贝，保证并发安全。
凭证状态由外部（CredentialManager）管理，凭证本身只负责"我是什么"。

## 5. 注册式扩展

Provider 和 Credential 通过 `init()` + 全局注册表扩展，零侵入式。
新增平台不需要修改任何现有代码。

## 6. Converter 分离

协议转换逻辑（`converter/`）独立于 Provider 核心逻辑，便于单独测试和替换。
ReqBuilder 负责请求转换，RespParser 负责响应转换。

## 7. Functional Options

构造器统一使用 Functional Options 模式，兼顾简洁默认值和灵活扩展。
