# Configuration Guide

aidev uses a **four-level configuration system** that gives you fine-grained control. Configuration values are resolved with clear precedence rules:

```
1. Environment Variables     (highest priority)
2. Project Config            (.aidev/config.json)
3. User Config               (~/.aidev/config.json)
4. Built-in Defaults         (lowest priority)
```

---

## Configuration Levels

### 1. User Config

Located at `~/.aidev/config.json`. This is your personal, global configuration.

```json
{
  "provider": "claude",
  "apiKey": "sk-ant-...",
  "theme": "dracula"
}
```

### 2. Project Config

Located at `.aidev/config.json` in your project root. Project-level overrides. **Never commit API keys to project config.**

```json
{
  "model": "gpt-4o",
  "temperature": 0.5
}
```

### 3. Environment Variables

Environment variables override all config files. They are ideal for CI/CD pipelines.

| Variable | Description |
|---|---|
| `AIDEV_PROVIDER` | Provider name (`claude`, `openai`, `deepseek`, `gemini`, `ollama`, `custom`) |
| `AIDEV_API_KEY` | API key (highest priority) |
| `AIDEV_MODEL` | Model identifier |
| `AIDEV_BASE_URL` | Custom API base URL |
| `AIDEV_TEMPERATURE` | Sampling temperature (0-2) |
| `AIDEV_MAX_TOKENS` | Maximum output tokens |
| `AIDEV_THEME` | UI theme name |

#### Provider-Specific Fallbacks

If `AIDEV_API_KEY` is not set, aidev falls back to these provider-specific variables:

| Provider | Fallback Variable |
|---|---|
| Claude | `ANTHROPIC_API_KEY` |
| OpenAI | `OPENAI_API_KEY` |
| DeepSeek | `DEEPSEEK_API_KEY` |
| Gemini | `GOOGLE_API_KEY` |

---

## Configuration Reference

### All Configuration Keys

| Key | Type | Default | Description |
|---|---|---|---|
| `provider` | `string` | `"claude"` | AI provider: `claude`, `openai`, `deepseek`, `gemini`, `ollama`, `custom` |
| `model` | `string` | `"claude-sonnet-4-20250514"` | Model identifier |
| `apiKey` | `string` | `""` | API key for the selected provider |
| `baseUrl` | `string` | `""` | Custom API base URL (for proxies or self-hosted) |
| `temperature` | `number` | `0.2` | Sampling temperature (0.0 - 2.0) |
| `maxTokens` | `number` | `4096` | Maximum tokens in the AI response |
| `thinkingEnabled` | `boolean` | `false` | Enable thinking/reasoning mode |
| `defaultPermission` | `string` | `"ask"` | Default tool permission: `allow`, `deny`, `ask` |
| `permissions` | `array` | `[]` | Per-tool permission overrides: `[{ "toolName": "runCommand", "mode": "ask" }]` |
| `theme` | `string` | `"dark"` | UI theme: `dark`, `light`, `monokai`, `dracula` |
| `stream` | `boolean` | `true` | Enable streaming response rendering |
| `showTokenCounts` | `boolean` | `true` | Show token usage after each response |
| `userName` | `string` | `"Developer"` | User display name |

---

## Configuration File Format

```json
{
  "provider": "claude",
  "model": "claude-sonnet-4-20250514",
  "apiKey": "sk-ant-...",
  "maxTokens": 8192,
  "temperature": 0.5,
  "theme": "dracula",
  "stream": true,
  "defaultPermission": "ask",
  "permissions": [
    { "toolName": "runCommand", "mode": "ask" },
    { "toolName": "readFile", "mode": "allow" }
  ]
}
```

---

## CI/CD Integration

```yaml
# GitHub Actions example
env:
  AIDEV_PROVIDER: claude
  AIDEV_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  AIDEV_MODEL: claude-sonnet-4-20250514
```

```bash
# Single question mode
aidev -q "Review the staged changes"
```

---

## Troubleshooting

**API key not found:**
Set `AIDEV_API_KEY` or the provider-specific variable (`ANTHROPIC_API_KEY`, etc.).

**Permission denied for tools:**
Check your `defaultPermission` and `permissions` array. Set to `"allow"` for auto-approve, `"ask"` for confirmation prompts.
