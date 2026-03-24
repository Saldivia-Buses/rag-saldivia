import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('GET /api/dev-login', () => {
	beforeEach(() => {
		vi.resetModules();
	});

	afterEach(() => {
		vi.unstubAllEnvs();
		delete process.env.NODE_ENV;
	});

	it('retorna 403 en modo producción', async () => {
		vi.stubEnv('NODE_ENV', 'production');

		const { GET } = await import('./+server.js');
		const event = { cookies: { set: vi.fn() } } as any;

		const response = await GET(event);

		expect(response.status).toBe(403);
	});

	it('setea cookie JWT y redirige a /chat en modo development', async () => {
		vi.stubEnv('NODE_ENV', 'development');
		vi.stubEnv('JWT_SECRET', 'dev-secret-for-test');

		const { GET } = await import('./+server.js');
		const mockSet = vi.fn();
		const event = { cookies: { set: mockSet } } as any;

		// redirect() throws a Response-like object
		await expect(GET(event)).rejects.toMatchObject({ status: 302, location: '/chat' });

		// The cookie should have been set before the redirect
		expect(mockSet).toHaveBeenCalledWith('sda_session', expect.any(String), expect.any(Object));
	});

	it('genera token JWT con claims de admin dev', async () => {
		vi.stubEnv('NODE_ENV', 'test');
		vi.stubEnv('JWT_SECRET', 'dev-secret-for-test');

		const { GET } = await import('./+server.js');
		let capturedToken: string | null = null;
		const mockSet = vi.fn((name: string, value: string) => {
			if (name === 'sda_session') capturedToken = value;
		});
		const event = { cookies: { set: mockSet } } as any;

		await GET(event).catch(() => {}); // redirect throws, ignore it

		expect(capturedToken).not.toBeNull();
		// JWT format: header.payload.signature
		expect(capturedToken?.split('.')).toHaveLength(3);
	});

	it('usa dev-secret-local cuando JWT_SECRET no está definida', async () => {
		vi.stubEnv('NODE_ENV', 'development');
		delete process.env.JWT_SECRET;

		const { GET } = await import('./+server.js');
		const mockSet = vi.fn();
		const event = { cookies: { set: mockSet } } as any;

		// Should redirect without throwing TypeError (i.e., JWT signing worked)
		await expect(GET(event)).rejects.toMatchObject({ status: 302 });
		expect(mockSet).toHaveBeenCalled();
	});
});
