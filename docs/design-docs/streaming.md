# 流式响应机制

## 架构

```
Provider.GenerateContentStream()
  │
  ├─ 创建 Queue[*Response]（defaultQueueSize=100）
  │
  ├─ 启动 goroutine ──────────────────────┐
  │                                        │
  ▼                                        ▼
返回 Consumer[*Response]          流式解析循环
  │                               ├─ 逐事件解析 → Response chunk
  ├─ Pop() 逐条消费               ├─ Push 到 queue
  ├─ Each() 遍历消费              └─ 最后发送 Done=true + Usage
  └─ 配合 Accumulator 累积
```

## Consumer/Producer 队列

基于 `queue/` 包的泛型 channel 队列：

- `Producer[T].Push(item)` — 生产者推入
- `Producer[T].Close()` — 关闭队列
- `Consumer[T].Pop()` — 消费一条（阻塞）
- `Consumer[T].Each(ctx, fn)` — 遍历所有（直到关闭或 ctx 取消）

## ResponseAccumulator

参考 openai-go 的 `ChatCompletionAccumulator` 设计。

### 状态机

```
empty → content → tool → finished
```

### 关键方法

| 方法 | 作用 |
|------|------|
| `AddChunk(chunk)` | 逐 chunk 累积到完整 Response |
| `JustFinishedContent()` | 文本内容完成事件检测（调用 AddChunk 后立即检查） |
| `JustFinishedToolCall()` | 工具调用完成事件检测 + 返回 `FinishedToolCall` |
| `Response()` | 获取完整累积结果 |

### ToolCall 两种传递方式

Accumulator 同时处理：
1. **Delta 模式** — 每个 chunk 携带增量 ToolCall（需要拼接 Arguments）
2. **Message 模式** — 某些平台在单个 chunk 中携带完整 ToolCall

### FunctionDefinitionParam.Arguments 序列化

`Arguments` 字段是 `[]byte`，但 JSON 序列化为 string（避免双重编码）。
通过自定义 `MarshalJSON` / `UnmarshalJSON` 实现。

## 非流式聚合模式

所有 Provider 的 `GenerateContent()` 统一实现：

```go
reader, err := p.GenerateContentStream(ctx, req)
acc := &providers.ResponseAccumulator{}
reader.Each(ctx, func(chunk) { acc.AddChunk(chunk) })
return acc.Response()
```

**禁止**为 `GenerateContent` 实现独立的非流式路径。
