import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('POST/DELETE /api/auth/session', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
		vi.stubEnv('SYSTEM_API_KEY', 'test-api-key');
	});

	afterEach(() => {
		vi.unstubAllEnvs();
		vi.unstubAllGlobals();
	});

	it('POST inicia sesión y devuelve el usuario', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ({
				token: 'jwt-token-abc',
				user: { id: 1, email: 'test@example.com', name: 'Test User', role: 'user', area_id: 1 },
			}),
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./+server.js');
		const mockCookies = { set: vi.fn(), delete: vi.fn(), get: vi.fn() };
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email: 'test@example.com', password: 'password123' }),
			}),
			cookies: mockCookies,
		} as any;

		const response = await POST(event);
		const body = await response.json();

		expect(response.status).toBe(200);
		expect(body.user.email).toBe('test@example.com');
		expect(mockCookies.set).toHaveBeenCalledWith('sda_session', 'jwt-token-abc', expect.any(Object));
	});

	it('POST retorna 401 cuando el gateway retorna 401', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			status: 401,
			text: async () => 'Unauthorized',
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email: 'bad@example.com', password: 'wrong' }),
			}),
			cookies: { set: vi.fn(), delete: vi.fn() },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('POST retorna 401 cuando el gateway retorna 403', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			status: 403,
			text: async () => 'Forbidden',
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email: 'banned@example.com', password: 'pass' }),
			}),
			cookies: { set: vi.fn(), delete: vi.fn() },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('POST retorna 503 cuando el gateway retorna otro error', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			status: 500,
			text: async () => 'Internal Server Error',
		});
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email: 'user@example.com', password: 'pass' }),
			}),
			cookies: { set: vi.fn(), delete: vi.fn() },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 503 });
	});

	it('POST retorna 503 cuando el gateway es inalcanzable', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));
		vi.stubGlobal('fetch', mockFetch);

		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost/', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email: 'user@example.com', password: 'pass' }),
			}),
			cookies: { set: vi.fn(), delete: vi.fn() },
		} as any;

		await expect(POST(event)).rejects.toMatchObject({ status: 503 });
	});

	it('DELETE cierra la sesión limpiando la cookie', async () => {
		vi.stubGlobal('fetch', vi.fn());

		const { DELETE } = await import('./+server.js');
		const mockCookies = { set: vi.fn(), delete: vi.fn() };
		const event = { cookies: mockCookies } as any;

		const response = await DELETE(event);
		const body = await response.json();

		expect(body.ok).toBe(true);
		expect(mockCookies.delete).toHaveBeenCalledWith('sda_session', { path: '/' });
	});
});
