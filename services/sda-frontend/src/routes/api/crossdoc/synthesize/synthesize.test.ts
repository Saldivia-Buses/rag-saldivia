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
});
