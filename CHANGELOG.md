# Changelog

All notable changes to NeoCode will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [2.3.0] - 2026-06-18

### Added

- Comprehensive test suite covering 8 packages: agent, config, model, permission, provider, session, skill, types
- Storm Breaker unit tests (breach threshold, different error signatures)
- Config 4-layer merge tests with environment variable override verification
- Model registry validation tests for all 26 models
- Permission policy tests for all 4 modes (Ask, Auto, Plan, Edit) with wildcard rules
- Provider retry logic tests (retryable errors, rate limiting, delay calculation)
- Session title generation and persistence tests
- Skill discovery, loading, project-level override, and resolution tests
- Tool registry tests (register, get, execute, read-only detection, parallel dispatch check)
- Type system validation tests (providers, roles, effort levels, permission modes)

### Changed

- README.md overhauled with full English documentation: Table of Contents, Features table, CLI reference, Configuration guide, MCP Plugins section, Architecture overview, API Reference (Provider/Tool interfaces), Security section, Contributing link
- README.zh-CN.md enhanced with matching comprehensive Chinese documentation
- README.en.md updated with full feature list, model pricing, and architecture details
- CONTRIBUTING.md rewritten for Go-based workflow: Make targets, Go style guide, testing guidelines, Conventional Commits, branch strategy
- CODE_OF_CONDUCT.md upgraded to full Contributor Covenant v2.0 with enforcement guidelines and severity levels
- SECURITY.md expanded with supported versions table, response time SLAs, detailed security measures (API key encryption, file sandboxing, shell injection prevention, SSRF protection, permission system)
- Version constants updated to 2.3.0 across main.go, npm build.mjs, and MCP client handshake

### Fixed

- CONTRIBUTING.md now correctly references Go tooling (make test, go vet) instead of legacy Node.js commands (npm test, npm run check)
- Bug report template references updated for Go binary (neocode --version) instead of Node.js

## [2.2.0] - 2026-06-17

### Added

- GitHub Actions CI/CD pipeline with lint, test, build matrix (6 platforms), and smoke test stages
- GoReleaser configuration for automated release packaging
- Desktop release workflow for macOS (.dmg), Windows (.exe), Linux builds via Wails v2
- npm publish workflow with platform-specific binary packages (@neocode/cli-darwin-arm64, etc.)
- Homebrew formula auto-update on release
- Issue templates (bug report, feature request) with structured fields
- Pull request template with type-of-change checklist
- FUNDING.yml for GitHub Sponsors
- CODEOWNERS file

### Changed

- Build system standardized on Makefile with cross-compilation targets
- npm build script updated for 6-platform distribution (darwin/linux x windows/arm64+amd64)

## [2.1.0] - 2026-06-16

### Added

- Sub-agent system with 4 built-in agents: explore, review, security, research
- Plan mode for read-only codebase analysis before making changes
- Context compaction with dynamic thresholds based on model context window size
- Tool result pruning for old, re-derivable outputs (readFile, grep, etc.)
- Token estimation with real-usage calibration
- Real-time usage display with cost estimation per model
- ccSwitch environment variable compatibility layer (62 provider presets)
- Skill system with YAML frontmatter + markdown body format
- Project-level skill override of global skills
- Storm Breaker death-spiral loop detection by (tool, error) signature
- Parallel dispatch for read-only tool batches
- Self-upgrade command with SHA256 verification and platform-specific binary detection

### Changed

- Model registry expanded from initial set to 26 models across 13 providers
- Provider abstraction unified under single Provider interface with SSE streaming
- Configuration system upgraded to 4-layer merge (env > project > user > defaults)
- Permission engine redesigned with 4 modes and per-tool wildcard rules

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
