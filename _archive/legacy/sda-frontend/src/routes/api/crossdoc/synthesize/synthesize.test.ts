import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockStream = new ReadableStream({
	start(c) {
		c.enqueue(new TextEncoder().encode('data: {"choices":[{"delta":{"content":"respuesta"}}]}\n'));
		c.close();
	}
});

const mockGateway = {
	gatewayGenerateStream: vi.fn(),
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

describe('POST /api/crossdoc/synthesize', () => {
	beforeEach(() => vi.resetAllMocks());

	it('returns 401 without auth', async () => {
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: 'test', results: [] }),
			}),
			locals: { user: null },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 401 });
	});

	it('retorna 400 cuando question está vacía o falta', async () => {
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ results: [] }), // question missing
			}),
			locals: { user: { id: 1 } },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 400 });
	});

	it('usa [] cuando results es null (results ?? [])', async () => {
		mockGateway.gatewayGenerateStream.mockResolvedValue(
			new Response(mockStream, { headers: { 'Content-Type': 'text/event-stream' } })
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: '¿test?', results: null }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		// Should not throw — null results fallback to []
		const res = await POST(event);
		expect(res.status).toBe(200);
	});

	it('retorna 502 cuando gatewayGenerateStream devuelve respuesta sin body', async () => {
		// Response with no body → !resp.body → throws error(502)
		mockGateway.gatewayGenerateStream.mockResolvedValue(
			new Response(null, { status: 200 })
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: '¿test?', results: [] }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 502 });
	});

	it('retorna SSE stream con Content-Type text/event-stream', async () => {
		mockGateway.gatewayGenerateStream.mockResolvedValue(
			new Response(mockStream, { headers: { 'Content-Type': 'text/event-stream' } })
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					question: '¿Cuál es la presión máxima?',
					results: [{ query: 'presión bomba', content: 'La presión es 12 bar', success: true }],
				}),
			}),
			locals: { user: { id: 1 } },
		} as any;
		const res = await POST(event);
		expect(res.status).toBe(200);
		expect(res.headers.get('Content-Type')).toContain('text/event-stream');
	});

	it('llama gatewayGenerateStream con el synthesis prompt correcto', async () => {
		mockGateway.gatewayGenerateStream.mockResolvedValue(
			new Response(mockStream, { headers: { 'Content-Type': 'text/event-stream' } })
		);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					question: '¿Cuál es la presión?',
					results: [{ query: 'presión bomba', content: 'La presión es 12 bar', success: true }],
				}),
			}),
			locals: { user: { id: 1 } },
		} as any;
		await POST(event);
		expect(mockGateway.gatewayGenerateStream).toHaveBeenCalledOnce();
		const callOpts = mockGateway.gatewayGenerateStream.mock.calls[0][0];
		expect(callOpts.messages[0].content).toContain('¿Cuál es la presión?');
		expect(callOpts.messages[0].content).toContain('La presión es 12 bar');
	});

	it('propaga el status de GatewayError', async () => {
		const err = new mockGateway.GatewayError(502, 'Gateway unavailable');
		mockGateway.gatewayGenerateStream.mockRejectedValue(err);
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: '¿test?', results: [] }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 502 });
	});

	it('retorna 502 en errores genéricos (no GatewayError)', async () => {
		mockGateway.gatewayGenerateStream.mockRejectedValue(new Error('Unexpected error'));
		const { POST } = await import('./+server.js');
		const event = {
			request: new Request('http://localhost', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ question: '¿test?', results: [] }),
			}),
			locals: { user: { id: 1 } },
		} as any;
		await expect(POST(event)).rejects.toMatchObject({ status: 502 });
	});
});
