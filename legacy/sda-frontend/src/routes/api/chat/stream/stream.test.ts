import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('POST /api/chat/stream/[id]', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.unstubAllEnvs();
		vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
		vi.stubEnv('SYSTEM_API_KEY', 'test-api-key');
	});

	afterEach(() => {
		vi.unstubAllEnvs();
		vi.resetModules(); // evita que un test con vi.resetModules() corrompa el cache para los siguientes
	});

	it('returns 401 when unauthenticated', async () => {
		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'test question', collection_names: ['test'] }),
			}),
			params: { id: 'session-123' },
			locals: { user: null },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('proxies SSE stream from gateway when authenticated', async () => {
		const mockStream = new ReadableStream({
			start(controller) {
				const encoder = new TextEncoder();
				controller.enqueue(encoder.encode('data: {"choices":[{"delta":{"content":"Hello"}}]}\n\n'));
				controller.enqueue(encoder.encode('data: [DONE]\n\n'));
				controller.close();
			},
		});

		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			body: mockStream,
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'example question', collection_names: ['example'] }),
			}),
			params: { id: 'test-id' },
			locals: { user: { id: 42, name: 'Test User', email: 'test@example.com', role: 'user', area_id: 1 } },
		} as any;

		const response = await POST(event);

		expect(response.status).toBe(200);
		expect(response.headers.get('content-type')).toBe('text/event-stream');
		expect(response.headers.get('cache-control')).toBe('no-cache');
		expect(mockFetch).toHaveBeenCalledWith(
			'http://gateway:9000/v1/generate',
			expect.objectContaining({
				method: 'POST',
				headers: expect.objectContaining({
					'Authorization': 'Bearer test-api-key',
					'Content-Type': 'application/json',
				}),
			})
		);
	});

	it('retorna 500 cuando SYSTEM_API_KEY no está configurado (línea 35)', async () => {
		// Cubre línea 35: if (!SYSTEM_API_KEY) throw error(500, 'Server misconfiguration')
		// SYSTEM_API_KEY es una constante de módulo → hay que resetear el módulo sin ella
		vi.resetModules();
		vi.unstubAllEnvs();
		vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
		// NO stubeamos SYSTEM_API_KEY → el módulo la lee como undefined en el load

		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'test', collection_names: ['test'] }),
			}),
			params: { id: 'session-abc' },
			locals: { user: { id: 1, name: 'Test', email: 'test@example.com', role: 'user', area_id: 1 } },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 500 });
	});

	it('returns 502 when gateway is unreachable', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'test question', collection_names: ['test'] }),
			}),
			params: { id: 'session-123' },
			locals: { user: { id: 1, name: 'Test', email: 'test@example.com', role: 'user', area_id: 1 } },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 502 });
	});

	it('returns 504 on timeout', async () => {
		const abortError = new Error('Aborted');
		(abortError as any).name = 'AbortError';
		const mockFetch = vi.fn().mockRejectedValue(abortError);
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'test question', collection_names: ['test'] }),
			}),
			params: { id: 'session-123' },
			locals: { user: { id: 1, name: 'Test', email: 'test@example.com', role: 'user', area_id: 1 } },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 504 });
	});

	it('retorna 502 cuando el gateway responde ok pero sin body (body es null)', async () => {
		// Cubre líneas 77-79: if (!gatewayResp.body) throw error(502, ...)
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			body: null,
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'test', collection_names: ['test'] }),
			}),
			params: { id: 'session-abc' },
			locals: { user: { id: 1, name: 'Test', email: 'test@example.com', role: 'user', area_id: 1 } },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 502 });
	});

	it('propagates gateway error status', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			status: 503,
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./[id]/+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'test question', collection_names: ['test'] }),
			}),
			params: { id: 'session-123' },
			locals: { user: { id: 1, name: 'Test', email: 'test@example.com', role: 'user', area_id: 1 } },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 503 });
	});
});
