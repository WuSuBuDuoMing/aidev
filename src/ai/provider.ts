/**
 * aidev CLI — AI Provider Abstraction
 *
 * Provides a unified interface over multiple AI back-ends:
 *   Claude (Anthropic) · OpenAI · DeepSeek · Gemini · Ollama · Custom (OpenAI-compatible)
 *
 * Every concrete provider implements `generate()` (full response) and `stream()`
 * (chunked response via callback).
 *
 * @module ai/provider
 */

import * as https from 'https';
import * as http from 'http';
import { URL } from 'url';
import {
  AIProviderEnum,
  Message,
  ModelConfig,
  StreamChunk,
  TokenUsage,
  estimateCost,
} from '../types';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Perform an HTTPS/HTTP request and return the full response body as a string. */
function httpRequest(
  url: string,
  options: https.RequestOptions,
  body?: string,
): Promise<{ status: number; body: string; headers: Record<string, string | string[] | undefined> }> {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    const transport = parsed.protocol === 'https:' ? https : http;
    const req = transport.request(url, options, (res) => {
      const chunks: Buffer[] = [];
      res.on('data', (chunk: Buffer) => chunks.push(chunk));
      res.on('end', () => {
        resolve({
          status: res.statusCode ?? 0,
          body: Buffer.concat(chunks).toString('utf-8'),
          headers: res.headers as Record<string, string | string[] | undefined>,
        });
      });
    });
    req.on('error', reject);
    if (body) req.write(body);
    req.end();
  });
}

/** Perform a streaming HTTP request, calling `onChunk` for each SSE line. */
function httpStreamRequest(
  url: string,
  options: https.RequestOptions,
  body: string,
  onChunk: (line: string) => void,
): Promise<{ status: number }> {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    const transport = parsed.protocol === 'https:' ? https : http;
    const req = transport.request(url, options, (res) => {
      if (res.statusCode && res.statusCode >= 400) {
        const chunks: Buffer[] = [];
        res.on('data', (c: Buffer) => chunks.push(c));
        res.on('end', () => {
          reject(new Error(`HTTP ${res.statusCode}: ${Buffer.concat(chunks).toString('utf-8')}`));
        });
        return;
      }
      let buffer = '';
      res.on('data', (chunk: Buffer) => {
        buffer += chunk.toString('utf-8');
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';
        for (const line of lines) {
          if (line.trim()) onChunk(line);
        }
      });
      res.on('end', () => {
        if (buffer.trim()) onChunk(buffer);
        resolve({ status: res.statusCode ?? 0 });
      });
    });
    req.on('error', reject);
    req.write(body);
    req.end();
  });
}

/** Build a messages array suitable for the API, optionally prepending a system message. */
function buildMessages(
  messages: Message[],
  systemPrompt?: string,
): { system?: string; messages: { role: string; content: string }[] } {
  const apiMessages = messages.map((m) => ({ role: m.role, content: m.content }));
  return { system: systemPrompt, messages: apiMessages };
}

// ---------------------------------------------------------------------------
// Abstract base
// ---------------------------------------------------------------------------

/** Unified response type from `generate()`. */
export interface GenerateResult {
  content: string;
  thinking?: string;
  usage: TokenUsage;
  estimatedCost: number;
}

/**
 * Abstract AI provider.
 * Concrete implementations must implement `doGenerate()` and `doStream()`.
 */
export abstract class AIProviderBase {
  protected config: ModelConfig;

  constructor(config: ModelConfig) {
    this.config = config;
  }

  /** Generate a full (non-streaming) response. */
  async generate(messages: Message[], systemPrompt?: string): Promise<GenerateResult> {
    const result = await this.doGenerate(messages, systemPrompt);
    result.estimatedCost = estimateCost(this.config.provider, this.config.model, result.usage);
    return result;
  }

  /** Stream a response, calling `onChunk` for every incremental piece. */
  async stream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
  ): Promise<GenerateResult> {
    const result = await this.doStream(messages, onChunk, systemPrompt);
    result.estimatedCost = estimateCost(this.config.provider, this.config.model, result.usage);
    return result;
  }

  protected abstract doGenerate(messages: Message[], systemPrompt?: string): Promise<GenerateResult>;
  protected abstract doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
  ): Promise<GenerateResult>;
}

// ---------------------------------------------------------------------------
// Claude (Anthropic)
// ---------------------------------------------------------------------------

class ClaudeProvider extends AIProviderBase {
  private get baseUrl(): string {
    return this.config.baseUrl || 'https://api.anthropic.com';
  }

  private get headers(): Record<string, string> {
    return {
      'Content-Type': 'application/json',
      'x-api-key': this.config.apiKey ?? '',
      'anthropic-version': this.config.apiVersion ?? '2023-06-01',
      ...(this.config.extraHeaders ?? {}),
    };
  }

  protected async doGenerate(messages: Message[], systemPrompt?: string): Promise<GenerateResult> {
    const { system, messages: apiMessages } = buildMessages(messages, systemPrompt);
    const body = JSON.stringify({
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      ...(system ? { system } : {}),
      messages: apiMessages,
    });

    const res = await httpRequest(`${this.baseUrl}/v1/messages`, {
      method: 'POST',
      headers: this.headers,
    }, body);

    if (res.status >= 400) {
      throw new Error(`Claude API error ${res.status}: ${res.body}`);
    }

    const json = JSON.parse(res.body) as {
      content: { type: string; text: string }[];
      usage: { input_tokens: number; output_tokens: number };
    };

    const textParts = json.content.filter((b) => b.type === 'text').map((b) => b.text);
    const thinkingParts = json.content.filter((b) => b.type === 'thinking').map((b) => b.text);

    return {
      content: textParts.join(''),
      thinking: thinkingParts.length > 0 ? thinkingParts.join('') : undefined,
      usage: {
        promptTokens: json.usage.input_tokens,
        completionTokens: json.usage.output_tokens,
        totalTokens: json.usage.input_tokens + json.usage.output_tokens,
      },
      estimatedCost: 0,
    };
  }

  protected async doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
  ): Promise<GenerateResult> {
    const { system, messages: apiMessages } = buildMessages(messages, systemPrompt);
    const body = JSON.stringify({
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      stream: true,
      ...(system ? { system } : {}),
      messages: apiMessages,
    });

    let fullContent = '';
    let fullThinking = '';
    let usage: TokenUsage = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };

    await httpStreamRequest(
      `${this.baseUrl}/v1/messages`,
      { method: 'POST', headers: this.headers },
      body,
      (line) => {
        if (!line.startsWith('data: ')) return;
        const data = line.slice(6).trim();
        if (data === '[DONE]') return;
        try {
          const evt = JSON.parse(data);
          if (evt.type === 'content_block_delta') {
            if (evt.delta?.type === 'text_delta') {
              fullContent += evt.delta.text;
              onChunk({ content: evt.delta.text, done: false });
            } else if (evt.delta?.type === 'thinking_delta') {
              fullThinking += evt.delta.thinking;
              onChunk({ content: '', thinking: evt.delta.thinking, done: false });
            }
          } else if (evt.type === 'message_delta' && evt.usage) {
            usage = {
              promptTokens: evt.usage.input_tokens ?? 0,
              completionTokens: evt.usage.output_tokens ?? 0,
              totalTokens: (evt.usage.input_tokens ?? 0) + (evt.usage.output_tokens ?? 0),
            };
          }
        } catch { /* ignore non-JSON SSE lines */ }
      },
    );

    onChunk({ content: '', done: true, usage });
    return { content: fullContent, thinking: fullThinking || undefined, usage, estimatedCost: 0 };
  }
}

// ---------------------------------------------------------------------------
// OpenAI-compatible (OpenAI, DeepSeek, Custom)
// ---------------------------------------------------------------------------

class OpenAICompatProvider extends AIProviderBase {
  private get baseUrl(): string {
    return this.config.baseUrl || 'https://api.openai.com';
  }

  private get headers(): Record<string, string> {
    return {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${this.config.apiKey ?? ''}`,
      ...(this.config.extraHeaders ?? {}),
    };
  }

  protected async doGenerate(messages: Message[], systemPrompt?: string): Promise<GenerateResult> {
    const { system, messages: apiMessages } = buildMessages(messages, systemPrompt);
    const allMessages = [
      ...(system ? [{ role: 'system', content: system }] : []),
      ...apiMessages,
    ];

    const body = JSON.stringify({
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      messages: allMessages,
    });

    const res = await httpRequest(`${this.baseUrl}/v1/chat/completions`, {
      method: 'POST',
      headers: this.headers,
    }, body);

    if (res.status >= 400) {
      throw new Error(`OpenAI-compat API error ${res.status}: ${res.body}`);
    }

    const json = JSON.parse(res.body) as {
      choices: { message: { content: string; reasoning_content?: string } }[];
      usage: { prompt_tokens: number; completion_tokens: number; total_tokens: number };
    };

    const choice = json.choices[0];
    return {
      content: choice.message.content ?? '',
      thinking: choice.message.reasoning_content,
      usage: {
        promptTokens: json.usage.prompt_tokens,
        completionTokens: json.usage.completion_tokens,
        totalTokens: json.usage.total_tokens,
      },
      estimatedCost: 0,
    };
  }

  protected async doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
  ): Promise<GenerateResult> {
    const { system, messages: apiMessages } = buildMessages(messages, systemPrompt);
    const allMessages = [
      ...(system ? [{ role: 'system', content: system }] : []),
      ...apiMessages,
    ];

    const body = JSON.stringify({
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      stream: true,
      messages: allMessages,
    });

    let fullContent = '';
    let fullThinking = '';
    let usage: TokenUsage = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };

    await httpStreamRequest(
      `${this.baseUrl}/v1/chat/completions`,
      { method: 'POST', headers: this.headers },
      body,
      (line) => {
        if (!line.startsWith('data: ')) return;
        const data = line.slice(6).trim();
        if (data === '[DONE]') return;
        try {
          const evt = JSON.parse(data) as {
            choices: { delta: { content?: string; reasoning_content?: string } }[];
            usage?: { prompt_tokens: number; completion_tokens: number; total_tokens: number };
          };
          const delta = evt.choices?.[0]?.delta;
          if (delta?.content) {
            fullContent += delta.content;
            onChunk({ content: delta.content, done: false });
          }
          if (delta?.reasoning_content) {
            fullThinking += delta.reasoning_content;
            onChunk({ content: '', thinking: delta.reasoning_content, done: false });
          }
          if (evt.usage) {
            usage = {
              promptTokens: evt.usage.prompt_tokens,
              completionTokens: evt.usage.completion_tokens,
              totalTokens: evt.usage.total_tokens,
            };
          }
        } catch { /* ignore non-JSON lines */ }
      },
    );

    onChunk({ content: '', done: true, usage });
    return { content: fullContent, thinking: fullThinking || undefined, usage, estimatedCost: 0 };
  }
}

// ---------------------------------------------------------------------------
// Gemini
// ---------------------------------------------------------------------------

class GeminiProvider extends AIProviderBase {
  private get apiKey(): string {
    return this.config.apiKey ?? '';
  }

  private get baseUrl(): string {
    return this.config.baseUrl || 'https://generativelanguage.googleapis.com';
  }

  private get modelPath(): string {
    return `v1beta/models/${this.config.model}`;
  }

  private toGeminiMessages(messages: Message[]): { role: string; parts: { text: string }[] }[] {
    return messages
      .filter((m) => m.role !== 'system')
      .map((m) => ({
        role: m.role === 'assistant' ? 'model' : 'user',
        parts: [{ text: m.content }],
      }));
  }

  protected async doGenerate(messages: Message[], systemPrompt?: string): Promise<GenerateResult> {
    const geminiMessages = this.toGeminiMessages(messages);
    const body = JSON.stringify({
      contents: geminiMessages,
      ...(systemPrompt
        ? { systemInstruction: { parts: [{ text: systemPrompt }] } }
        : {}),
      generationConfig: {
        temperature: this.config.temperature,
        maxOutputTokens: this.config.maxTokens,
      },
    });

    const url = `${this.baseUrl}/${this.modelPath}:generateContent?key=${this.apiKey}`;
    const res = await httpRequest(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...(this.config.extraHeaders ?? {}) },
    }, body);

    if (res.status >= 400) {
      throw new Error(`Gemini API error ${res.status}: ${res.body}`);
    }

    const json = JSON.parse(res.body) as {
      candidates: { content: { parts: { text: string; thought?: boolean }[] } }[];
      usageMetadata: { promptTokenCount: number; candidatesTokenCount: number; totalTokenCount: number };
    };

    const parts = json.candidates[0]?.content?.parts ?? [];
    const textContent = parts.filter((p) => !p.thought).map((p) => p.text).join('');
    const thinkingContent = parts.filter((p) => p.thought).map((p) => p.text).join('');

    return {
      content: textContent,
      thinking: thinkingContent || undefined,
      usage: {
        promptTokens: json.usageMetadata.promptTokenCount,
        completionTokens: json.usageMetadata.candidatesTokenCount,
        totalTokens: json.usageMetadata.totalTokenCount,
      },
      estimatedCost: 0,
    };
  }

  protected async doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
  ): Promise<GenerateResult> {
    const geminiMessages = this.toGeminiMessages(messages);
    const body = JSON.stringify({
      contents: geminiMessages,
      ...(systemPrompt
        ? { systemInstruction: { parts: [{ text: systemPrompt }] } }
        : {}),
      generationConfig: {
        temperature: this.config.temperature,
        maxOutputTokens: this.config.maxTokens,
      },
    });

    const url = `${this.baseUrl}/${this.modelPath}:streamGenerateContent?alt=sse&key=${this.apiKey}`;
    let fullContent = '';
    let fullThinking = '';
    let usage: TokenUsage = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };

    await httpStreamRequest(
      url,
      { method: 'POST', headers: { 'Content-Type': 'application/json', ...(this.config.extraHeaders ?? {}) } },
      body,
      (line) => {
        if (!line.startsWith('data: ')) return;
        try {
          const evt = JSON.parse(line.slice(6)) as {
            candidates?: { content?: { parts?: { text: string; thought?: boolean }[] } }[];
            usageMetadata?: { promptTokenCount: number; candidatesTokenCount: number; totalTokenCount: number };
          };
          const parts = evt.candidates?.[0]?.content?.parts ?? [];
          for (const p of parts) {
            if (p.thought) {
              fullThinking += p.text;
              onChunk({ content: '', thinking: p.text, done: false });
            } else {
              fullContent += p.text;
              onChunk({ content: p.text, done: false });
            }
          }
          if (evt.usageMetadata) {
            usage = {
              promptTokens: evt.usageMetadata.promptTokenCount,
              completionTokens: evt.usageMetadata.candidatesTokenCount,
              totalTokens: evt.usageMetadata.totalTokenCount,
            };
          }
        } catch { /* ignore non-JSON lines */ }
      },
    );

    onChunk({ content: '', done: true, usage });
    return { content: fullContent, thinking: fullThinking || undefined, usage, estimatedCost: 0 };
  }
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create an AI provider instance for the given model configuration.
 *
 * @param config - The model configuration including provider, model name, and credentials.
 * @returns A concrete `AIProviderBase` instance.
 */
export function createProvider(config: ModelConfig): AIProviderBase {
  switch (config.provider) {
    case AIProviderEnum.Claude:
      return new ClaudeProvider(config);
    case AIProviderEnum.Gemini:
      return new GeminiProvider(config);
    case AIProviderEnum.OpenAI:
    case AIProviderEnum.DeepSeek:
    case AIProviderEnum.Ollama:
    case AIProviderEnum.Custom:
    default:
      return new OpenAICompatProvider(config);
  }
}

/** Infer the default base URL for a provider. */
export function defaultBaseUrl(provider: AIProviderEnum): string {
  switch (provider) {
    case AIProviderEnum.Claude:
      return 'https://api.anthropic.com';
    case AIProviderEnum.OpenAI:
      return 'https://api.openai.com';
    case AIProviderEnum.DeepSeek:
      return 'https://api.deepseek.com';
    case AIProviderEnum.Gemini:
      return 'https://generativelanguage.googleapis.com';
    case AIProviderEnum.Ollama:
      return 'http://localhost:11434';
    default:
      return '';
  }
}

/** Infer a sensible default model identifier for a provider. */
export function defaultModel(provider: AIProviderEnum): string {
  switch (provider) {
    case AIProviderEnum.Claude:
      return 'claude-sonnet-4-20250514';
    case AIProviderEnum.OpenAI:
      return 'gpt-4o';
    case AIProviderEnum.DeepSeek:
      return 'deepseek-chat';
    case AIProviderEnum.Gemini:
      return 'gemini-2.5-flash';
    case AIProviderEnum.Ollama:
      return 'llama3.1';
    default:
      return '';
  }
}
