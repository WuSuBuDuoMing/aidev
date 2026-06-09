# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-06-09

### Added

#### Core
- Multi-provider AI support: Claude, GPT, DeepSeek, Gemini, Ollama, and any OpenAI-compatible API
- Streaming response rendering with real-time syntax highlighting
- Three-level configuration system (environment variables > project > user)
- Persistent conversation history with full-text search
- Context-aware project indexing with automatic file relationship detection

#### Commands
- Full Commander.js-powered CLI with `aidev` as the entry point
- One-liner usage: `aidev "your prompt"` for quick queries
- Pipe support: `cat file.ts | aidev "explain this"`
- Interactive REPL mode with tab completion

#### Slash Commands
- `/help` -- Show all available commands and shortcuts
- `/clear` -- Clear the current conversation
- `/compact` -- Compact conversation to save context window
- `/config` -- View or modify configuration
- `/context` -- Manage project context files
- `/cost` -- Show token usage and cost estimates
- `/diff` -- Show a diff of the last AI-generated changes
- `/doctor` -- Run diagnostics to check aidev health
- `/edit` -- Edit a file with AI assistance
- `/explain` -- Explain selected code or a file
- `/export` -- Export conversation history to markdown
- `/fix` -- Fix errors in the current file or selection
- `/git` -- Git operations (commit, branch, status, diff)
- `/init` -- Initialize aidev configuration for the current project
- `/login` -- Authenticate with an AI provider
- `/logout` -- Remove stored credentials
- `/model` -- Switch the active AI model
- `/plugins` -- List, install, or manage plugins
- `/review` -- AI-powered code review of staged changes
- `/search` -- Search across the project with AI assistance
- `/skills` -- List, run, or manage skills
- `/test` -- Generate or run tests for the current file
- `/theme` -- Change the output theme

#### Skills System
- Skill definition using Markdown frontmatter format
- Built-in skills: code-review, refactor, test-gen, doc-gen, explain, optimize
- Custom skill creation with template variables and input parameters
- Skill discovery from project and user directories

#### Plugin Architecture
- Plugin API with hooks for commands, providers, and middleware
- Plugin lifecycle management (install, enable, disable, uninstall)
- Plugin scaffolding with `aidev plugins create`
- Plugin marketplace discovery

#### Git Integration
- `git status`, `git diff`, `git log` within the REPL
- AI-generated conventional commit messages
- Branch creation and management
- Staged change review with `/review`

#### Code Review
- AI-powered code review with severity ratings (critical, warning, suggestion)
- Inline comments with specific line references
- Summary of findings with actionable recommendations

#### UI & Themes
- Streaming markdown rendering in the terminal
- Syntax highlighting for all major languages via cli-highlight
- Multiple built-in themes (default, monokai, dracula, solarized, nord)
- Custom theme support via JSON configuration
- Spinner animations for long-running operations

#### Configuration
- User-level configuration at `~/.aidev/config.json`
- Project-level configuration at `.aidev/config.json`
- Environment variable overrides for CI/CD integration
- `aidev config set/get/list/reset` commands
- API key secure storage

#### Provider Support
- **Claude**: claude-sonnet-4-20250514, claude-3-5-sonnet, claude-3-opus, claude-3-haiku
- **GPT**: gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-3.5-turbo
- **DeepSeek**: deepseek-chat, deepseek-coder
- **Gemini**: gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash
- **Ollama**: Any locally running model (llama3.1, codellama, mistral, etc.)
- **OpenAI-compatible**: Any API that implements the OpenAI chat completions interface

#### Developer Experience
- TypeScript-first codebase with strict type checking
- Vitest test framework with comprehensive test coverage
- ESLint + Prettier for code formatting
- GitHub Actions CI pipeline (lint, build, test) on Node.js 18/20
- Comprehensive documentation: configuration, skills, plugins, providers

### Security
- API keys stored in user config, never in project configuration
- Permission control system for file read/write/execute operations
- No telemetry or data collection

[1.0.0]: https://github.com/WuSuBuDuoMing/aidev/releases/tag/v1.0.0
