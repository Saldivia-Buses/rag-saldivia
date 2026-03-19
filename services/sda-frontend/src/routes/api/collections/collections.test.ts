import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGateway = {
    gatewayListCollections: vi.fn(),
    gatewayCreateCollection: vi.fn(),
    gatewayDeleteCollection: vi.fn(),
    GatewayError: class GatewayError extends Error {
        status: number; detail: string;
        constructor(status: number, detail: string) { super(detail); this.status = status; this.detail = detail; }
    },
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('GET /api/collections', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 200 with collections list', async () => {
        mockGateway.gatewayListCollections.mockResolvedValue({ collections: ['col-a', 'col-b'] });
        const { GET } = await import('./+server.js');
        const event = { locals: { user: { id: 1 } } } as any;
        const res = await GET(event);
        expect(res.status).toBe(200);
        const body = await res.json();
        expect(body.collections).toEqual(['col-a', 'col-b']);
    });

    it('returns 401 when not authenticated', async () => {
        const { GET } = await import('./+server.js');
        const event = { locals: { user: null } } as any;
        await expect(GET(event)).rejects.toMatchObject({ status: 401 });
    });
});

describe('POST /api/collections', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 201 with created collection name', async () => {
        mockGateway.gatewayCreateCollection.mockResolvedValue({ name: 'nueva-col' });
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost/api/collections', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: 'nueva-col', schema: 'default' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        expect(res.status).toBe(201);
        const body = await res.json();
        expect(body.name).toBe('nueva-col');
    });

    it('returns 400 when name is missing', async () => {
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost/api/collections', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ schema: 'default' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 400 });
    });
});

describe('DELETE /api/collections/[name]', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 204 on successful delete', async () => {
        mockGateway.gatewayDeleteCollection.mockResolvedValue({ ok: true });
        const { DELETE } = await import('./[name]/+server.js');
        const event = {
            params: { name: 'col-a' },
            locals: { user: { id: 1 } },
        } as any;
        const res = await DELETE(event);
        expect(res.status).toBe(204);
    });

    it('returns 401 when not authenticated', async () => {
        const { DELETE } = await import('./[name]/+server.js');
        const event = { params: { name: 'col-a' }, locals: { user: null } } as any;
        await expect(DELETE(event)).rejects.toMatchObject({ status: 401 });
    });
});
