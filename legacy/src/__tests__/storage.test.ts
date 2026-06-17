/**
 * aidev CLI — Storage Session Tests
 */

import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

// Use a temp DB to avoid polluting the real one
const TMP_DB = path.join(os.tmpdir(), `aidev-test-${Date.now()}.db`);

// Override the DB path before importing
process.env.AIDEV_DB_PATH = TMP_DB;

import { getDatabase, closeDatabase } from '../storage/database';
import { saveSession, loadSession, listSessions, deleteSession } from '../storage/session';
import { Conversation, Message } from '../types';

function makeTestConversation(id: string, msgCount = 3): Conversation {
  const messages: Message[] = [];
  for (let i = 0; i < msgCount; i++) {
    messages.push({
      role: i % 2 === 0 ? 'user' : 'assistant',
      content: `Message ${i}: ${i % 2 === 0 ? 'Hello' : 'Hi there!'}`,
      timestamp: Date.now() + i,
    });
  }
  return {
    id,
    title: `Test ${id}`,
    messages,
    createdAt: Date.now(),
    updatedAt: Date.now(),
    model: 'claude-sonnet-4-20250514',
    provider: 'claude' as any,
  };
}

describe('Storage: Database', () => {
  it('creates database and tables', () => {
    const db = getDatabase();
    expect(db).toBeDefined();
    const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all() as { name: string }[];
    const names = tables.map((t) => t.name);
    expect(names).toContain('sessions');
    expect(names).toContain('messages');
  });
});

describe('Storage: Session CRUD', () => {
  const conv1 = makeTestConversation('test-001', 2);
  const conv2 = makeTestConversation('test-002', 5);

  it('saves a session', () => {
    saveSession(conv1);
    const loaded = loadSession('test-001');
    expect(loaded).not.toBeNull();
    expect(loaded!.id).toBe('test-001');
    expect(loaded!.title).toBe('Test test-001');
    expect(loaded!.messages).toHaveLength(2);
  });

  it('loads messages in order', () => {
    const loaded = loadSession('test-001');
    expect(loaded!.messages[0].role).toBe('user');
    expect(loaded!.messages[1].role).toBe('assistant');
    expect(loaded!.messages[0].content).toBe('Message 0: Hello');
  });

  it('upserts on re-save', () => {
    saveSession(conv1);
    saveSession({ ...conv1, title: 'Updated Title' });
    const loaded = loadSession('test-001');
    expect(loaded!.title).toBe('Updated Title');
  });

  it('lists sessions with message count', () => {
    saveSession(conv2);
    const sessions = listSessions(10);
    expect(sessions.length).toBeGreaterThanOrEqual(2);
    const found = sessions.find((s) => s.id === 'test-002');
    expect(found).toBeDefined();
    expect(found!.message_count).toBe(5);
  });

  it('deletes a session', () => {
    const deleted = deleteSession('test-002');
    expect(deleted).toBe(true);
    expect(loadSession('test-002')).toBeNull();
  });

  it('returns null for non-existent session', () => {
    expect(loadSession('non-existent')).toBeNull();
  });
});

afterAll(() => {
  closeDatabase();
  try { fs.unlinkSync(TMP_DB); } catch { /* ignore */ }
  try { fs.unlinkSync(TMP_DB + '-wal'); } catch { /* ignore */ }
  try { fs.unlinkSync(TMP_DB + '-shm'); } catch { /* ignore */ }
});
