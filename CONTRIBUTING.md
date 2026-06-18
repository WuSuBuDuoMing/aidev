# Contributing to NeoCode (aidev)

Thank you for your interest in contributing to NeoCode! This document provides guidelines and information about contributing to this project.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Style Guide](#style-guide)
- [Testing](#testing)
- [License](#license)

---

## Code of Conduct

We are committed to providing a welcoming and inclusive experience for everyone. Please read our [Code of Conduct](./CODE_OF_CONDUCT.md) before participating.

---

## Getting Started

There are many ways to contribute to NeoCode:

- **Report bugs** -- Use the [bug report template](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=bug_report.md)
- **Suggest features** -- Use the [feature request template](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md)
- **Submit code** -- Fix bugs or implement new features
- **Improve docs** -- Fix typos, add examples, or write new guides
- **Create skills** -- Share useful skill templates with the community
- **Build MCP plugins** -- Extend NeoCode with new tool capabilities

---

## Development Setup

### Prerequisites

- **Go** >= 1.26
- **Git**
- **Node.js** >= 18 (for npm packaging and desktop frontend)

### Steps

```bash
# 1. Fork the repository on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/aidev.git
cd aidev

# 3. Add upstream remote
git remote add upstream https://github.com/WuSuBuDuoMing/aidev.git

# 4. Build
make build

# 5. Run tests
make test

# 6. Lint
make lint

# 7. Run locally
make run
```

### Available Make Targets

```bash
make build      # Build the neocode binary
make run        # Build and run
make test       # Run all tests with race detector
make lint       # Run go vet
make fmt        # Format code with go fmt
make tidy       # Tidy go modules
make cross      # Cross-compile for all 6 platforms
make clean      # Remove build artifacts
make install    # Install to $GOPATH/bin
```

---

## Project Structure

```
aidev/
|-- cmd/neocode/           # CLI entry point (main.go)
|-- internal/              # Core kernel (not importable externally)
|   |-- agent/             # Agent loop, Storm Breaker, sub-agents, planner
|   |-- config/            # TOML config, ccSwitch, model registry
|   |   +-- model/         # Model specs (pricing, context windows)
|   |-- provider/          # AI providers (OpenAI-compat, Anthropic, Gemini)
|   |-- tool/              # Tool interface and registry
|   |   |-- builtin/       # 10 built-in tools
|   |   +-- mcp/           # MCP client (stdio/HTTP/SSE)
|   |-- permission/        # Permission policy engine
|   |-- session/           # SQLite session persistence
|   |-- skill/             # Skill discovery and loading
|   |-- storage/           # SQLite database layer
|   |-- types/             # Core type definitions
|   |-- cli/               # Self-upgrade logic
|   +-- util/              # Shared utilities
|-- desktop/               # Wails v2 desktop app (Go + React)
|-- npm/                   # npm distribution build script
|-- scripts/               # Shell installer + Homebrew formula
|-- legacy/                # TypeScript v1 code (archived)
|-- docs/                  # Design docs + configuration reference
|-- .github/               # CI/CD, issue templates, PR template
|-- .goreleaser.yaml       # GoReleaser configuration
+-- Makefile               # Build automation
```

---

## Making Changes

### Branching Strategy

```bash
# Create a feature branch from main
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-number-description
```

### Branch Naming

| Type | Prefix | Example |
|------|--------|---------|
| New feature | `feature/` | `feature/add-mistral-provider` |
| Bug fix | `fix/` | `fix/streaming-render-crash` |
| Documentation | `docs/` | `docs/update-provider-guide` |
| Refactor | `refactor/` | `refactor/extract-session-manager` |
| Test | `test/` | `test/add-config-unit-tests` |

### Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Formatting, no code change |
| `refactor` | Code refactoring |
| `test` | Adding or updating tests |
| `chore` | Maintenance tasks |
| `perf` | Performance improvement |

**Examples:**

```
feat(provider): add Mistral AI provider support
fix(streaming): handle network timeout during response streaming
docs(readme): add Ollama setup instructions for Windows
test(config): add unit tests for 4-layer config resolution
perf(agent): parallelize read-only tool dispatch
```

---

## Pull Request Process

1. **Ensure quality** -- Run all checks before submitting:
   ```bash
   make lint         # go vet
   make test         # go test -race ./...
   make build        # verify build
   ```

2. **Update documentation** -- If you changed behavior, update the relevant docs.

3. **Write a clear PR description** -- Explain what your changes do and why.

4. **Link related issues** -- Use `Closes #123` or `Fixes #123`.

5. **Wait for review** -- A maintainer will review your PR.

6. **Address feedback** -- Make requested changes promptly.

### PR Checklist

- [ ] Code compiles without errors (`make build`)
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] New tests added for new functionality
- [ ] Documentation updated if behavior changed
- [ ] Commit messages follow Conventional Commits
- [ ] Branch is up to date with main

---

## Reporting Bugs

When reporting a bug, please include:

1. **NeoCode version** -- Run `neocode version`
2. **Go version** -- Run `go version`
3. **Operating system** -- OS and version
4. **AI provider** -- Which provider and model
5. **Steps to reproduce** -- Clear steps to trigger the bug
6. **Expected behavior** -- What you expected to happen
7. **Actual behavior** -- What actually happened
8. **Logs/error output** -- Any error messages

---

## Suggesting Features

Feature suggestions are welcome! Please use the [feature request template](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md) and include:

1. **Problem description** -- What problem does this solve?
2. **Proposed solution** -- How should it work?
3. **Alternatives considered** -- Other approaches you considered
4. **Use case** -- How would you use this in your workflow?

---

## Style Guide

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (run `make fmt`)
- Exported types and functions must have doc comments
- Use table-driven tests when testing multiple cases
- Prefer `context.Context` as the first parameter for functions that may be long-running
- Handle errors explicitly -- do not use `_` for error returns except in tests

### Naming

- **Packages**: lowercase, single-word (e.g., `config`, `provider`, `agent`)
- **Interfaces**: describe behavior (e.g., `Provider`, `Tool`)
- **Structs**: PascalCase (e.g., `OpenAIProvider`, `StormBreaker`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase for exported (e.g., `MaxToolRounds`), camelCase for unexported

### Error Handling

```go
// Good: wrap errors with context
if err != nil {
    return fmt.Errorf("load config: %w", err)
}

// Bad: ignore errors
result, _ := someFunction()
```

---

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/config/...

# Run a specific test
go test -run TestGetSpec ./internal/config/model/

# Run tests with verbose output
go test -v ./...

# Run tests with race detector (default in Makefile)
go test -race ./...
```

### Writing Tests

- Use table-driven tests for multiple input/output cases
- Use `t.TempDir()` for tests that need temporary directories
- Use `t.Helper()` in test utility functions
- Test both success and error paths
- Test edge cases (empty input, nil, boundary values)

---

## License

By contributing to NeoCode, you agree that your contributions will be licensed under the [MIT License](./LICENSE).

---

**Thank you for contributing!**
