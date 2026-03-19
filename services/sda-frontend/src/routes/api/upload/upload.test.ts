import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('POST /api/upload', () => {
    beforeEach(() => vi.resetAllMocks());

    it('forwards multipart al gateway y retorna su respuesta', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const mockFetch = vi.fn().mockResolvedValue({
            ok: true,
            status: 202,
            json: () => Promise.resolve({ job_id: 'abc123', status: 'queued' }),
        });
        vi.stubGlobal('fetch', mockFetch);

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['contenido pdf'], 'doc.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'mi-coleccion');

        const event = {
            request: new Request('http://localhost/api/upload', {
                method: 'POST',
                body: formData,
            }),
            locals: { user: { id: 42, email: 'test@test.com', role: 'user', area_id: 1, name: 'Test' } },
        } as any;

        const res = await POST(event);
        expect(res.status).toBe(202);

        expect(mockFetch).toHaveBeenCalledWith(
            'http://gateway:9000/v1/documents',
            expect.objectContaining({
                method: 'POST',
                headers: expect.objectContaining({
                    'Authorization': 'Bearer test-key',
                    'X-User-Id': '42',
                }),
            })
        );
    });

    it('retorna 401 cuando no hay sesión', async () => {
        const { POST } = await import('./+server.js');
        const formData = new FormData();
        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: null },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 401 });
    });

    it('retorna 400 cuando falta archivo', async () => {
        const { POST } = await import('./+server.js');
        const formData = new FormData();
        formData.append('collection', 'mi-col');
        // No file appended
        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1 } },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 400 });
    });
});
