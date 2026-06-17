/**
 * aidev CLI — Session Persistence
 *
 * Saves and loads conversation sessions to/from SQLite.
 *
 * @module storage/session
 */

import { Conversation, Message } from '../types';
import { getDatabase } from './database';

/** Summary of a session (without messages) for listing. */
export interface SessionSummary {
  id: string;
  title: string;
  provider: string | null;
  model: string | null;
  created_at: string;
  updated_at: string;
  message_count: number;
}

/** Save or update a conversation in the database. */
export function saveSession(conversation: Conversation): void {
  const db = getDatabase();

  db.prepare(`
    INSERT INTO sessions (id, title, provider, model, created_at, updated_at)
    VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
    ON CONFLICT(id) DO UPDATE SET
      title = excluded.title,
      provider = excluded.provider,
      model = excluded.model,
      updated_at = datetime('now')
  `).run(
    conversation.id,
    conversation.title,
    conversation.provider,
    conversation.model,
  );

  // Delete existing messages for this session (full replace)
  db.prepare('DELETE FROM messages WHERE session_id = ?').run(conversation.id);

  // Insert all messages
  const insertMsg = db.prepare(`
    INSERT INTO messages (id, session_id, role, content, tool_calls, tool_call_id, tool_name, timestamp)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `);

  const insertMany = db.transaction((messages: Message[]) => {
    for (let i = 0; i < messages.length; i++) {
      const msg = messages[i];
      insertMsg.run(
        `${conversation.id}-msg-${i}`,
        conversation.id,
        msg.role,
        msg.content,
        msg.toolCalls ? JSON.stringify(msg.toolCalls) : null,
        msg.toolCallId ?? null,
        msg.toolName ?? null,
        msg.timestamp,
      );
    }
  });

  insertMany(conversation.messages);
}

/** Load a full conversation by ID. Returns null if not found. */
export function loadSession(sessionId: string): Conversation | null {
  const db = getDatabase();

  const session = db.prepare('SELECT * FROM sessions WHERE id = ?').get(sessionId) as {
    id: string; title: string; provider: string; model: string;
    created_at: string; updated_at: string;
  } | undefined;

  if (!session) return null;

  const rows = db.prepare(
    'SELECT * FROM messages WHERE session_id = ? ORDER BY timestamp ASC'
  ).all(sessionId) as {
    role: string; content: string; tool_calls: string | null;
    tool_call_id: string | null; tool_name: string | null; timestamp: number;
  }[];

  const messages: Message[] = rows.map((row) => ({
    role: row.role as Message['role'],
    content: row.content ?? '',
    timestamp: row.timestamp,
    ...(row.tool_calls ? { toolCalls: JSON.parse(row.tool_calls) } : {}),
    ...(row.tool_call_id ? { toolCallId: row.tool_call_id } : {}),
    ...(row.tool_name ? { toolName: row.tool_name } : {}),
  }));

  return {
    id: session.id,
    title: session.title,
    messages,
    createdAt: new Date(session.created_at).getTime(),
    updatedAt: new Date(session.updated_at).getTime(),
    model: session.model,
    provider: session.provider as Conversation['provider'],
  };
}

/** List recent sessions (most recent first). */
export function listSessions(limit = 20): SessionSummary[] {
  const db = getDatabase();

  return db.prepare(`
    SELECT s.*, COUNT(m.id) as message_count
    FROM sessions s
    LEFT JOIN messages m ON m.session_id = s.id
    GROUP BY s.id
    ORDER BY s.updated_at DESC
    LIMIT ?
  `).all(limit) as SessionSummary[];
}

/** Delete a session and its messages. */
export function deleteSession(sessionId: string): boolean {
  const db = getDatabase();
  const result = db.prepare('DELETE FROM sessions WHERE id = ?').run(sessionId);
  return result.changes > 0;
}
