// feedback.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('POST /api/chat/sessions/[id]/messages/[msgId]/feedback', () => {
    beforeEach(() => {
        vi.resetModules();
        vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
        vi.stubEnv('SYSTEM_API_KEY', 'test-key');
    });
    afterEach(() => {
        vi.unstubAllEnvs();
        vi.unstubAllGlobals();
    });

    const mockUser = { id: 7, name: 'Test', email: 't@test.com', role: 'user', area_id: 1 };

    it('POST retorna 401 sin usuario', async () => {
        vi.stubGlobal('fetch', vi.fn());
        const { POST } = await import('./+server.js');
        const event = {
            params: { id: 'ses-1', msgId: '42' },
            locals: { user: null },
            request: { json: async () => ({ rating: 'up' }) },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 401 });
    });

    it('POST guarda el feedback y retorna ok', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true }),
        }));
        const { POST } = await import('./+server.js');
        const event = {
            params: { id: 'ses-1', msgId: '42' },
            locals: { user: mockUser },
            request: { json: async () => ({ rating: 'up' }) },
        } as any;
        const response = await POST(event);
        const body = await response.json();
        expect(response.status).toBe(200);
        expect(body.ok).toBe(true);
    });

    it('POST propaga error del gateway', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 404,
            text: async () => 'Message not found',
        }));
        const { POST } = await import('./+server.js');
        const event = {
            params: { id: 'ses-1', msgId: '999' },
            locals: { user: mockUser },
            request: { json: async () => ({ rating: 'down' }) },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 404 });
    });
});
