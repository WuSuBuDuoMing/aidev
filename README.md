<div align="center">

# aidev

### AI-Powered Terminal Coding Assistant

[![npm version](https://img.shields.io/npm/v/aidev-cli.svg?style=flat-square&color=blue)](https://www.npmjs.com/package/aidev-cli)
[![license](https://img.shields.io/npm/l/aidev-cli.svg?style=flat-square&color=green)](https://github.com/WuSuBuDuoMing/aidev/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/WuSuBuDuoMing/aidev?style=flat-square&color=yellow)](https://github.com/WuSuBuDuoMing/aidev/stargazers)
[![GitHub contributors](https://img.shields.io/github/contributors/WuSuBuDuoMing/aidev?style=flat-square&color=purple)](https://github.com/WuSuBuDuoMing/aidev/graphs/contributors)
[![GitHub issues](https://img.shields.io/github/issues/WuSuBuDuoMing/aidev?style=flat-square&color=red)](https://github.com/WuSuBuDuoMing/aidev/issues)
[![CI](https://img.shields.io/github/actions/workflow/status/WuSuBuDuoMing/aidev/ci.yml?style=flat-square&label=CI)](https://github.com/WuSuBuDuoMing/aidev/actions)

A superior AI terminal coding assistant that brings the power of multiple AI providers directly into your terminal. Supports **Claude**, **GPT**, **DeepSeek**, **Gemini**, **Ollama**, and any **OpenAI-compatible API** -- all from a single, unified CLI.

[English](./README.md) | [简体中文](./README.zh-CN.md)

</div>

---

## Why aidev?

aidev is not just another AI CLI wrapper. It is a fully-featured, extensible coding assistant designed for professional developers who live in the terminal.

### Feature Comparison

| Feature | aidev | deepcode-cli | GitHub Copilot CLI | cursor-cli |
|---|:---:|:---:|:---:|:---:|
| Multi-provider support | **5+** | 1 | 1 | 1 |
| Local model support (Ollama) | **Yes** | No | No | No |
| OpenAI-compatible API | **Yes** | Limited | No | No |
| Streaming with syntax highlighting | **Yes** | Yes | No | Yes |
| Skills system | **Yes** | No | No | No |
| Plugin architecture | **Yes** | No | No | Limited |
| Code review / diff preview | **Yes** | Limited | No | Yes |
| Git integration | **Full** | Basic | Basic | Basic |
| Custom themes | **Yes** | No | No | No |
| Permission control | **Yes** | No | No | No |
| Conversation history | **Yes** | Yes | No | Yes |
| Multi-level configuration | **Yes** | Basic | Basic | Basic |
| Context-aware coding | **Yes** | Yes | Limited | Yes |

---

## Features

- **Multi-Provider Support** -- Seamlessly switch between Claude, GPT, DeepSeek, Gemini, Ollama, and any OpenAI-compatible API. Use the best model for each task.
- **Context-Aware Coding** -- Automatically indexes your project and understands file relationships, imports, and project structure.
- **Streaming with Syntax Highlighting** -- Real-time streaming responses with beautiful syntax highlighting for all major languages.
- **Skills System** -- Create, share, and compose reusable skill templates to automate repetitive coding patterns.
- **Plugin Architecture** -- Extend aidev with custom plugins that add new commands, providers, or integrations.
- **Git Integration** -- Stage, commit, create branches, view diffs, and generate conventional commit messages -- all from within aidev.
- **Code Review** -- AI-powered code review with inline comments and actionable suggestions.
- **Diff Preview** -- Review AI-generated changes with a clear, color-coded diff before applying them.
- **Custom Themes** -- Personalize your terminal experience with built-in and custom color themes.
- **Permission Control** -- Fine-grained control over what aidev can read, write, and execute on your system.
- **Conversation History** -- Persistent conversation sessions with full history, search, and context restoration.
- **Multi-Level Configuration** -- Configure aidev at the environment, project, and user level with clear precedence rules.

---

## Quick Start

### Installation

```bash
# Install globally via npm
npm install -g aidev-cli

# Or use yarn
yarn global add aidev-cli

# Or use pnpm
pnpm add -g aidev-cli
```

### First Run

```bash
# Launch aidev
aidev

# Set up your preferred provider
aidev config set provider claude
aidev config set apiKey "your-api-key"

# Start coding
aidev "Help me refactor this function to use async/await"
```

### One-liner Usage

```bash
# Ask a quick question
aidev "What does this regex match? /^[a-z]+-\d{3}$/"

# Pipe input
cat src/app.ts | aidev "Review this code for potential bugs"

# Generate code
aidev "Create a TypeScript function that parses CSV files with proper error handling"
```

---

## Configuration

aidev uses a **three-level configuration system** with clear precedence:

```
Environment Variables  >  Project Config (.aidev/config.json)  >  User Config (~/.aidev/config.json)
```

### Quick Configuration

```bash
# Set provider
aidev config set provider claude

# Set API key (stored securely in user config)
aidev config set apiKey "sk-..."

# Set model
aidev config set model claude-sonnet-4-20250514

# Set output theme
aidev config set theme monokai

# View all settings
aidev config list

# Reset to defaults
aidev config reset
```

See [docs/configuration.md](./docs/configuration.md) for the full configuration reference.

---

## Supported Providers

| Provider | Default Model | Streaming | Function Calling | Local |
|---|---|:---:|:---:|:---:|
| **Claude** (Anthropic) | `claude-sonnet-4-20250514` | Yes | Yes | No |
| **GPT** (OpenAI) | `gpt-4o` | Yes | Yes | No |
| **DeepSeek** | `deepseek-chat` | Yes | Yes | No |
| **Gemini** (Google) | `gemini-2.5-pro` | Yes | Yes | No |
| **Ollama** | `llama3.1` | Yes | Varies | Yes |
| **OpenAI-compatible** | `default` | Yes | Varies | Varies |

See [docs/providers.md](./docs/providers.md) for detailed setup instructions for each provider.

---

## Slash Commands

aidev supports a rich set of slash commands for quick actions:

| Command | Description |
|---|---|
| `/help` | Show all available commands and shortcuts |
| `/clear` | Clear the current conversation |
| `/compact` | Compact conversation to save context window |
| `/config` | View or modify configuration |
| `/context` | Manage project context files |
| `/cost` | Show token usage and cost estimates for the session |
| `/diff` | Show a diff of the last AI-generated changes |
| `/doctor` | Run diagnostics to check aidev health |
| `/edit` | Edit a file with AI assistance |
| `/explain` | Explain selected code or a file |
| `/export` | Export conversation history to markdown |
| `/fix` | Fix errors in the current file or selection |
| `/git` | Git operations (commit, branch, status, diff) |
| `/init` | Initialize aidev configuration for the current project |
| `/login` | Authenticate with an AI provider |
| `/logout` | Remove stored credentials |
| `/model` | Switch the active AI model |
| `/plugins` | List, install, or manage plugins |
| `/review` | AI-powered code review of staged changes |
| `/search` | Search across the project with AI assistance |
| `/skills` | List, run, or manage skills |
| `/test` | Generate or run tests for the current file |
| `/theme` | Change the output theme |

---

## Keyboard Shortcuts

| Shortcut | Action |
|---|---|
| `Ctrl+C` | Cancel current generation |
| `Ctrl+D` | Exit aidev |
| `Ctrl+L` | Clear screen |
| `Ctrl+K` | Clear current input line |
| `Ctrl+R` | Search conversation history |
| `Up / Down` | Navigate input history |
| `Tab` | Autocomplete commands and file paths |
| `Escape` | Cancel current operation |
| `Ctrl+Enter` | Submit multi-line input |

---

## Skills System

Skills are reusable prompt templates that encode expert knowledge for common tasks.

### Built-in Skills

- **code-review** -- Comprehensive code review with severity ratings
- **refactor** -- Suggest and apply refactoring improvements
- **test-gen** -- Generate unit tests with full coverage analysis
- **doc-gen** -- Generate documentation for functions and modules
- **explain** -- Explain code with varying levels of detail
- **optimize** -- Identify and fix performance bottlenecks

### Using Skills

```bash
# List available skills
/skills list

# Run a skill
/skills run code-review

# Run a skill on specific files
/skills run test-gen src/utils/parser.ts
```

### Creating Custom Skills

Create a `.aidev/skills/my-skill.md` file:

```markdown
---
name: my-skill
description: A custom skill for my workflow
version: 1.0.0
input:
  - name: file
    type: string
    description: Target file path
    required: true
---

You are an expert TypeScript developer.

Analyze the file at {{file}} and:

1. Check for type safety issues
2. Identify potential runtime errors
3. Suggest improvements for readability
```

See [docs/skills.md](./docs/skills.md) for full documentation.

---

## Plugin System

Extend aidev with plugins that add new capabilities:

```bash
# Install a plugin
aidev plugins install aidev-plugin-docker

# List installed plugins
aidev plugins list

# Create your own plugin
aidev plugins create my-plugin
```

See [docs/plugins.md](./docs/plugins.md) for the plugin API and development guide.

---

## Architecture

```
+------------------------------------------------------------------+
|                          aidev CLI                                |
+------------------------------------------------------------------+
|                                                                    |
|  +-------------+  +-------------+  +-------------+  +----------+ |
|  |   Commands   |  |  Slash Cmds |  |   Skills    |  |  Config  | |
|  |  (Commander) |  |  /help etc  |  |  (.md files) |  |  (3-lvl) | |
|  +------+------+  +------+------+  +------+------+  +----+-----+ |
|         |                |                |              |        |
|  +------v----------------v----------------v--------------v-----+  |
|  |                     Core Engine                              |  |
|  |  +------------+  +------------+  +-----------+  +---------+ |  |
|  |  |  Context   |  |  Session   |  |  History   |  |  Git    | |  |
|  |  |  Manager   |  |  Manager   |  |  Manager   |  |  Tools  | |  |
|  |  +------+-----+  +-----+------+  +-----+-----+  +----+----+ |  |
|  +---------|--------------|--------------|--------------|-------+  |
|            |              |              |              |           |
|  +---------v--------------v--------------v--------------v-------+  |
|  |                     AI Abstraction Layer                     |  |
|  |  +--------+  +-----+  +---------+  +-------+  +----------+ |  |
|  |  | Claude |  | GPT |  | DeepSeek |  | Gemini|  | Ollama   | |  |
|  |  +--------+  +-----+  +---------+  +-------+  +----------+ |  |
|  +--------------------------------------------------------------+  |
|                                                                    |
|  +--------------------------------------------------------------+  |
|  |                      Plugin System                            |  |
|  |  +------------+  +-------------+  +------------------------+ |  |
|  |  |  Commands   |  |  Providers  |  |  Hooks & Middleware    | |  |
|  |  +------------+  +-------------+  +------------------------+ |  |
|  +--------------------------------------------------------------+  |
|                                                                    |
|  +--------------------------------------------------------------+  |
|  |                      UI Layer                                |  |
|  |  +------------+  +-------------+  +-------------+            |  |
|  |  |  Streaming  |  |   Syntax    |  |   Themes    |            |  |
|  |  |  Renderer   |  | Highlighter |  |   Engine    |            |  |
|  |  +------------+  +-------------+  +-------------+            |  |
|  +--------------------------------------------------------------+  |
+------------------------------------------------------------------+
```

---

## Contributing

We welcome contributions of all kinds! Whether it's reporting a bug, suggesting a feature, improving documentation, or submitting code -- every contribution matters.

Please read our [Contributing Guide](./CONTRIBUTING.md) before getting started.

```bash
# Clone the repository
git clone https://github.com/WuSuBuDuoMing/aidev.git
cd aidev

# Install dependencies
npm install

# Start development mode
npm run dev

# Run tests
npm test

# Run linting
npm run check
```

---

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for a detailed history of changes.

---

## License

aidev is released under the [MIT License](./LICENSE).

Copyright (c) 2024-2026 WuSuBuDuoMing

---

<div align="center">

**Built with care by developers, for developers.**

[Report a Bug](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=bug_report.md) | [Request a Feature](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md) | [Join the Discussion](https://github.com/WuSuBuDuoMing/aidev/discussions)

</div>
