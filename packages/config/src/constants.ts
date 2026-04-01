/**
 * Centralized constants — single source of truth for all timeouts, TTLs, and limits.
 *
 * Import from @rag-saldivia/config instead of hardcoding values.
 * Plan 26: extracted from 8 files across the monorepo.
 */

// ── RAG ────────────────────────────────────────────────────────────────────
/** Default timeout for RAG server requests (overridable via RAG_TIMEOUT_MS env) */
export const RAG_TIMEOUT_MS = 120_000
/** Timeout for non-streaming RAG requests (collections, documents) */
export const RAG_FETCH_TIMEOUT_MS = 10_000

// ── Redis cache TTLs ───────────────────────────────────────────────────────
/** Collections list cache TTL in seconds */
export const COLLECTIONS_CACHE_TTL_S = 60
/** Admin dashboard stats cache TTL in seconds */
export const ADMIN_STATS_CACHE_TTL_S = 60

// ── WebSocket ──────────────────────────────────────────────────────────────
/** Base delay for reconnect backoff */
export const WS_RECONNECT_BASE_MS = 1_000
/** Maximum delay for reconnect backoff */
export const WS_RECONNECT_MAX_MS = 30_000

// ── Presence ───────────────────────────────────────────────────────────────
/** Redis TTL for presence keys */
export const PRESENCE_TTL_S = 30
/** Heartbeat interval for presence updates */
export const PRESENCE_HEARTBEAT_MS = 15_000

// ── UI ─────────────────────────────────────────────────────────────────────
/** Auto-expire typing indicator */
export const TYPING_TIMEOUT_MS = 3_000

// ── Logger ─────────────────────────────────────────────────────────────────
/** Client logger batch size before flush */
export const LOGGER_BATCH_SIZE = 20
/** Client logger flush interval */
export const LOGGER_FLUSH_INTERVAL_MS = 5_000

// ── Auth ───────────────────────────────────────────────────────────────────
/** Access token lifetime (short-lived) */
export const ACCESS_TOKEN_EXPIRY = "15m"
/** Access token max-age in seconds (15 minutes) */
export const ACCESS_TOKEN_MAX_AGE_S = 15 * 60
/** Refresh token lifetime (long-lived) */
export const REFRESH_TOKEN_EXPIRY = "7d"
/** Refresh token max-age in seconds (7 days) */
export const REFRESH_TOKEN_MAX_AGE_S = 7 * 86_400
