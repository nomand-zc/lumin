# 凭证生命周期

## 凭证状态机

```
                ┌──────────────┐
                │  Available   │ ← 初始状态 / 刷新成功 / 恢复
                └──┬──┬──┬──┬─┘
                   │  │  │  │
        ┌──────────┘  │  │  └──────────┐
        ▼             ▼  ▼             ▼
  ┌──────────┐ ┌──────────┐ ┌────────────────┐
  │ Expired  │ │  Banned  │ │ UsageLimited   │
  │  (2)     │ │  (4)     │ │  (5)           │
  └────┬─────┘ └──────────┘ └───────┬────────┘
       │        ❌ 需人工修复         │
       ▼                            ▼
  Refresh()                   等待窗口重置
       │                            │
       ├── 成功 → Available          └── → Available
       └── 失败 ↓
             ┌──────────────┐  ┌────────────────────┐
             │ Invalidated  │  │ ReauthRequired     │
             │  (3)         │  │  (6)               │
             └──────────────┘  └────────────────────┘
              ❌ 永久失效          ❌ 需用户重新授权
```

## 6 种状态

| 状态 | 值 | 含义 | 恢复方式 |
|------|---|------|---------|
| `StatusAvailable` | 1 | 正常可用 | — |
| `StatusExpired` | 2 | Token 过期 | 调用 Refresh() |
| `StatusInvalidated` | 3 | 永久失效 | 不可恢复 |
| `StatusBanned` | 4 | 被封禁 | 需人工修复 |
| `StatusUsageLimited` | 5 | 用量限制 | 等待窗口重置 |
| `StatusReauthRequired` | 6 | 需重新授权 | 需人工操作 |

## Credential 接口

```go
type Credential interface {
    Clone() Credential              // 深拷贝（并发安全）
    Validate() error                // 验证必需字段
    GetAccessToken() string
    GetRefreshToken() string
    GetExpiresAt() *time.Time
    IsExpired() bool
    GetUserInfo() (UserInfo, error)
    ToMap() map[string]any          // 序列化（配合 GetValue[V]() 泛型取值）
}
```

## 新增凭证实现

1. 创建 `credentials/<platform>/credentials.go`
2. 实现 `Credential` 接口所有方法
3. `init()` 中注册工厂：`credentials.Register("<platform>", factory)`
4. `Validate()` 使用 `errs.go` 中预定义的错误常量报告缺失字段

预定义错误常量：`ErrAccessTokenEmpty` / `ErrRefreshTokenEmpty` / `ErrExpiresAtEmpty` / `ErrRegionEmpty` / `ErrClientIDEmpty` 等
