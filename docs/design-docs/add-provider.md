# 新增 Provider 适配器指南

## 前置条件

- 熟悉目标平台的 API 文档和流协议
- 了解 `providers/interface.go` 的接口定义

## 步骤

### 1. 创建目录结构

```
providers/<platform>/
├── <platform>_provider.go    # Provider 实现
├── doc.go                    # 包文档（协议对齐参考来源）
├── models.go                 # ModelList + ModelMap
├── options.go                # Functional Options
├── errors.go                 # 平台错误 → HTTPError 映射
├── token_refresher.go        # Token 刷新
├── usage_limiter.go          # 用量规则
├── converter/
│   ├── req_builder.go        # 统一 Request → 平台 Request
│   └── resp_parser.go        # 平台 Response → 统一 Response
├── sse/                      # 流协议解析
│   ├── stream.go
│   └── parser.go
└── types/
    ├── request.go            # 平台 DTO
    └── response.go
```

### 2. 实现 Provider

参考 `providers/codex/codex_provider.go` 标准范式：

```go
func (p *xxxProvider) GenerateContentStream(ctx context.Context, req *providers.Request) (queue.Consumer[*providers.Response], error) {
    // 1. EnsureInvocationContext(ctx)
    // 2. req.Credential 类型断言
    // 3. 模型别名映射
    // 4. Converter 构建平台请求
    // 5. 构建 HTTP 请求（Header、Auth）
    // 6. httpClient.Do()
    // 7. 错误检查 → 归一化 HTTPError
    // 8. 启动 goroutine 流式解析 → Queue
    // 9. 返回 Consumer
}
```

`GenerateContent()` **必须**复用 `GenerateContentStream()` + `ResponseAccumulator`。

### 3. 注册

在 `init()` 中注册：
```go
func init() {
    providers.Register(&xxxProvider{})
}
```

### 4. 实现凭证

创建 `credentials/<platform>/credentials.go`，实现 `Credential` 接口。
在 `init()` 中注册工厂：`credentials.Register("<platform>", factory)`。

### 5. 文档

`doc.go` 中注明协议对齐的参考来源（如官方 API 文档 URL、SDK 源码路径）。

## 检查清单

- [ ] 实现 `Provider` 接口所有方法
- [ ] 错误归一化为 5 种 HTTPError 类型
- [ ] `GenerateContent` 由流式聚合实现
- [ ] `init()` 注册 Provider 和 Credential 工厂
- [ ] models.go 定义模型列表和别名映射
- [ ] doc.go 注明协议参考来源
