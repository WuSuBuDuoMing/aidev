/**
 * aidev CLI — Configuration Manager
 *
 * Loads and merges configuration from three layers:
 *   1. Built-in defaults
 *   2. Global config   (~/.aidev/config.json)
 *   3. Project config  (<project>/.aidev/config.json)
 *   4. Environment variables (highest priority)
 *
 * @module config
 */

import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { AIProviderEnum, AppConfig, PartialConfig, PermissionMode } from '../types';

// ---------------------------------------------------------------------------
// Paths
// ---------------------------------------------------------------------------

const GLOBAL_DIR = path.join(os.homedir(), '.aidev');
const GLOBAL_CONFIG_PATH = path.join(GLOBAL_DIR, 'config.json');

function projectConfigPath(cwd: string): string {
  return path.join(cwd, '.aidev', 'config.json');
}

// ---------------------------------------------------------------------------
// Defaults
// ---------------------------------------------------------------------------

const DEFAULT_CONFIG: AppConfig = {
  provider: AIProviderEnum.Claude,
  model: 'claude-sonnet-4-20250514',
  apiKey: '',
  baseUrl: '',
  temperature: 0.2,
  maxTokens: 4096,
  thinkingEnabled: false,
  defaultPermission: 'ask' as PermissionMode,
  permissions: [],
  theme: 'dark',
  stream: true,
  showTokenCounts: true,
  userName: 'Developer',
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Safely read and parse a JSON file. Returns `null` on any error. */
function readJsonFile(filePath: string): Record<string, unknown> | null {
  try {
    if (!fs.existsSync(filePath)) return null;
    const raw = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return null;
  }
}

/** Deep-merge source into target (source wins on conflict). */
function deepMerge<T extends Record<string, unknown>>(target: T, source: Partial<T>): T {
  const result = { ...target };
  for (const key of Object.keys(source) as (keyof T)[]) {
    if (
      source[key] &&
      typeof source[key] === 'object' &&
      !Array.isArray(source[key]) &&
      target[key] &&
      typeof target[key] === 'object' &&
      !Array.isArray(target[key])
    ) {
      result[key] = deepMerge(
        target[key] as Record<string, unknown>,
        source[key] as Record<string, unknown>,
      ) as T[keyof T];
    } else if (source[key] !== undefined) {
      result[key] = source[key] as T[keyof T];
    }
  }
  return result;
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

const VALID_PROVIDERS = Object.values(AIProviderEnum) as string[];
const VALID_THEMES = ['dark', 'light', 'monokai', 'dracula'];
const VALID_PERMISSIONS = ['allow', 'deny', 'ask'];

/** Validate a partial config and return an array of error messages. */
export function validateConfig(cfg: PartialConfig): string[] {
  const errors: string[] = [];
  if (cfg.provider && !VALID_PROVIDERS.includes(cfg.provider)) {
    errors.push(`Invalid provider "${cfg.provider}". Valid: ${VALID_PROVIDERS.join(', ')}`);
  }
  if (cfg.temperature !== undefined && (cfg.temperature < 0 || cfg.temperature > 2)) {
    errors.push('temperature must be between 0 and 2');
  }
  if (cfg.maxTokens !== undefined && cfg.maxTokens < 1) {
    errors.push('maxTokens must be >= 1');
  }
  if (cfg.theme && !VALID_THEMES.includes(cfg.theme)) {
    errors.push(`Invalid theme "${cfg.theme}". Valid: ${VALID_THEMES.join(', ')}`);
  }
  if (cfg.defaultPermission && !VALID_PERMISSIONS.includes(cfg.defaultPermission)) {
    errors.push(`Invalid defaultPermission "${cfg.defaultPermission}". Valid: ${VALID_PERMISSIONS.join(', ')}`);
  }
  return errors;
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/** Ensure the global config directory exists. */
export function ensureGlobalDir(): void {
  if (!fs.existsSync(GLOBAL_DIR)) {
    fs.mkdirSync(GLOBAL_DIR, { recursive: true });
  }
}

/** Save configuration to the global config file. */
export function saveGlobalConfig(cfg: PartialConfig): void {
  ensureGlobalDir();
  const existing = readJsonFile(GLOBAL_CONFIG_PATH) ?? {};
  const merged = deepMerge(existing, cfg as Record<string, unknown>);
  fs.writeFileSync(GLOBAL_CONFIG_PATH, JSON.stringify(merged, null, 2), 'utf-8');
}

/** Load and merge all configuration layers into a single AppConfig. */
export function loadConfig(cwd?: string): AppConfig {
  // Layer 1: defaults
  let config: Record<string, unknown> = { ...DEFAULT_CONFIG };

  // Layer 2: global config
  const globalCfg = readJsonFile(GLOBAL_CONFIG_PATH);
  if (globalCfg) {
    config = deepMerge(config, globalCfg);
  }

  // Layer 3: project config
  const projectDir = cwd ?? process.cwd();
  const projectCfg = readJsonFile(projectConfigPath(projectDir));
  if (projectCfg) {
    config = deepMerge(config, projectCfg);
  }

  // Layer 4: environment variables (highest priority)
  const envProvider = process.env.AIDEV_PROVIDER;
  if (envProvider && VALID_PROVIDERS.includes(envProvider)) {
    config.provider = envProvider;
  }

  const envApiKey = process.env.AIDEV_API_KEY;
  if (envApiKey) {
    config.apiKey = envApiKey;
  }

  // Provider-specific env vars
  if (!config.apiKey) {
    const providerSpecificKey =
      process.env.ANTHROPIC_API_KEY ||
      process.env.OPENAI_API_KEY ||
      process.env.DEEPSEEK_API_KEY ||
      process.env.GOOGLE_API_KEY;
    if (providerSpecificKey) {
      config.apiKey = providerSpecificKey;
    }
  }

  const envModel = process.env.AIDEV_MODEL;
  if (envModel) {
    config.model = envModel;
  }

  const envBaseUrl = process.env.AIDEV_BASE_URL;
  if (envBaseUrl) {
    config.baseUrl = envBaseUrl;
  }

  const envTemp = process.env.AIDEV_TEMPERATURE;
  if (envTemp) {
    const parsed = parseFloat(envTemp);
    if (!isNaN(parsed)) config.temperature = parsed;
  }

  const envMaxTokens = process.env.AIDEV_MAX_TOKENS;
  if (envMaxTokens) {
    const parsed = parseInt(envMaxTokens, 10);
    if (!isNaN(parsed)) config.maxTokens = parsed;
  }

  const envTheme = process.env.AIDEV_THEME;
  if (envTheme && VALID_THEMES.includes(envTheme)) {
    config.theme = envTheme;
  }

  // Final validation
  const errors = validateConfig(config as PartialConfig);
  if (errors.length > 0) {
    throw new Error(`Configuration errors:\n${errors.map((e) => `  - ${e}`).join('\n')}`);
  }

  return config as AppConfig;
}

/** Get the global config directory path. */
export function getGlobalDir(): string {
  return GLOBAL_DIR;
}

/** Get the global skills directory path. */
export function getGlobalSkillsDir(): string {
  return path.join(GLOBAL_DIR, 'skills');
}

/** Get the project-level config directory path. */
export function getProjectConfigDir(cwd?: string): string {
  return path.join(cwd ?? process.cwd(), '.aidev');
}

/** Get the project-level skills directory path. */
export function getProjectSkillsDir(cwd?: string): string {
  return path.join(cwd ?? process.cwd(), '.aidev', 'skills');
}
