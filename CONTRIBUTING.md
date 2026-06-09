# Contributing to aidev / 贡献指南

Thank you for your interest in contributing to aidev! This document provides guidelines and information about contributing to this project.

感谢您对 aidev 项目的关注！本文档提供了贡献指南和相关信息。

---

## Table of Contents / 目录

- [Code of Conduct / 行为准则](#code-of-conduct--行为准则)
- [Getting Started / 开始贡献](#getting-started--开始贡献)
- [Development Setup / 开发环境搭建](#development-setup--开发环境搭建)
- [Project Structure / 项目结构](#project-structure--项目结构)
- [Making Changes / 提交更改](#making-changes--提交更改)
- [Pull Request Process / Pull Request 流程](#pull-request-process--pull-request-流程)
- [Reporting Bugs / 报告 Bug](#reporting-bugs--报告-bug)
- [Suggesting Features / 建议功能](#suggesting-features--建议功能)
- [Style Guide / 代码风格](#style-guide--代码风格)
- [License / 许可证](#license--许可证)

---

## Code of Conduct / 行为准则

We are committed to providing a welcoming and inclusive experience for everyone. Please be respectful and constructive in all interactions.

我们致力于为每个人提供友好和包容的体验。请在所有互动中保持尊重和建设性。

---

## Getting Started / 开始贡献

There are many ways to contribute to aidev:

您可以通过多种方式为 aidev 做出贡献：

- **Report bugs** / 报告 bug -- Use the [bug report template](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=bug_report.md)
- **Suggest features** / 建议功能 -- Use the [feature request template](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md)
- **Submit code** / 提交代码 -- Fix bugs or implement new features
- **Improve docs** / 改进文档 -- Fix typos, add examples, or write new guides
- **Create skills** / 创建技能 -- Share useful skill templates with the community
- **Build plugins** / 构建插件 -- Extend aidev with new capabilities

---

## Development Setup / 开发环境搭建

### Prerequisites / 前置条件

- **Node.js** >= 18.0.0
- **npm** >= 9.0.0
- **Git**

### Steps / 步骤

```bash
# 1. Fork the repository on GitHub
#    在 GitHub 上 Fork 仓库

# 2. Clone your fork
#    克隆你的 fork
git clone https://github.com/YOUR_USERNAME/aidev.git
cd aidev

# 3. Add upstream remote
#    添加上游远程仓库
git remote add upstream https://github.com/WuSuBuDuoMing/aidev.git

# 4. Install dependencies
#    安装依赖
npm install

# 5. Build the project
#    构建项目
npm run build

# 6. Run tests to verify setup
#    运行测试验证环境
npm test

# 7. Start development mode
#    启动开发模式
npm run dev
```

---

## Project Structure / 项目结构

```
aidev/
+-- src/
|   +-- ai/             # AI provider implementations / AI 提供商实现
|   +-- commands/       # CLI commands / CLI 命令
|   +-- config/         # Configuration system / 配置系统
|   +-- context/        # Project context management / 项目上下文管理
|   +-- plugins/        # Plugin system / 插件系统
|   +-- skills/         # Skills system / 技能系统
|   +-- tools/          # Built-in tools (git, file ops) / 内置工具
|   +-- types/          # TypeScript type definitions / 类型定义
|   +-- ui/             # Terminal UI (streaming, themes) / 终端 UI
|   +-- index.ts        # Entry point / 入口文件
+-- docs/               # Documentation / 文档
+-- .aidev/             # Local config & skills / 本地配置和技能
+-- package.json
+-- tsconfig.json
+-- vitest.config.ts
```

---

## Making Changes / 提交更改

### Branching Strategy / 分支策略

```bash
# Create a feature branch from main
# 从 main 创建功能分支
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name

# Or for bug fixes
# 或者用于修复 bug
git checkout -b fix/issue-number-description
```

### Branch Naming / 分支命名规范

| Type / 类型 | Prefix / 前缀 | Example / 示例 |
|---|---|---|
| New feature / 新功能 | `feature/` | `feature/add-mistral-provider` |
| Bug fix / Bug 修复 | `fix/` | `fix/streaming-render-crash` |
| Documentation / 文档 | `docs/` | `docs/update-provider-guide` |
| Refactor / 重构 | `refactor/` | `refactor/extract-session-manager` |
| Test / 测试 | `test/` | `test/add-config-unit-tests` |

### Commit Messages / 提交消息

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

我们遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types / 类型:**

| Type / 类型 | Description / 描述 |
|---|---|
| `feat` | New feature / 新功能 |
| `fix` | Bug fix / Bug 修复 |
| `docs` | Documentation / 文档 |
| `style` | Formatting, no code change / 格式化，无代码变更 |
| `refactor` | Code refactoring / 代码重构 |
| `test` | Adding or updating tests / 添加或更新测试 |
| `chore` | Maintenance tasks / 维护任务 |

**Examples / 示例:**

```
feat(providers): add Mistral AI provider support
fix(streaming): handle network timeout during response streaming
docs(providers): add Ollama setup instructions for Windows
test(config): add unit tests for three-level config resolution
```

---

## Pull Request Process / Pull Request 流程

1. **Ensure quality** -- Run all checks before submitting:
   ```bash
   # 确保质量 -- 提交前运行所有检查
   npm run check    # Lint + type check / 代码检查 + 类型检查
   npm test         # Run tests / 运行测试
   npm run build    # Verify build / 验证构建
   ```

2. **Update documentation** -- If you changed behavior, update the relevant docs.
   **更新文档** -- 如果您更改了行为，请更新相关文档。

3. **Write a clear PR description** -- Explain what your changes do and why.
   **编写清晰的 PR 描述** -- 解释您的更改内容和原因。

4. **Link related issues** -- Use `Closes #123` or `Fixes #123`.
   **关联相关 issue** -- 使用 `Closes #123` 或 `Fixes #123`。

5. **Wait for review** -- A maintainer will review your PR.
   **等待审查** -- 维护者将审查您的 PR。

6. **Address feedback** -- Make requested changes promptly.
   **处理反馈** -- 及时进行要求的更改。

### PR Checklist / PR 检查清单

- [ ] Code compiles without errors / 代码编译无误
- [ ] All tests pass / 所有测试通过
- [ ] Linting passes / 代码检查通过
- [ ] New tests added for new functionality / 为新功能添加了测试
- [ ] Documentation updated / 文档已更新
- [ ] Commit messages follow Conventional Commits / 提交消息遵循规范

---

## Reporting Bugs / 报告 Bug

When reporting a bug, please include:

报告 bug 时，请包含以下信息：

1. **aidev version** -- Run `aidev --version` / 运行 `aidev --version` 获取版本
2. **Node.js version** -- Run `node --version` / 运行 `node --version` 获取版本
3. **Operating system** -- OS and version / 操作系统及版本
4. **AI provider** -- Which provider and model / 使用的提供商和模型
5. **Steps to reproduce** -- Clear steps to trigger the bug / 重现 bug 的清晰步骤
6. **Expected behavior** -- What you expected to happen / 预期行为
7. **Actual behavior** -- What actually happened / 实际行为
8. **Logs** -- Any error messages or logs / 任何错误消息或日志

---

## Suggesting Features / 建议功能

Feature suggestions are welcome! Please use the [feature request template](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md) and include:

欢迎提出功能建议！请使用[功能请求模板](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md)，并包含：

1. **Problem description** -- What problem does this solve?
   **问题描述** -- 这个功能解决什么问题？
2. **Proposed solution** -- How should it work?
   **提议的解决方案** -- 它应该如何工作？
3. **Alternatives considered** -- Other approaches you considered
   **考虑过的替代方案** -- 您考虑过的其他方案
4. **Additional context** -- Screenshots, examples, references
   **附加上下文** -- 截图、示例、参考资料

---

## Style Guide / 代码风格

### TypeScript

- Use **strict TypeScript** with explicit types for public APIs
  使用**严格 TypeScript**，公共 API 需显式类型声明
- Prefer `interface` over `type` for object shapes
  对于对象结构，优先使用 `interface` 而非 `type`
- Use `const` by default; use `let` only when reassignment is needed
  默认使用 `const`，仅在需要重新赋值时使用 `let`
- Avoid `any` -- use `unknown` and narrow with type guards
  避免使用 `any`，使用 `unknown` 配合类型守卫

### Naming / 命名规范

- **Files**: `kebab-case.ts` / 文件名使用 `kebab-case.ts`
- **Classes**: `PascalCase` / 类名使用 `PascalCase`
- **Functions/variables**: `camelCase` / 函数和变量使用 `camelCase`
- **Constants**: `UPPER_SNAKE_CASE` / 常量使用 `UPPER_SNAKE_CASE`
- **Interfaces**: `PascalCase` with descriptive names / 接口使用 `PascalCase`

### Formatting / 格式化

We use **Prettier** for automatic formatting. Run:

我们使用 **Prettier** 进行自动格式化。运行：

```bash
npm run format
```

---

## License / 许可证

By contributing to aidev, you agree that your contributions will be licensed under the [MIT License](./LICENSE).

通过向 aidev 贡献代码，您同意您的贡献将在 [MIT 许可证](./LICENSE) 下获得许可。

---

<div align="center">

**Thank you for contributing! / 感谢您的贡献！**

</div>
