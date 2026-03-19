import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('GET/DELETE /api/chat/sessions/[id]', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
		vi.stubEnv('SYSTEM_API_KEY', 'test-api-key');
	});

	afterEach(() => {
		vi.unstubAllEnvs();
		vi.unstubAllGlobals();
	});

	const mockUser = { id: 7, name: 'Test User', email: 'test@example.com', role: 'user', area_id: 1 };

	it('GET retorna 401 cuando no hay usuario autenticado', async () => {
		vi.stubGlobal('fetch', vi.fn());

		const { GET } = await import('./+server.js');
		const event = {
			params: { id: 'session-abc' },
			locals: { user: null },
		} as any;

		await expect(GET(event)).rejects.toMatchObject({ status: 401 });
	});

	it('GET devuelve el detalle de la sesión', async () => {
		const mockSession = {
			id: 'session-abc',
			title: 'Mi conversación',
			collection: 'documentos',
			crossdoc: false,
			updated_at: '2026-03-19T10:00:00Z',
			messages: [
				{ role: 'user', content: '¿Qué dice el manual?', timestamp: '2026-03-19T10:01:00Z' },
				{ role: 'assistant', content: 'El manual indica...', timestamp: '2026-03-19T10:01:05Z' },
			],
		};
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => mockSession,
		}));

		const { GET } = await import('./+server.js');
		const event = {
			params: { id: 'session-abc' },
			locals: { user: mockUser },
		} as any;

		const response = await GET(event);
		const body = await response.json();

		expect(response.status).toBe(200);
		expect(body.id).toBe('session-abc');
		expect(body.messages).toHaveLength(2);
	});

	it('GET propaga el error de gateway con status y detalle', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 404,
			text: async () => 'Session not found',
		}));

		const { GET } = await import('./+server.js');
		const event = {
			params: { id: 'nonexistent' },
			locals: { user: mockUser },
		} as any;

		await expect(GET(event)).rejects.toMatchObject({ status: 404 });
	});

	it('DELETE retorna 401 cuando no hay usuario autenticado', async () => {
		vi.stubGlobal('fetch', vi.fn());

		const { DELETE } = await import('./+server.js');
		const event = {
			params: { id: 'session-abc' },
			locals: { user: null },
		} as any;

		await expect(DELETE(event)).rejects.toMatchObject({ status: 401 });
	});

	it('DELETE elimina la sesión y retorna ok', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ({ ok: true }),
		}));

		const { DELETE } = await import('./+server.js');
		const event = {
			params: { id: 'session-abc' },
			locals: { user: mockUser },
		} as any;

		const response = await DELETE(event);
		const body = await response.json();

		expect(response.status).toBe(200);
		expect(body.ok).toBe(true);
	});

	it('DELETE propaga el status de error del gateway', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 503,
			text: async () => 'Service unavailable',
		}));

		const { DELETE } = await import('./+server.js');
		const event = {
			params: { id: 'session-abc' },
			locals: { user: mockUser },
		} as any;

		await expect(DELETE(event)).rejects.toMatchObject({ status: 503 });
	});

	it('DELETE retorna 503 cuando gateway lanza Error genérico (no GatewayError)', async () => {
		// Cubre línea 23: `err instanceof GatewayError ? err.status : 503` → rama false (503)
		vi.resetModules();
		vi.doMock('$lib/server/gateway', () => ({
			gatewayGetSession: vi.fn(),
			gatewayDeleteSession: vi.fn().mockRejectedValue(new Error('Error interno inesperado')),
			GatewayError: class GatewayError extends Error {
				status: number; detail: string;
				constructor(s: number, d: string) { super(d); this.status = s; this.detail = d; }
			},
		}));

		const { DELETE } = await import('./+server.js');
		const event = {
			params: { id: 'session-abc' },
			locals: { user: mockUser },
		} as any;

		await expect(DELETE(event)).rejects.toMatchObject({ status: 503 });
	});

	it('GET retorna 404 cuando el gateway retorna sesión nula', async () => {
		// Gateway returns null session → route throws error(404) → caught by catch → rethrown
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => null,
		}));

		const { GET } = await import('./+server.js');
		const event = {
			params: { id: 'session-null' },
			locals: { user: mockUser },
		} as any;

		await expect(GET(event)).rejects.toMatchObject({ status: 404 });
	});
});
