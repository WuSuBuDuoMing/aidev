import { describe, it, expect } from 'vitest';
import { createProvider, defaultBaseUrl, defaultModel } from '../ai/provider';
import { AIProviderEnum } from '../types';
import type { ModelConfig } from '../types';

function makeConfig(provider: AIProviderEnum, overrides?: Partial<ModelConfig>): ModelConfig {
  return {
    provider,
    model: defaultModel(provider),
    apiKey: 'test-key',
    temperature: 0.2,
    maxTokens: 4096,
    thinkingEnabled: false,
    ...overrides,
  };
}

describe('createProvider', () => {
  it('returns ClaudeProvider for Claude', () => {
    const provider = createProvider(makeConfig(AIProviderEnum.Claude));
    expect(provider).toBeDefined();
    expect(provider.constructor.name).toBe('ClaudeProvider');
  });

  it('returns OpenAICompatProvider for OpenAI', () => {
    const provider = createProvider(makeConfig(AIProviderEnum.OpenAI));
    expect(provider.constructor.name).toBe('OpenAICompatProvider');
  });

  it('returns OpenAICompatProvider for DeepSeek', () => {
    const provider = createProvider(makeConfig(AIProviderEnum.DeepSeek));
    expect(provider.constructor.name).toBe('OpenAICompatProvider');
  });

  it('returns OpenAICompatProvider for Ollama', () => {
    const provider = createProvider(makeConfig(AIProviderEnum.Ollama));
    expect(provider.constructor.name).toBe('OpenAICompatProvider');
  });

  it('returns OpenAICompatProvider for Custom', () => {
    const provider = createProvider(makeConfig(AIProviderEnum.Custom));
    expect(provider.constructor.name).toBe('OpenAICompatProvider');
  });

  it('returns GeminiProvider for Gemini', () => {
    const provider = createProvider(makeConfig(AIProviderEnum.Gemini));
    expect(provider.constructor.name).toBe('GeminiProvider');
  });
});

describe('defaultBaseUrl', () => {
  it('returns anthropic URL for Claude', () => {
    expect(defaultBaseUrl(AIProviderEnum.Claude)).toBe('https://api.anthropic.com');
  });

  it('returns openai URL for OpenAI', () => {
    expect(defaultBaseUrl(AIProviderEnum.OpenAI)).toBe('https://api.openai.com');
  });

  it('returns deepseek URL for DeepSeek', () => {
    expect(defaultBaseUrl(AIProviderEnum.DeepSeek)).toBe('https://api.deepseek.com');
  });

  it('returns google URL for Gemini', () => {
    expect(defaultBaseUrl(AIProviderEnum.Gemini)).toBe('https://generativelanguage.googleapis.com');
  });

  it('returns localhost for Ollama', () => {
    expect(defaultBaseUrl(AIProviderEnum.Ollama)).toBe('http://localhost:11434');
  });

  it('returns empty string for Custom', () => {
    expect(defaultBaseUrl(AIProviderEnum.Custom)).toBe('');
  });
});

describe('defaultModel', () => {
  it('returns a model for each known provider', () => {
    const providers = [
      AIProviderEnum.Claude,
      AIProviderEnum.OpenAI,
      AIProviderEnum.DeepSeek,
      AIProviderEnum.Gemini,
      AIProviderEnum.Ollama,
    ];
    for (const p of providers) {
      expect(defaultModel(p)).toBeTruthy();
    }
  });

  it('returns empty string for Custom', () => {
    expect(defaultModel(AIProviderEnum.Custom)).toBe('');
  });

  it('returns correct Claude model', () => {
    expect(defaultModel(AIProviderEnum.Claude)).toBe('claude-sonnet-4-20250514');
  });

  it('returns correct OpenAI model', () => {
    expect(defaultModel(AIProviderEnum.OpenAI)).toBe('gpt-4o');
  });
});
