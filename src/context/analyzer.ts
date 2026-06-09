/**
 * aidev CLI — Project Context Analyzer
 *
 * Scans the current project to build a structured context that the AI can
 * use to give more accurate, project-aware answers.
 *
 * @module context/analyzer
 */

import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';
import { ProjectContext, GitInfo } from '../types';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** Max depth when building directory tree. */
const MAX_TREE_DEPTH = 3;

/** Max number of files to list in the context. */
const MAX_FILES = 500;

/** Files and directories to always ignore (supplements .gitignore). */
const IGNORED_PATTERNS = new Set([
  'node_modules',
  '.git',
  'dist',
  'build',
  '.next',
  '__pycache__',
  '.venv',
  'venv',
  '.tox',
  'target',
  'vendor',
  '.idea',
  '.vscode',
  '.DS_Store',
]);

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Safe execSync wrapper — returns stdout or empty string on error. */
function exec(cmd: string, cwd: string): string {
  try {
    return execSync(cmd, { cwd, timeout: 10_000, encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] }).trim();
  } catch {
    return '';
  }
}

/** Load .gitignore patterns (simple implementation — covers 90% of cases). */
function loadGitignorePatterns(cwd: string): Set<string> {
  const patterns = new Set<string>();
  const gitignorePath = path.join(cwd, '.gitignore');
  if (!fs.existsSync(gitignorePath)) return patterns;

  const lines = fs.readFileSync(gitignorePath, 'utf-8').split('\n');
  for (const raw of lines) {
    const line = raw.trim();
    if (!line || line.startsWith('#')) continue;
    // Normalize: strip trailing slash, leading slash
    const normalized = line.replace(/\/$/, '').replace(/^\//, '');
    if (normalized) patterns.add(normalized);
  }
  return patterns;
}

/** Check if a file/dir name should be ignored. */
function shouldIgnore(name: string, gitignorePatterns: Set<string>): boolean {
  if (IGNORED_PATTERNS.has(name)) return true;
  if (name.startsWith('.') && name !== '.') return true;
  if (gitignorePatterns.has(name)) return true;
  return false;
}

// ---------------------------------------------------------------------------
// Detection
// ---------------------------------------------------------------------------

/** Detect the primary programming language of the project. */
function detectLanguage(cwd: string): string {
  const indicators: [string, string][] = [
    ['package.json', 'JavaScript/TypeScript'],
    ['tsconfig.json', 'TypeScript'],
    ['requirements.txt', 'Python'],
    ['pyproject.toml', 'Python'],
    ['setup.py', 'Python'],
    ['go.mod', 'Go'],
    ['Cargo.toml', 'Rust'],
    ['pom.xml', 'Java'],
    ['build.gradle', 'Java'],
    ['build.gradle.kts', 'Kotlin'],
    ['Gemfile', 'Ruby'],
    ['composer.json', 'PHP'],
    ['CMakeLists.txt', 'C/C++'],
    ['Makefile', 'C/C++'],
    ['mix.exs', 'Elixir'],
    ['pubspec.yaml', 'Dart'],
  ];

  for (const [file, lang] of indicators) {
    if (fs.existsSync(path.join(cwd, file))) return lang;
  }
  return 'Unknown';
}

/** Detect the framework being used. */
function detectFramework(cwd: string): string {
  const pkgPath = path.join(cwd, 'package.json');
  if (fs.existsSync(pkgPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf-8'));
      const deps = { ...pkg.dependencies, ...pkg.devDependencies };
      if (deps.next) return 'Next.js';
      if (deps.nuxt) return 'Nuxt';
      if (deps.svelte) return 'Svelte';
      if (deps.vue) return 'Vue';
      if (deps.angular) return 'Angular';
      if (deps.react) return 'React';
      if (deps.express) return 'Express';
      if (deps.fastify) return 'Fastify';
      if (deps.hono) return 'Hono';
      if (deps.astro) return 'Astro';
      if (deps.remix) return 'Remix';
    } catch { /* malformed package.json */ }
  }

  // Python
  const reqPath = path.join(cwd, 'requirements.txt');
  if (fs.existsSync(reqPath)) {
    const content = fs.readFileSync(reqPath, 'utf-8').toLowerCase();
    if (content.includes('django')) return 'Django';
    if (content.includes('flask')) return 'Flask';
    if (content.includes('fastapi')) return 'FastAPI';
  }

  // Go
  const goModPath = path.join(cwd, 'go.mod');
  if (fs.existsSync(goModPath)) {
    const content = fs.readFileSync(goModPath, 'utf-8');
    if (content.includes('gin-gonic')) return 'Gin';
    if (content.includes('labstack/echo')) return 'Echo';
    if (content.includes('gofiber')) return 'Fiber';
  }

  return 'None detected';
}

/** Read the manifest file for the detected language. */
function readManifest(cwd: string): { file: string | null; content: string | null } {
  const candidates = [
    'package.json',
    'requirements.txt',
    'pyproject.toml',
    'go.mod',
    'Cargo.toml',
    'pom.xml',
    'Gemfile',
    'composer.json',
  ];
  for (const name of candidates) {
    const filePath = path.join(cwd, name);
    if (fs.existsSync(filePath)) {
      try {
        return { file: name, content: fs.readFileSync(filePath, 'utf-8') };
      } catch { /* unreadable */ }
    }
  }
  return { file: null, content: null };
}

// ---------------------------------------------------------------------------
// Git
// ---------------------------------------------------------------------------

/** Gather git information for the project. */
function gatherGitInfo(cwd: string): GitInfo | null {
  const isRepo = exec('git rev-parse --is-inside-work-tree', cwd);
  if (isRepo !== 'true') return null;

  const branch = exec('git branch --show-current', cwd) || exec('git rev-parse --abbrev-ref HEAD', cwd);
  const lastCommit = exec('git log -1 --oneline', cwd);
  const recentCommitsRaw = exec('git log --oneline -10', cwd);
  const recentCommits = recentCommitsRaw ? recentCommitsRaw.split('\n') : [];
  const isDirty = exec('git status --porcelain', cwd).length > 0;

  const statusRaw = exec('git status --porcelain', cwd);
  const stagedFiles: string[] = [];
  const unstagedFiles: string[] = [];
  const untrackedFiles: string[] = [];

  for (const line of statusRaw.split('\n')) {
    if (!line.trim()) continue;
    const indexStatus = line[0];
    const workTreeStatus = line[1];
    const fileName = line.slice(3).trim();

    if (indexStatus === '?' && workTreeStatus === '?') {
      untrackedFiles.push(fileName);
    } else {
      if (indexStatus !== ' ' && indexStatus !== '?') stagedFiles.push(fileName);
      if (workTreeStatus !== ' ' && workTreeStatus !== '?') unstagedFiles.push(fileName);
    }
  }

  return { branch, lastCommit, recentCommits, isDirty, stagedFiles, unstagedFiles, untrackedFiles };
}

// ---------------------------------------------------------------------------
// File scanning
// ---------------------------------------------------------------------------

/** Recursively collect file paths, respecting ignore rules. */
function collectFiles(dir: string, gitignorePatterns: Set<string>, depth = 0): string[] {
  if (depth > MAX_TREE_DEPTH) return [];

  const files: string[] = [];
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return [];
  }

  for (const entry of entries) {
    if (shouldIgnore(entry.name, gitignorePatterns)) continue;
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...collectFiles(fullPath, gitignorePatterns, depth + 1));
    } else {
      files.push(fullPath);
    }
    if (files.length >= MAX_FILES) break;
  }
  return files;
}

/** Build a human-readable directory tree string. */
function buildTree(dir: string, gitignorePatterns: Set<string>, prefix = '', depth = 0): string {
  if (depth > MAX_TREE_DEPTH) return '';

  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return '';
  }

  entries = entries.filter((e) => !shouldIgnore(e.name, gitignorePatterns));
  entries.sort((a, b) => {
    if (a.isDirectory() !== b.isDirectory()) return a.isDirectory() ? -1 : 1;
    return a.name.localeCompare(b.name);
  });

  let tree = '';
  for (let i = 0; i < entries.length; i++) {
    const entry = entries[i];
    const isLast = i === entries.length - 1;
    const connector = isLast ? '└── ' : '├── ';
    const nextPrefix = prefix + (isLast ? '    ' : '│   ');

    if (entry.isDirectory()) {
      tree += `${prefix}${connector}${entry.name}/\n`;
      tree += buildTree(path.join(dir, entry.name), gitignorePatterns, nextPrefix, depth + 1);
    } else {
      tree += `${prefix}${connector}${entry.name}\n`;
    }
  }
  return tree;
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Analyze the current project directory and return structured context.
 *
 * @param cwd - The project root directory. Defaults to `process.cwd()`.
 * @returns A `ProjectContext` object with files, structure, git info, etc.
 */
export function analyzeProject(cwd?: string): ProjectContext {
  const dir = cwd ?? process.cwd();
  const gitignorePatterns = loadGitignorePatterns(dir);

  const language = detectLanguage(dir);
  const framework = detectFramework(dir);
  const { file: manifestFile, content: manifestContent } = readManifest(dir);

  let packageJson: Record<string, unknown> | null = null;
  const pkgPath = path.join(dir, 'package.json');
  if (fs.existsSync(pkgPath)) {
    try {
      packageJson = JSON.parse(fs.readFileSync(pkgPath, 'utf-8')) as Record<string, unknown>;
    } catch { /* ignore */ }
  }

  const allFiles = collectFiles(dir, gitignorePatterns);
  const relativeFiles = allFiles.map((f) => path.relative(dir, f).replace(/\\/g, '/'));
  const structure = buildTree(dir, gitignorePatterns);
  const gitInfo = gatherGitInfo(dir);

  return {
    files: relativeFiles,
    structure,
    gitInfo,
    language,
    framework,
    packageJson,
    manifestFile,
    manifestContent,
  };
}

/**
 * Build a concise text summary of the project context suitable for injection
 * into a system prompt.
 */
export function contextSummary(ctx: ProjectContext): string {
  const parts: string[] = [];
  parts.push(`Language: ${ctx.language}`);
  parts.push(`Framework: ${ctx.framework}`);
  if (ctx.gitInfo) {
    parts.push(`Branch: ${ctx.gitInfo.branch}`);
    parts.push(`Last commit: ${ctx.gitInfo.lastCommit}`);
    if (ctx.gitInfo.isDirty) parts.push('Working tree: dirty');
  }
  parts.push(`Total files: ${ctx.files.length}`);
  if (ctx.structure) {
    parts.push(`\nDirectory tree:\n${ctx.structure}`);
  }
  return parts.join('\n');
}
