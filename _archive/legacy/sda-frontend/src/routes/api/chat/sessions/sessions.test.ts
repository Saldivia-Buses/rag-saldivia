import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('GET/POST /api/chat/sessions', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
		vi.stubEnv('SYSTEM_API_KEY', 'test-api-key');
	});

	afterEach(() => {
		vi.unstubAllEnvs();
		vi.unstubAllGlobals();
	});

	const mockUser = { id: 5, name: 'Test User', email: 'test@example.com', role: 'user', area_id: 1 };

	it('GET retorna 401 cuando no hay usuario autenticado', async () => {
		vi.stubGlobal('fetch', vi.fn());

		const { GET } = await import('./+server.js');
		const event = { locals: { user: null } } as any;

		await expect(GET(event)).rejects.toMatchObject({ status: 401 });
	});

	it('GET devuelve la lista de sesiones del usuario', async () => {
		const mockSessions = {
			sessions: [
				{ id: 's1', title: 'Mi sesión', collection: 'test-coll', crossdoc: false, updated_at: '2026-03-19T10:00:00Z' },
			],
		};
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => mockSessions,
		}));

		const { GET } = await import('./+server.js');
		const event = { locals: { user: mockUser } } as any;

		const response = await GET(event);
		const body = await response.json();

		expect(response.status).toBe(200);
		expect(body.sessions).toHaveLength(1);
		expect(body.sessions[0].id).toBe('s1');
	});

	it('GET propaga el status de error del gateway', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 502,
			text: async () => 'Bad Gateway',
		}));

		const { GET } = await import('./+server.js');
		const event = { locals: { user: mockUser } } as any;

		await expect(GET(event)).rejects.toMatchObject({ status: 502 });
	});

	it('GET retorna 502 en error de red (gateway wraps como GatewayError 502)', async () => {
		vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Connection refused')));

		const { GET } = await import('./+server.js');
		const event = { locals: { user: mockUser } } as any;

		// gateway.ts wraps network errors as GatewayError(502), route propagates that status
		await expect(GET(event)).rejects.toMatchObject({ status: 502 });
	});

	it('POST retorna 401 cuando no hay usuario autenticado', async () => {
		vi.stubGlobal('fetch', vi.fn());

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collection: 'test', crossdoc: false }),
			}),
			locals: { user: null },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('POST crea una nueva sesión y retorna 201', async () => {
		const mockSession = { id: 's-new', title: 'Nueva sesión', collection: 'main-coll' };
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => mockSession,
		}));

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collection: 'main-coll', crossdoc: false }),
			}),
			locals: { user: mockUser },
		} as any;

		const response = await POST(event);
		const body = await response.json();

		expect(response.status).toBe(201);
		expect(body.id).toBe('s-new');
	});

	it('POST propaga el status de error del gateway', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 404,
			text: async () => 'Collection not found',
		}));

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collection: 'noexiste', crossdoc: false }),
			}),
			locals: { user: mockUser },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 404 });
	});

	it('GET retorna 503 cuando gateway lanza Error genérico (no GatewayError)', async () => {
		// Cubre línea 11: `err instanceof GatewayError ? err.status : 503` → rama false (503)
		// Requiere que la función del gateway lance un Error plano, no un GatewayError
		vi.resetModules();
		vi.doMock('$lib/server/gateway', () => ({
			gatewayListSessions: vi.fn().mockRejectedValue(new Error('Error interno inesperado')),
			GatewayError: class GatewayError extends Error {
				status: number; detail: string;
				constructor(s: number, d: string) { super(d); this.status = s; this.detail = d; }
			},
		}));

		const { GET } = await import('./+server.js');
		const event = { locals: { user: mockUser } } as any;

		await expect(GET(event)).rejects.toMatchObject({ status: 503 });
	});

	it('POST retorna 503 cuando gateway lanza Error genérico (no GatewayError)', async () => {
		// Cubre línea 23: mismo patrón en POST
		vi.resetModules();
		vi.doMock('$lib/server/gateway', () => ({
			gatewayCreateSession: vi.fn().mockRejectedValue(new Error('Error interno inesperado')),
			GatewayError: class GatewayError extends Error {
				status: number; detail: string;
				constructor(s: number, d: string) { super(d); this.status = s; this.detail = d; }
			},
		}));

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collection: 'test', crossdoc: false }),
			}),
			locals: { user: mockUser },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 503 });
	});
});
