# NeoCode 技术设计文档

> 版本: 2.0.0 | 日期: 2026-06-13 | 状态: 设计阶段

## 1. 项目概述

NeoCode 是一个用 Go 语言重写的原生 AI 编码代理，专注于**国产模型深度优化 + 国际模型完整支持**。产出零依赖单二进制，支持 CLI + 桌面端（Wails）双形态分发。

### 核心定位

- **国产模型原生优化**：DeepSeek V4/Pro/Flash、MiMo v2.5/Pro、Qwen、GLM-5.1、Kimi-K2.7、StepFun、MiniMax
- **国际模型完整支持**：Claude Fable 5/Opus 4.8/Sonnet 4.6/Haiku 4.5、GPT-5.5/o3、Gemini 2.5 Pro/Flash (1M 上下文)
- **ccSwitch 生态兼容**：内置 62 个供应商预设，环境变量/配置格式完全兼容
- **多平台分发**：CLI（6 平台二进制 + npm + Homebrew）+ 桌面（Wails: .exe/.dmg/.tar.gz）

### 项目信息

| 字段 | 值 |
|------|-----|
| 项目名称 | NeoCode |
| CLI 命令 | `neocode` |
| Go module | `github.com/user/neocode` |
| npm 包 | `@neocode/cli` |
| 版本 | v2.0.0（全量重写） |
| License | MIT |
| 参考项目 | MiMo-Code、Reasonix、ccSwitch |

---

## 2. 架构设计

### 核心原则

1. **单一内核，多前端**：`internal/` 包含所有核心逻辑，CLI 和 Desktop 只是 I/O 表面
2. **CGO_ENABLED=0 CLI**：零 C 依赖单二进制，6 平台交叉编译
3. **前缀缓存优化**：追加式消息历史，byte-identical 前缀最大化 DeepSeek/MiMo 缓存命中
4. **1M 上下文感知**：根据模型的 context_window 动态调整压缩阈值
5. **ccSwitch 兼容**：读写标准环境变量，支持 ccSwitch 62 个供应商预设
6. **实时用量可见**：每个模型调用后显示 token + 费用 + 余额
7. **API Key 透明**：本地加密存储，首次使用时明确告知路径和安全说明

### 目录结构

```text
neocode/
├── cmd/neocode/main.go           # CLI 入口
├── internal/
│   ├── agent/                    # 核心循环、压缩、提示、会话、子代理
│   ├── config/                   # TOML 配置、Effort、模型注册表、ccSwitch
│   ├── provider/                 # Provider（OpenAI兼容/Anthropic/Google/注册/重试/流/缓存/计费）
│   ├── tool/                     # Tool 接口 + 内置工具 + MCP 适配
│   │   └── builtin/              # bash/read/write/edit/glob/grep/git/ls/webfetch
│   ├── permission/               # 权限策略 + bash 只读检测
│   ├── session/                  # 会话 CRUD + 历史 + 快照
│   ├── storage/                  # SQLite + API Key 加密存储
│   ├── context/                  # 项目上下文 + Git 状态
│   ├── monitor/                  # 实时用量 + 余额 + 上下文进度条
│   ├── cli/                      # 交互聊天 + 斜杠命令 + 升级 + 状态
│   ├── skill/                    # 技能发现 + YAML 加载
│   ├── i18n/                     # 中文/英文
│   └── util/                     # 路径安全 + Token 估算 + ANSI
├── desktop/                      # Wails 桌面端（独立 go.mod）
│   ├── frontend/                 # React + TailwindCSS
│   └── build/                    # 平台打包配置
├── npm/                          # npm 交叉编译 + 启动器
├── .goreleaser.yaml
├── Makefile
├── neocode.toml
├── docs/
├── README.md / README.zh-CN.md
└── .github/workflows/
```

---

## 3. 模型支持

### 最新模型注册表（2026-06）

#### Claude

| 模型 ID | 上下文 | 推理 | 价格 (in/out per 1M) |
|---------|--------|------|---------------------|
| claude-fable-5 | 200K | ✅ | $15 / $75 |
| claude-opus-4-8 | 200K | ✅ | $15 / $75 |
| claude-sonnet-4-6 | 200K | ✅ | $3 / $15 |
| claude-haiku-4-5 | 200K | ❌ | $0.80 / $4 |

#### DeepSeek

| 模型 ID | 上下文 | 推理 | 价格 |
|---------|--------|------|------|
| deepseek-v4 | 128K | ✅ | ¥1 / ¥2 |
| deepseek-v4-pro | 128K | ✅ | ¥2 / ¥8 |
| deepseek-v4-flash | 128K | ✅ | ¥0.5 / ¥2 |

#### OpenAI

| 模型 ID | 上下文 | 推理 | 价格 |
|---------|--------|------|------|
| gpt-5.5 | 200K | ✅ | $15 / $60 |
| gpt-5.4-mini | 200K | ✅ | $1 / $4 |
| gpt-o3 | 200K | ✅ | $10 / $40 |
| gpt-o3-mini | 200K | ✅ | $1.10 / $4.40 |
| gpt-4o | 128K | ❌ | $2.50 / $10 |

#### Gemini

| 模型 ID | 上下文 | 推理 | 价格 |
|---------|--------|------|------|
| gemini-2.5-pro | **1M** | ✅ | $1.25 / $10 |
| gemini-2.5-flash | **1M** | ✅ | $0.15 / $0.60 |
| gemini-3.5-flash | **1M** | ✅ | $0.075 / $0.30 |

#### 国产模型

| 模型 ID | Provider | 特殊处理 |
|---------|----------|---------|
| mimo-v2.5-pro | MiMo | 128K 输出上限 + X-Mimo-Source |
| glm-5.1 | 智谱 | Anthropic 兼容 |
| kimi-k2.7-code | 月之暗面 | Anthropic 兼容 |
| step-3.5-flash-2603 | StepFun | Anthropic 兼容 |
| minimax-m2.7 | MiniMax | Anthropic 兼容 |
| qianfan-code-latest | 百度千帆 | Anthropic 兼容 |
| ark-code-latest | 火山引擎 | Anthropic 兼容 |
| doubao-seed-2-0-code | 豆包 | OpenAI 兼容 |
| ling-2.5-1t | 百聆 | Anthropic 兼容 |

### 1M 上下文处理

```go
func CompactThreshold(ctxWindow int) int {
    switch {
    case ctxWindow >= 500_000:  return 400_000  // 1M 模型
    case ctxWindow >= 100_000:  return 80_000   // 128K/200K
    default:                    return 40_000   // 小上下文
    }
}
```

---

## 4. 核心系统

### 4 种权限模式

| 模式 | 行为 | 适用场景 |
|------|------|---------|
| **Ask** | 每个工具调用前弹出确认 | 默认，最安全 |
| **Auto** | 只读自动允许，写入根据配置 | 日常编码 |
| **Plan** | 只读分析，不执行修改 | 复杂任务 |
| **Edit** | 自动编辑，无需确认 | 信任场景 |

### 思考程度控制 (Effort)

| 级别 | Claude | DeepSeek | OpenAI | Gemini |
|------|--------|----------|--------|--------|
| Low | budget: 1K | - | effort: low | budget: 1K |
| Medium | budget: 8K | - | effort: medium | budget: 8K |
| High | budget: 32K | - | effort: high | budget: 32K |
| XHigh | budget: 100K | thinking: true | effort: high+tools | budget: 100K |
| **Ultracode** | budget: max + tools | thinking + 3 retries | o3-pro + tools | budget: max + code exec |

### 实时用量 + 余额

```text
╭─ neocode ──────────────────────────────────────────────────────────────────╮
│  Model: deepseek-v4-pro (DeepSeek) | Effort: High | Mode: Auto           │
│  Tokens: 12,340 in / 2,150 out | Cost: $0.0082 | Balance: ¥48.50        │
│  Context: ████████████░░░░░░░░ 62% (79K/128K)                            │
╰───────────────────────────────────────────────────────────────────────────╯
```

### ccSwitch 供应商兼容（62 个预设）

**国内直连**：DeepSeek、MiMo、GLM、Kimi、阿里百炼、百度千帆、火山引擎、豆包、StepFun、MiniMax、百聆、ModelScope

**国际**：Claude、OpenAI、Gemini、GitHub Copilot、AWS Bedrock、OpenRouter、NVIDIA

**中转服务**：NewAPI、AiHubMix、SiliconFlow、DMXAPI、PackyCode、CCSub、RunAPI、APIKEY.FUN 等 20+ 个

### MCP / Skills / Agents

- **MCP 插件**：stdio/HTTP/SSE 三种传输，`.mcp.json` 自动发现
- **Skills 技能**：YAML frontmatter + markdown body
- **Agents 子代理**：explore（只读）、review（审查）、security（安全审计）

### 自动更新

- CLI：`neocode upgrade` — GitHub Releases + SHA256 + 二进制替换
- 桌面端：Minisign 签名 + `latest.json` + 平台特定策略

### 历史聊天记录

- SQLite 持久化
- `/history` 浏览 + `/history search` 搜索
- 会话自动标题

### API Key 安全

```text
~/.neocode/
├── config.toml       # 配置
├── keystore.enc      # 加密 API Keys (AES-256-GCM)
├── history.db        # SQLite 会话
├── skills/           # 全局技能
└── plugins/          # MCP 插件
```

首次使用明确告知存储路径和安全说明。

---

## 5. 分阶段实施

| 阶段 | 内容 | 周期 |
|------|------|------|
| Phase 1 | 脚手架 + 配置 + Provider + SSE 流 + 基础 CLI | Week 1-2 |
| Phase 2 | 工具系统 + 权限 + Agent 循环 + 上下文压缩 | Week 2-3 |
| Phase 3 | 会话管理 + CLI 完善 + 技能 + 国际化 + 升级 | Week 3-4 |
| Phase 4 | MCP 插件 + Plan 模式 + 子代理 + 内存系统 | Week 4-5 |
| Phase 5 | Wails 桌面端 + 打包 + 自动更新 | Week 5-7 |
| Phase 6 | 分发 + CI/CD + 文档 | Week 7-8 |

---

## 6. 分发方式

| 形态 | 平台 | 格式 |
|------|------|------|
| CLI | macOS (Intel+ARM) | tar.gz + Homebrew |
| CLI | Windows (x64+ARM) | zip |
| CLI | Linux (amd64+ARM) | tar.gz + .deb |
| CLI | npm | `@neocode/cli` (6 平台子包) |
| CLI | Shell 安装器 | `curl -fsSL https://get.neocode.dev \| bash` |
| Desktop | macOS | .dmg (Universal Binary) |
| Desktop | Windows | NSIS 安装器 (.exe) |
| Desktop | Linux | .tar.gz + .deb |
