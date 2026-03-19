import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGateway = {
	gatewayGenerateText: vi.fn(),
	GatewayError: class GatewayError extends Error {
		status: number;
		detail: string;
		constructor(status: number, detail: string) {
			super(detail);
			this.status = status;
			this.detail = detail;
		}
	},
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('POST /api/crossdoc/subquery', () => {
	beforeEach(() => vi.resetAllMocks());

	it('returns 401 without auth', async () => {
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'presión bomba' }),
			}),
			locals: { user: null },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('success: true cuando gateway devuelve contenido útil', async () => {
		mockGateway.gatewayGenerateText.mockResolvedValue(
			'La presión máxima es 12 bar según la ficha técnica.'
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'presión máxima bomba' }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		const res = await POST(event);
		const body = await res.json();
		expect(body.success).toBe(true);
		expect(body.content).toBe('La presión máxima es 12 bar según la ficha técnica.');
	});

	it('success: false cuando gateway devuelve respuesta vacía', async () => {
		mockGateway.gatewayGenerateText.mockResolvedValue('No information found');
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ query: 'query sin resultados' }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		const res = await POST(event);
		const body = await res.json();
		expect(body.success).toBe(false);
	});
});
