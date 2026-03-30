/**
 * FTS5 setup for messaging full-text search.
 *
 * Creates a virtual table and triggers to keep it in sync with msg_messages.
 * Call setupFTS5() after database migrations to ensure the FTS5 table exists.
 *
 * Note: FTS5 virtual tables cannot be managed by Drizzle migrations,
 * so we create them programmatically via raw SQL.
 */

import { sql } from "drizzle-orm"
import { getDb } from "./connection"

/**
 * Create FTS5 virtual table and sync triggers for messaging search.
 * Safe to call multiple times (IF NOT EXISTS).
 */
export async function setupFTS5(): Promise<void> {
  const db = getDb()

  await db.run(sql`CREATE VIRTUAL TABLE IF NOT EXISTS msg_messages_fts
    USING fts5(content, content=msg_messages, content_rowid=rowid)`)

  await db.run(sql`CREATE TRIGGER IF NOT EXISTS msg_fts_insert AFTER INSERT ON msg_messages BEGIN
    INSERT INTO msg_messages_fts(rowid, content) VALUES (new.rowid, new.content);
  END`)

  await db.run(sql`CREATE TRIGGER IF NOT EXISTS msg_fts_delete AFTER DELETE ON msg_messages BEGIN
    INSERT INTO msg_messages_fts(msg_messages_fts, rowid, content)
      VALUES('delete', old.rowid, old.content);
  END`)

  await db.run(sql`CREATE TRIGGER IF NOT EXISTS msg_fts_update AFTER UPDATE OF content ON msg_messages BEGIN
    INSERT INTO msg_messages_fts(msg_messages_fts, rowid, content)
      VALUES('delete', old.rowid, old.content);
    INSERT INTO msg_messages_fts(rowid, content) VALUES (new.rowid, new.content);
  END`)
}
