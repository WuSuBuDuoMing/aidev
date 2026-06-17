/**
 * aidev CLI — Skills System
 *
 * Skills are reusable prompt templates stored as YAML-frontmatter files.
 * They can live in:
 *   - Project-level:  <project>/.aidev/skills/*.md
 *   - Global-level:   ~/.aidev/skills/*.md
 *
 * File format (example):
 * ```yaml
 * ---
 * name: security-audit
 * description: Perform a security audit of the codebase
 * trigger: /security
 * ---
 * You are a security engineer ...
 * ```
 *
 * @module skills
 */

import * as fs from 'fs';
import * as path from 'path';
import { Skill } from '../types';
import { getGlobalSkillsDir, getProjectSkillsDir } from '../config';

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

/**
 * Parse a skill file (YAML frontmatter + markdown body).
 *
 * @param filePath - Absolute path to the skill file.
 * @returns A `Skill` object, or `null` if parsing fails.
 */
export function parseSkillFile(filePath: string): Skill | null {
  try {
    const raw = fs.readFileSync(filePath, 'utf-8');
    return parseSkillContent(raw, filePath);
  } catch {
    return null;
  }
}

/**
 * Parse skill content from a raw string.
 */
export function parseSkillContent(raw: string, filePath?: string): Skill | null {
  // Normalize Windows line endings
  const normalized = raw.replace(/\r\n/g, '\n');
  // Must start with YAML frontmatter delimited by ---
  const match = normalized.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  if (!match) return null;

  const frontmatter = match[1];
  const prompt = match[2].trim();

  // Simple key-value YAML parser (no external dep needed for flat key-value)
  const meta: Record<string, string> = {};
  for (const line of frontmatter.split('\n')) {
    const kv = line.match(/^(\w[\w-]*)\s*:\s*(.+)$/);
    if (kv) {
      meta[kv[1].trim()] = kv[2].trim();
    }
  }

  if (!meta.name || !prompt) return null;

  return {
    name: meta.name,
    description: meta.description ?? '',
    trigger: meta.trigger ?? `/${meta.name}`,
    prompt,
  };
}

// ---------------------------------------------------------------------------
// Discovery
// ---------------------------------------------------------------------------

/** Scan a directory for skill files (.md) and return parsed skills. */
function loadSkillsFromDir(dir: string): Skill[] {
  if (!fs.existsSync(dir)) return [];

  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return [];
  }

  const skills: Skill[] = [];
  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith('.md')) continue;
    const skill = parseSkillFile(path.join(dir, entry.name));
    if (skill) skills.push(skill);
  }
  return skills;
}

/**
 * Discover and load all skills from both global and project-level directories.
 *
 * Project-level skills override global skills with the same name.
 *
 * @param cwd - Project root. Defaults to `process.cwd()`.
 */
export function loadAllSkills(cwd?: string): Skill[] {
  const globalSkills = loadSkillsFromDir(getGlobalSkillsDir());
  const projectSkills = loadSkillsFromDir(getProjectSkillsDir(cwd));

  // Merge: project overrides global by name
  const map = new Map<string, Skill>();
  for (const s of globalSkills) map.set(s.name, s);
  for (const s of projectSkills) map.set(s.name, s);

  return Array.from(map.values());
}

/**
 * Resolve a skill by its trigger command (e.g. "/security") or name.
 */
export function resolveSkill(skills: Skill[], input: string): Skill | undefined {
  const normalized = input.startsWith('/') ? input : `/${input}`;
  return (
    skills.find((s) => s.trigger === normalized) ??
    skills.find((s) => s.name === normalized.replace(/^\//, ''))
  );
}
