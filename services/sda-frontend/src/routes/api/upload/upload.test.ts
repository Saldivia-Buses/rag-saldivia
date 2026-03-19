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

    it('retorna 400 cuando falta la colección en el formulario', async () => {
        // Cubre líneas 12-13: if (!collection || ...) → collection es null
        const { POST } = await import('./+server.js');
        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        // No se agrega 'collection'
        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1 } },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 400 });
    });

    it('retorna 400 cuando colección es string de solo espacios', async () => {
        // Cubre líneas 12-13: !collection.trim() → '' después de trim es falsy
        const { POST } = await import('./+server.js');
        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        formData.append('collection', '   '); // solo espacios → trim() = ''
        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1 } },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 400 });
    });

    it('retorna 503 cuando gateway responde con error', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const mockFetch = vi.fn().mockResolvedValue({
            ok: false,
            status: 503,
            json: () => Promise.resolve({ error: 'Service unavailable' }),
        });
        vi.stubGlobal('fetch', mockFetch);

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'test-col');

        const event = {
            request: new Request('http://localhost/api/upload', {
                method: 'POST',
                body: formData,
            }),
            locals: { user: { id: 1, email: 'test@example.com', role: 'user', area_id: 1, name: 'Test' } },
        } as any;

        const res = await POST(event);
        // Handler propagates the gateway's status
        expect(res.status).toBe(503);
    });

    it('retorna 502 cuando gateway es inalcanzable', async () => {
        process.env.GATEWAY_URL = 'http://unreachable:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));
        vi.stubGlobal('fetch', mockFetch);

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'test-col');

        const event = {
            request: new Request('http://localhost/api/upload', {
                method: 'POST',
                body: formData,
            }),
            locals: { user: { id: 1, email: 'test@example.com', role: 'user', area_id: 1, name: 'Test' } },
        } as any;

        await expect(POST(event)).rejects.toMatchObject({ status: 502 });
    });

    it('retorna 504 cuando el gateway agota el timeout (AbortError)', async () => {
        // Cubre línea 39: if ((err as any)?.name === 'AbortError') throw error(504, ...)
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const abortError = Object.assign(new Error('AbortError'), { name: 'AbortError' });
        vi.stubGlobal('fetch', vi.fn().mockRejectedValue(abortError));

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'test-col');

        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1, email: 'test@example.com', role: 'user', area_id: 1, name: 'Test' } },
        } as any;

        await expect(POST(event)).rejects.toMatchObject({ status: 504 });
    });

    it('retorna 503 cuando SYSTEM_API_KEY no está configurado', async () => {
        // Cubre línea 18: if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.')
        // La var se lee dentro del handler (no a nivel módulo), por eso podemos mutarla
        process.env.GATEWAY_URL = 'http://gateway:9000';
        const savedKey = process.env.SYSTEM_API_KEY;
        delete process.env.SYSTEM_API_KEY;

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'test-col');

        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1 } },
        } as any;

        try {
            await expect(POST(event)).rejects.toMatchObject({ status: 503 });
        } finally {
            // Restaurar para no afectar otros tests
            if (savedKey !== undefined) process.env.SYSTEM_API_KEY = savedKey;
        }
    });

    it('usa localhost:9000 como URL por defecto cuando GATEWAY_URL no está configurado', async () => {
        // Cubre línea 16: const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000'
        const savedUrl = process.env.GATEWAY_URL;
        delete process.env.GATEWAY_URL;
        process.env.SYSTEM_API_KEY = 'test-key';

        const localMock = vi.fn().mockResolvedValue({
            ok: true,
            status: 200,
            json: () => Promise.resolve({ ok: true }),
        });
        vi.stubGlobal('fetch', localMock);

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['data'], 'test.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'test-col');

        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1, email: 'test@example.com', role: 'user', area_id: 1, name: 'Test' } },
        } as any;

        await POST(event);

        expect(localMock).toHaveBeenCalledWith(
            expect.stringContaining('http://localhost:9000'),
            expect.any(Object)
        );

        // Restaurar
        if (savedUrl !== undefined) process.env.GATEWAY_URL = savedUrl;
    });
});
