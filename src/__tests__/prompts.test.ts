import { describe, it, expect } from 'vitest';
import {
  codingAssistantPrompt,
  codeReviewPrompt,
  refactorPrompt,
  explainPrompt,
  testGenerationPrompt,
  commitMessagePrompt,
  prDescriptionPrompt,
  getSpecializedPrompts,
} from '../ai/prompts';
import type { ProjectContext } from '../types';

describe('codingAssistantPrompt', () => {
  it('returns a non-empty string', () => {
    expect(codingAssistantPrompt().length).toBeGreaterThan(100);
  });

  it('includes core principles', () => {
    const prompt = codingAssistantPrompt();
    expect(prompt).toContain('Core Principles');
    expect(prompt).toContain('Capabilities');
    expect(prompt).toContain('Safety');
  });

  it('includes project context when provided', () => {
    const project: ProjectContext = {
      files: ['src/index.ts'],
      structure: 'src/\n  index.ts',
      gitInfo: {
        branch: 'main',
        lastCommit: 'abc123',
        recentCommits: ['abc123 first commit'],
        isDirty: true,
        stagedFiles: [],
        unstagedFiles: ['src/index.ts'],
        untrackedFiles: [],
      },
      language: 'TypeScript',
      framework: 'Node.js',
      packageJson: null,
      manifestFile: 'package.json',
      manifestContent: '{}',
    };
    const prompt = codingAssistantPrompt(project);
    expect(prompt).toContain('TypeScript');
    expect(prompt).toContain('Node.js');
    expect(prompt).toContain('main');
    expect(prompt).toContain('dirty');
  });

  it('omits project section when no project given', () => {
    const prompt = codingAssistantPrompt();
    expect(prompt).not.toContain('Project Context');
  });
});

describe('specialized prompts', () => {
  it('codeReviewPrompt mentions severity levels', () => {
    expect(codeReviewPrompt()).toContain('critical');
    expect(codeReviewPrompt()).toContain('APPROVE');
  });

  it('refactorPrompt mentions code smells', () => {
    expect(refactorPrompt()).toContain('code smells');
  });

  it('explainPrompt mentions step by step', () => {
    expect(explainPrompt()).toContain('step by step');
  });

  it('testGenerationPrompt mentions AAA pattern', () => {
    expect(testGenerationPrompt()).toContain('AAA');
  });

  it('commitMessagePrompt mentions Conventional Commits', () => {
    expect(commitMessagePrompt()).toContain('Conventional Commits');
  });

  it('prDescriptionPrompt mentions Summary', () => {
    expect(prDescriptionPrompt()).toContain('Summary');
  });
});

describe('getSpecializedPrompts', () => {
  it('returns a map with all 6 commands', () => {
    const map = getSpecializedPrompts();
    expect(Object.keys(map)).toEqual(
      expect.arrayContaining(['review', 'refactor', 'explain', 'test', 'commit', 'pr']),
    );
  });

  it('each value is a function returning a string', () => {
    const map = getSpecializedPrompts();
    for (const fn of Object.values(map)) {
      const result = fn();
      expect(typeof result).toBe('string');
      expect(result.length).toBeGreaterThan(50);
    }
  });
});
