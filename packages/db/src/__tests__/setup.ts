/**
 * Helpers compartidos para tests de packages/db.
 *
 * Patrón de uso:
 *   const { client, db } = createTestDb()
 *   beforeAll(async () => { await initSchema(client); _injectDbForTesting(db) })
 *   afterAll(() => _resetDbForTesting())
 *   afterEach(async () => { await client.executeMultiple("DELETE FROM ...") })
 */

import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import * as schema from "../schema"

export function createTestDb() {
  const client = createClient({ url: ":memory:" })
  const db = drizzle(client, { schema })
  return { client, db }
}

export type TestDb = ReturnType<typeof createTestDb>["db"]

/** Inicializa el schema completo en el cliente de test. */
export async function initSchema(client: ReturnType<typeof createClient>) {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      email TEXT NOT NULL UNIQUE,
      name TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'user',
      api_key_hash TEXT NOT NULL,
      password_hash TEXT,
      preferences TEXT NOT NULL DEFAULT '{}',
      active INTEGER NOT NULL DEFAULT 1,
      onboarding_completed INTEGER NOT NULL DEFAULT 0,
      sso_provider TEXT,
      sso_subject TEXT,
      created_at INTEGER NOT NULL,
      last_login INTEGER,
      last_seen INTEGER
    );
    CREATE TABLE IF NOT EXISTS areas (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT NOT NULL DEFAULT '',
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS user_areas (
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
      PRIMARY KEY (user_id, area_id)
    );
    CREATE TABLE IF NOT EXISTS area_collections (
      area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
      collection_name TEXT NOT NULL,
      permission TEXT NOT NULL DEFAULT 'read',
      PRIMARY KEY (area_id, collection_name)
    );
    CREATE TABLE IF NOT EXISTS chat_sessions (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      collection TEXT NOT NULL,
      crossdoc INTEGER NOT NULL DEFAULT 0,
      forked_from TEXT,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS chat_messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      role TEXT NOT NULL,
      content TEXT NOT NULL,
      sources TEXT,
      timestamp INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS message_feedback (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      message_id INTEGER NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
      user_id INTEGER NOT NULL REFERENCES users(id),
      rating TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      UNIQUE(message_id, user_id)
    );
    CREATE TABLE IF NOT EXISTS events (
      id TEXT PRIMARY KEY,
      ts INTEGER NOT NULL,
      source TEXT NOT NULL,
      level TEXT NOT NULL,
      type TEXT NOT NULL,
      user_id INTEGER REFERENCES users(id),
      session_id TEXT,
      payload TEXT NOT NULL DEFAULT '{}',
      sequence INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS user_memory (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      key TEXT NOT NULL,
      value TEXT NOT NULL,
      source TEXT NOT NULL DEFAULT 'explicit',
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      UNIQUE(user_id, key)
    );
    CREATE TABLE IF NOT EXISTS annotations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
      selected_text TEXT NOT NULL,
      note TEXT,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS session_tags (
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      tag TEXT NOT NULL,
      PRIMARY KEY (session_id, tag)
    );
    CREATE TABLE IF NOT EXISTS session_shares (
      id TEXT PRIMARY KEY,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      token TEXT NOT NULL UNIQUE,
      expires_at INTEGER NOT NULL,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS webhooks (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      url TEXT NOT NULL,
      events TEXT NOT NULL DEFAULT '[]',
      secret TEXT NOT NULL,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS scheduled_reports (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      query TEXT NOT NULL,
      collection TEXT NOT NULL,
      schedule TEXT NOT NULL,
      destination TEXT NOT NULL,
      email TEXT,
      active INTEGER NOT NULL DEFAULT 1,
      last_run INTEGER,
      next_run INTEGER NOT NULL,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS collection_history (
      id TEXT PRIMARY KEY,
      collection TEXT NOT NULL,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      action TEXT NOT NULL,
      filename TEXT,
      doc_count INTEGER,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS prompt_templates (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      prompt TEXT NOT NULL,
      focus_mode TEXT NOT NULL DEFAULT 'detallado',
      created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS rate_limits (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      target_type TEXT NOT NULL,
      target_id INTEGER NOT NULL,
      max_queries_per_hour INTEGER NOT NULL,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS projects (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      name TEXT NOT NULL,
      description TEXT NOT NULL DEFAULT '',
      instructions TEXT NOT NULL DEFAULT '',
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS project_sessions (
      project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      PRIMARY KEY (project_id, session_id)
    );
    CREATE TABLE IF NOT EXISTS project_collections (
      project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
      collection_name TEXT NOT NULL,
      PRIMARY KEY (project_id, collection_name)
    );
    CREATE TABLE IF NOT EXISTS saved_responses (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
      content TEXT NOT NULL,
      session_title TEXT,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS external_sources (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      provider TEXT NOT NULL,
      name TEXT NOT NULL,
      credentials TEXT NOT NULL DEFAULT '{}',
      collection_dest TEXT NOT NULL,
      schedule TEXT NOT NULL DEFAULT 'daily',
      active INTEGER NOT NULL DEFAULT 1,
      last_sync INTEGER,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_events_sequence ON events(sequence);
    CREATE UNIQUE INDEX IF NOT EXISTS idx_session_shares_token ON session_shares(token);
    CREATE UNIQUE INDEX IF NOT EXISTS idx_user_memory_unique ON user_memory(user_id, key);
  `)
}

// ── Helpers de inserción para datos de test ───────────────────────────────

export async function insertUser(
  db: TestDb,
  email = "user@test.com",
  role: "admin" | "area_manager" | "user" = "user"
) {
  const [user] = await db
    .insert(schema.users)
    .values({ email, name: "Test User", role, apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  if (!user) throw new Error("insertUser failed")
  return user
}

export async function insertSession(
  db: TestDb,
  userId: number,
  id = crypto.randomUUID(),
  title = "Test Session"
) {
  const now = Date.now()
  const [session] = await db
    .insert(schema.chatSessions)
    .values({ id, userId, title, collection: "test-col", crossdoc: false, createdAt: now, updatedAt: now })
    .returning()
  if (!session) throw new Error("insertSession failed")
  return session
}

export async function insertMessage(
  db: TestDb,
  sessionId: string,
  role: "user" | "assistant" | "system" = "user",
  content = "Test message"
) {
  const [msg] = await db
    .insert(schema.chatMessages)
    .values({ sessionId, role, content, timestamp: Date.now() })
    .returning()
  if (!msg) throw new Error("insertMessage failed")
  return msg
}
