<div align="center">

# NeoCode

**AI coding agent -- optimized for Chinese and international models.**

A Go-based native AI coding agent with zero-dependency single-binary distribution.
Deep optimization for DeepSeek, MiMo, Qwen, GLM, Kimi, and full support for Claude, GPT, Gemini.

[![CI](https://img.shields.io/github/actions/workflow/status/WuSuBuDuoMing/aidev/ci.yml?style=flat-square&label=CI)](https://github.com/WuSuBuDuoMing/aidev/actions)
[![Go Report](https://goreportcard.com/badge/github.com/WuSuBuDuoMing/aidev?style=flat-square)](https://goreportcard.com/report/github.com/WuSuBuDuoMing/aidev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](./LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/WuSuBuDuoMing/aidev?style=flat-square&color=yellow)](https://github.com/WuSuBuDuoMing/aidev/stargazers)
[![npm](https://img.shields.io/npm/v/@neocode/cli?style=flat-square&color=red)](https://www.npmjs.com/package/@neocode/cli)

English | [Chinese](README.zh-CN.md)

</div>

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Supported Models](#supported-models)
- [CLI Commands](#cli-commands)
- [Configuration](#configuration)
- [MCP Plugins](#mcp-plugins)
- [Architecture](#architecture)
- [Desktop Apps](#desktop-apps)
- [Security](#security)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Changelog](#changelog)
- [License](#license)

---

## Features

| Feature | Description |
|---------|-------------|
| **62 built-in provider presets** | DeepSeek, Claude, GPT, Gemini, MiMo, Qwen, GLM, Kimi + 20+ relay services |
| **ccSwitch compatible** | Reads all standard ccSwitch environment variables for seamless migration |
| **1M context support** | Gemini 2.5 Pro/Flash with dynamic context compaction |
| **Tool calling** | 10 built-in tools: file read/write/edit, shell commands, git, search, web fetch |
| **MCP plugin protocol** | stdio/HTTP/SSE transports with auto-discovery from `.mcp.json` |
| **4 permission modes** | Ask, Auto, Plan, Edit -- fine-grained control over tool execution |
| **5 effort levels** | Low to Ultracode + Workflows -- control reasoning depth per task |
| **Real-time usage** | Token counts, cost estimation, and balance queries after every response |
| **Sub-agent system** | Built-in explore, review, security, and research sub-agents |
| **Session persistence** | SQLite-backed conversations with full-text search |
| **Skills system** | YAML frontmatter + markdown body, project-level override of global skills |
| **Auto-update** | `neocode upgrade` one-click upgrade with SHA256 verification |
| **Desktop apps** | Windows (.exe), macOS (.dmg), Linux (.deb) via Wails v2 |
| **Storm Breaker** | Death-spiral loop detection prevents runaway tool calls |

---

## Quick Start

### Installation

```bash
# npm (recommended)
npm i -g @neocode/cli

# Homebrew (macOS / Linux)
brew install neocode/neocode/neocode

# Shell installer
curl -fsSL https://get.neocode.dev | bash

# Go install from source
go install github.com/WuSuBuDuoMing/aidev/cmd/neocode@latest
```

### Configure API Key

```bash
# DeepSeek (default)
export DEEPSEEK_API_KEY=sk-xxx

# Anthropic Claude
export ANTHROPIC_API_KEY=sk-ant-xxx

# OpenAI GPT
export OPENAI_API_KEY=sk-xxx

# Google Gemini
export GOOGLE_API_KEY=xxx

# Or set a unified key
export NEOCODE_API_KEY=sk-xxx
```

### Run

```bash
# Interactive chat
neocode

# Single question mode
neocode "How do I optimize this Go function for performance?"

# Show status
neocode status

# List providers
neocode providers
```

---

## Supported Models

| Provider | Models | Context | Pricing (in/out per 1M tokens) |
|----------|--------|---------|--------------------------------|
| **DeepSeek** | v4, v4-pro, v4-flash | 128K | $0.10/$0.40 ~ $0.55/$2.19 |
| **Claude** | Fable 5, Opus 4.8, Sonnet 4.6, Haiku 4.5 | 200K | $0.80/$4 ~ $15/$75 |
| **OpenAI** | GPT-5.5, GPT-5.4-mini, GPT-4o, o3, o3-mini | 128K-200K | $1/$4 ~ $15/$60 |
| **Gemini** | 2.5 Pro/Flash, 3.5 Flash | **1M** | $0.075/$0.30 ~ $1.25/$10 |
| **MiMo** | v2.5, v2.5-pro | 128K | Free |
| **GLM** | GLM-5.1 | 128K | -- |
| **Kimi** | K2.7-code | 128K | -- |
| **StepFun** | step-3.5-flash-2603 | 128K | -- |
| **MiniMax** | minimax-m2.7 | 128K | -- |
| **Baidu** | qianfan-code-latest | 128K | -- |
| **Volcengine** | ark-code-latest, doubao-seed-2.0-code | 128K | -- |
| **BaiLing** | ling-2.5-1t | 128K | -- |
| **Ollama** | llama3.1, qwen2.5, deepseek-v3 (local) | 128K | Free |

---

## CLI Commands

### Command Line Usage

```bash
neocode                         # Start interactive chat
neocode "question"              # Single question mode
neocode status                  # Show configuration and status
neocode providers               # List available providers
neocode version                 # Show version
neocode upgrade                 # Self-update to latest release
```

### Command Line Options

```
-p, --provider <name>      Set provider (deepseek, claude, openai, gemini, ...)
-m, --model <id>           Set model ID
-k, --api-key <key>        Set API key
--effort <level>           Set effort (low, medium, high, xhigh, ultracode)
--mode <mode>              Set permission mode (ask, auto, plan, edit)
--no-stream                Disable streaming output
--thinking                 Enable thinking/reasoning mode
```

### Slash Commands (Interactive)

```
/help              Show all commands
/status            Show current configuration
/providers         List available providers
/model <id>        Switch model at runtime
/provider <id>     Switch provider at runtime
/effort <lvl>      Set effort level (low/medium/high/xhigh/ultracode)
/mode <mode>       Set permission mode (ask/auto/plan/edit)
/plan <task>       Analyze a task in read-only plan mode
/agents            List available sub-agents
/history [n]       List recent sessions
/history search X  Search sessions by content
/resume <id>       Resume a past session
/skills            List available skills
/new               Start a new conversation
/upgrade           Check for updates
/clear             Clear screen
/exit              Exit
```

---

## Configuration

NeoCode uses a 4-layer configuration system with clear priority rules:

```
Environment variables  >  Project (.neocode/config.toml)  >  User (~/.neocode/config.toml)  >  Defaults
```

### Environment Variables

```bash
# NeoCode-specific (highest priority)
NEOCODE_PROVIDER=deepseek
NEOCODE_MODEL=deepseek-v4
NEOCODE_API_KEY=sk-xxx
NEOCODE_BASE_URL=https://custom.api.com/v1
NEOCODE_THEME=dark

# Provider-specific keys
DEEPSEEK_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
GOOGLE_API_KEY=xxx
MI_API_KEY=xxx
QWEN_API_KEY=xxx
GLM_API_KEY=xxx
MOONSHOT_API_KEY=xxx
```

### Config File (`~/.neocode/config.toml`)

```toml
default_provider = "deepseek"
default_model = "deepseek-v4"
temperature = 0.2
max_tokens = 4096
theme = "dark"
stream = true
thinking_enabled = false
language = "en"
permission_mode = "ask"
effort = "medium"
show_token_counts = true

[sandbox]
write_roots = ["."]

[update]
auto_check = true
channel = "stable"
```

### ccSwitch Compatible

NeoCode reads all standard ccSwitch environment variables. Works seamlessly with ccSwitch for provider switching.

---

## MCP Plugins

Create `.mcp.json` in your project root to auto-discover MCP servers:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"]
    },
    "custom-server": {
      "command": "python",
      "args": ["-m", "my_mcp_server"]
    }
  }
}
```

NeoCode auto-discovers and connects on startup. Tools from MCP servers are prefixed with `mcp__<server>__<tool>`.

Supported transports:
- **stdio** -- local processes (default)
- **HTTP/SSE** -- remote servers via `"url"` field

---

## Architecture

```
aidev/
|-- cmd/neocode/           # CLI entry point
|-- internal/              # Core kernel
|   |-- agent/             # Agent loop + Storm Breaker + sub-agents + planner
|   |-- cli/               # Self-upgrade from GitHub Releases
|   |-- config/            # TOML config + ccSwitch + model registry (30+ models)
|   |-- config/model/      # Model specs, context thresholds, pricing
|   |-- context/           # Context compaction with dynamic thresholds
|   |-- i18n/              # Bilingual (EN/CN) UI string management
|   |-- monitor/           # Real-time token usage tracking and cost estimation
|   |-- permission/        # Permission policy engine (4 modes)
|   |-- provider/          # 3 providers (OpenAI-compat, Anthropic, Gemini) + SSE + retry
|   |-- session/           # SQLite session persistence (WAL mode)
|   |-- skill/             # YAML+Markdown skill system
|   |-- storage/           # SQLite database layer with migrations
|   |-- tool/              # 10 built-in tools + MCP client (stdio/HTTP/SSE)
|   |   |-- builtin/       # Built-in tool implementations with unit tests
|   |   +-- mcp/           # MCP protocol client and tool adapter
|   |-- types/             # Core type definitions
|   +-- util/              # Shared utilities
|-- desktop/               # Wails v2 desktop app (React + TailwindCSS)
|-- npm/                   # npm distribution (6 platform packages + meta)
|-- scripts/               # Shell installer + Homebrew formula
|-- legacy/                # TypeScript v1 code (archived)
+-- docs/                  # Design docs + configuration reference
```

### Key Design Decisions

- **Single binary**: `CGO_ENABLED=0` for zero-dependency distribution
- **Provider abstraction**: Unified `Provider` interface with SSE streaming
- **Parallel dispatch**: Read-only tools run concurrently; write tools run sequentially
- **Storm Breaker**: Detects death-spiral loops by (tool, error) signature
- **Context compaction**: Automatic conversation summarization when approaching context limits
- **Tool result pruning**: Old, re-derivable outputs replaced with placeholders to save tokens

---

## Desktop Apps

```bash
# macOS
open NeoCode.dmg

# Windows
NeoCode-installer.exe

# Linux
tar xzf NeoCode-linux-amd64.tar.gz
./neocode
```

Desktop app is built with Wails v2 (Go backend + React frontend).

---

## Security

- **API key encryption**: Keys stored encrypted at `~/.neocode/keystore.enc` (AES-256-GCM)
- **Explicit storage disclosure**: Storage path shown on first use
- **Path sandboxing**: Write operations restricted to project directory
- **No shell injection**: Shell commands use argument arrays, not string interpolation
- **Permission modes**: Fine-grained control over which tools can execute
- **Plan mode**: Read-only analysis without any file modifications
- **SSRF protection**: Web fetch operations are sandboxed

See [SECURITY.md](./SECURITY.md) for our vulnerability reporting policy.

---

## API Reference

### Provider Interface

```go
type Provider interface {
    Generate(ctx context.Context, messages []Message, tools []ToolDefinition, systemPrompt string) (*GenerateResult, error)
    Stream(ctx context.Context, messages []Message, tools []ToolDefinition, systemPrompt string, onChunk func(StreamChunk)) (*GenerateResult, error)
    Name() string
    DefaultModel() string
}
```

### Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    Schema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
    ReadOnly() bool
}
```

### Built-in Tools

| Tool | Name | Read-Only | Description |
|------|------|:---------:|-------------|
| ReadTool | `readFile` | Yes | Read file contents with optional line range |
| WriteTool | `writeFile` | No | Write content to file, creating parent dirs |
| EditTool | `editFile` | No | Find-and-replace edit with uniqueness check |
| GlobTool | `searchFiles` | Yes | Find files matching glob patterns |
| GrepTool | `grep` | Yes | Regex search across file contents |
| BashTool | `runCommand` | No | Execute shell commands with timeout |
| GitStatusTool | `gitStatus` | Yes | Show git working tree status |
| GitDiffTool | `gitDiff` | Yes | Show git diff (staged or unstaged) |
| GitCommitTool | `gitCommit` | No | Create git commit with auto-staging |
| LsTool | `listDir` | Yes | List directory contents recursively |

### Sub-Agents

| Agent | Description |
|-------|-------------|
| `explore` | Read-only codebase exploration |
| `review` | Code review with severity ratings |
| `security` | Security vulnerability audit |
| `research` | Deep codebase investigation |

---

## Contributing

We welcome contributions of all kinds -- bug reports, feature requests, documentation improvements, and code.

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/WuSuBuDuoMing/aidev.git
cd aidev

# Build
make build

# Run tests
make test

# Lint
make lint

# Run locally
make run

# Cross-compile for all platforms
make cross
```

---

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for the full history of changes.

---

## License

NeoCode is released under the [MIT License](./LICENSE).

Copyright (c) 2024-2026 WuSuBuDuoMing

---

<div align="center">

**Built with care, for developers who live in the terminal.**

[Report Bug](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=bug_report.md) | [Request Feature](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md) | [Discussions](https://github.com/WuSuBuDuoMing/aidev/discussions)

</div>
