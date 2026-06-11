import { describe, it, expect } from 'vitest';
import { parseSkillContent, resolveSkill } from '../skills';
import type { Skill } from '../types';

describe('parseSkillContent', () => {
  it('parses a valid skill file', () => {
    const raw = `---
name: test-skill
description: A test skill
trigger: /test
---
This is the prompt body.`;
    const skill = parseSkillContent(raw);
    expect(skill).not.toBeNull();
    expect(skill!.name).toBe('test-skill');
    expect(skill!.description).toBe('A test skill');
    expect(skill!.trigger).toBe('/test');
    expect(skill!.prompt).toBe('This is the prompt body.');
  });

  it('uses default trigger when not specified', () => {
    const raw = `---
name: my-skill
description: desc
---
Body text.`;
    const skill = parseSkillContent(raw);
    expect(skill!.trigger).toBe('/my-skill');
  });

  it('returns null for missing frontmatter', () => {
    expect(parseSkillContent('No frontmatter here')).toBeNull();
  });

  it('returns null for missing name', () => {
    const raw = `---
description: no name
---
Body`;
    expect(parseSkillContent(raw)).toBeNull();
  });

  it('returns null for empty body', () => {
    const raw = `---
name: empty
description: desc
---`;
    expect(parseSkillContent(raw)).toBeNull();
  });

  it('handles Windows line endings in delimiters', () => {
    const raw = '---\r\nname: win-skill\r\ndescription: desc\r\n---\r\nBody.';
    const skill = parseSkillContent(raw);
    expect(skill).not.toBeNull();
    expect(skill!.name).toBe('win-skill');
  });

  it('handles multiline prompt body', () => {
    const raw = `---
name: multi
description: desc
---
Line 1

Line 3

- bullet`;
    const skill = parseSkillContent(raw);
    expect(skill!.prompt).toContain('Line 1');
    expect(skill!.prompt).toContain('Line 3');
    expect(skill!.prompt).toContain('- bullet');
  });
});

describe('resolveSkill', () => {
  const skills: Skill[] = [
    { name: 'review', description: 'Code review', trigger: '/review', prompt: 'review prompt' },
    { name: 'test-gen', description: 'Test generation', trigger: '/test', prompt: 'test prompt' },
    { name: 'explain', description: 'Explain code', trigger: '/explain', prompt: 'explain prompt' },
  ];

  it('resolves by trigger with slash', () => {
    expect(resolveSkill(skills, '/review')!.name).toBe('review');
  });

  it('resolves by trigger without slash prefix', () => {
    expect(resolveSkill(skills, 'review')!.name).toBe('review');
  });

  it('resolves by name', () => {
    expect(resolveSkill(skills, 'test-gen')!.name).toBe('test-gen');
  });

  it('returns undefined for unknown skill', () => {
    expect(resolveSkill(skills, 'nonexistent')).toBeUndefined();
  });

  it('prefers trigger match over name match', () => {
    const ambiguous: Skill[] = [
      { name: 'foo', description: '', trigger: '/bar', prompt: 'p1' },
      { name: 'bar', description: '', trigger: '/baz', prompt: 'p2' },
    ];
    expect(resolveSkill(ambiguous, '/bar')!.name).toBe('foo');
  });
});
