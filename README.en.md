# NeoCode

**AI coding agent -- optimized for Chinese and international models.**

A Go-based native AI coding agent with zero-dependency single-binary distribution. Deep optimization for DeepSeek, MiMo, Qwen, GLM, Kimi, and full support for Claude, GPT, Gemini.

English | [Chinese](README.zh-CN.md)

## Quick Start

```bash
# npm (recommended)
npm i -g @neocode/cli

# Homebrew (macOS / Linux)
brew install neocode/neocode/neocode

# Shell installer
curl -fsSL https://get.neocode.dev | bash

# Go install from source
go install github.com/WuSuBuDuoMing/aidev/cmd/neocode@latest

# Configure
export DEEPSEEK_API_KEY=sk-xxx
# or
export ANTHROPIC_API_KEY=sk-ant-xxx

# Run
neocode
```

## Features

- **62 built-in provider presets** -- DeepSeek, Claude, GPT, Gemini, MiMo, Qwen, GLM, Kimi + 20+ relay services
- **ccSwitch compatible** -- reads all standard ccSwitch environment variables
- **1M context support** -- Gemini 2.5 Pro/Flash with dynamic compaction
- **Tool calling** -- 10 built-in tools: file read/write/edit, shell commands, git, search, web fetch
- **MCP plugin protocol** -- stdio/HTTP/SSE transports with auto-discovery
- **4 permission modes** -- Ask, Auto, Plan, Edit
- **5 effort levels** -- Low, Medium, High, XHigh, Ultracode + Workflows
- **Real-time usage** -- token counts, cost estimation, balance queries
- **Sub-agent system** -- built-in explore, review, security, research agents
- **Session persistence** -- SQLite-backed conversations with full-text search
- **Skills system** -- YAML+Markdown skill templates with project-level override
- **Auto-update** -- `neocode upgrade` one-click upgrade with SHA256 verification
- **Desktop apps** -- Windows (.exe), macOS (.dmg), Linux (.deb) via Wails v2

## Supported Models

| Provider | Models | Context | Pricing (in/out per 1M) |
|----------|--------|---------|-------------------------|
| DeepSeek | v4, v4-pro, v4-flash | 128K | $0.10/$0.40 ~ $0.55/$2.19 |
| Claude | Fable 5, Opus 4.8, Sonnet 4.6, Haiku 4.5 | 200K | $0.80/$4 ~ $15/$75 |
| OpenAI | GPT-5.5, GPT-5.4-mini, GPT-4o, o3, o3-mini | 128K-200K | $1/$4 ~ $15/$60 |
| Gemini | 2.5 Pro/Flash, 3.5 Flash | **1M** | $0.075/$0.30 ~ $1.25/$10 |
| MiMo | v2.5, v2.5-pro | 128K | Free |
| GLM | GLM-5.1 | 128K | -- |
| Kimi | K2.7-code | 128K | -- |
| StepFun | step-3.5-flash-2603 | 128K | -- |
| MiniMax | minimax-m2.7 | 128K | -- |
| Baidu | qianfan-code-latest | 128K | -- |
| Volcengine | ark-code-latest | 128K | -- |
| BaiLing | ling-2.5-1t | 128K | -- |
| Ollama | llama3.1, qwen2.5, deepseek-v3 (local) | 128K | Free |

## CLI Commands

```bash
neocode                         # Interactive chat
neocode "question"              # Single question mode
neocode status                  # Show configuration
neocode providers               # List providers
neocode version                 # Show version
neocode upgrade                 # Self-update
```

### Slash Commands

```
/help    /status    /providers    /model <id>    /provider <id>
/plan    /agents    /history      /resume        /skills
/effort  /mode      /new          /upgrade       /clear    /exit
```

## Configuration

Config file at `~/.neocode/config.toml`, 4-layer merge:

```
Env vars > Project (.neocode/config.toml) > User (~/.neocode/config.toml) > Defaults
```

### Environment Variables

```bash
NEOCODE_PROVIDER=deepseek    # Highest priority
NEOCODE_MODEL=deepseek-v4
NEOCODE_API_KEY=sk-xxx
DEEPSEEK_API_KEY=sk-xxx      # Provider-specific
ANTHROPIC_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
GOOGLE_API_KEY=xxx
```

### ccSwitch Compatible

NeoCode reads all standard ccSwitch environment variables. Works seamlessly with ccSwitch.

## MCP Plugins

Create `.mcp.json` in project root:

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

NeoCode auto-discovers and connects on startup. Tools are prefixed with `mcp__<server>__<tool>`.

## Architecture

```
aidev/
|-- cmd/neocode/           # CLI entry point
|-- internal/              # Core kernel
|   |-- agent/             # Agent loop + Storm Breaker + sub-agents
|   |-- config/            # TOML config + ccSwitch + model registry
|   |-- provider/          # 3 providers (OpenAI-compat, Anthropic, Gemini) + SSE + retry
|   |-- tool/              # 10 built-in tools + MCP client
|   |-- permission/        # Permission policy engine (4 modes)
|   |-- session/           # SQLite session persistence
|   |-- skill/             # Skill system (YAML + Markdown)
|   |-- storage/           # SQLite database layer
|   |-- types/             # Core type definitions
|   +-- cli/               # Self-upgrade
|-- desktop/               # Wails v2 desktop app
|-- npm/                   # npm distribution
|-- scripts/               # Shell installer + Homebrew
+-- docs/                  # Design docs
```

## Security

- API keys encrypted at `~/.neocode/keystore.enc` (AES-256-GCM)
- Explicit storage path disclosure on first use
- Write operations sandboxed to project directory
- Shell commands use argument arrays (no injection)
- Read-only tools parallelized, write operations sequential
- Plan mode for read-only analysis

## Desktop

```bash
# macOS
open NeoCode.dmg

# Windows
NeoCode-installer.exe

# Linux
tar xzf NeoCode-linux-amd64.tar.gz
./neocode
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT -- see [LICENSE](./LICENSE).
