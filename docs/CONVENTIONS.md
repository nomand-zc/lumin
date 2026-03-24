# 编码规范和命名约定

## 命名

| 类别 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写单词 | `providers`, `credentials`, `httpclient` |
| 接口 | 行为命名，无 `I` 前缀 | `Provider`, `Credential`, `TokenCounter` |
| Provider 类型 | 小写私有结构体 | `kiroProvider`, `codexProvider` |
| 常量 | CamelCase | `ErrorTypeBadRequest` |
| 文件名 | 小写下划线 | `token_conter.go`, `http_errors.go` |

## 设计模式

| 模式 | 使用场景 |
|------|---------|
| Functional Options | 构造器（Provider / HTTPClient / Response / TokenCounter） |
| Builder / Parser | `converter/` 中的请求构建和响应解析 |
| Registry | Provider 和 Credential 的全局注册 |
| 洋葱中间件 | HTTPClient 的 `RoundTripperMiddleware` |
| Consumer / Producer | 泛型队列，流式传输 |
| Context 传值 | `Invocation` 通过 `context.WithValue` 透传 |

## 测试

- 测试文件 `_test.go`，与被测文件同包
- 使用 `github.com/stretchr/testify` 断言
- 测试数据放 `testdata/`
- 运行：`go test ./...`
