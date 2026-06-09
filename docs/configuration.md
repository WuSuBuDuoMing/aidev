# Configuration Guide

aidev uses a **three-level configuration system** that gives you fine-grained control over every aspect of the tool. Configuration values are resolved with clear precedence rules, allowing you to set sensible defaults at the user level, project-specific overrides at the project level, and sensitive values via environment variables.

---

## Configuration Levels

Configuration is resolved in the following order (highest priority first):

```
1. Environment Variables     (highest priority)
2. Project Config            (.aidev/config.json)
3. User Config               (~/.aidev/config.json)
4. Built-in Defaults         (lowest priority)
```

### 1. User Config

Located at `~/.aidev/config.json`. This is your personal, global configuration. API keys and personal preferences are stored here.

```bash
# View user config location
aidev config path --user

# Edit user config
aidev config set provider claude
aidev config set apiKey "sk-ant-..."
```

### 2. Project Config

Located at `.aidev/config.json` in your project root. This allows per-project overrides. It is safe to commit non-sensitive settings (provider, model, theme) to version control. **Never commit API keys to project config.**

```bash
# Initialize project config
aidev init

# Edit project config
aidev config set model gpt-4o --project
```

### 3. Environment Variables

Environment variables override all config files. They are ideal for CI/CD pipelines and scripts.

```bash
# Linux / macOS
export AIDEV_PROVIDER=claude
export AIDEV_API_KEY="sk-ant-..."
export AIDEV_MODEL="claude-sonnet-4-20250514"

# Windows (PowerShell)
$env:AIDEV_PROVIDER = "claude"
$env:AIDEV_API_KEY = "sk-ant-..."
$env:AIDEV_MODEL = "claude-sonnet-4-20250514"
```

---

## Configuration Reference

### Provider Settings

| Key | Type | Default | Description |
|---|---|---|---|
| `provider` | `string` | `"claude"` | Active AI provider (`claude`, `gpt`, `deepseek`, `gemini`, `ollama`, or custom) |
| `apiKey` | `string` | `""` | API key for the selected provider |
| `model` | `string` | (provider-dependent) | Model identifier for the selected provider |
| `baseUrl` | `string` | `""` | Custom API base URL (for OpenAI-compatible providers) |
| `maxTokens` | `number` | `4096` | Maximum tokens in the AI response |
| `temperature` | `number` | `0.7` | Sampling temperature (0.0 - 2.0) |
| `topP` | `number` | `1.0` | Nucleus sampling parameter |
| `systemPrompt` | `string` | `""` | Custom system prompt prepended to conversations |

### Provider-Specific Environment Variables

| Provider | API Key Variable | Base URL Variable |
|---|---|---|
| Claude | `AIDEV_CLAUDE_API_KEY` or `ANTHROPIC_API_KEY` | `AIDEV_CLAUDE_BASE_URL` |
| GPT | `AIDEV_OPENAI_API_KEY` or `OPENAI_API_KEY` | `AIDEV_OPENAI_BASE_URL` |
| DeepSeek | `AIDEV_DEEPSEEK_API_KEY` | `AIDEV_DEEPSEEK_BASE_URL` |
| Gemini | `AIDEV_GEMINI_API_KEY` or `GOOGLE_API_KEY` | `AIDEV_GEMINI_BASE_URL` |
| Ollama | (none required) | `AIDEV_OLLAMA_BASE_URL` (default: `http://localhost:11434`) |

### Context Settings

| Key | Type | Default | Description |
|---|---|---|---|
| `context.maxFiles` | `number` | `50` | Maximum files to include in project context |
| `context.maxSize` | `number` | `100000` | Maximum total characters of context |
| `context.ignorePatterns` | `string[]` | `["node_modules", "dist", ".git"]` | Glob patterns to exclude from context |
| `context.includePatterns` | `string[]` | `[]` | Glob patterns to explicitly include (overrides ignore) |

### UI Settings

| Key | Type | Default | Description |
|---|---|---|---|
| `theme` | `string` | `"default"` | Output color theme (`default`, `monokai`, `dracula`, `solarized`, `nord`, or custom) |
| `streaming` | `boolean` | `true` | Enable streaming response rendering |
| `syntaxHighlight` | `boolean` | `true` | Enable syntax highlighting for code blocks |
| `markdown` | `boolean` | `true` | Enable markdown rendering in output |
| `spinner` | `boolean` | `true` | Show spinner animation during generation |
| `lineNumbers` | `boolean` | `true` | Show line numbers in code blocks |
| `wordWrap` | `boolean` | `false` | Enable word wrapping in output |

### History & Session Settings

| Key | Type | Default | Description |
|---|---|---|---|
| `history.maxSessions` | `number` | `100` | Maximum number of conversation sessions to retain |
| `history.autoSave` | `boolean` | `true` | Automatically save conversation history |
| `history.contextWindow` | `number` | `200000` | Maximum context window size (in tokens) |
| `session.autoCompact` | `boolean` | `true` | Automatically compact when nearing context limit |
| `session.persistOnExit` | `boolean` | `true` | Save session state when exiting |

### Git Settings

| Key | Type | Default | Description |
|---|---|---|---|
| `git.autoCommit` | `boolean` | `false` | Automatically stage and commit AI-generated changes |
| `git.commitStyle` | `string` | `"conventional"` | Commit message style (`conventional`, `simple`, `custom`) |
| `git.branchPrefix` | `string` | `"aidev/"` | Default prefix for AI-created branches |

### Permission Settings

| Key | Type | Default | Description |
|---|---|---|---|
| `permissions.readFile` | `string` | `"ask"` | File read permission (`allow`, `ask`, `deny`) |
| `permissions.writeFile` | `string` | `"ask"` | File write permission (`allow`, `ask`, `deny`) |
| `permissions.executeCommand` | `string` | `"ask"` | Command execution permission (`allow`, `ask`, `deny`) |
| `permissions.allowedPaths` | `string[]` | `[]` | Paths where operations are always allowed |
| `permissions.deniedPaths` | `string[]` | `[]` | Paths where operations are always denied |

---

## Configuration Commands

```bash
# View all configuration
aidev config list

# Get a specific value
aidev config get provider

# Set a value (user level by default)
aidev config set provider gpt

# Set a value at project level
aidev config set model gpt-4o --project

# Reset a single key to default
aidev config reset provider

# Reset all configuration
aidev config reset --all

# Show config file paths
aidev config path

# Show effective configuration (merged result)
aidev config effective
```

---

## Configuration File Format

Configuration files use standard JSON format:

```json
{
  "provider": "claude",
  "model": "claude-sonnet-4-20250514",
  "maxTokens": 8192,
  "temperature": 0.5,
  "theme": "dracula",
  "streaming": true,
  "context": {
    "maxFiles": 100,
    "ignorePatterns": ["node_modules", "dist", ".git", "*.test.ts"]
  },
  "permissions": {
    "readFile": "allow",
    "writeFile": "ask",
    "executeCommand": "ask"
  },
  "git": {
    "commitStyle": "conventional",
    "branchPrefix": "aidev/"
  }
}
```

---

## Custom Themes

Create a theme file at `~/.aidev/themes/my-theme.json`:

```json
{
  "name": "my-theme",
  "colors": {
    "primary": "#61afef",
    "success": "#98c379",
    "warning": "#e5c07b",
    "error": "#e06c75",
    "muted": "#5c6370",
    "code": {
      "keyword": "#c678dd",
      "string": "#98c379",
      "number": "#d19a66",
      "comment": "#5c6370",
      "function": "#61afef"
    }
  }
}
```

Then activate it:

```bash
aidev config set theme my-theme
```

---

## CI/CD Integration

For CI/CD pipelines, use environment variables to avoid storing secrets in config files:

```yaml
# GitHub Actions example
env:
  AIDEV_PROVIDER: claude
  AIDEV_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  AIDEV_MODEL: claude-sonnet-4-20250514
```

```bash
# Use aidev in scripts
aidev "Review the staged changes" --no-interactive
```

---

## Troubleshooting

### Common Issues

**Config not being applied:**
Run `aidev config effective` to see the merged configuration and identify which level is overriding your setting.

**API key not found:**
Ensure the key is set in user config (`~/.aidev/config.json`) or as an environment variable. Project config does not store API keys for security reasons.

**Permission denied errors:**
Check `aidev config get permissions` to verify your permission settings. Adjust `allowedPaths` and `deniedPaths` as needed.

**Theme not loading:**
Verify the theme file exists at `~/.aidev/themes/<name>.json` and contains valid JSON with the required `name` and `colors` fields.
