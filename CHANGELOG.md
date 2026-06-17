# Changelog

All notable changes to NeoCode will be documented in this file.

## [2.0.0] - 2026-06-16

### Added

- Complete rewrite from TypeScript to Go
- Single static binary (CGO_ENABLED=0), zero dependencies
- 33 model registry (Claude Fable 5, GPT-5.5, DeepSeek V4, Gemini 2.5/3.5, MiMo v2.5, GLM-5.1, Kimi K2.7)
- 10 built-in tools (bash, read, write, edit, glob, grep, git status/diff/commit, ls)
- MCP client (stdio/HTTP/SSE transports, .mcp.json auto-discovery)
- SQLite session persistence (WAL mode)
- Skills system (YAML frontmatter + markdown body)
- 4 permission modes (ask, auto, plan, edit)
- 5 effort levels (low, medium, high, xhigh, ultracode)
- Storm Breaker (death-spiral loop detection by error signature)
- Parallel read-only tool dispatch
- Context compaction (1M-aware dynamic thresholds)
- Tool result pruning (re-derivable outputs replaced with placeholders)
- Token estimation with real-usage calibration
- Real-time usage/balance display
- ccSwitch environment variable compatibility (62 provider presets)
- Self-upgrade from GitHub Releases
- 12 slash commands (/help, /status, /providers, /model, /provider, /effort, /mode, /history, /resume, /skills, /new, /upgrade)
- Wails v2 desktop app (React + TailwindCSS frontend)
- Cross-platform distribution (6 CLI targets + 3 desktop platforms)
- npm package (@neocode/cli with platform-specific binaries)
- Shell installer (curl -fsSL https://get.neocode.dev | bash)
- Homebrew formula
- GitHub Actions CI/CD (lint, test, build matrix, release)
- Bilingual documentation (English + Chinese)
- API key encrypted local storage (AES-256-GCM)
- Path sandboxing for write operations
- SSRF protection for web fetch

### Security

- API keys never stored in config files, encrypted in keystore.enc
- Write operations sandboxed to project directory
- Shell commands use argument arrays (no injection)
- Explicit storage path disclosure on first use
