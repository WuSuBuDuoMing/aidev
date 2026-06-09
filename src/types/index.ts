/**
 * aidev CLI — Core Type Definitions
 * @module types
 */

/** Available AI providers */
export enum AIProviderEnum {
  Claude = 'claude',
  OpenAI = 'openai',
  DeepSeek = 'deepseek',
  Gemini = 'gemini',
  Ollama = 'ollama',
  Custom = 'custom',
}

/** Role of a conversation message */
export type MessageRole = 'system' | 'user' | 'assistant' | 'tool';

/** A single tool call embedded in an assistant message */
export interface ToolCall {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
}

/** A single message in a conversation */
export interface Message {
  role: MessageRole;
  content: string;
  timestamp: number;
  toolCalls?: ToolCall[];
  toolCallId?: string;
}

/** A full conversation with metadata */
export interface Conversation {
  id: string;
  title: string;
  messages: Message[];
  createdAt: number;
  updatedAt: number;
  model: string;
  provider: AIProviderEnum;
}

/** Configuration for an AI model */
export interface ModelConfig {
  provider: AIProviderEnum;
  model: string;
  apiKey?: string;
  baseUrl?: string;
  temperature: number;
  maxTokens: number;
  thinkingEnabled: boolean;
  /** System-level API version override (provider-specific) */
  apiVersion?: string;
  /** Extra headers to send with every request */
  extraHeaders?: Record<string, string>;
}

/** Parameters schema for a tool (JSON Schema fragment) */
export interface ToolParameters {
  type: 'object';
  properties: Record<string, unknown>;
  required?: string[];
}

/** Result returned by a tool execution */
export interface ToolResult {
  success: boolean;
  output: string;
  error?: string;
}

/** A callable tool that the AI can invoke */
export interface Tool {
  name: string;
  description: string;
  parameters: ToolParameters;
  execute: (args: Record<string, unknown>) => Promise<ToolResult>;
}

/** Permission mode for a tool */
export type PermissionMode = 'allow' | 'deny' | 'ask';

/** A permission rule mapping a tool to a mode */
export interface Permission {
  toolName: string;
  mode: PermissionMode;
}

/** A registered plugin */
export interface Plugin {
  name: string;
  version: string;
  description: string;
  commands: PluginCommand[];
  tools: Tool[];
}

/** A command contributed by a plugin */
export interface PluginCommand {
  name: string;
  description: string;
  handler: (args: string[]) => Promise<void>;
}

/** Information about the git state of the project */
export interface GitInfo {
  branch: string;
  lastCommit: string;
  recentCommits: string[];
  isDirty: boolean;
  stagedFiles: string[];
  unstagedFiles: string[];
  untrackedFiles: string[];
}

/** Summary of the current project */
export interface ProjectContext {
  files: string[];
  structure: string;
  gitInfo: GitInfo | null;
  language: string;
  framework: string;
  packageJson: Record<string, unknown> | null;
  /** e.g. requirements.txt, go.mod, Cargo.toml */
  manifestFile: string | null;
  manifestContent: string | null;
}

/** An AI skill (reusable prompt template) */
export interface Skill {
  name: string;
  description: string;
  trigger: string;
  prompt: string;
}

/** Global application configuration */
export interface AppConfig {
  /** Currently selected provider */
  provider: AIProviderEnum;
  /** Currently selected model identifier */
  model: string;
  /** API key for the current provider */
  apiKey: string;
  /** Override base URL (useful for proxies or self-hosted) */
  baseUrl: string;
  /** Sampling temperature (0-2) */
  temperature: number;
  /** Maximum tokens to generate */
  maxTokens: number;
  /** Enable extended thinking / reasoning mode */
  thinkingEnabled: boolean;
  /** Default permission mode for tools */
  defaultPermission: PermissionMode;
  /** Per-tool permission overrides */
  permissions: Permission[];
  /** Active UI theme */
  theme: 'dark' | 'light' | 'monokai' | 'dracula';
  /** Whether to stream responses */
  stream: boolean;
  /** Whether to show token counts */
  showTokenCounts: boolean;
  /** User display name */
  userName: string;
}

/** Partial config used for merging / overrides */
export type PartialConfig = Partial<AppConfig>;

/** Streaming chunk returned by providers */
export interface StreamChunk {
  /** Incremental text content */
  content: string;
  /** Reasoning/thinking text (provider-specific) */
  thinking?: string;
  /** Whether this is the final chunk */
  done: boolean;
  /** Usage information (may only appear in final chunk) */
  usage?: TokenUsage;
}

/** Token usage statistics */
export interface TokenUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  /** Estimated cost in USD */
  estimatedCost?: number;
}

/** Cost table per provider (USD per 1M tokens) */
export const COST_TABLE: Record<AIProviderEnum, Record<string, { input: number; output: number }>> = {
  [AIProviderEnum.Claude]: {
    'claude-sonnet-4-20250514': { input: 3, output: 15 },
    'claude-3-5-sonnet-20241022': { input: 3, output: 15 },
    'claude-3-5-haiku-20241022': { input: 0.80, output: 4 },
    'claude-3-opus-20240229': { input: 15, output: 75 },
  },
  [AIProviderEnum.OpenAI]: {
    'gpt-4o': { input: 2.5, output: 10 },
    'gpt-4o-mini': { input: 0.15, output: 0.60 },
    'o1': { input: 15, output: 60 },
    'o3-mini': { input: 1.1, output: 4.4 },
  },
  [AIProviderEnum.DeepSeek]: {
    'deepseek-chat': { input: 0.27, output: 1.1 },
    'deepseek-reasoner': { input: 0.55, output: 2.19 },
  },
  [AIProviderEnum.Gemini]: {
    'gemini-2.5-pro': { input: 1.25, output: 10 },
    'gemini-2.5-flash': { input: 0.15, output: 0.60 },
    'gemini-2.0-flash': { input: 0.10, output: 0.40 },
  },
  [AIProviderEnum.Ollama]: {},
  [AIProviderEnum.Custom]: {},
};

/** Estimate cost for a given provider, model, and token usage */
export function estimateCost(
  provider: AIProviderEnum,
  model: string,
  usage: TokenUsage,
): number {
  const table = COST_TABLE[provider]?.[model];
  if (!table) return 0;
  return (usage.promptTokens * table.input + usage.completionTokens * table.output) / 1_000_000;
}
