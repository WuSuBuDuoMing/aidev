import { describe, it, expect } from 'vitest';
import { AIProviderEnum, estimateCost, COST_TABLE } from '../types';
import type { TokenUsage } from '../types';

describe('AIProviderEnum', () => {
  it('has all expected providers', () => {
    expect(Object.values(AIProviderEnum)).toEqual([
      'claude', 'openai', 'deepseek', 'gemini', 'ollama', 'custom',
    ]);
  });
});

describe('COST_TABLE', () => {
  it('has entries for Claude models', () => {
    const claude = COST_TABLE[AIProviderEnum.Claude];
    expect(claude['claude-sonnet-4-20250514']).toBeDefined();
    expect(claude['claude-sonnet-4-20250514'].input).toBe(3);
    expect(claude['claude-sonnet-4-20250514'].output).toBe(15);
  });

  it('has entries for OpenAI models', () => {
    const openai = COST_TABLE[AIProviderEnum.OpenAI];
    expect(openai['gpt-4o']).toBeDefined();
    expect(openai['gpt-4o-mini']).toBeDefined();
  });

  it('has entries for DeepSeek models', () => {
    const ds = COST_TABLE[AIProviderEnum.DeepSeek];
    expect(ds['deepseek-chat']).toBeDefined();
    expect(ds['deepseek-reasoner']).toBeDefined();
  });

  it('has entries for Gemini models', () => {
    const gemini = COST_TABLE[AIProviderEnum.Gemini];
    expect(gemini['gemini-2.5-pro']).toBeDefined();
    expect(gemini['gemini-2.5-flash']).toBeDefined();
  });

  it('has empty tables for Ollama and Custom', () => {
    expect(Object.keys(COST_TABLE[AIProviderEnum.Ollama])).toHaveLength(0);
    expect(Object.keys(COST_TABLE[AIProviderEnum.Custom])).toHaveLength(0);
  });
});

describe('estimateCost', () => {
  it('calculates cost correctly for Claude', () => {
    const usage: TokenUsage = {
      promptTokens: 1_000_000,
      completionTokens: 1_000_000,
      totalTokens: 2_000_000,
    };
    const cost = estimateCost(AIProviderEnum.Claude, 'claude-sonnet-4-20250514', usage);
    expect(cost).toBe(18); // 3 + 15
  });

  it('returns 0 for unknown model', () => {
    const usage: TokenUsage = { promptTokens: 1000, completionTokens: 1000, totalTokens: 2000 };
    const cost = estimateCost(AIProviderEnum.Claude, 'unknown-model', usage);
    expect(cost).toBe(0);
  });

  it('returns 0 for Ollama (local, no cost)', () => {
    const usage: TokenUsage = { promptTokens: 1000, completionTokens: 1000, totalTokens: 2000 };
    const cost = estimateCost(AIProviderEnum.Ollama, 'llama3.1', usage);
    expect(cost).toBe(0);
  });

  it('scales proportionally with token count', () => {
    const usage: TokenUsage = { promptTokens: 500_000, completionTokens: 500_000, totalTokens: 1_000_000 };
    const cost = estimateCost(AIProviderEnum.OpenAI, 'gpt-4o', usage);
    expect(cost).toBe(6.25); // (500k * 2.5 + 500k * 10) / 1M
  });

  it('returns fractional cost for small usage', () => {
    const usage: TokenUsage = { promptTokens: 1000, completionTokens: 1000, totalTokens: 2000 };
    const cost = estimateCost(AIProviderEnum.DeepSeek, 'deepseek-chat', usage);
    expect(cost).toBeCloseTo(0.00137, 5); // (1000 * 0.27 + 1000 * 1.1) / 1M
  });
});
