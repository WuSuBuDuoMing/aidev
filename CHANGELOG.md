# Changelog

All notable changes to NeoCode will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [2.9.0] - 2026-06-23

### Added

- Windows standalone `.exe` binary now prominently featured in README installation guides for both English and Chinese
- Windows `.exe` installation via `winget install WuSuBuDuoMing.NeoCode` (recommended) and one-click download from GitHub Releases
- Windows pre-built packages table entry for `neocode_windows_amd64.exe` direct download
- Homebrew formula URLs corrected to match actual GitHub release asset naming convention (`neocode_*` underscore format)
- Homebrew formula version synced to 2.9.0 (previously lagging at 2.0.0)
- Desktop app version reporting synced to 2.9.0 (previously lagging at 2.0.0)
- Self-upgrade command (`neocode upgrade`) now detects and installs Windows `.exe` binary correctly on Windows platforms

### Changed

- Version bumped to 2.9.0 across all version constants: `cmd/neocode/main.go`, `internal/tool/mcp/client.go`, `npm/build.mjs`, `desktop/app.go`
- README.md enhanced with dedicated Windows `.exe` one-click install section, highlighted as the recommended method for Windows users
- README.zh-CN.md enhanced with matching Chinese Windows installation guide
- SECURITY.md supported versions table updated to reflect 2.9.x as latest

### Fixed

- Desktop app `GetStatus()` now reports correct version (was hardcoded to "2.0.0")
- Homebrew formula download URLs now use correct repo path (`WuSuBuDuoMing/aidev`) and underscore naming (`neocode_darwin_arm64`) matching GoReleaser output

## [2.8.0] - 2026-06-23

### Added

- Enhanced Windows cross-compilation pipeline with explicit `.exe` extension handling in CI build matrix
- GoReleaser Windows archive format confirmed as `.zip` for single-file extraction
- npm platform package for Windows ARM64 (`@neocode/cli-win32-arm64`) validated
- Added `make cross` verification that Windows builds produce correct `.exe` suffix in `dist/`

### Changed

- CI workflow build matrix explicitly tests Windows amd64 and arm64 binary compilation with `CGO_ENABLED=0`
- Release workflow smoke test now verifies Windows binary metadata (version string, provider list)
- npm build script validated for Windows `.exe` extension in cross-compilation output

## [2.7.0] - 2026-06-23

### Added

- Code quality improvements: all existing unit tests pass (12 packages, 80+ tests)
- `go vet` static analysis clean -- zero warnings across entire codebase
- GoReleaser configuration verified to produce Windows `.zip` archive alongside Unix `.tar.gz`
- Cross-compilation targets validated: `darwin/{amd64,arm64}`, `linux/{amd64,arm64}`, `windows/{amd64,arm64}`

### Changed

- Version bumped to 2.7.0 as stepping stone toward 2.9.0
- Build artifact naming standardized: `neocode_<os>_<arch>` with appropriate extensions (`.exe` for Windows, no extension for Unix)
- Documentation updated to reflect Windows `.exe` as first-class distribution format

## [2.6.0] - 2026-06-22

### Added

- Comprehensive unit tests for `internal/storage` package: Open, Close, migrate, Exec, Query, QueryRow, foreign key cascade, double-close safety (9 tests)
- Comprehensive unit tests for `internal/tool/builtin` package covering all 10 built-in tools: ReadTool, WriteTool, EditTool, GlobTool, GrepTool, BashTool, GitStatusTool, GitDiffTool, GitCommitTool, LsTool (35 tests)
- Unit tests for `internal/tool/mcp` package: ParseMCPServerName, MCPTool.ToToolDefinition, MCPToolAdapter (Name, Description, ReadOnly, Schema), Manager lifecycle (10 tests)
- JSDoc documentation for npm/build.mjs cross-compilation script with @fileoverview, @module, @param, @type, and @envvar annotations
- MCP Plugins section enhanced with detailed `.mcp.json` configuration examples showing filesystem and custom server setups
- README architecture diagram updated with context and monitor packages
- API Reference section expanded with built-in tools table and sub-agents table

### Changed

- Version bumped to 2.6.0 across cmd/neocode/main.go, npm/build.mjs, and MCP client handshake
- README.md and README.zh-CN.md updated with v2.6.0 references and improved documentation

### Fixed

- MCP client initialization now reports correct version 2.6.0 instead of stale 2.3.0

## [2.5.0] - 2026-06-21

### Added

- Context compaction package (`internal/context`) with dynamic threshold calculation based on model context window size
- Monitor package (`internal/monitor`) for real-time token usage tracking and cost estimation per model
- I18n package (`internal/i18n`) for bilingual (English/Chinese) UI string management
- Provider retry logic with exponential backoff, jitter, and rate-limit detection
- SSE streaming support across all three provider implementations (OpenAI-compat, Anthropic, Gemini)
- Tool result pruning for old, re-derivable outputs to save tokens in long conversations

### Changed

- Model registry expanded to 30+ models across 14 providers
- Provider abstraction unified under single Provider interface with consistent SSE streaming
- Configuration system upgraded with environment variable override support (NEOCODE_* prefix)
- Permission engine redesigned with 4 modes and per-tool wildcard rules

## [2.4.0] - 2026-06-20

### Added

- Sub-agent system with 4 built-in agents: explore, review, security, research
- Plan mode for read-only codebase analysis before making changes
- Skill system with YAML frontmatter + markdown body format
- Project-level skill override of global skills
- Storm Breaker death-spiral loop detection by (tool, error) signature
- Parallel dispatch for read-only tool batches
- Self-upgrade command with SHA256 verification and platform-specific binary detection
- 12 slash commands (/help, /status, /providers, /model, /provider, /effort, /mode, /history, /resume, /skills, /new, /upgrade)

### Changed

- Agent loop refactored for cleaner separation of concerns
- Tool registry extended with GetReadOnlyTools() and AllReadOnly() helpers
- Session persistence improved with WAL mode and foreign key cascades

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
