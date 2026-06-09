# Provider Setup Guide

This guide covers how to set up and configure each AI provider supported by aidev.

---

## Overview

aidev supports six categories of AI providers:

| Category | Providers | Auth Method |
|---|---|---|
| Claude (Anthropic) | Anthropic API | API Key |
| GPT (OpenAI) | OpenAI API | API Key |
| DeepSeek | DeepSeek API | API Key |
| Gemini (Google) | Google AI Studio / Vertex AI | API Key |
| Ollama | Local Ollama server | None (local) |
| OpenAI-compatible | Any compatible API | API Key / Custom |

---

## Claude (Anthropic)

### Getting an API Key

1. Visit [console.anthropic.com](https://console.anthropic.com)
2. Create an account or sign in
3. Navigate to **API Keys**
4. Click **Create Key**
5. Copy the key (starts with `sk-ant-...`)

### Configuration

```bash
# Set provider
aidev config set provider claude

# Set API key
aidev config set apiKey "sk-ant-api03-..."

# Set model (optional, defaults to claude-sonnet-4-20250514)
aidev config set model claude-sonnet-4-20250514
```

Or via environment variables:

```bash
export AIDEV_PROVIDER=claude
export AIDEV_API_KEY="sk-ant-api03-..."
# Or use the standard Anthropic variable
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

### Available Models

| Model ID | Context Window | Best For |
|---|---|---|
| `claude-sonnet-4-20250514` | 200K tokens | General coding, balanced speed/quality |
| `claude-3-5-sonnet-20241022` | 200K tokens | Previous generation, still very capable |
| `claude-3-opus-20240229` | 200K tokens | Complex reasoning, large codebases |
| `claude-3-haiku-20240307` | 200K tokens | Fast responses, simple tasks |

### Features

- Streaming: Yes
- Function calling: Yes
- Vision (image input): Yes
- Max output tokens: 4096 (default), up to 8192

---

## GPT (OpenAI)

### Getting an API Key

1. Visit [platform.openai.com](https://platform.openai.com)
2. Create an account or sign in
3. Navigate to **API Keys** in the sidebar
4. Click **Create new secret key**
5. Copy the key (starts with `sk-...`)

### Configuration

```bash
# Set provider
aidev config set provider gpt

# Set API key
aidev config set apiKey "sk-proj-..."

# Set model (optional, defaults to gpt-4o)
aidev config set model gpt-4o
```

Or via environment variables:

```bash
export AIDEV_PROVIDER=gpt
export AIDEV_API_KEY="sk-proj-..."
# Or use the standard OpenAI variable
export OPENAI_API_KEY="sk-proj-..."
```

### Available Models

| Model ID | Context Window | Best For |
|---|---|---|
| `gpt-4o` | 128K tokens | Best overall quality and speed |
| `gpt-4o-mini` | 128K tokens | Cost-effective, fast responses |
| `gpt-4-turbo` | 128K tokens | Previous generation flagship |
| `gpt-3.5-turbo` | 16K tokens | Legacy, very fast, lower quality |

### Features

- Streaming: Yes
- Function calling: Yes
- Vision (image input): Yes (gpt-4o, gpt-4-turbo)
- JSON mode: Yes

---

## DeepSeek

### Getting an API Key

1. Visit [platform.deepseek.com](https://platform.deepseek.com)
2. Create an account or sign in
3. Navigate to **API Keys**
4. Click **Create API Key**
5. Copy the key

### Configuration

```bash
# Set provider
aidev config set provider deepseek

# Set API key
aidev config set apiKey "sk-..."

# Set model (optional, defaults to deepseek-chat)
aidev config set model deepseek-chat
```

Or via environment variables:

```bash
export AIDEV_PROVIDER=deepseek
export AIDEV_DEEPSEEK_API_KEY="sk-..."
```

### Available Models

| Model ID | Context Window | Best For |
|---|---|---|
| `deepseek-chat` | 64K tokens | General conversation and coding |
| `deepseek-coder` | 64K tokens | Specialized code generation |

### Features

- Streaming: Yes
- Function calling: Yes
- OpenAI-compatible API (uses same chat completions format)

---

## Gemini (Google)

### Getting an API Key

1. Visit [aistudio.google.com](https://aistudio.google.com)
2. Sign in with your Google account
3. Click **Get API key** in the left sidebar
4. Create a new API key or use an existing project
5. Copy the key (starts with `AIza...`)

### Configuration

```bash
# Set provider
aidev config set provider gemini

# Set API key
aidev config set apiKey "AIza..."

# Set model (optional, defaults to gemini-2.5-pro)
aidev config set model gemini-2.5-pro
```

Or via environment variables:

```bash
export AIDEV_PROVIDER=gemini
export AIDEV_GEMINI_API_KEY="AIza..."
# Or use the standard Google variable
export GOOGLE_API_KEY="AIza..."
```

### Available Models

| Model ID | Context Window | Best For |
|---|---|---|
| `gemini-2.5-pro` | 1M tokens | Largest context, complex analysis |
| `gemini-2.5-flash` | 1M tokens | Fast with large context |
| `gemini-2.0-flash` | 1M tokens | Fastest, good for simple tasks |

### Features

- Streaming: Yes
- Function calling: Yes
- Vision (image input): Yes
- Largest context window (1M tokens)

---

## Ollama (Local Models)

Ollama allows you to run AI models locally on your machine with zero API costs and complete privacy.

### Installation

```bash
# macOS / Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows
# Download from https://ollama.ai/download
```

### Pull a Model

```bash
# Pull a coding-focused model
ollama pull codellama:13b

# Or a general model
ollama pull llama3.1

# Or a smaller model for faster responses
ollama pull llama3.1:8b
```

### Configuration

```bash
# Set provider
aidev config set provider ollama

# Set model (default: llama3.1)
aidev config set model codellama:13b

# Set base URL if Ollama is not on localhost
aidev config set baseUrl "http://192.168.1.100:11434"
```

Or via environment variables:

```bash
export AIDEV_PROVIDER=ollama
export AIDEV_OLLAMA_BASE_URL="http://localhost:11434"
```

### Recommended Models for Coding

| Model | Size | VRAM | Best For |
|---|---|---|---|
| `codellama:13b` | 13B | 8 GB | Code generation and explanation |
| `deepseek-coder-v2:16b` | 16B | 10 GB | Advanced code tasks |
| `llama3.1:8b` | 8B | 6 GB | General purpose, fast |
| `llama3.1:70b` | 70B | 48 GB | Highest quality (needs powerful GPU) |
| `mistral:7b` | 7B | 5 GB | Fast, good balance |
| `qwen2.5-coder:7b` | 7B | 5 GB | Excellent code understanding |

### Features

- Streaming: Yes
- Function calling: Varies by model
- Zero cost, full privacy
- Requires Ollama server running locally (or on your network)

### Troubleshooting Ollama

**Connection refused:**
Ensure Ollama is running: `ollama serve` (it usually starts automatically).

**Slow responses:**
Use a smaller model (8B instead of 70B), or ensure your GPU has enough VRAM.

**Model not found:**
Pull the model first: `ollama pull <model-name>`. Check available models with `ollama list`.

---

## OpenAI-Compatible Providers

Any API that implements the OpenAI chat completions interface can be used with aidev. This includes:

- **Together AI** (together.ai)
- **Groq** (groq.com)
- **Fireworks AI** (fireworks.ai)
- **Mistral AI** (mistral.ai)
- **Perplexity** (perplexity.ai)
- **OpenRouter** (openrouter.ai)
- **vLLM** (self-hosted)
- **LocalAI** (self-hosted)
- **llama.cpp server** (self-hosted)

### Configuration

```bash
# Set provider to openai-compatible
aidev config set provider openai

# Set the base URL
aidev config set baseUrl "https://api.together.xyz/v1"

# Set API key
aidev config set apiKey "your-api-key"

# Set model name (provider-specific)
aidev config set model "meta-llama/Llama-3-70b-chat-hf"
```

### Examples

#### Together AI

```bash
aidev config set provider openai
aidev config set baseUrl "https://api.together.xyz/v1"
aidev config set apiKey "your-together-api-key"
aidev config set model "meta-llama/Llama-3-70b-chat-hf"
```

#### Groq

```bash
aidev config set provider openai
aidev config set baseUrl "https://api.groq.com/openai/v1"
aidev config set apiKey "gsk_..."
aidev config set model "llama-3.1-70b-versatile"
```

#### OpenRouter

```bash
aidev config set provider openai
aidev config set baseUrl "https://openrouter.ai/api/v1"
aidev config set apiKey "sk-or-..."
aidev config set model "anthropic/claude-sonnet-4-20250514"
```

#### Self-hosted (vLLM / LocalAI)

```bash
aidev config set provider openai
aidev config set baseUrl "http://localhost:8000/v1"
aidev config set apiKey "not-needed"
aidev config set model "your-model-name"
```

---

## Switching Providers

You can switch providers at any time:

### During a Session

```bash
# Switch in the REPL
/model claude-sonnet-4-20250514

# Or switch provider entirely
/model gpt-4o
```

### Persistently

```bash
# Change default provider
aidev config set provider gpt
aidev config set model gpt-4o
```

### Per-Project

```bash
# Use Claude for one project, GPT for another
cd project-a
aidev config set provider claude --project

cd project-b
aidev config set provider gpt --project
```

---

## Provider Comparison

| Feature | Claude | GPT | DeepSeek | Gemini | Ollama | OpenAI-compat |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Streaming | Yes | Yes | Yes | Yes | Yes | Yes |
| Function calling | Yes | Yes | Yes | Yes | Varies | Varies |
| Vision | Yes | Yes | No | Yes | Varies | Varies |
| Max context | 200K | 128K | 64K | 1M | Varies | Varies |
| Cost per 1M input | $3-15 | $2.5-10 | $0.14-2 | $1.25-10 | Free | Varies |
| Local | No | No | No | No | Yes | Varies |
| Speed | Fast | Fast | Fast | Fast | Varies | Varies |
| Code quality | Excellent | Excellent | Very Good | Very Good | Varies | Varies |

---

## Troubleshooting

### General

**"Provider not configured" error:**
Run `aidev config set provider <name>` and `aidev config set apiKey <key>`.

**"Invalid API key" error:**
Verify your API key is correct. Check that it hasn't expired or been revoked.

**"Rate limit exceeded" error:**
Wait a moment and retry. Consider upgrading your API plan or switching to a different provider temporarily.

### Network Issues

**Connection timeout:**
Check your internet connection. For Ollama, ensure the server is running on the expected port.

**Proxy configuration:**
Set the standard `HTTP_PROXY` and `HTTPS_PROXY` environment variables:

```bash
export HTTP_PROXY="http://proxy.example.com:8080"
export HTTPS_PROXY="http://proxy.example.com:8080"
```

### Provider-Specific

**Claude: "Invalid model" error:**
Ensure you're using a valid model ID (see the Claude models table above). Model IDs are case-sensitive.

**GPT: "Insufficient quota" error:**
Check your OpenAI billing at [platform.openai.com/account/billing](https://platform.openai.com/account/billing).

**Gemini: "API not enabled" error:**
Enable the Generative Language API in your Google Cloud Console.

**Ollama: "Model not loaded" error:**
The model may still be downloading. Check with `ollama list` and wait for the download to complete.
