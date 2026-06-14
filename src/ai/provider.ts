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
  ToolCall,
  ToolDefinition,
  estimateCost,
} from '../types';
import { withRetry, getRetryableReason } from './retry';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Default request timeout in ms (60 seconds). */
const REQUEST_TIMEOUT_MS = 60_000;

/** Default per-chunk stream timeout in ms (3 minutes). */
const STREAM_CHUNK_TIMEOUT_MS = 180_000;

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
    req.setTimeout(REQUEST_TIMEOUT_MS, () => {
      req.destroy(new Error(`Request timed out after ${REQUEST_TIMEOUT_MS}ms`));
    });
    req.on('error', reject);
    if (body) req.write(body);
    req.end();
  });
}

/** Perform a streaming HTTP request, calling `onChunk` for each SSE line.
 *  Includes per-chunk timeout to detect hung connections. */
function httpStreamRequest(
  url: string,
  options: https.RequestOptions,
  body: string,
  onChunk: (line: string) => void,
): Promise<{ status: number }> {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    const transport = parsed.protocol === 'https:' ? https : http;
    let chunkTimer: ReturnType<typeof setTimeout> | null = null;

    const resetChunkTimer = () => {
      if (chunkTimer) clearTimeout(chunkTimer);
      chunkTimer = setTimeout(() => {
        req.destroy(new Error('SSE read timed out'));
      }, STREAM_CHUNK_TIMEOUT_MS);
    };

    const req = transport.request(url, options, (res) => {
      if (res.statusCode && res.statusCode >= 400) {
        const chunks: Buffer[] = [];
        res.on('data', (c: Buffer) => chunks.push(c));
        res.on('end', () => {
          if (chunkTimer) clearTimeout(chunkTimer);
          reject(new Error(`HTTP ${res.statusCode}: ${Buffer.concat(chunks).toString('utf-8')}`));
        });
        return;
      }

      resetChunkTimer();
      let buffer = '';
      res.on('data', (chunk: Buffer) => {
        resetChunkTimer();
        buffer += chunk.toString('utf-8');
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';
        for (const line of lines) {
          if (line.trim()) onChunk(line);
        }
      });
      res.on('end', () => {
        if (chunkTimer) clearTimeout(chunkTimer);
        if (buffer.trim()) onChunk(buffer);
        resolve({ status: res.statusCode ?? 0 });
      });
    });

    req.setTimeout(REQUEST_TIMEOUT_MS, () => {
      req.destroy(new Error(`Request timed out after ${REQUEST_TIMEOUT_MS}ms`));
    });
    req.on('error', (err) => {
      if (chunkTimer) clearTimeout(chunkTimer);
      reject(err);
    });
    req.write(body);
    req.end();
  });
}

/** Build a messages array suitable for the API, optionally prepending a system message. */
function buildMessages(
  messages: Message[],
  systemPrompt?: string,
): { system?: string; messages: { role: string; content: string; tool_call_id?: string; toolCalls?: ToolCall[] }[] } {
  const apiMessages = messages.map((m) => ({
    role: m.role,
    content: m.content,
    ...(m.toolCallId ? { tool_call_id: m.toolCallId } : {}),
    ...(m.toolCalls ? { toolCalls: m.toolCalls } : {}),
  }));
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
  /** Tool calls requested by the AI (if any) */
  toolCalls?: ToolCall[];
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

  /** Generate a full (non-streaming) response. Automatically retries transient errors. */
  async generate(messages: Message[], systemPrompt?: string, tools?: ToolDefinition[]): Promise<GenerateResult> {
    const result = await withRetry(
      () => this.doGenerate(messages, systemPrompt, tools),
      {
        onRetry: (attempt, delay, reason) => {
          process.stderr.write(`\n  [Retry ${attempt}: ${reason}, waiting ${Math.round(delay / 1000)}s]\n`);
        },
      },
    );
    result.estimatedCost = estimateCost(this.config.provider, this.config.model, result.usage);
    return result;
  }

  /** Stream a response, calling `onChunk` for every incremental piece. Automatically retries transient errors. */
  async stream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
    tools?: ToolDefinition[],
  ): Promise<GenerateResult> {
    const result = await withRetry(
      () => this.doStream(messages, onChunk, systemPrompt, tools),
      {
        onRetry: (attempt, delay, reason) => {
          process.stderr.write(`\n  [Retry ${attempt}: ${reason}, waiting ${Math.round(delay / 1000)}s]\n`);
        },
      },
    );
    result.estimatedCost = estimateCost(this.config.provider, this.config.model, result.usage);
    return result;
  }

  protected abstract doGenerate(messages: Message[], systemPrompt?: string, tools?: ToolDefinition[]): Promise<GenerateResult>;
  protected abstract doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
    tools?: ToolDefinition[],
  ): Promise<GenerateResult>;
}

// ---------------------------------------------------------------------------
// Claude (Anthropic)
// ---------------------------------------------------------------------------

/** Convert aidev messages to Claude's content block format. */
function toClaudeMessages(messages: Message[]): { role: string; content: unknown[] }[] {
  const result: { role: string; content: unknown[] }[] = [];

  for (const msg of messages) {
    if (msg.role === 'tool') {
      // Tool results are wrapped in a user message with tool_result content blocks
      result.push({
        role: 'user',
        content: [{
          type: 'tool_result',
          tool_use_id: msg.toolCallId,
          content: msg.content,
        }],
      });
    } else if (msg.role === 'assistant' && msg.toolCalls && msg.toolCalls.length > 0) {
      // Assistant message with tool calls: text + tool_use blocks
      const content: unknown[] = [];
      if (msg.content) content.push({ type: 'text', text: msg.content });
      for (const tc of msg.toolCalls) {
        content.push({ type: 'tool_use', id: tc.id, name: tc.name, input: tc.arguments });
      }
      result.push({ role: 'assistant', content });
    } else {
      result.push({ role: msg.role, content: [{ type: 'text', text: msg.content }] });
    }
  }
  return result;
}

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

  protected async doGenerate(messages: Message[], systemPrompt?: string, tools?: ToolDefinition[]): Promise<GenerateResult> {
    const claudeMessages = toClaudeMessages(messages);
    const body: Record<string, unknown> = {
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      messages: claudeMessages,
    };
    if (systemPrompt) body.system = systemPrompt;
    if (tools && tools.length > 0) {
      body.tools = tools.map((t) => ({
        name: t.name,
        description: t.description,
        input_schema: t.parameters,
      }));
    }

    const res = await httpRequest(`${this.baseUrl}/v1/messages`, {
      method: 'POST',
      headers: this.headers,
    }, JSON.stringify(body));

    if (res.status >= 400) {
      throw new Error(`Claude API error ${res.status}: ${res.body}`);
    }

    const json = JSON.parse(res.body) as {
      content: { type: string; text?: string; id?: string; name?: string; input?: Record<string, unknown> }[];
      usage: { input_tokens: number; output_tokens: number };
      stop_reason?: string;
    };

    const textParts = json.content.filter((b) => b.type === 'text').map((b) => b.text ?? '');
    const thinkingParts = json.content.filter((b) => b.type === 'thinking').map((b) => b.text ?? '');
    const toolUseBlocks = json.content.filter((b) => b.type === 'tool_use');

    const toolCalls: ToolCall[] | undefined = toolUseBlocks.length > 0
      ? toolUseBlocks.map((b) => ({ id: b.id!, name: b.name!, arguments: b.input ?? {} }))
      : undefined;

    return {
      content: textParts.join(''),
      thinking: thinkingParts.length > 0 ? thinkingParts.join('') : undefined,
      usage: {
        promptTokens: json.usage.input_tokens,
        completionTokens: json.usage.output_tokens,
        totalTokens: json.usage.input_tokens + json.usage.output_tokens,
      },
      estimatedCost: 0,
      toolCalls,
    };
  }

  protected async doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
    tools?: ToolDefinition[],
  ): Promise<GenerateResult> {
    const claudeMessages = toClaudeMessages(messages);
    const body: Record<string, unknown> = {
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      stream: true,
      messages: claudeMessages,
    };
    if (systemPrompt) body.system = systemPrompt;
    if (tools && tools.length > 0) {
      body.tools = tools.map((t) => ({
        name: t.name,
        description: t.description,
        input_schema: t.parameters,
      }));
    }

    let fullContent = '';
    let fullThinking = '';
    let usage: TokenUsage = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };

    // Accumulate tool calls across streaming chunks
    const toolCallMap = new Map<string, { id: string; name: string; inputJson: string }>();

    await httpStreamRequest(
      `${this.baseUrl}/v1/messages`,
      { method: 'POST', headers: this.headers },
      JSON.stringify(body),
      (line) => {
        if (!line.startsWith('data: ')) return;
        const data = line.slice(6).trim();
        if (data === '[DONE]') return;
        try {
          const evt = JSON.parse(data);
          if (evt.type === 'content_block_start') {
            if (evt.content_block?.type === 'tool_use') {
              toolCallMap.set(evt.index, {
                id: evt.content_block.id,
                name: evt.content_block.name,
                inputJson: '',
              });
            }
          } else if (evt.type === 'content_block_delta') {
            if (evt.delta?.type === 'text_delta') {
              fullContent += evt.delta.text;
              onChunk({ content: evt.delta.text, done: false });
            } else if (evt.delta?.type === 'thinking_delta') {
              fullThinking += evt.delta.thinking;
              onChunk({ content: '', thinking: evt.delta.thinking, done: false });
            } else if (evt.delta?.type === 'input_json_delta') {
              const existing = toolCallMap.get(String(evt.index));
              if (existing) {
                existing.inputJson += evt.delta.partial_json ?? '';
              }
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

    const toolCalls: ToolCall[] | undefined = toolCallMap.size > 0
      ? Array.from(toolCallMap.values()).map((tc) => ({
        id: tc.id,
        name: tc.name,
        arguments: (() => { try { return JSON.parse(tc.inputJson || '{}'); } catch { return {}; } })(),
      }))
      : undefined;

    onChunk({ content: '', done: true, usage });
    return {
      content: fullContent,
      thinking: fullThinking || undefined,
      usage,
      estimatedCost: 0,
      toolCalls,
    };
  }
}

// ---------------------------------------------------------------------------
// OpenAI-compatible (OpenAI, DeepSeek, Custom)
// ---------------------------------------------------------------------------

/** Convert aidev messages to OpenAI's message format (including tool calls/results). */
function toOpenAIMessages(messages: Message[]): unknown[] {
  const result: unknown[] = [];
  for (const msg of messages) {
    if (msg.role === 'tool') {
      result.push({ role: 'tool', tool_call_id: msg.toolCallId, content: msg.content });
    } else if (msg.role === 'assistant' && msg.toolCalls && msg.toolCalls.length > 0) {
      result.push({
        role: 'assistant',
        content: msg.content || null,
        tool_calls: msg.toolCalls.map((tc) => ({
          id: tc.id,
          type: 'function',
          function: { name: tc.name, arguments: JSON.stringify(tc.arguments) },
        })),
      });
    } else {
      result.push({ role: msg.role, content: msg.content });
    }
  }
  return result;
}

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

  protected async doGenerate(messages: Message[], systemPrompt?: string, tools?: ToolDefinition[]): Promise<GenerateResult> {
    const openAiMessages = toOpenAIMessages(messages);
    if (systemPrompt) {
      openAiMessages.unshift({ role: 'system', content: systemPrompt });
    }

    const body: Record<string, unknown> = {
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      messages: openAiMessages,
    };
    if (tools && tools.length > 0) {
      body.tools = tools.map((t) => ({
        type: 'function',
        function: { name: t.name, description: t.description, parameters: t.parameters },
      }));
    }

    const res = await httpRequest(`${this.baseUrl}/v1/chat/completions`, {
      method: 'POST',
      headers: this.headers,
    }, JSON.stringify(body));

    if (res.status >= 400) {
      throw new Error(`OpenAI-compat API error ${res.status}: ${res.body}`);
    }

    const json = JSON.parse(res.body) as {
      choices: {
        message: {
          content: string;
          reasoning_content?: string;
          tool_calls?: { id: string; type: string; function: { name: string; arguments: string } }[];
        };
      }[];
      usage: { prompt_tokens: number; completion_tokens: number; total_tokens: number };
    };

    const choice = json.choices[0];
    const toolCalls: ToolCall[] | undefined = choice.message.tool_calls?.map((tc) => ({
      id: tc.id,
      name: tc.function.name,
      arguments: (() => { try { return JSON.parse(tc.function.arguments); } catch { return {}; } })(),
    }));

    return {
      content: choice.message.content ?? '',
      thinking: choice.message.reasoning_content,
      usage: {
        promptTokens: json.usage.prompt_tokens,
        completionTokens: json.usage.completion_tokens,
        totalTokens: json.usage.total_tokens,
      },
      estimatedCost: 0,
      toolCalls,
    };
  }

  protected async doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
    tools?: ToolDefinition[],
  ): Promise<GenerateResult> {
    const openAiMessages = toOpenAIMessages(messages);
    if (systemPrompt) {
      openAiMessages.unshift({ role: 'system', content: systemPrompt });
    }

    const body: Record<string, unknown> = {
      model: this.config.model,
      max_tokens: this.config.maxTokens,
      temperature: this.config.temperature,
      stream: true,
      messages: openAiMessages,
    };
    if (tools && tools.length > 0) {
      body.tools = tools.map((t) => ({
        type: 'function',
        function: { name: t.name, description: t.description, parameters: t.parameters },
      }));
    }

    let fullContent = '';
    let fullThinking = '';
    let usage: TokenUsage = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };

    // Accumulate tool calls across streaming chunks
    const toolCallMap = new Map<number, { id: string; name: string; argumentsJson: string }>();

    await httpStreamRequest(
      `${this.baseUrl}/v1/chat/completions`,
      { method: 'POST', headers: this.headers },
      JSON.stringify(body),
      (line) => {
        if (!line.startsWith('data: ')) return;
        const data = line.slice(6).trim();
        if (data === '[DONE]') return;
        try {
          const evt = JSON.parse(data) as {
            choices: {
              delta: {
                content?: string;
                reasoning_content?: string;
                tool_calls?: { index: number; id?: string; function?: { name?: string; arguments?: string } }[];
              };
            }[];
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
          if (delta?.tool_calls) {
            for (const tc of delta.tool_calls) {
              if (!toolCallMap.has(tc.index)) {
                toolCallMap.set(tc.index, { id: tc.id ?? '', name: tc.function?.name ?? '', argumentsJson: '' });
              }
              const existing = toolCallMap.get(tc.index)!;
              if (tc.id) existing.id = tc.id;
              if (tc.function?.name) existing.name = tc.function.name;
              if (tc.function?.arguments) existing.argumentsJson += tc.function.arguments;
            }
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

    const toolCalls: ToolCall[] | undefined = toolCallMap.size > 0
      ? Array.from(toolCallMap.values()).map((tc) => ({
        id: tc.id,
        name: tc.name,
        arguments: (() => { try { return JSON.parse(tc.argumentsJson || '{}'); } catch { return {}; } })(),
      }))
      : undefined;

    onChunk({ content: '', done: true, usage });
    return {
      content: fullContent,
      thinking: fullThinking || undefined,
      usage,
      estimatedCost: 0,
      toolCalls,
    };
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

  private toGeminiMessages(messages: Message[]): { role: string; parts: unknown[] }[] {
    const result: { role: string; parts: unknown[] }[] = [];
    for (const msg of messages) {
      if (msg.role === 'system') continue; // handled via systemInstruction
      if (msg.role === 'tool') {
        // Tool results are functionResponse parts in a user message
        result.push({
          role: 'user',
          parts: [{
            functionResponse: {
              name: msg.toolName ?? msg.toolCallId ?? 'unknown',
              response: { content: msg.content },
            },
          }],
        });
      } else if (msg.role === 'assistant' && msg.toolCalls && msg.toolCalls.length > 0) {
        // Assistant message with function calls
        const parts: unknown[] = [];
        if (msg.content) parts.push({ text: msg.content });
        for (const tc of msg.toolCalls) {
          parts.push({ functionCall: { name: tc.name, args: tc.arguments } });
        }
        result.push({ role: 'model', parts });
      } else {
        result.push({
          role: msg.role === 'assistant' ? 'model' : 'user',
          parts: [{ text: msg.content }],
        });
      }
    }
    return result;
  }

  protected async doGenerate(messages: Message[], systemPrompt?: string, tools?: ToolDefinition[]): Promise<GenerateResult> {
    const geminiMessages = this.toGeminiMessages(messages);
    const body: Record<string, unknown> = {
      contents: geminiMessages,
      generationConfig: {
        temperature: this.config.temperature,
        maxOutputTokens: this.config.maxTokens,
      },
    };
    if (systemPrompt) {
      body.systemInstruction = { parts: [{ text: systemPrompt }] };
    }
    if (tools && tools.length > 0) {
      body.tools = [{
        function_declarations: tools.map((t) => ({
          name: t.name,
          description: t.description,
          parameters: t.parameters,
        })),
      }];
    }

    const url = `${this.baseUrl}/${this.modelPath}:generateContent?key=${this.apiKey}`;
    const res = await httpRequest(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...(this.config.extraHeaders ?? {}) },
    }, JSON.stringify(body));

    if (res.status >= 400) {
      throw new Error(`Gemini API error ${res.status}: ${res.body}`);
    }

    const json = JSON.parse(res.body) as {
      candidates: { content: { parts: { text?: string; thought?: boolean; functionCall?: { name: string; args?: Record<string, unknown> } }[] } }[];
      usageMetadata: { promptTokenCount: number; candidatesTokenCount: number; totalTokenCount: number };
    };

    const parts = json.candidates[0]?.content?.parts ?? [];
    const textContent = parts.filter((p) => !p.thought && !p.functionCall).map((p) => p.text ?? '').join('');
    const thinkingContent = parts.filter((p) => p.thought).map((p) => p.text ?? '').join('');

    const toolCalls: ToolCall[] | undefined = parts.some((p) => p.functionCall)
      ? parts.filter((p) => p.functionCall).map((p, i) => ({
        id: `gemini-${Date.now()}-${i}`,
        name: p.functionCall!.name,
        arguments: p.functionCall!.args ?? {},
      }))
      : undefined;

    return {
      content: textContent,
      thinking: thinkingContent || undefined,
      usage: {
        promptTokens: json.usageMetadata.promptTokenCount,
        completionTokens: json.usageMetadata.candidatesTokenCount,
        totalTokens: json.usageMetadata.totalTokenCount,
      },
      estimatedCost: 0,
      toolCalls,
    };
  }

  protected async doStream(
    messages: Message[],
    onChunk: (chunk: StreamChunk) => void,
    systemPrompt?: string,
    tools?: ToolDefinition[],
  ): Promise<GenerateResult> {
    const geminiMessages = this.toGeminiMessages(messages);
    const body: Record<string, unknown> = {
      contents: geminiMessages,
      generationConfig: {
        temperature: this.config.temperature,
        maxOutputTokens: this.config.maxTokens,
      },
    };
    if (systemPrompt) {
      body.systemInstruction = { parts: [{ text: systemPrompt }] };
    }
    if (tools && tools.length > 0) {
      body.tools = [{
        function_declarations: tools.map((t) => ({
          name: t.name,
          description: t.description,
          parameters: t.parameters,
        })),
      }];
    }

    const url = `${this.baseUrl}/${this.modelPath}:streamGenerateContent?alt=sse&key=${this.apiKey}`;
    let fullContent = '';
    let fullThinking = '';
    let usage: TokenUsage = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };
    const toolCallsList: ToolCall[] = [];

    await httpStreamRequest(
      url,
      { method: 'POST', headers: { 'Content-Type': 'application/json', ...(this.config.extraHeaders ?? {}) } },
      JSON.stringify(body),
      (line) => {
        if (!line.startsWith('data: ')) return;
        try {
          const evt = JSON.parse(line.slice(6)) as {
            candidates?: { content?: { parts?: { text?: string; thought?: boolean; functionCall?: { name: string; args?: Record<string, unknown> } }[] } }[];
            usageMetadata?: { promptTokenCount: number; candidatesTokenCount: number; totalTokenCount: number };
          };
          const parts = evt.candidates?.[0]?.content?.parts ?? [];
          for (const p of parts) {
            if (p.functionCall) {
              toolCallsList.push({
                id: `gemini-${Date.now()}-${toolCallsList.length}`,
                name: p.functionCall.name,
                arguments: p.functionCall.args ?? {},
              });
            } else if (p.thought) {
              fullThinking += p.text ?? '';
              onChunk({ content: '', thinking: p.text ?? '', done: false });
            } else if (p.text) {
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

    const toolCalls: ToolCall[] | undefined = toolCallsList.length > 0 ? toolCallsList : undefined;

    onChunk({ content: '', done: true, usage });
    return {
      content: fullContent,
      thinking: fullThinking || undefined,
      usage,
      estimatedCost: 0,
      toolCalls,
    };
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
