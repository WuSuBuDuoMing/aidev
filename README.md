# NeoCode (aidev v2)

**AI 编码代理 — 为国产模型和国际模型深度优化。**

基于 Go 语言重写的原生 AI 编码代理（v2.0.0），零依赖单二进制分发。

[English](README.en.md) | 中文

## 快速开始

```bash
# npm 安装（推荐）
npm i -g @neocode/cli

# Shell 安装
curl -fsSL https://get.neocode.dev | bash

# 配置 API Key
export DEEPSEEK_API_KEY=sk-xxx
# 或
export ANTHROPIC_API_KEY=sk-ant-xxx

# 运行
neocode
```

## 功能特性

- **62 个内置供应商预设** — DeepSeek、Claude、GPT、Gemini、MiMo、Qwen、GLM、Kimi + 20+ 中转服务
- **ccSwitch 兼容** — 读取所有标准 ccSwitch 环境变量
- **1M 上下文支持** — Gemini 2.5 Pro/Flash 动态压缩
- **工具调用** — 文件读写编辑、Shell 命令、Git、搜索、Web 获取
- **MCP 插件协议** — stdio/HTTP/SSE 三种传输
- **4 种权限模式** — Ask / Auto / Plan / Edit
- **5 级思考程度** — Low → Medium → High → XHigh → Ultracode + Workflows
- **实时用量** — Token 计数、费用估算、余额查询
- **自动更新** — `neocode upgrade` 一键升级
- **桌面端** — Windows (.exe)、macOS (.dmg)、Linux (.deb)

## 支持的模型

| Provider | 模型 | 上下文 |
|----------|------|--------|
| DeepSeek | v4, v4-pro, v4-flash | 128K |
| Claude | Fable 5, Opus 4.8, Sonnet 4.6, Haiku 4.5 | 200K |
| OpenAI | GPT-5.5, GPT-4o, o3, o3-mini | 200K |
| Gemini | 2.5 Pro/Flash, 3.5 Flash | **1M** |
| MiMo | v2.5, v2.5-pro | 128K |
| GLM | GLM-5.1 | 128K |
| Kimi | K2.7-code | 128K |

## 架构

```
aidev/
├── cmd/neocode/           # CLI 入口
├── internal/              # 核心内核
│   ├── agent/             # Agent 循环 + Storm Breaker + 子代理
│   ├── config/            # TOML 配置 + ccSwitch + 模型注册表
│   ├── provider/          # 3个 Provider + SSE流 + 重试
│   ├── tool/              # 10个内置工具 + MCP 客户端
│   ├── permission/        # 权限策略 (4种模式)
│   ├── session/           # SQLite 会话持久化
│   ├── skill/             # 技能系统
│   └── cli/               # 自升级
├── desktop/               # Wails v2 桌面端
├── npm/                   # npm 分发
├── scripts/               # Shell 安装器 + Homebrew
├── legacy/                # TypeScript v1 代码（存档）
└── docs/                  # 设计文档
```

## License

MIT
