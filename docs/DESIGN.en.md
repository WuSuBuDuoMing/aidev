# NeoCode Technical Design Document

> Version: 2.0.0 | Date: 2026-06-13 | Status: Design Phase

## 1. Project Overview

NeoCode is a native AI coding agent written in Go, purpose-built for **deep optimization of Chinese domestic models + full support for international models**. It produces zero-dependency single binaries and ships as both CLI and Desktop (Wails) applications.

### Core Positioning

- **Chinese Model Native Optimization**: DeepSeek V4/Pro/Flash, MiMo v2.5/Pro, Qwen, GLM-5.1, Kimi-K2.7, StepFun, MiniMax
- **International Model Full Support**: Claude Fable 5/Opus 4.8/Sonnet 4.6/Haiku 4.5, GPT-5.5/o3, Gemini 2.5 Pro/Flash (1M context)
- **ccSwitch Ecosystem Compatible**: 62 built-in provider presets, full env var and config format compatibility
- **Multi-Platform Distribution**: CLI (6-platform binary + npm + Homebrew) + Desktop (Wails: .exe/.dmg/.tar.gz)

### Project Info

| Field | Value |
|-------|-------|
| Project Name | NeoCode |
| CLI Command | `neocode` |
| Go Module | `github.com/user/neocode` |
| npm Package | `@neocode/cli` |
| Version | v2.0.0 (ground-up rewrite) |
| License | MIT |
| Reference Projects | MiMo-Code, Reasonix, ccSwitch |

---

## 2. Architecture

### Core Principles

1. **Single Kernel, Multiple Frontends**: `internal/` contains all core logic; CLI and Desktop are just I/O surfaces
2. **CGO_ENABLED=0 CLI**: Zero C dependencies, single static binary, cross-compiled to 6 platforms
3. **Prefix Cache Optimization**: Append-only message history with byte-identical prefixes for maximum DeepSeek/MiMo cache hit rates
4. **1M Context Awareness**: Dynamic compaction thresholds based on model's context_window
5. **ccSwitch Compatible**: Read/write standard environment variables, support all 62 ccSwitch provider presets
6. **Real-Time Usage Visibility**: Every API call displays tokens + cost + balance (where supported)
7. **API Key Transparency**: Local encrypted storage, explicit path and security disclosure on first use

### Directory Structure

```text
neocode/
├── cmd/neocode/main.go           # CLI entry point
├── internal/
│   ├── agent/                    # Agent core (loop, compaction, prompts, sessions, subagents, planner)
│   ├── config/                   # Configuration (TOML, effort, model registry, ccSwitch compat)
│   ├── provider/                 # AI providers (OpenAI-compat, Anthropic, Google, registry, retry, stream, cache, billing)
│   ├── tool/                     # Tool system (interface, registry, builtin tools, MCP adapter)
│   │   └── builtin/              # bash, read, write, edit, glob, grep, git, ls, webfetch
│   ├── permission/               # Permission policy (deny > ask > allow) + bash readonly detection
│   ├── session/                  # Session CRUD (SQLite), history, checkpoints
│   ├── storage/                  # SQLite (WAL), API key keystore, migrations
│   ├── context/                  # Project context analyzer, git state
│   ├── monitor/                  # Real-time usage, balance queries, context progress bar
│   ├── cli/                      # Interactive chat, slash commands, upgrade, status
│   ├── skill/                    # Skill discovery + YAML loader
│   ├── i18n/                     # Chinese/English internationalization
│   └── util/                     # Path safety, token estimation, ANSI handling
├── desktop/                      # Wails desktop app (separate go.mod)
│   ├── frontend/                 # React + TailwindCSS
│   └── build/                    # Platform packaging configs
├── npm/                          # npm cross-compilation + launcher
├── .goreleaser.yaml              # GoReleaser config (6 platforms)
├── Makefile                      # Build orchestration
├── neocode.toml                  # Default config template
└── .github/workflows/            # CI, CLI release, desktop release
```

---

## 3. Model Support

### Model Registry (2026-06)

#### Claude (Anthropic)

| Model ID | Context | Max Output | Reasoning | Pricing (in/out per 1M) |
|----------|---------|------------|-----------|-------------------------|
| claude-fable-5 | 200K | 32K | Yes | $15 / $75 |
| claude-opus-4-8 | 200K | 32K | Yes | $15 / $75 |
| claude-sonnet-4-6 | 200K | 64K | Yes | $3 / $15 |
| claude-haiku-4-5 | 200K | 8K | No | $0.80 / $4 |

#### DeepSeek

| Model ID | Context | Max Output | Reasoning | Pricing |
|----------|---------|------------|-----------|---------|
| deepseek-v4 | 128K | 64K | Yes | ¥1 / ¥2 per 1M |
| deepseek-v4-pro | 128K | 64K | Yes | ¥2 / ¥8 per 1M |
| deepseek-v4-flash | 128K | 64K | Yes | ¥0.5 / ¥2 per 1M |

#### OpenAI

| Model ID | Context | Max Output | Reasoning | Pricing |
|----------|---------|------------|-----------|---------|
| gpt-5.5 | 200K | 100K | Yes | $15 / $60 |
| gpt-5.4-mini | 200K | 32K | Yes | $1 / $4 |
| gpt-o3 | 200K | 100K | Yes | $10 / $40 |
| gpt-o3-mini | 200K | 100K | Yes | $1.10 / $4.40 |
| gpt-4o | 128K | 16K | No | $2.50 / $10 |

#### Gemini (Google)

| Model ID | Context | Max Output | Reasoning | Pricing |
|----------|---------|------------|-----------|---------|
| gemini-2.5-pro | **1M** | 64K | Yes | $1.25 / $10 |
| gemini-2.5-flash | **1M** | 64K | Yes | $0.15 / $0.60 |
| gemini-3.5-flash | **1M** | 64K | Yes | $0.075 / $0.30 |

#### Chinese Domestic Models

| Model ID | Provider | Context | Notes |
|----------|----------|---------|-------|
| mimo-v2.5-pro | MiMo (Xiaomi) | 128K | 128K output cap + X-Mimo-Source header |
| glm-5.1 | Zhipu AI | 128K | Anthropic-compatible |
| kimi-k2.7-code | Moonshot | 128K | Anthropic-compatible |
| step-3.5-flash-2603 | StepFun | 128K | Anthropic-compatible |
| minimax-m2.7 | MiniMax | 128K | Anthropic-compatible |
| qianfan-code-latest | Baidu Qianfan | 128K | Anthropic-compatible |
| ark-code-latest | Volcengine | 128K | Anthropic-compatible |
| doubao-seed-2-0-code | Doubao | 128K | OpenAI-compatible |
| ling-2.5-1t | BaiLing | 128K | Anthropic-compatible |

### 1M Context Handling

```go
func CompactThreshold(ctxWindow int) int {
    switch {
    case ctxWindow >= 500_000:  return 400_000  // 1M models
    case ctxWindow >= 100_000:  return 80_000   // 128K/200K models
    default:                    return 40_000   // Small context models
    }
}
```

---

## 4. Core Systems

### Four Permission Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Ask** | Confirmation prompt before every tool call | Default, safest |
| **Auto** | Read-only auto-allowed, writes per config | Daily coding |
| **Plan** | Read-only analysis, no modifications | Complex tasks |
| **Edit** | Auto-edit without confirmation | Trusted scenarios |

### Effort Control

| Level | Claude | DeepSeek | OpenAI | Gemini |
|-------|--------|----------|--------|--------|
| Low | budget: 1K | - | effort: low | budget: 1K |
| Medium | budget: 8K | - | effort: medium | budget: 8K |
| High | budget: 32K | - | effort: high | budget: 32K |
| XHigh | budget: 100K | thinking: true | effort: high+tools | budget: 100K |
| **Ultracode** | budget: max + tools | thinking + 3 retries | o3-pro + tools | budget: max + code exec |

### Real-Time Usage + Balance

```text
╭─ neocode ──────────────────────────────────────────────────────────────────╮
│  Model: deepseek-v4-pro (DeepSeek) | Effort: High | Mode: Auto           │
│  Tokens: 12,340 in / 2,150 out | Cost: $0.0082 | Balance: ¥48.50        │
│  Context: ████████████░░░░░░░░ 62% (79K/128K)                            │
╰───────────────────────────────────────────────────────────────────────────╯
```

### ccSwitch Provider Compatibility (62 Presets)

All ccSwitch ecosystem providers are built-in. View full list via `neocode providers list`.

**Domestic Direct**: DeepSeek, MiMo, GLM, Kimi, Alibaba Bailian, Baidu Qianfan, Volcengine, Doubao, StepFun, MiniMax, BaiLing, ModelScope

**International**: Claude, OpenAI, Gemini, GitHub Copilot, AWS Bedrock, OpenRouter, NVIDIA

**Relay Services**: NewAPI, AiHubMix, SiliconFlow, DMXAPI, PackyCode, CCSub, RunAPI, APIKEY.FUN, and 20+ more

### MCP / Skills / Agents

- **MCP Plugins**: stdio/HTTP/SSE transports, auto-discover `.mcp.json`
- **Skills**: YAML frontmatter + markdown body, project/global discovery
- **Sub-Agents**: explore (read-only search), review (code review), security (security audit), each with independent context budgets

### Auto-Update

- CLI: `neocode upgrade` — GitHub Releases download + SHA256 verification + binary replacement
- Desktop: Minisign signature verification + `latest.json` manifest + platform-specific update strategies

### Chat History

- SQLite persistent storage
- `/history` browse recent sessions
- `/history search <query>` search
- Auto-titled sessions (from first user message)

### API Key Security

```text
~/.neocode/
├── config.toml       # Configuration
├── keystore.enc      # Encrypted API Keys (AES-256-GCM)
├── history.db        # SQLite session database
├── skills/           # Global skills
└── plugins/          # MCP plugins
```

Explicit path and security disclosure on first use.

---

## 5. Implementation Phases

| Phase | Content | Timeline |
|-------|---------|----------|
| Phase 1 | Scaffolding + Config + Providers + SSE streaming + Basic CLI | Week 1-2 |
| Phase 2 | Tool system + Permissions + Agent loop + Context compaction | Week 2-3 |
| Phase 3 | Session management + CLI polish + Skills + i18n + Upgrade | Week 3-4 |
| Phase 4 | MCP plugins + Plan mode + Sub-agents + Memory system | Week 4-5 |
| Phase 5 | Wails desktop + Packaging + Auto-update | Week 5-7 |
| Phase 6 | Distribution + CI/CD + Documentation | Week 7-8 |

---

## 6. Distribution

| Form | Platform | Format |
|------|----------|--------|
| CLI | macOS (Intel+ARM) | tar.gz + Homebrew |
| CLI | Windows (x64+ARM) | zip |
| CLI | Linux (amd64+ARM) | tar.gz + .deb |
| CLI | npm | `@neocode/cli` (6 platform sub-packages) |
| CLI | Shell installer | `curl -fsSL https://get.neocode.dev \| bash` |
| Desktop | macOS | .dmg (Universal Binary) |
| Desktop | Windows | NSIS installer (.exe) |
| Desktop | Linux | .tar.gz + .deb |
