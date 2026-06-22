<div align="center">

# NeoCode

**AI 编码代理 -- 为国产模型和国际模型深度优化。**

基于 Go 语言重写的原生 AI 编码代理，零依赖单二进制分发。
深度优化 DeepSeek、MiMo、Qwen、GLM、Kimi，完整支持 Claude、GPT、Gemini。

[![CI](https://img.shields.io/github/actions/workflow/status/WuSuBuDuoMing/aidev/ci.yml?style=flat-square&label=CI)](https://github.com/WuSuBuDuoMing/aidev/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](./LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/WuSuBuDuoMing/aidev?style=flat-square&color=yellow)](https://github.com/WuSuBuDuoMing/aidev/stargazers)
[![npm](https://img.shields.io/npm/v/@neocode/cli?style=flat-square&color=red)](https://www.npmjs.com/package/@neocode/cli)

[English](README.md) | 简体中文

</div>

---

## 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [支持的模型](#支持的模型)
- [CLI 命令](#cli-命令)
- [配置系统](#配置系统)
- [MCP 插件](#mcp-插件)
- [架构](#架构)
- [桌面应用](#桌面应用)
- [安全特性](#安全特性)
- [API 参考](#api-参考)
- [贡献指南](#贡献指南)
- [更新日志](#更新日志)
- [许可证](#许可证)

---

## 功能特性

| 功能 | 描述 |
|------|------|
| **62 个内置供应商预设** | DeepSeek、Claude、GPT、Gemini、MiMo、Qwen、GLM、Kimi + 20+ 中转服务 |
| **ccSwitch 兼容** | 读取所有标准 ccSwitch 环境变量，无缝迁移 |
| **1M 上下文支持** | Gemini 2.5 Pro/Flash 动态上下文压缩 |
| **工具调用** | 10 个内置工具：文件读写编辑、Shell 命令、Git、搜索、Web 获取 |
| **MCP 插件协议** | stdio/HTTP/SSE 三种传输，`.mcp.json` 自动发现 |
| **4 种权限模式** | Ask / Auto / Plan / Edit -- 精细控制工具执行 |
| **5 级思考程度** | Low 到 Ultracode + Workflows -- 按任务控制推理深度 |
| **实时用量统计** | 每次响应后显示 Token 计数、费用估算、余额查询 |
| **子代理系统** | 内置探索、审查、安全、研究子代理 |
| **会话持久化** | SQLite 存储对话历史，支持全文搜索 |
| **技能系统** | YAML 前置元数据 + Markdown 正文，项目级技能覆盖全局技能 |
| **自动更新** | `neocode upgrade` 一键升级，SHA256 校验 |
| **桌面应用** | Windows (.exe)、macOS (.dmg)、Linux (.deb)，基于 Wails v2 |
| **Storm Breaker** | 死循环检测，防止工具调用失控 |

---

## 快速开始

## 安装

NeoCode 支持 **macOS**、**Linux** 和 **Windows**，选择适合你的方式。

### npm（全平台 -- 推荐）

```bash
npm i -g @neocode/cli
```

### Homebrew（macOS / Linux）

```bash
brew install neocode/neocode/neocode
```

### Shell 安装脚本（macOS / Linux）

```bash
curl -fsSL https://get.neocode.dev | bash
```

### Go 安装（全平台）

需要 Go 1.22+。

```bash
go install github.com/WuSuBuDuoMing/aidev/cmd/neocode@latest
```

### macOS -- 独立二进制

从 [GitHub Releases](https://github.com/WuSuBuDuoMing/aidev/releases) 下载最新版本：

```bash
# Intel 芯片
curl -fsSL -o neocode.tar.gz https://github.com/WuSuBuDuoMing/aidev/releases/latest/download/neocode_darwin_amd64.tar.gz
tar xzf neocode.tar.gz
sudo mv neocode /usr/local/bin/

# Apple Silicon（M1/M2/M3/M4）
curl -fsSL -o neocode.tar.gz https://github.com/WuSuBuDuoMing/aidev/releases/latest/download/neocode_darwin_arm64.tar.gz
tar xzf neocode.tar.gz
sudo mv neocode /usr/local/bin/
```

### Linux -- 独立二进制

```bash
# x86_64
curl -fsSL -o neocode.tar.gz https://github.com/WuSuBuDuoMing/aidev/releases/latest/download/neocode_linux_amd64.tar.gz
tar xzf neocode.tar.gz
sudo mv neocode /usr/local/bin/

# ARM64
curl -fsSL -o neocode.tar.gz https://github.com/WuSuBuDuoMing/aidev/releases/latest/download/neocode_linux_arm64.tar.gz
tar xzf neocode.tar.gz
sudo mv neocode /usr/local/bin/
```

### Linux -- APT（Debian/Ubuntu）

```bash
# 从 GitHub Releases 下载 .deb 包，然后：
sudo dpkg -i neocode_linux_amd64.deb
```

### Windows -- 独立二进制 (.exe)

从 [GitHub Releases](https://github.com/WuSuBuDuoMing/aidev/releases) 下载 `neocode_windows_amd64.zip`，解压后即可使用 `neocode.exe`。

```powershell
# 一键下载安装（推荐）
Invoke-WebRequest -Uri "https://github.com/WuSuBuDuoMing/aidev/releases/latest/download/neocode_windows_amd64.zip" -OutFile neocode.zip
Expand-Archive -Path neocode.zip -DestinationPath C:\neocode
# 将 C:\neocode 添加到系统 PATH

# 或通过 winget（即将上线）
winget install WuSuBuDuoMing.NeoCode
```

**Windows 用户快速上手：**

```powershell
# 安装完成后，直接运行：
neocode version          # 验证安装
neocode                  # 启动交互式聊天
neocode "你好，请帮我优化这段代码"   # 单次提问
```

Windows 构建支持 **x86_64**（`neocode_windows_amd64.zip`）和 **ARM64**（`neocode_windows_arm64.zip`），零外部依赖。

### Docker

```bash
# 本地构建
docker build -t neocode .
docker run -it --rm \
  -e DEEPSEEK_API_KEY=$DEEPSEEK_API_KEY \
  -v $(pwd):/workspace \
  neocode

# 或从 GitHub Container Registry 拉取（即将上线）
docker run -it --rm \
  -e DEEPSEEK_API_KEY=$DEEPSEEK_API_KEY \
  -v $(pwd):/workspace \
  ghcr.io/wusubuduoming/neocode:latest
```

### 预构建包汇总

| 平台 | 格式 | 文件名 |
|------|------|--------|
| macOS (Intel) | `.tar.gz` | `neocode_darwin_amd64.tar.gz` |
| macOS (Apple Silicon) | `.tar.gz` | `neocode_darwin_arm64.tar.gz` |
| Linux (x86_64) | `.tar.gz` | `neocode_linux_amd64.tar.gz` |
| Linux (ARM64) | `.tar.gz` | `neocode_linux_arm64.tar.gz` |
| Linux (Debian/Ubuntu) | `.deb` | `neocode_linux_amd64.deb` |
| Windows (x86_64) | `.zip` | `neocode_windows_amd64.zip` |
| Windows (ARM64) | `.zip` | `neocode_windows_arm64.zip` |
| Docker | 镜像 | `ghcr.io/wusubuduoming/neocode` |

所有二进制均可在 [Releases](https://github.com/WuSuBuDuoMing/aidev/releases) 页面获取，附带 SHA256 校验和。

### 配置 API Key

```bash
# DeepSeek（默认）
export DEEPSEEK_API_KEY=sk-xxx

# Anthropic Claude
export ANTHROPIC_API_KEY=sk-ant-xxx

# OpenAI GPT
export OPENAI_API_KEY=sk-xxx

# Google Gemini
export GOOGLE_API_KEY=xxx

# 或使用统一 Key
export NEOCODE_API_KEY=sk-xxx
```

### 运行

```bash
# 交互式聊天
neocode

# 单次提问
neocode "如何优化这个 Go 函数的性能？"

# 查看状态
neocode status

# 列出供应商
neocode providers
```

---

## 支持的模型

| 供应商 | 模型 | 上下文 | 定价（输入/输出 每百万 Token） |
|--------|------|--------|-------------------------------|
| **DeepSeek** | v4, v4-pro, v4-flash | 128K | $0.10/$0.40 ~ $0.55/$2.19 |
| **Claude** | Fable 5, Opus 4.8, Sonnet 4.6, Haiku 4.5 | 200K | $0.80/$4 ~ $15/$75 |
| **OpenAI** | GPT-5.5, GPT-5.4-mini, GPT-4o, o3, o3-mini | 128K-200K | $1/$4 ~ $15/$60 |
| **Gemini** | 2.5 Pro/Flash, 3.5 Flash | **1M** | $0.075/$0.30 ~ $1.25/$10 |
| **MiMo** | v2.5, v2.5-pro | 128K | 免费 |
| **GLM** | GLM-5.1 | 128K | -- |
| **Kimi** | K2.7-code | 128K | -- |
| **StepFun** | step-3.5-flash-2603 | 128K | -- |
| **MiniMax** | minimax-m2.7 | 128K | -- |
| **百度** | qianfan-code-latest | 128K | -- |
| **火山引擎** | ark-code-latest, doubao-seed-2.0-code | 128K | -- |
| **百灵** | ling-2.5-1t | 128K | -- |
| **Ollama** | llama3.1, qwen2.5, deepseek-v3（本地） | 128K | 免费 |

---

## CLI 命令

### 命令行用法

```bash
neocode                         # 启动交互式聊天
neocode "问题"                   # 单次提问模式
neocode status                  # 显示配置和状态
neocode providers               # 列出可用供应商
neocode version                 # 显示版本
neocode upgrade                 # 自动更新到最新版本
```

### 斜杠命令（交互式）

```
/help              显示所有命令
/status            显示当前配置
/providers         列出可用供应商
/model <id>        运行时切换模型
/provider <id>     运行时切换供应商
/effort <lvl>      设置思考级别
/mode <mode>       设置权限模式
/plan <task>       以只读计划模式分析任务
/agents            列出可用子代理
/history [n]       列出最近会话
/resume <id>       恢复历史会话
/skills            列出可用技能
/new               开始新对话
/upgrade           检查更新
/clear             清屏
/exit              退出
```

---

## 配置系统

NeoCode 使用 4 层配置系统，具有清晰的优先级规则：

```
环境变量 > 项目配置 (.neocode/config.toml) > 用户配置 (~/.neocode/config.toml) > 默认值
```

### 环境变量

```bash
NEOCODE_PROVIDER=deepseek        # 最高优先级
NEOCODE_MODEL=deepseek-v4
NEOCODE_API_KEY=sk-xxx
DEEPSEEK_API_KEY=sk-xxx          # 供应商特定
ANTHROPIC_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
GOOGLE_API_KEY=xxx
```

### 配置文件 (`~/.neocode/config.toml`)

```toml
default_provider = "deepseek"
default_model = "deepseek-v4"
temperature = 0.2
max_tokens = 4096
theme = "dark"
stream = true
language = "zh-CN"
permission_mode = "ask"
effort = "medium"
show_token_counts = true
```

---

## MCP 插件

在项目根目录创建 `.mcp.json` 以自动发现 MCP 服务器：

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"]
    }
  }
}
```

NeoCode 启动时自动发现并连接。MCP 工具以 `mcp__<服务器名>__<工具名>` 命名。

---

## 架构

```
aidev/
|-- cmd/neocode/           # CLI 入口
|-- internal/              # 核心内核
|   |-- agent/             # Agent 循环 + Storm Breaker + 子代理 + 规划器
|   |-- cli/               # 从 GitHub Releases 自升级
|   |-- config/            # TOML 配置 + ccSwitch + 模型注册表（30+ 模型）
|   |-- config/model/      # 模型规格、上下文阈值、定价
|   |-- context/           # 动态阈值上下文压缩
|   |-- i18n/              # 双语（中/英）UI 字符串管理
|   |-- monitor/           # 实时 Token 用量跟踪与费用估算
|   |-- permission/        # 权限策略引擎（4 种模式）
|   |-- provider/          # 3 个 Provider（OpenAI 兼容、Anthropic、Gemini）+ SSE + 重试
|   |-- session/           # SQLite 会话持久化（WAL 模式）
|   |-- skill/             # YAML+Markdown 技能系统
|   |-- storage/           # SQLite 数据库层（含迁移）
|   |-- tool/              # 10 个内置工具 + MCP 客户端（stdio/HTTP/SSE）
|   |   |-- builtin/       # 内置工具实现（含单元测试）
|   |   +-- mcp/           # MCP 协议客户端与工具适配器
|   |-- types/             # 核心类型定义
|   +-- util/              # 共享工具函数
|-- desktop/               # Wails v2 桌面应用（React + TailwindCSS）
|-- npm/                   # npm 分发（6 平台包 + 元包）
|-- scripts/               # Shell 安装器 + Homebrew formula
|-- legacy/                # TypeScript v1 代码（存档）
+-- docs/                  # 设计文档 + 配置参考
```

---

## 桌面应用

```bash
# macOS
open NeoCode.dmg

# Windows
NeoCode-installer.exe

# Linux
tar xzf NeoCode-linux-amd64.tar.gz
./neocode
```

桌面应用基于 Wails v2 构建（Go 后端 + React 前端）。

---

## 安全特性

- **API Key 加密存储**: 使用 AES-256-GCM 加密存储于 `~/.neocode/keystore.enc`
- **路径沙箱**: 写操作限制在项目目录内
- **无 Shell 注入**: Shell 命令使用参数数组，非字符串拼接
- **权限模式**: 细粒度控制哪些工具可以执行
- **Plan 模式**: 只读分析，不修改任何文件
- **SSRF 防护**: Web 请求操作受到沙箱保护

详见 [SECURITY.md](./SECURITY.md) 了解漏洞报告政策。

---

## API 参考

### Provider 接口

```go
type Provider interface {
    Generate(ctx context.Context, messages []Message, tools []ToolDefinition, systemPrompt string) (*GenerateResult, error)
    Stream(ctx context.Context, messages []Message, tools []ToolDefinition, systemPrompt string, onChunk func(StreamChunk)) (*GenerateResult, error)
    Name() string
    DefaultModel() string
}
```

### Tool 接口

```go
type Tool interface {
    Name() string
    Description() string
    Schema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
    ReadOnly() bool
}
```

---

## 贡献指南

我们欢迎各种形式的贡献 -- Bug 报告、功能请求、文档改进和代码提交。

详见 [CONTRIBUTING.md](./CONTRIBUTING.md)。

---

## 更新日志

详见 [CHANGELOG.md](./CHANGELOG.md)。

---

## 许可证

NeoCode 基于 [MIT 许可证](./LICENSE) 发布。

Copyright (c) 2024-2026 WuSuBuDuoMing

---

<div align="center">

**用心打造，为终端开发者服务。**

[报告 Bug](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=bug_report.md) | [功能建议](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md) | [参与讨论](https://github.com/WuSuBuDuoMing/aidev/discussions)

</div>
