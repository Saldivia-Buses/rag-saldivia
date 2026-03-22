import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('GET /api/ingestion/[jobId]/status', () => {
    beforeEach(() => vi.resetAllMocks());

    it('proxea al gateway y devuelve estado del job', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const mockResponse = {
            job_id: 'j1', state: 'running', progress: 42,
            tier: 'medium', page_count: 180, filename: 'doc.pdf',
            collection: 'col', created_at: '2026-03-22T10:00:00'
        };

        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            status: 200,
            json: () => Promise.resolve(mockResponse),
        }));

        const { GET } = await import('./+server.js');
        const event = {
            params: { jobId: 'j1' },
            locals: { user: { id: 1, token: 'user-jwt' } },
        } as any;

        const res = await GET(event);
        const body = await res.json();

        expect(res.status).toBe(200);
        expect(body.progress).toBe(42);
        expect(body.state).toBe('running');
    });

    it('retorna 401 sin sesión', async () => {
        const { GET } = await import('./+server.js');
        const event = { params: { jobId: 'j1' }, locals: { user: null } } as any;
        await expect(GET(event)).rejects.toMatchObject({ status: 401 });
    });

    it('propaga 404 cuando el job no existe', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false, status: 404,
            text: () => Promise.resolve('Not found'),
            json: () => Promise.resolve({ detail: 'Job not found' }),
        }));

        const { GET } = await import('./+server.js');
        const event = {
            params: { jobId: 'nonexistent' },
            locals: { user: { id: 1, token: 'jwt' } },
        } as any;

        const res = await GET(event);
        expect(res.status).toBe(404);
    });
});
