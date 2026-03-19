// src/lib/server/gateway.ts
// Typed wrapper for all gateway API calls. Uses SYSTEM_API_KEY Bearer auth.
const GATEWAY_URL = process.env.GATEWAY_URL ?? 'http://localhost:9000';

/** Default timeout for normal API calls (ms) */
const DEFAULT_TIMEOUT_MS = 10_000;

function getSystemApiKey(): string {
    const key = process.env.SYSTEM_API_KEY;
    if (!key) throw new Error('SYSTEM_API_KEY environment variable is required');
    return key;
}

const headers = () => ({
    'Authorization': `Bearer ${getSystemApiKey()}`,
    'Content-Type': 'application/json',
});

export class GatewayError extends Error {
    status: number;
    detail: string;
    constructor(status: number, detail: string) {
        super(`Gateway error (${status}): ${detail}`);
        this.name = 'GatewayError';
        this.status = status;
        this.detail = detail;
    }
}

async function gw<T>(path: string, init?: RequestInit & { timeoutMs?: number }): Promise<T> {
    const { timeoutMs, ...fetchInit } = init ?? {};
    const timeout = timeoutMs ?? DEFAULT_TIMEOUT_MS;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), timeout);

    try {
        const res = await fetch(`${GATEWAY_URL}${path}`, {
            ...fetchInit,
            headers: { ...headers(), ...(fetchInit?.headers ?? {}) },
            signal: controller.signal,
        });
        if (!res.ok) {
            const detail = await res.text();
            throw new GatewayError(res.status, detail);
        }
        return res.json() as Promise<T>;
    } catch (err) {
        if (err instanceof GatewayError) throw err;
        if ((err as any)?.name === 'AbortError') {
            throw new GatewayError(504, `Gateway timeout after ${timeout}ms on ${path}`);
        }
        throw new GatewayError(502, `Gateway unreachable: ${(err as Error).message}`);
    } finally {
        clearTimeout(timer);
    }
}

// Auth
export async function gatewayLogin(email: string, password: string) {
    return gw<{ token: string; user: SessionUser }>(
        '/auth/session',
        { method: 'POST', body: JSON.stringify({ email, password }) }
    );
}

export async function gatewayGetMe(userId: number) {
    return gw<SessionUser>(`/auth/me?user_id=${userId}`);
}

export async function gatewayRefreshKey(userId: number) {
    return gw<{ api_key: string }>(`/auth/refresh-key?user_id=${userId}`, { method: 'POST' });
}

// Collections
export async function gatewayListCollections() {
    return gw<{ collections: string[] }>('/v1/collections');
}

export async function gatewayCollectionStats(name: string) {
    return gw<CollectionStats>(`/v1/collections/${name}/stats`);
}

export async function gatewayCreateCollection(name: string, schema = 'default') {
    return gw<{ name: string }>(
        '/v1/collections',
        { method: 'POST', body: JSON.stringify({ name, schema }) }
    );
}

export async function gatewayDeleteCollection(name: string) {
    return gw<{ ok: boolean }>(`/v1/collections/${name}`, { method: 'DELETE' });
}

// Chat sessions
export async function gatewayListSessions(userId: number) {
    return gw<{ sessions: ChatSessionSummary[] }>(`/chat/sessions?user_id=${userId}`);
}

export async function gatewayCreateSession(userId: number, collection: string, crossdoc = false) {
    return gw<{ id: string; title: string; collection: string }>(
        `/chat/sessions?user_id=${userId}`,
        { method: 'POST', body: JSON.stringify({ collection, crossdoc }) }
    );
}

export async function gatewayGetSession(sessionId: string, userId: number) {
    return gw<ChatSessionDetail>(`/chat/sessions/${sessionId}?user_id=${userId}`);
}

export async function gatewayDeleteSession(sessionId: string, userId: number) {
    return gw<{ ok: boolean }>(`/chat/sessions/${sessionId}?user_id=${userId}`, { method: 'DELETE' });
}

// Admin
export async function gatewayListUsers() {
    return gw<{ users: AdminUser[] }>('/admin/users');
}

export async function gatewayCreateUser(data: CreateUserData) {
    return gw<{ id: number; email: string; api_key: string }>(
        '/admin/users', { method: 'POST', body: JSON.stringify(data) }
    );
}

export async function gatewayUpdateUser(id: number, data: Partial<AdminUser>) {
    return gw<{ ok: boolean }>(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export async function gatewayDeleteUser(id: number) {
    return gw<{ ok: boolean }>(`/admin/users/${id}`, { method: 'DELETE' });
}

export async function gatewayListAreas() {
    return gw<{ areas: AreaSummary[] }>('/admin/areas');
}

export async function gatewayGetAreaCollections(areaId: number) {
    return gw<{ collections: AreaCollection[] }>(`/admin/areas/${areaId}/collections`);
}

export async function gatewayGrantCollection(areaId: number, collectionName: string, permission = 'read') {
    return gw<{ ok: boolean }>(
        `/admin/areas/${areaId}/collections`,
        { method: 'POST', body: JSON.stringify({ collection_name: collectionName, permission }) }
    );
}

export async function gatewayRevokeCollection(areaId: number, collectionName: string) {
    return gw<{ ok: boolean }>(`/admin/areas/${areaId}/collections/${collectionName}`, { method: 'DELETE' });
}

export async function gatewayGetAudit(params: AuditParams) {
    const qs = new URLSearchParams(
        Object.entries(params).filter(([, v]) => v != null).map(([k, v]) => [k, String(v)])
    ).toString();
    return gw<{ entries: AuditEntry[] }>(`/admin/audit${qs ? '?' + qs : ''}`);
}

// Types
export interface SessionUser {
    id: number; email: string; name: string; role: string; area_id: number;
}
export interface CollectionStats {
    collection: string; entity_count: number; document_count?: number;
    index_type?: string; has_sparse?: boolean;
}
export interface ChatSessionSummary {
    id: string; title: string; collection: string; crossdoc: boolean; updated_at: string;
}
export interface ChatSessionDetail extends ChatSessionSummary {
    messages: ChatMessage[];
}
export interface ChatMessage {
    role: 'user' | 'assistant'; content: string; sources?: Source[]; timestamp: string;
}
export interface Source {
    document: string; page?: number; excerpt: string;
}
export interface AdminUser {
    id: number; email: string; name: string; area_id: number; role: string;
    active: boolean; last_login: string | null;
}
export interface CreateUserData {
    email: string; name: string; area_id: number; role: string; password?: string;
}
export interface AreaSummary {
    id: number; name: string; description: string;
}
export interface AreaCollection {
    name: string; permission: string;
}
export interface AuditEntry {
    id: number; user_id: number; action: string; collection: string | null;
    query_preview: string | null; ip_address: string; timestamp: string;
}
export interface AuditParams {
    user_id?: number; action?: string; collection?: string;
    from?: string; to?: string; limit?: number;
}
