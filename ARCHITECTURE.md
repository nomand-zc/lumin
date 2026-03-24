# ARCHITECTURE.md — lumin-client 架构文档

## 系统定位

lumin-client 是 LUMIN 生态最底层的核心 SDK，为 lumin-acpool、lumin-proxy、lumin-desktop 提供统一的 AI 模型调用能力。

## Provider 接口体系

```
Provider（顶层接口）
├── Type() string              — 类型标识（"kiro" / "codex" / "geminicli"）
├── Name() string              — 实例名（通常 "default"）
├── Model
│   ├── GenerateContent()      — 同步生成
│   └── GenerateContentStream()— 流式生成 → Consumer[*Response]
├── CredentialManager
│   └── Refresh()              — 刷新凭证
└── UsageLimiter
    ├── Models() / ListModels()
    ├── DefaultUsageRules()
    ├── GetUsageRules()
    └── GetUsageStats()
```

## 请求/响应数据流

```
调用者
  │
  ├─ 构建 Request（Credential + Messages + GenerationConfig + Tools）
  │
  ▼
Provider.GenerateContentStream()
  │
  ├─ 1. EnsureInvocationContext(ctx) → 初始化调用上下文
  ├─ 2. req.Credential 类型断言 → 平台特定凭证
  ├─ 3. Converter.ReqBuilder: 统一 Request → 平台请求 DTO
  ├─ 4. HTTPClient.Do(): 发送请求（经过中间件管道）
  ├─ 5. 错误检查 → 归一化为 HTTPError
  ├─ 6. 启动 goroutine 流式解析
  │     └─ 逐事件解析 → Response chunk → Push 到 Queue
  │
  ▼
Consumer[*Response]
  │
  ├─ Pop() / Each() 逐 chunk 消费
  └─ ResponseAccumulator.AddChunk() 累积为完整 Response
```

## 核心模块关系

```
┌─────────────────────────────────────────────┐
│                  调用者                       │
└──────────────────┬──────────────────────────┘
                   │
         ┌─────────▼─────────┐
         │  providers/       │ 统一接口 + 数据模型
         │  ├── interface    │ Provider / Model / CredentialManager
         │  ├── request      │ Request / GenerationConfig / Tool
         │  ├── response     │ Response / Choice / Usage
         │  ├── message      │ Message（多模态）
         │  └── accumulator  │ ResponseAccumulator
         └────┬────┬────┬────┘
              │    │    │
    ┌─────────▼┐ ┌─▼────────┐ ┌▼──────────┐
    │ kiro/    │ │ codex/   │ │ geminicli/ │  平台适配器
    │ converter│ │ converter│ │ converter  │  各自的 converter/ + sse/ + types/
    │ sse/     │ │ sse/     │ │ sse/       │
    └────┬─────┘ └────┬─────┘ └────┬───────┘
         │            │            │
         └──────┬─────┘────────────┘
                │
    ┌───────────▼───────────┐
    │    httpclient/        │  洋葱模型中间件 HTTP 客户端
    └───────────────────────┘
                │
    ┌───────────▼───────────┐
    │    credentials/       │  凭证接口 + 状态机 + 工厂注册
    └───────────────────────┘
                │
    ┌───────────▼───────────┐
    │  queue/ / pool/ /     │  基础设施
    │  usagerule/ / log/    │
    └───────────────────────┘
```

## 注册机制

- **Provider 注册**: 每个 Provider 在 `init()` 调用 `providers.Register()` → 全局注册表
- **查找**: `providers.GetProvider(type, name)`，name 未找到时回退到 "default"
- **凭证注册**: `credentials.Register(name, factory)` → 凭证工厂注册表

## 关键依赖

| 依赖 | 用途 |
|------|------|
| `aws-sdk-go-v2/aws/protocol/eventstream` | Kiro event-stream 协议解码 |
| `google/uuid` | UUID 生成 |
| `juju/errors` | 错误包装 |
| `panjf2000/ants/v2` | 高性能协程池 |
| `spf13/cobra` | CLI 命令行框架 |
| `stretchr/testify` | 测试断言 |
| `tiktoken-go/tokenizer` | Token 计数 |
| `uber-go/zap` | 结构化日志 |
