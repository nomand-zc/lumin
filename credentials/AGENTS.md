# AGENTS.md — credentials/

> 凭证层入口。详细设计请跟随链接。

## 职责

定义统一 `Credential` 接口 + 凭证状态管理 + 各平台凭证实现。

## 文件速查

| 文件 | 职责 |
|------|------|
| `interface.go` | `Credential` 接口 + `UserInfo` + `GetValue[V]()` |
| `credential_status.go` | 6 种状态枚举 |
| `errs.go` | 预定义验证错误常量 |
| `register.go` | 凭证工厂注册表 |

## 关键约束

1. 凭证必须支持 `Clone()` 深拷贝（并发安全）
2. `Validate()` 使用 `errs.go` 中的预定义错误常量
3. 工厂必须在 `init()` 中注册

## 深入阅读

- [凭证生命周期](../docs/design-docs/credential-lifecycle.md)
