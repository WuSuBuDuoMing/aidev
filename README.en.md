# NeoCode

**AI coding agent — optimized for Chinese and international models.**

A Go-based native AI coding agent with zero-dependency single-binary distribution. Deep optimization for DeepSeek, MiMo, Qwen, GLM, Kimi, and full support for Claude, GPT, Gemini.

English | [中文](README.md)

## Quick Start

```bash
# npm (recommended)
npm i -g @neocode/cli

# Homebrew
brew install neocode/neocode/neocode

# Shell installer
curl -fsSL https://get.neocode.dev | bash

# Configure
export DEEPSEEK_API_KEY=sk-xxx
# or
export ANTHROPIC_API_KEY=sk-ant-xxx

# Run
neocode
```

## Features

- **62 built-in provider presets** — DeepSeek, Claude, GPT, Gemini, MiMo, Qwen, GLM, Kimi + 20+ relay services
- **ccSwitch compatible** — reads all standard ccSwitch environment variables
- **1M context support** — Gemini 2.5 Pro/Flash with dynamic compaction
- **Tool calling** — file read/write/edit, shell commands, git, search, web fetch
- **MCP plugin protocol** — stdio/HTTP/SSE transports
- **4 permission modes** — Ask, Auto, Plan, Edit
- **5 effort levels** — Low → Medium → High → XHigh → Ultracode + Workflows
- **Real-time usage** — token counts, cost estimation, balance queries
- **Auto-update** — `neocode upgrade` one-click upgrade
- **Desktop apps** — Windows (.exe), macOS (.dmg), Linux (.deb)

## Supported Models

| Provider | Models | Context | Pricing (in/out per 1M) |
|----------|--------|---------|-------------------------|
| DeepSeek | v4, v4-pro, v4-flash | 128K | ¥1/¥2 ~ ¥2/¥8 |
| Claude | Fable 5, Opus 4.8, Sonnet 4.6, Haiku 4.5 | 200K | $0.80/$4 ~ $15/$75 |
| OpenAI | GPT-5.5, GPT-4o, o3, o3-mini | 200K | $1/$4 ~ $15/$60 |
| Gemini | 2.5 Pro/Flash, 3.5 Flash | **1M** | $0.075/$0.30 ~ $1.25/$10 |
| MiMo | v2.5, v2.5-pro | 128K | Free |
| GLM | GLM-5.1 | 128K | - |
| Kimi | K2.7-code | 128K | - |

## CLI Commands

```bash
neocode                    # Interactive chat
neocode "question"         # Single question mode
neocode status             # Show configuration
neocode providers          # List providers
neocode upgrade            # Self-update
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

NeoCode auto-discovers and connects on startup.

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

## Security

- API Keys encrypted at `~/.neocode/keystore.enc`
- Explicit storage path disclosure on first use
- Write operations sandboxed to project directory
- Shell commands use argument arrays (no injection)
- Read-only tools parallelized, write operations sequential

## License

MIT
