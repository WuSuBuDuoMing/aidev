import { describe, it, expect } from 'vitest';
import { checkPermission, ToolRegistry, createBuiltinRegistry } from '../tools';
import type { Tool, Permission, ToolResult } from '../types';

function makeTool(name: string): Tool {
  return {
    name,
    description: `Tool ${name}`,
    parameters: { type: 'object', properties: {} },
    async execute(): Promise<ToolResult> {
      return { success: true, output: `executed ${name}` };
    },
  };
}

describe('checkPermission', () => {
  it('allows when mode is allow', async () => {
    const result = await checkPermission('test', [], 'allow');
    expect(result).toBe(true);
  });

  it('denies when mode is deny', async () => {
    const result = await checkPermission('test', [], 'deny');
    expect(result).toBe(false);
  });

  it('uses per-tool rule over default', async () => {
    const perms: Permission[] = [{ toolName: 'test', mode: 'deny' }];
    const result = await checkPermission('test', perms, 'allow');
    expect(result).toBe(false);
  });

  it('falls back to default when no rule matches', async () => {
    const perms: Permission[] = [{ toolName: 'other', mode: 'deny' }];
    const result = await checkPermission('test', perms, 'allow');
    expect(result).toBe(true);
  });

  it('calls askUser when mode is ask', async () => {
    const askUser = async (name: string) => name === 'test';
    const result = await checkPermission('test', [], 'ask', askUser);
    expect(result).toBe(true);
  });

  it('returns false for ask mode without askUser callback', async () => {
    const result = await checkPermission('test', [], 'ask');
    expect(result).toBe(false);
  });

  it('returns false for ask when user declines', async () => {
    const askUser = async () => false;
    const result = await checkPermission('test', [], 'ask', askUser);
    expect(result).toBe(false);
  });
});

describe('ToolRegistry', () => {
  it('registers and retrieves a tool', () => {
    const registry = new ToolRegistry();
    const tool = makeTool('myTool');
    registry.register(tool);
    expect(registry.get('myTool')).toBe(tool);
  });

  it('returns undefined for unknown tool', () => {
    const registry = new ToolRegistry();
    expect(registry.get('unknown')).toBeUndefined();
  });

  it('overwrites duplicate registration', () => {
    const registry = new ToolRegistry();
    const first = makeTool('dup');
    const second = makeTool('dup');
    registry.register(first);
    registry.register(second);
    expect(registry.get('dup')).toBe(second);
  });

  it('getAll returns all tools', () => {
    const registry = new ToolRegistry();
    registry.register(makeTool('a'));
    registry.register(makeTool('b'));
    expect(registry.getAll()).toHaveLength(2);
  });

  it('getDefinitions returns name/description/parameters', () => {
    const registry = new ToolRegistry();
    registry.register(makeTool('x'));
    const defs = registry.getDefinitions();
    expect(defs).toHaveLength(1);
    expect(defs[0].name).toBe('x');
    expect(defs[0].description).toBe('Tool x');
    expect(defs[0].parameters).toEqual({ type: 'object', properties: {} });
  });

  it('execute runs tool when allowed', async () => {
    const registry = new ToolRegistry();
    registry.register(makeTool('run'));
    const result = await registry.execute('run', {}, [], 'allow');
    expect(result.success).toBe(true);
    expect(result.output).toBe('executed run');
  });

  it('execute returns error for unknown tool', async () => {
    const registry = new ToolRegistry();
    const result = await registry.execute('missing', {}, [], 'allow');
    expect(result.success).toBe(false);
    expect(result.error).toContain('Unknown tool');
  });

  it('execute returns error when permission denied', async () => {
    const registry = new ToolRegistry();
    registry.register(makeTool('blocked'));
    const result = await registry.execute('blocked', {}, [], 'deny');
    expect(result.success).toBe(false);
    expect(result.error).toContain('Permission denied');
  });
});

describe('createBuiltinRegistry', () => {
  it('registers all 9 built-in tools', () => {
    const registry = createBuiltinRegistry();
    const tools = registry.getAll();
    expect(tools).toHaveLength(9);
    const names = tools.map((t) => t.name);
    expect(names).toContain('readFile');
    expect(names).toContain('writeFile');
    expect(names).toContain('editFile');
    expect(names).toContain('searchFiles');
    expect(names).toContain('runCommand');
    expect(names).toContain('gitStatus');
    expect(names).toContain('gitDiff');
    expect(names).toContain('gitCommit');
    expect(names).toContain('listDir');
  });

  it('each tool has a description and parameters', () => {
    const registry = createBuiltinRegistry();
    for (const tool of registry.getAll()) {
      expect(tool.description).toBeTruthy();
      expect(tool.parameters.type).toBe('object');
    }
  });
});
