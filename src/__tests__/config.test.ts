import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { validateConfig, getGlobalDir, getGlobalSkillsDir, getProjectConfigDir, getProjectSkillsDir, loadConfig } from '../config';
import { AIProviderEnum } from '../types';
import type { PartialConfig } from '../types';

describe('validateConfig', () => {
  it('returns empty array for valid config', () => {
    const errors = validateConfig({
      provider: AIProviderEnum.Claude,
      temperature: 0.5,
      maxTokens: 2048,
      theme: 'dark',
      defaultPermission: 'ask',
    });
    expect(errors).toEqual([]);
  });

  it('rejects invalid provider', () => {
    const errors = validateConfig({ provider: 'invalid' as AIProviderEnum });
    expect(errors).toHaveLength(1);
    expect(errors[0]).toContain('Invalid provider');
  });

  it('rejects temperature out of range', () => {
    expect(validateConfig({ temperature: -1 })).toHaveLength(1);
    expect(validateConfig({ temperature: 3 })).toHaveLength(1);
    expect(validateConfig({ temperature: 0 })).toHaveLength(0);
    expect(validateConfig({ temperature: 2 })).toHaveLength(0);
  });

  it('rejects maxTokens < 1', () => {
    expect(validateConfig({ maxTokens: 0 })).toHaveLength(1);
    expect(validateConfig({ maxTokens: -5 })).toHaveLength(1);
    expect(validateConfig({ maxTokens: 1 })).toHaveLength(0);
  });

  it('rejects invalid theme', () => {
    const errors = validateConfig({ theme: 'neon' });
    expect(errors).toHaveLength(1);
    expect(errors[0]).toContain('Invalid theme');
  });

  it('rejects invalid defaultPermission', () => {
    const errors = validateConfig({ defaultPermission: 'maybe' });
    expect(errors).toHaveLength(1);
    expect(errors[0]).toContain('Invalid defaultPermission');
  });

  it('collects multiple errors', () => {
    const errors = validateConfig({
      provider: 'bad' as AIProviderEnum,
      temperature: 99,
      theme: 'x',
    });
    expect(errors.length).toBeGreaterThanOrEqual(3);
  });

  it('accepts empty partial config', () => {
    expect(validateConfig({})).toEqual([]);
  });
});

describe('path helpers', () => {
  it('getGlobalDir returns a string containing .aidev', () => {
    expect(getGlobalDir()).toContain('.aidev');
  });

  it('getGlobalSkillsDir appends skills', () => {
    expect(getGlobalSkillsDir()).toMatch(/\.aidev[/\\]skills$/);
  });

  it('getProjectConfigDir uses cwd', () => {
    const dir = getProjectConfigDir('/tmp/myproject');
    expect(dir).toContain('.aidev');
    expect(dir).toContain('myproject');
  });

  it('getProjectSkillsDir uses cwd', () => {
    const dir = getProjectSkillsDir('/tmp/myproject');
    expect(dir).toMatch(/myproject[/\\]\.aidev[/\\]skills$/);
  });
});

describe('loadConfig', () => {
  const originalEnv = { ...process.env };

  beforeEach(() => {
    vi.stubEnv('AIDEV_PROVIDER', undefined as unknown as string);
    vi.stubEnv('AIDEV_API_KEY', undefined as unknown as string);
    vi.stubEnv('AIDEV_MODEL', undefined as unknown as string);
    vi.stubEnv('AIDEV_BASE_URL', undefined as unknown as string);
    vi.stubEnv('AIDEV_TEMPERATURE', undefined as unknown as string);
    vi.stubEnv('AIDEV_MAX_TOKENS', undefined as unknown as string);
    vi.stubEnv('AIDEV_THEME', undefined as unknown as string);
    vi.stubEnv('ANTHROPIC_API_KEY', undefined as unknown as string);
    vi.stubEnv('OPENAI_API_KEY', undefined as unknown as string);
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('returns defaults when no env vars set', () => {
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.provider).toBe(AIProviderEnum.Claude);
    expect(cfg.model).toBe('claude-sonnet-4-20250514');
    expect(cfg.temperature).toBe(0.2);
    expect(cfg.maxTokens).toBe(4096);
    expect(cfg.theme).toBe('dark');
    expect(cfg.stream).toBe(true);
  });

  it('overrides provider from env', () => {
    vi.stubEnv('AIDEV_PROVIDER', 'openai');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.provider).toBe(AIProviderEnum.OpenAI);
  });

  it('overrides apiKey from env', () => {
    vi.stubEnv('AIDEV_API_KEY', 'test-key-123');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.apiKey).toBe('test-key-123');
  });

  it('overrides model from env', () => {
    vi.stubEnv('AIDEV_MODEL', 'gpt-4o');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.model).toBe('gpt-4o');
  });

  it('overrides temperature from env', () => {
    vi.stubEnv('AIDEV_TEMPERATURE', '0.8');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.temperature).toBe(0.8);
  });

  it('overrides maxTokens from env', () => {
    vi.stubEnv('AIDEV_MAX_TOKENS', '8192');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.maxTokens).toBe(8192);
  });

  it('overrides theme from env', () => {
    vi.stubEnv('AIDEV_THEME', 'monokai');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.theme).toBe('monokai');
  });

  it('picks up provider-specific API key', () => {
    vi.stubEnv('OPENAI_API_KEY', 'sk-test');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.apiKey).toBe('sk-test');
  });

  it('silently ignores invalid env provider', () => {
    vi.stubEnv('AIDEV_PROVIDER', 'nonexistent');
    const cfg = loadConfig('/nonexistent/path');
    expect(cfg.provider).toBe(AIProviderEnum.Claude); // falls back to default
  });
});
