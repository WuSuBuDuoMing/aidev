/**
 * aidev CLI — Context Analyzer Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { analyzeProject, contextSummary } from '../context/analyzer';
import { ProjectContext } from '../types';

describe('contextSummary', () => {
  const mockContext: ProjectContext = {
    files: ['src/index.ts', 'src/utils.ts', 'package.json'],
    structure: '├── src/\n│   ├── index.ts\n│   └── utils.ts\n└── package.json',
    gitInfo: {
      branch: 'main',
      lastCommit: 'abc1234 feat: add feature',
      recentCommits: ['abc1234 feat: add feature'],
      isDirty: true,
      stagedFiles: ['src/index.ts'],
      unstagedFiles: [],
      untrackedFiles: [],
    },
    language: 'TypeScript',
    framework: 'Node.js',
    packageJson: { name: 'test-project', version: '1.0.0' },
    manifestFile: 'package.json',
    manifestContent: '{"name":"test-project"}',
  };

  it('includes language and framework', () => {
    const summary = contextSummary(mockContext);
    expect(summary).toContain('Language: TypeScript');
    expect(summary).toContain('Framework: Node.js');
  });

  it('includes git info when dirty', () => {
    const summary = contextSummary(mockContext);
    expect(summary).toContain('Branch: main');
    expect(summary).toContain('dirty');
  });

  it('includes file count', () => {
    const summary = contextSummary(mockContext);
    expect(summary).toContain('Total files: 3');
  });

  it('includes directory structure', () => {
    const summary = contextSummary(mockContext);
    expect(summary).toContain('src/');
  });

  it('handles null git info', () => {
    const ctx = { ...mockContext, gitInfo: null };
    const summary = contextSummary(ctx);
    expect(summary).not.toContain('Branch');
  });
});

describe('analyzeProject', () => {
  it('returns a ProjectContext object', () => {
    const ctx = analyzeProject();
    expect(ctx).toBeDefined();
    expect(typeof ctx.language).toBe('string');
    expect(typeof ctx.framework).toBe('string');
    expect(Array.isArray(ctx.files)).toBe(true);
    expect(typeof ctx.structure).toBe('string');
  });

  it('detects language from existing project files', () => {
    const ctx = analyzeProject();
    // The aidev project itself should be detected as TypeScript
    expect(['TypeScript', 'JavaScript', 'JavaScript/TypeScript', 'Unknown']).toContain(ctx.language);
  });
});
