<div align="center">

# aidev

### AI 驱动的终端编程助手

[![npm version](https://img.shields.io/npm/v/aidev-cli.svg?style=flat-square&color=blue)](https://www.npmjs.com/package/aidev-cli)
[![license](https://img.shields.io/npm/l/aidev-cli.svg?style=flat-square&color=green)](https://github.com/WuSuBuDuoMing/aidev/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/WuSuBuDuoMing/aidev?style=flat-square&color=yellow)](https://github.com/WuSuBuDuoMing/aidev/stargazers)
[![GitHub contributors](https://img.shields.io/github/contributors/WuSuBuDuoMing/aidev?style=flat-square&color=purple)](https://github.com/WuSuBuDuoMing/aidev/graphs/contributors)
[![GitHub issues](https://img.shields.io/github/issues/WuSuBuDuoMing/aidev?style=flat-square&color=red)](https://github.com/WuSuBuDuoMing/aidev/issues)
[![CI](https://img.shields.io/github/actions/workflow/status/WuSuBuDuoMing/aidev/ci.yml?style=flat-square&label=CI)](https://github.com/WuSuBuDuoMing/aidev/actions)

一款卓越的 AI 终端编程助手，将多种 AI 模型的能力直接带入您的终端。支持 **Claude**、**GPT**、**DeepSeek**、**Gemini**、**Ollama** 以及任何 **OpenAI 兼容 API** -- 一切尽在单一命令行工具中。

[English](./README.md) | [简体中文](./README.zh-CN.md)

</div>

---

## 为什么选择 aidev？

aidev 不仅仅是又一个 AI 命令行封装工具。它是一个功能完善、可扩展的编程助手，专为生活在终端中的专业开发者设计。

### 功能对比

| 功能特性 | aidev | deepcode-cli | GitHub Copilot CLI | cursor-cli |
|---|:---:|:---:|:---:|:---:|
| 多模型提供商支持 | **5+** | 1 | 1 | 1 |
| 本地模型支持 (Ollama) | **是** | 否 | 否 | 否 |
| OpenAI 兼容 API | **是** | 有限 | 否 | 否 |
| 流式输出 + 语法高亮 | **是** | 是 | 否 | 是 |
| 技能系统 | **是** | 否 | 否 | 否 |
| 插件架构 | **是** | 否 | 否 | 有限 |
| 代码审查 / 差异预览 | **是** | 有限 | 否 | 是 |
| Git 集成 | **完整** | 基础 | 基础 | 基础 |
| 自定义主题 | **是** | 否 | 否 | 否 |
| 权限控制 | **是** | 否 | 否 | 否 |
| 对话历史 | **是** | 是 | 否 | 是 |
| 多级配置 | **是** | 基础 | 基础 | 基础 |
| 上下文感知编程 | **是** | 是 | 有限 | 是 |

---

## 功能特性

- **多提供商支持** -- 在 Claude、GPT、DeepSeek、Gemini、Ollama 及任何 OpenAI 兼容 API 之间无缝切换，为每个任务选择最优模型。
- **上下文感知编程** -- 自动索引项目结构，理解文件关系、导入路径和项目组织方式。
- **流式输出 + 语法高亮** -- 实时流式响应，支持所有主流语言的精美语法高亮。
- **技能系统** -- 创建、分享和组合可复用的技能模板，自动化重复性编码模式。
- **插件架构** -- 通过自定义插件扩展 aidev，添加新命令、新提供商或新集成。
- **Git 集成** -- 在 aidev 中直接暂存、提交、创建分支、查看差异，以及生成规范提交消息。
- **代码审查** -- AI 驱动的代码审查，提供内联注释和可操作的改进建议。
- **差异预览** -- 在应用 AI 生成的更改之前，以清晰的彩色差异视图进行审阅。
- **自定义主题** -- 使用内置和自定义配色方案个性化终端体验。
- **权限控制** -- 精细控制 aidev 对系统文件的读取、写入和执行权限。
- **对话历史** -- 持久化对话会话，支持完整历史记录、搜索和上下文恢复。
- **多级配置** -- 在环境变量、项目和用户三个层级配置 aidev，具有清晰的优先级规则。

---

## 快速开始

### 安装

```bash
# 通过 npm 全局安装
npm install -g aidev-cli

# 或使用 yarn
yarn global add aidev-cli

# 或使用 pnpm
pnpm add -g aidev-cli
```

### 首次运行

```bash
# 启动 aidev
aidev

# 设置首选提供商
aidev config set provider claude
aidev config set apiKey "your-api-key"

# 开始编程
aidev "帮我把这个函数重构为 async/await 风格"
```

### 单行用法

```bash
# 快速提问
aidev "这个正则表达式匹配什么？ /^[a-z]+-\d{3}$/"

# 管道输入
cat src/app.ts | aidev "审查这段代码中的潜在 bug"

# 生成代码
aidev "创建一个 TypeScript 函数，用于解析 CSV 文件并包含完善的错误处理"
```

---

## 配置

aidev 采用**三级配置系统**，具有清晰的优先级规则：

```
环境变量  >  项目配置 (.aidev/config.json)  >  用户配置 (~/.aidev/config.json)
```

### 快速配置

```bash
# 设置提供商
aidev config set provider claude

# 设置 API 密钥（安全存储在用户配置中）
aidev config set apiKey "sk-..."

# 设置模型
aidev config set model claude-sonnet-4-20250514

# 设置输出主题
aidev config set theme monokai

# 查看所有设置
aidev config list

# 重置为默认值
aidev config reset
```

完整的配置参考请参阅 [docs/configuration.md](./docs/configuration.md)。

---

## 支持的提供商

| 提供商 | 默认模型 | 流式输出 | 函数调用 | 本地部署 |
|---|---|:---:|:---:|:---:|
| **Claude** (Anthropic) | `claude-sonnet-4-20250514` | 是 | 是 | 否 |
| **GPT** (OpenAI) | `gpt-4o` | 是 | 是 | 否 |
| **DeepSeek** | `deepseek-chat` | 是 | 是 | 否 |
| **Gemini** (Google) | `gemini-2.5-pro` | 是 | 是 | 否 |
| **Ollama** | `llama3.1` | 是 | 视模型而定 | 是 |
| **OpenAI 兼容** | `default` | 是 | 视模型而定 | 视情况而定 |

每个提供商的详细设置说明请参阅 [docs/providers.md](./docs/providers.md)。

---

## 斜杠命令

aidev 支持丰富的斜杠命令用于快速操作：

| 命令 | 描述 |
|---|---|
| `/help` | 显示所有可用命令和快捷键 |
| `/clear` | 清除当前对话 |
| `/compact` | 压缩对话以节省上下文窗口 |
| `/config` | 查看或修改配置 |
| `/context` | 管理项目上下文文件 |
| `/cost` | 显示当前会话的 token 用量和费用估算 |
| `/diff` | 显示 AI 最近生成的更改差异 |
| `/doctor` | 运行诊断检查 aidev 健康状态 |
| `/edit` | 使用 AI 辅助编辑文件 |
| `/explain` | 解释选中的代码或文件 |
| `/export` | 将对话历史导出为 Markdown |
| `/fix` | 修复当前文件或选区中的错误 |
| `/git` | Git 操作（提交、分支、状态、差异） |
| `/init` | 为当前项目初始化 aidev 配置 |
| `/login` | 验证 AI 提供商身份 |
| `/logout` | 删除已存储的凭证 |
| `/model` | 切换当前 AI 模型 |
| `/plugins` | 列出、安装或管理插件 |
| `/review` | 对暂存更改进行 AI 代码审查 |
| `/search` | 使用 AI 辅助搜索项目内容 |
| `/skills` | 列出、运行或管理技能 |
| `/test` | 为当前文件生成或运行测试 |
| `/theme` | 更改输出主题 |

---

## 键盘快捷键

| 快捷键 | 操作 |
|---|---|
| `Ctrl+C` | 取消当前生成 |
| `Ctrl+D` | 退出 aidev |
| `Ctrl+L` | 清屏 |
| `Ctrl+K` | 清除当前输入行 |
| `Ctrl+R` | 搜索对话历史 |
| `上/下` | 浏览输入历史 |
| `Tab` | 命令和文件路径自动补全 |
| `Escape` | 取消当前操作 |
| `Ctrl+Enter` | 提交多行输入 |

---

## 技能系统

技能是可复用的提示模板，封装了常见任务的专家知识。

### 内置技能

- **code-review** -- 带有严重等级评定的全面代码审查
- **refactor** -- 提出并应用重构改进建议
- **test-gen** -- 生成带有完整覆盖率分析的单元测试
- **doc-gen** -- 为函数和模块生成文档
- **explain** -- 以不同详细程度解释代码
- **optimize** -- 识别并修复性能瓶颈

### 使用技能

```bash
# 列出可用技能
/skills list

# 运行技能
/skills run code-review

# 对特定文件运行技能
/skills run test-gen src/utils/parser.ts
```

### 创建自定义技能

在 `.aidev/skills/my-skill.md` 文件中创建：

```markdown
---
name: my-skill
description: 针对我工作流的自定义技能
version: 1.0.0
input:
  - name: file
    type: string
    description: 目标文件路径
    required: true
---

你是一位资深的 TypeScript 开发者。

分析 {{file}} 处的文件并：

1. 检查类型安全问题
2. 识别潜在的运行时错误
3. 提出可读性改进建议
```

完整文档请参阅 [docs/skills.md](./docs/skills.md)。

---

## 插件系统

通过插件扩展 aidev 的功能：

```bash
# 安装插件
aidev plugins install aidev-plugin-docker

# 列出已安装插件
aidev plugins list

# 创建你自己的插件
aidev plugins create my-plugin
```

插件 API 和开发指南请参阅 [docs/plugins.md](./docs/plugins.md)。

---

## 架构

```
+------------------------------------------------------------------+
|                          aidev CLI                                |
+------------------------------------------------------------------+
|                                                                    |
|  +-------------+  +-------------+  +-------------+  +----------+ |
|  |    命令      |  |  斜杠命令   |  |    技能      |  |   配置   | |
|  |  (Commander) |  |  /help 等   |  |  (.md 文件)  |  |  (三级)  | |
|  +------+------+  +------+------+  +------+------+  +----+-----+ |
|         |                |                |              |        |
|  +------v----------------v----------------v--------------v-----+  |
|  |                     核心引擎                                 |  |
|  |  +------------+  +------------+  +-----------+  +---------+ |  |
|  |  |  上下文     |  |  会话      |  |  历史     |  |  Git    | |  |
|  |  |  管理器     |  |  管理器    |  |  管理器   |  |  工具   | |  |
|  |  +------+-----+  +-----+------+  +-----+-----+  +----+----+ |  |
|  +---------|--------------|--------------|--------------|-------+  |
|            |              |              |              |           |
|  +---------v--------------v--------------v--------------v-------+  |
|  |                     AI 抽象层                                |  |
|  |  +--------+  +-----+  +---------+  +-------+  +----------+ |  |
|  |  | Claude |  | GPT |  | DeepSeek|  | Gemini|  | Ollama   | |  |
|  |  +--------+  +-----+  +---------+  +-------+  +----------+ |  |
|  +--------------------------------------------------------------+  |
|                                                                    |
|  +--------------------------------------------------------------+  |
|  |                      插件系统                                 |  |
|  |  +------------+  +-------------+  +------------------------+ |  |
|  |  |   命令      |  |   提供商    |  |   钩子与中间件          | |  |
|  |  +------------+  +-------------+  +------------------------+ |  |
|  +--------------------------------------------------------------+  |
|                                                                    |
|  +--------------------------------------------------------------+  |
|  |                      UI 层                                    |  |
|  |  +------------+  +-------------+  +-------------+            |  |
|  |  |  流式       |  |   语法      |  |   主题      |            |  |
|  |  |  渲染器     |  |   高亮器    |  |   引擎      |            |  |
|  |  +------------+  +-------------+  +-------------+            |  |
|  +--------------------------------------------------------------+  |
+------------------------------------------------------------------+
```

---

## 贡献

我们欢迎各种形式的贡献！无论是报告 bug、建议功能、改进文档还是提交代码 -- 每一份贡献都很重要。

开始之前请阅读我们的[贡献指南](./CONTRIBUTING.md)。

```bash
# 克隆仓库
git clone https://github.com/WuSuBuDuoMing/aidev.git
cd aidev

# 安装依赖
npm install

# 启动开发模式
npm run dev

# 运行测试
npm test

# 运行代码检查
npm run check
```

---

## 更新日志

详细的变更历史请参阅 [CHANGELOG.md](./CHANGELOG.md)。

---

## 许可证

aidev 基于 [MIT 许可证](./LICENSE) 发布。

Copyright (c) 2024-2026 WuSuBuDuoMing

---

<div align="center">

**由开发者用心打造，为开发者服务。**

[报告 Bug](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=bug_report.md) | [功能建议](https://github.com/WuSuBuDuoMing/aidev/issues/new?template=feature_request.md) | [参与讨论](https://github.com/WuSuBuDuoMing/aidev/discussions)

</div>
