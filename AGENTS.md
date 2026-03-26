# AGENTS.md — lumin-client

> 这是一张地图，不是说明书。详细内容请跟随链接。

## 这是什么

统一多平台 AI 模型调用 SDK（Go），屏蔽 Kiro / Codex / GeminiCLI / iFlow 协议差异。
模块路径: `github.com/nomand-zc/lumin-client` · Go 1.24.11

## 快速导航

| 你想做什么 | 去哪里看 |
|-----------|---------|
| 理解整体架构和数据流 | → [ARCHITECTURE.md](ARCHITECTURE.md) |
| 了解核心设计决策和原则 | → [docs/design-docs/core-beliefs.md](docs/design-docs/core-beliefs.md) |
| 新增一个 Provider 适配器 | → [docs/design-docs/add-provider.md](docs/design-docs/add-provider.md) |
| 理解错误处理规范 | → [docs/design-docs/error-handling.md](docs/design-docs/error-handling.md) |
| 理解凭证管理 | → [docs/design-docs/credential-lifecycle.md](docs/design-docs/credential-lifecycle.md) |
| 理解流式响应机制 | → [docs/design-docs/streaming.md](docs/design-docs/streaming.md) |
| 查看编码规范和约定 | → [docs/CONVENTIONS.md](docs/CONVENTIONS.md) |
| 了解各平台差异对照 | → [docs/references/platform-comparison.md](docs/references/platform-comparison.md) |

## 目录结构（一句话速览）

```
providers/          → 核心接口 + 统一模型 + 4 个平台适配器
credentials/        → 凭证接口 + 状态机 + 各平台凭证实现
httpclient/         → 洋葱模型中间件 HTTP 客户端
queue/              → 泛型 Consumer/Producer 队列（流式传输）
pool/               → 内存池 + 协程池
usagerule/          → 多粒度时间窗口用量限制
log/                → Logger 接口（zap）
utils/              → 工具函数
cli/                → 命令行工具（cobra）
```

## 关键约束（必须遵守）

1. **所有 HTTP 错误必须归一化**为 5 种 `HTTPError` 类型（400/401/403/429/500），见 [error-handling.md](docs/design-docs/error-handling.md)
2. **`GenerateContent` 必须由 `GenerateContentStream` + `ResponseAccumulator` 聚合实现**，不允许独立实现
3. **Provider 必须在 `init()` 中注册**到全局注册表，凭证工厂同理
4. **`req.Credential` 类型断言**：传入错误类型会 panic，调用者负责确保匹配
5. **提交前必须运行 pre-commit 检查**：安装后自动执行，详见下方 Pre-commit 配置

## 子目录 AGENTS.md

各模块有自己的精简 AGENTS.md，只描述该模块的边界和关键约定：

- [providers/AGENTS.md](providers/AGENTS.md) — Provider 层入口
- [credentials/AGENTS.md](credentials/AGENTS.md) — 凭证层入口

## 文档索引

```
docs/
├── design-docs/              ← 设计文档（架构决策和模式）
│   ├── index.md              ← 设计文档总览
│   ├── core-beliefs.md       ← 核心设计原则
│   ├── add-provider.md       ← 新增 Provider 指南
│   ├── error-handling.md     ← 错误处理规范
│   ├── credential-lifecycle.md ← 凭证生命周期
│   └── streaming.md          ← 流式响应机制
├── references/               ← 参考资料
│   └── platform-comparison.md ← 平台差异对照
└── CONVENTIONS.md            ← 编码规范和命名约定
```

## Pre-commit 配置

项目使用 [pre-commit](https://pre-commit.com/) 进行代码提交前的自动化质量检查，确保代码符合项目的统一规范。

### 安装

```bash
# 检查是否已安装 pre-commit
pre-commit --version

# 如果未安装，使用以下命令安装：
# 方式一：使用 pip（推荐）
pip install pre-commit

# 方式二：使用 Homebrew（macOS）
brew install pre-commit

# 方式三：使用 apt（Ubuntu/Debian）
sudo apt install pre-commit

# 安装完成后，在项目根目录安装 git hooks
pre-commit install
```

### 检查项说明

| 类别 | 检查项 | 说明 |
|------|--------|------|
| **Go 代码** | `go fmt` | 格式化 Go 代码 |
| | `goimports` | 导入语句排序和格式化 |
| | `go vet` | 静态分析检查 |
| | `golangci-lint` | 综合代码检查 |
| | `go mod tidy` | 依赖整理 |
| **通用检查** | `trailing-whitespace` | 行尾空白检查 |
| | `end-of-file-fixer` | 文件末尾换行检查 |
| | `check-added-large-files` | 大文件检查（≤500KB） |
| | `check-yaml` / `check-json` | YAML/JSON 语法检查 |
| | `check-merge-conflict` | 合并冲突标记检查 |
| **文档** | `pymarkdown` | Markdown 格式化 |

### 使用方式

```bash
# 对所有文件运行检查
pre-commit run --all-files

# 只检查暂存的文件（推荐）
pre-commit run

# 手动触发特定检查
pre-commit run go-vet --all-files
```

### 跳过检查

**仅在紧急情况下使用**：

```bash
# 跳过所有检查提交（不推荐）
git commit --no-verify
```

### 配置文件

详细配置见 [.pre-commit-config.yaml](.pre-commit-config.yaml)
