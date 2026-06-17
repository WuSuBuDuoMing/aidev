/**
 * aidev CLI — SQLite Database
 *
 * Initializes and manages the SQLite database for session persistence.
 * Uses better-sqlite3 for synchronous, zero-dependency SQLite access.
 *
 * @module storage/database
 */

import Database from 'better-sqlite3';
import * as path from 'path';
import * as os from 'os';
import * as fs from 'fs';

const DB_DIR = path.join(os.homedir(), '.aidev');
const DB_PATH = path.join(DB_DIR, 'history.db');

let db: Database.Database | null = null;

/**
 * Get or create the database connection.
 * Initializes tables on first call.
 */
export function getDatabase(): Database.Database {
  if (db) return db;

  // Ensure directory exists
  if (!fs.existsSync(DB_DIR)) {
    fs.mkdirSync(DB_DIR, { recursive: true });
  }

  db = new Database(DB_PATH);

  // Performance pragmas (inspired by MiMo-Code)
  db.pragma('journal_mode = WAL');
  db.pragma('synchronous = NORMAL');
  db.pragma('busy_timeout = 5000');
  db.pragma('foreign_keys = ON');

  // Initialize schema
  db.exec(`
    CREATE TABLE IF NOT EXISTS sessions (
      id TEXT PRIMARY KEY,
      title TEXT NOT NULL DEFAULT 'New Conversation',
      provider TEXT,
      model TEXT,
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now'))
    );

    CREATE TABLE IF NOT EXISTS messages (
      id TEXT PRIMARY KEY,
      session_id TEXT NOT NULL,
      role TEXT NOT NULL,
      content TEXT,
      tool_calls TEXT,
      tool_call_id TEXT,
      tool_name TEXT,
      timestamp INTEGER NOT NULL,
      FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
  `);

  return db;
}

/** Close the database connection. */
export function closeDatabase(): void {
  if (db) {
    db.close();
    db = null;
  }
}
