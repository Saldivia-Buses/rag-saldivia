#!/usr/bin/env bun
/**
 * Inicialización directa de la base de datos con SQL puro.
 * No requiere drizzle-kit. Usa bun:sqlite directamente.
 *
 * Uso: bun packages/db/src/init.ts
 *      bun run db:migrate
 */

import { createClient } from "@libsql/client"
import { mkdirSync } from "fs"
import { join, dirname, resolve } from "path"

const dbPath = process.env["DATABASE_PATH"] ?? join(process.cwd(), "data", "app.db")
const dbUrl = dbPath === ":memory:" ? ":memory:" : `file:${resolve(dbPath)}`

if (dbPath !== ":memory:") {
  try { mkdirSync(dirname(resolve(dbPath)), { recursive: true }) } catch { /* ya existe */ }
}

const client = createClient({ url: dbUrl })

// Helper para ejecutar SQL de setup
async function exec(sql: string) {
  await client.executeMultiple(sql)
}

console.log(`Inicializando base de datos: ${dbPath}`)

console.log(`Inicializando base de datos: ${dbPath}`)

await exec(`
  -- Áreas
  CREATE TABLE IF NOT EXISTS areas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL
  );

  -- Usuarios
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    api_key_hash TEXT NOT NULL,
    password_hash TEXT,
    preferences TEXT NOT NULL DEFAULT '{}',
    active INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    last_login INTEGER
  );
  CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key_hash);

  -- Usuario ↔ Áreas (many-to-many)
  CREATE TABLE IF NOT EXISTS user_areas (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, area_id)
  );

  -- Área ↔ Colecciones
  CREATE TABLE IF NOT EXISTS area_collections (
    area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
    collection_name TEXT NOT NULL,
    permission TEXT NOT NULL DEFAULT 'read',
    PRIMARY KEY (area_id, collection_name)
  );

  -- Audit log
  CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    action TEXT NOT NULL,
    collection TEXT,
    query_preview TEXT,
    ip_address TEXT NOT NULL DEFAULT '',
    timestamp INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id);
  CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);

  -- Sesiones de chat
  CREATE TABLE IF NOT EXISTS chat_sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    collection TEXT NOT NULL,
    crossdoc INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_chat_sessions_user ON chat_sessions(user_id);

  -- Mensajes de chat
  CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    sources TEXT,
    timestamp INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id);

  -- Feedback de mensajes
  CREATE TABLE IF NOT EXISTS message_feedback (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    rating TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    UNIQUE(message_id, user_id)
  );

  -- Proyectos (F3.41)
  CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    instructions TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);

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

  -- Webhooks (F2.38)
  CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    events TEXT NOT NULL DEFAULT '[]',
    secret TEXT NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(active);

  -- Rate limits (F2.36)
  CREATE TABLE IF NOT EXISTS rate_limits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    max_queries_per_hour INTEGER NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_rate_limits_target ON rate_limits(target_type, target_id);

  -- Informes programados (F2.33)
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
  CREATE INDEX IF NOT EXISTS idx_reports_active_next_run ON scheduled_reports(active, next_run);

  -- Historial de colecciones (F2.32)
  CREATE TABLE IF NOT EXISTS collection_history (
    id TEXT PRIMARY KEY,
    collection TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    filename TEXT,
    doc_count INTEGER,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_collection_history_collection ON collection_history(collection);

  -- Templates de query (F2.28)
  CREATE TABLE IF NOT EXISTS prompt_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    prompt TEXT NOT NULL,
    focus_mode TEXT NOT NULL DEFAULT 'detallado',
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    active INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_prompt_templates_active ON prompt_templates(active);

  -- Compartir sesiones (F2.25)
  CREATE TABLE IF NOT EXISTS session_shares (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL
  );
  CREATE UNIQUE INDEX IF NOT EXISTS idx_session_shares_token ON session_shares(token);

  -- Etiquetas de sesiones (F2.24)
  CREATE TABLE IF NOT EXISTS session_tags (
    session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (session_id, tag)
  );
  CREATE INDEX IF NOT EXISTS idx_session_tags_tag ON session_tags(tag);

  -- Anotaciones de fragmentos (F2.22)
  CREATE TABLE IF NOT EXISTS annotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    selected_text TEXT NOT NULL,
    note TEXT,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_annotations_user ON annotations(user_id);
  CREATE INDEX IF NOT EXISTS idx_annotations_session ON annotations(session_id);

  -- Respuestas guardadas (F1.10)
  CREATE TABLE IF NOT EXISTS saved_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    session_title TEXT,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS idx_saved_responses_user ON saved_responses(user_id);

  -- Jobs de ingesta
  CREATE TABLE IF NOT EXISTS ingestion_jobs (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    task_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    collection TEXT NOT NULL,
    tier TEXT NOT NULL,
    page_count INTEGER,
    state TEXT NOT NULL DEFAULT 'pending',
    progress INTEGER NOT NULL DEFAULT 0,
    file_hash TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_checked INTEGER,
    created_at INTEGER NOT NULL,
    completed_at INTEGER
  );
  CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_user ON ingestion_jobs(user_id);
  CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_state ON ingestion_jobs(state);

  -- Alertas de ingesta
  CREATE TABLE IF NOT EXISTS ingestion_alerts (
    id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    filename TEXT NOT NULL,
    collection TEXT NOT NULL,
    tier TEXT NOT NULL,
    page_count INTEGER,
    file_hash TEXT,
    error TEXT,
    retry_count INTEGER,
    progress_at_failure INTEGER,
    gateway_version TEXT,
    created_at INTEGER NOT NULL,
    resolved_at INTEGER,
    resolved_by TEXT,
    notes TEXT
  );

  -- Cola de ingesta (reemplaza Redis)
  CREATE TABLE IF NOT EXISTS ingestion_queue (
    id TEXT PRIMARY KEY,
    collection TEXT NOT NULL,
    file_path TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    priority INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    locked_at INTEGER,
    locked_by TEXT,
    created_at INTEGER NOT NULL,
    started_at INTEGER,
    completed_at INTEGER,
    error TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0
  );
  CREATE INDEX IF NOT EXISTS idx_queue_status ON ingestion_queue(status, priority);

  -- Black Box — eventos del sistema
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
  CREATE INDEX IF NOT EXISTS idx_events_ts ON events(ts);
  CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
  CREATE INDEX IF NOT EXISTS idx_events_level ON events(level);
  CREATE INDEX IF NOT EXISTS idx_events_sequence ON events(sequence);
`)

// FTS5 virtual tables for universal search (F3.39)
// Nota: executeMultiple no soporta múltiples statements que incluyen FTS — ejecutar por separado
try {
  await client.executeMultiple(`
    CREATE VIRTUAL TABLE IF NOT EXISTS sessions_fts USING fts5(
      session_id UNINDEXED,
      user_id UNINDEXED,
      title,
      collection UNINDEXED,
      content=chat_sessions,
      content_rowid=rowid
    );

    CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
      message_id UNINDEXED,
      session_id UNINDEXED,
      body,
      content=chat_messages,
      content_rowid=rowid
    );
  `)

  // Triggers para mantener FTS sincronizado con chat_sessions
  await client.executeMultiple(`
    CREATE TRIGGER IF NOT EXISTS sessions_fts_insert AFTER INSERT ON chat_sessions BEGIN
      INSERT INTO sessions_fts(rowid, session_id, user_id, title, collection) VALUES (new.rowid, new.id, new.user_id, new.title, new.collection);
    END;

    CREATE TRIGGER IF NOT EXISTS sessions_fts_delete AFTER DELETE ON chat_sessions BEGIN
      INSERT INTO sessions_fts(sessions_fts, rowid, session_id, user_id, title, collection) VALUES ('delete', old.rowid, old.id, old.user_id, old.title, old.collection);
    END;

    CREATE TRIGGER IF NOT EXISTS sessions_fts_update AFTER UPDATE ON chat_sessions BEGIN
      INSERT INTO sessions_fts(sessions_fts, rowid, session_id, user_id, title, collection) VALUES ('delete', old.rowid, old.id, old.user_id, old.title, old.collection);
      INSERT INTO sessions_fts(rowid, session_id, user_id, title, collection) VALUES (new.rowid, new.id, new.user_id, new.title, new.collection);
    END;
  `)

  // Triggers para chat_messages
  await client.executeMultiple(`
    CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON chat_messages BEGIN
      INSERT INTO messages_fts(rowid, message_id, session_id, body) VALUES (new.rowid, new.id, new.session_id, new.content);
    END;

    CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON chat_messages BEGIN
      INSERT INTO messages_fts(messages_fts, rowid, message_id, session_id, body) VALUES ('delete', old.rowid, old.id, old.session_id, old.content);
    END;
  `)
} catch (err) {
  // FTS5 puede no estar disponible en todas las compilaciones de SQLite
  console.warn("FTS5 no disponible — búsqueda universal desactivada:", String(err).slice(0, 100))
}

console.log("Base de datos inicializada correctamente (libsql)")
console.log("Tablas creadas: areas, users, user_areas, area_collections,")
console.log("  audit_log, chat_sessions, chat_messages, message_feedback,")
console.log("  ingestion_jobs, ingestion_alerts, ingestion_queue, events")
