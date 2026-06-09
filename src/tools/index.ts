/**
 * aidev CLI — Tool System
 *
 * Provides a registry of tools the AI can invoke to interact with the user's
 * project: reading/writing files, running commands, searching code, etc.
 *
 * Every tool follows the same interface and respects the permission system.
 *
 * @module tools
 */

import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';
import { glob } from 'glob';
import * as Diff from 'diff';
import {
  Tool,
  ToolParameters,
  ToolResult,
  Permission,
  PermissionMode,
} from '../types';

// ---------------------------------------------------------------------------
// Permission Engine
// ---------------------------------------------------------------------------

/**
 * Decide whether a tool invocation is allowed.
 *
 * @param toolName  - Name of the tool being invoked.
 * @param permissions - Array of per-tool permission rules.
 * @param defaultMode - Fallback when no rule matches.
 * @param askUser    - Async callback invoked when the mode is "ask".
 * @returns `true` if the invocation should proceed, `false` to skip.
 */
export async function checkPermission(
  toolName: string,
  permissions: Permission[],
  defaultMode: PermissionMode,
  askUser?: (toolName: string) => Promise<boolean>,
): Promise<boolean> {
  const rule = permissions.find((p) => p.toolName === toolName);
  const mode = rule?.mode ?? defaultMode;

  switch (mode) {
    case 'allow':
      return true;
    case 'deny':
      return false;
    case 'ask':
      return askUser ? askUser(toolName) : false;
    default:
      return false;
  }
}

// ---------------------------------------------------------------------------
// Built-in tools
// ---------------------------------------------------------------------------

function createReadFileTool(): Tool {
  return {
    name: 'readFile',
    description: 'Read the contents of a file at the given path.',
    parameters: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'Absolute or relative path to the file.' },
        startLine: { type: 'number', description: 'Optional: start line (1-based, inclusive).' },
        endLine: { type: 'number', description: 'Optional: end line (1-based, inclusive).' },
      },
      required: ['path'],
    } as ToolParameters,
    async execute(args) {
      try {
        const filePath = path.resolve(String(args.path));
        if (!fs.existsSync(filePath)) {
          return { success: false, output: '', error: `File not found: ${filePath}` };
        }
        const content = fs.readFileSync(filePath, 'utf-8');
        const lines = content.split('\n');
        const start = args.startLine ? Math.max(1, Number(args.startLine)) : 1;
        const end = args.endLine ? Math.min(lines.length, Number(args.endLine)) : lines.length;
        const selected = lines.slice(start - 1, end);
        const numbered = selected.map((l, i) => `${start + i}\t${l}`).join('\n');
        return { success: true, output: numbered };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createWriteFileTool(): Tool {
  return {
    name: 'writeFile',
    description: 'Write content to a file. Creates parent directories if needed.',
    parameters: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'File path.' },
        content: { type: 'string', description: 'Content to write.' },
      },
      required: ['path', 'content'],
    } as ToolParameters,
    async execute(args) {
      try {
        const filePath = path.resolve(String(args.path));
        fs.mkdirSync(path.dirname(filePath), { recursive: true });
        fs.writeFileSync(filePath, String(args.content), 'utf-8');
        return { success: true, output: `Wrote ${filePath}` };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createEditFileTool(): Tool {
  return {
    name: 'editFile',
    description:
      'Apply a targeted edit to a file by replacing `oldText` with `newText`. ' +
      'The oldText must match exactly. Shows a diff preview before applying.',
    parameters: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'File path.' },
        oldText: { type: 'string', description: 'Exact text to find and replace.' },
        newText: { type: 'string', description: 'Replacement text.' },
      },
      required: ['path', 'oldText', 'newText'],
    } as ToolParameters,
    async execute(args) {
      try {
        const filePath = path.resolve(String(args.path));
        if (!fs.existsSync(filePath)) {
          return { success: false, output: '', error: `File not found: ${filePath}` };
        }
        const content = fs.readFileSync(filePath, 'utf-8');
        const oldText = String(args.oldText);
        const newText = String(args.newText);

        if (!content.includes(oldText)) {
          return { success: false, output: '', error: 'oldText not found in file.' };
        }

        const occurrences = content.split(oldText).length - 1;
        if (occurrences > 1) {
          return {
            success: false,
            output: '',
            error: `oldText matches ${occurrences} times. Please provide a more specific snippet.`,
          };
        }

        const newContent = content.replace(oldText, newText);
        const diff = Diff.createPatch(filePath, content, newContent);
        fs.writeFileSync(filePath, newContent, 'utf-8');

        return { success: true, output: `Edited ${filePath}\n\nDiff:\n${diff}` };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createSearchFilesTool(): Tool {
  return {
    name: 'searchFiles',
    description: 'Search for a pattern across files in the project using regex.',
    parameters: {
      type: 'object',
      properties: {
        pattern: { type: 'string', description: 'Regex pattern to search for.' },
        glob: { type: 'string', description: 'Glob filter (e.g. "**/*.ts"). Default: all files.' },
        path: { type: 'string', description: 'Root directory. Default: cwd.' },
      },
      required: ['pattern'],
    } as ToolParameters,
    async execute(args) {
      try {
        const root = path.resolve(args.path ? String(args.path) : process.cwd());
        const globPattern = args.glob ? String(args.glob) : '**/*';
        const files = await glob(globPattern, { cwd: root, nodir: true, ignore: ['node_modules/**', '.git/**'] });
        const regex = new RegExp(String(args.pattern), 'g');
        const matches: string[] = [];

        for (const file of files) {
          const fullPath = path.join(root, file);
          try {
            const content = fs.readFileSync(fullPath, 'utf-8');
            const lines = content.split('\n');
            for (let i = 0; i < lines.length; i++) {
              if (regex.test(lines[i])) {
                matches.push(`${file}:${i + 1}: ${lines[i].trim()}`);
                regex.lastIndex = 0; // reset for global flag
              }
            }
          } catch { /* skip binary / unreadable */ }
          if (matches.length >= 200) break;
        }

        return { success: true, output: matches.length > 0 ? matches.join('\n') : 'No matches found.' };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createRunCommandTool(): Tool {
  return {
    name: 'runCommand',
    description: 'Execute a shell command and return its output.',
    parameters: {
      type: 'object',
      properties: {
        command: { type: 'string', description: 'The shell command to run.' },
        cwd: { type: 'string', description: 'Working directory. Default: process.cwd().' },
        timeout: { type: 'number', description: 'Timeout in milliseconds. Default: 30000.' },
      },
      required: ['command'],
    } as ToolParameters,
    async execute(args) {
      try {
        const cwd = args.cwd ? path.resolve(String(args.cwd)) : process.cwd();
        const timeout = args.timeout ? Number(args.timeout) : 30_000;
        const output = execSync(String(args.command), {
          cwd,
          timeout,
          encoding: 'utf-8',
          stdio: ['pipe', 'pipe', 'pipe'],
          maxBuffer: 1024 * 1024 * 5, // 5 MB
        });
        return { success: true, output: output.trim() };
      } catch (err: unknown) {
        const e = err as { status?: number; stderr?: string; stdout?: string; message?: string };
        return {
          success: false,
          output: e.stdout?.trim() ?? '',
          error: `Exit code ${e.status ?? 'unknown'}: ${e.stderr?.trim() ?? e.message ?? String(err)}`,
        };
      }
    },
  };
}

function createGitStatusTool(): Tool {
  return {
    name: 'gitStatus',
    description: 'Show the current git status of the repository.',
    parameters: { type: 'object', properties: {} } as ToolParameters,
    async execute() {
      try {
        const output = execSync('git status --short --branch', {
          encoding: 'utf-8',
          stdio: ['pipe', 'pipe', 'pipe'],
        }).trim();
        return { success: true, output: output || 'Working tree clean.' };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createGitDiffTool(): Tool {
  return {
    name: 'gitDiff',
    description: 'Show git diff. Optionally diff staged changes or a specific file.',
    parameters: {
      type: 'object',
      properties: {
        staged: { type: 'boolean', description: 'If true, show staged changes.' },
        file: { type: 'string', description: 'Optional file path to diff.' },
      },
    } as ToolParameters,
    async execute(args) {
      try {
        let cmd = 'git diff';
        if (args.staged) cmd += ' --staged';
        if (args.file) cmd += ` -- ${String(args.file)}`;
        const output = execSync(cmd, { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] }).trim();
        return { success: true, output: output || 'No changes.' };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createGitCommitTool(): Tool {
  return {
    name: 'gitCommit',
    description: 'Create a git commit with the given message. Stages all changes first.',
    parameters: {
      type: 'object',
      properties: {
        message: { type: 'string', description: 'Commit message.' },
        all: { type: 'boolean', description: 'If true, run git add -A before committing.' },
      },
      required: ['message'],
    } as ToolParameters,
    async execute(args) {
      try {
        if (args.all) {
          execSync('git add -A', { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] });
        }
        const msg = String(args.message).replace(/'/g, "'\\''");
        execSync(`git commit -m '${msg}'`, { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] });
        return { success: true, output: 'Commit created.' };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

function createListDirTool(): Tool {
  return {
    name: 'listDir',
    description: 'List files and directories at the given path.',
    parameters: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'Directory path. Default: cwd.' },
        depth: { type: 'number', description: 'Recursion depth. Default: 1.' },
      },
    } as ToolParameters,
    async execute(args) {
      try {
        const dir = path.resolve(args.path ? String(args.path) : process.cwd());
        const depth = args.depth ? Number(args.depth) : 1;

        function listRecursive(d: string, currentDepth: number): string[] {
          if (currentDepth > depth) return [];
          let entries: fs.Dirent[];
          try {
            entries = fs.readdirSync(d, { withFileTypes: true });
          } catch {
            return [];
          }
          const results: string[] = [];
          const indent = '  '.repeat(currentDepth - 1);
          for (const entry of entries) {
            if (entry.name === 'node_modules' || entry.name === '.git') continue;
            if (entry.isDirectory()) {
              results.push(`${indent}${entry.name}/`);
              results.push(...listRecursive(path.join(d, entry.name), currentDepth + 1));
            } else {
              results.push(`${indent}${entry.name}`);
            }
          }
          return results;
        }

        const output = listRecursive(dir, 1).join('\n');
        return { success: true, output: output || '(empty directory)' };
      } catch (err) {
        return { success: false, output: '', error: String(err) };
      }
    },
  };
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

/** Tool registry — stores and provides access to all registered tools. */
export class ToolRegistry {
  private tools: Map<string, Tool> = new Map();

  /** Register a tool. Overwrites if already registered. */
  register(tool: Tool): void {
    this.tools.set(tool.name, tool);
  }

  /** Get a tool by name. */
  get(name: string): Tool | undefined {
    return this.tools.get(name);
  }

  /** Get all registered tools as an array. */
  getAll(): Tool[] {
    return Array.from(this.tools.values());
  }

  /** Get tool definitions formatted for AI provider tool-use APIs. */
  getDefinitions(): { name: string; description: string; parameters: ToolParameters }[] {
    return this.getAll().map((t) => ({
      name: t.name,
      description: t.description,
      parameters: t.parameters,
    }));
  }

  /** Execute a tool by name with permission checking. */
  async execute(
    name: string,
    args: Record<string, unknown>,
    permissions: Permission[],
    defaultPermission: PermissionMode,
    askUser?: (toolName: string) => Promise<boolean>,
  ): Promise<ToolResult> {
    const tool = this.tools.get(name);
    if (!tool) {
      return { success: false, output: '', error: `Unknown tool: ${name}` };
    }

    const allowed = await checkPermission(name, permissions, defaultPermission, askUser);
    if (!allowed) {
      return { success: false, output: '', error: `Permission denied for tool: ${name}` };
    }

    try {
      return await tool.execute(args);
    } catch (err) {
      return { success: false, output: '', error: `Tool execution error: ${String(err)}` };
    }
  }
}

/**
 * Create a registry pre-loaded with all built-in tools.
 */
export function createBuiltinRegistry(): ToolRegistry {
  const registry = new ToolRegistry();
  registry.register(createReadFileTool());
  registry.register(createWriteFileTool());
  registry.register(createEditFileTool());
  registry.register(createSearchFilesTool());
  registry.register(createRunCommandTool());
  registry.register(createGitStatusTool());
  registry.register(createGitDiffTool());
  registry.register(createGitCommitTool());
  registry.register(createListDirTool());
  return registry;
}
