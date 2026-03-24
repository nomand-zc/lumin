# 错误处理规范

## 原则

所有平台的 HTTP 错误**必须**归一化为 `HTTPError`，上层依赖语义类型做决策。

## 5 种语义错误类型

| 类型 | 状态码 | 含义 | 可重试 |
|------|--------|------|--------|
| `bad_request` | 400 | 参数错误 | ❌ 不可重试 |
| `unauthorized` | 401 | Token 过期 | ✅ 刷新后重试 |
| `forbidden` | 403 | 权限不足 / 封禁 | ❌ 需人工介入 |
| `rate_limit` | 429 | 限流 / 限额 | ✅ 等待 CooldownUntil |
| `server_error` | 500 | 服务器错误 | ✅ 可重试 |

## HTTPError 结构

```go
type HTTPError struct {
    ErrorType     ErrorType   // 语义类型（上层依赖此字段）
    Message       string      // 人类可读描述
    CooldownUntil *time.Time  // 限流错误的冷却截止时间
    RawStatusCode int         // 原始状态码（仅调试）
    RawBody       string      // 原始响应体（仅调试）
}
```

## 非标准状态码映射

某些平台的状态码不是标准的 HTTP 语义，需要手动映射：

| 平台 | 原始状态码/错误 | 映射为 |
|------|----------------|--------|
| Kiro | 402（月度配额耗尽） | `rate_limit(429)` |
| GeminiCLI | VALIDATION_REQUIRED | `forbidden(403)` |

## 实现要求

每个 Provider 的 `errors.go` 中实现错误解析函数：
1. 读取 HTTP 响应状态码和响应体
2. 根据平台特定的错误格式解析错误信息
3. 映射为上述 5 种语义类型之一
4. 对 429 类错误，尽可能提取 `Retry-After` → `CooldownUntil`

## 禁止事项

- ❌ 上层逻辑不得依赖 `RawStatusCode` 或 `RawBody`
- ❌ 不允许透传平台原始错误类型
- ❌ 不允许新增第 6 种错误类型
