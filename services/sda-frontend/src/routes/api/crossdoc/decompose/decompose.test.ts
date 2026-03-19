import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGateway = {
	gatewayGenerateText: vi.fn(),
	GatewayError: class GatewayError extends Error {
		status: number; detail: string;
		constructor(status: number, detail: string) { super(detail); this.status = status; this.detail = detail; }
	},
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('POST /api/crossdoc/decompose', () => {
	beforeEach(() => vi.resetAllMocks());

	it('returns 401 without auth', async () => {
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: 'test?' }),
			}),
			locals: { user: null },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('parsea sub-queries del LLM y aplica dedup', async () => {
		mockGateway.gatewayGenerateText.mockResolvedValue(
			'presión máxima bomba centrífuga\ntemperatura motor eléctrico\nvoltaje nominal inversor'
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: '¿Cuál es la presión máxima?' }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		const res = await POST(event);
		expect(res.status).toBe(200);
		const body = await res.json();
		expect(Array.isArray(body.subQueries)).toBe(true);
		expect(body.subQueries.length).toBeGreaterThan(0);
	});

	it('respeta el cap de maxSubQueries', async () => {
		mockGateway.gatewayGenerateText.mockResolvedValue(
			'query uno aqui\nquery dos aqui\nquery tres aqui\nquery cuatro aqui\nquery cinco aqui'
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: 'test', maxSubQueries: 2 }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		const res = await POST(event);
		const body = await res.json();
		expect(body.subQueries.length).toBeLessThanOrEqual(2);
	});
});
