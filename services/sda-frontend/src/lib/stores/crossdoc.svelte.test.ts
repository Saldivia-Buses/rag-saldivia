import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CrossdocStore } from './crossdoc.svelte.js';
import { ChatStore } from './chat.svelte.js';
import { DEFAULT_CROSSDOC_OPTIONS } from '$lib/crossdoc/types';

// Mock fetch global
const mockFetch = vi.fn();
global.fetch = mockFetch;

function makeSSEResponse(text: string) {
    const stream = new ReadableStream({
        start(c) {
            c.enqueue(new TextEncoder().encode(`data: {"choices":[{"delta":{"content":"${text}"}}]}\n`));
            c.enqueue(new TextEncoder().encode('data: [DONE]\n'));
            c.close();
        },
    });
    return new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } });
}

describe('CrossdocStore', () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it('options tienen defaults correctos', () => {
        const store = new CrossdocStore();
        expect(store.options).toEqual(DEFAULT_CROSSDOC_OPTIONS);
    });

    it('progress empieza en null', () => {
        const store = new CrossdocStore();
        expect(store.progress).toBeNull();
    });

    it('reset() limpia el progress', () => {
        const store = new CrossdocStore();
        // @ts-ignore — direct state access for test
        store.progress = { phase: 'done', subQueries: [], completed: 0, total: 0, results: [] };
        store.reset();
        expect(store.progress).toBeNull();
    });

    it('run() llama decompose → subquery → synthesize en orden', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            // decompose call
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query uno test', 'query dos test'] }), { status: 200 }))
            // subquery call × 2
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado uno', success: true }), { status: 200 }))
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado dos', success: true }), { status: 200 }))
            // synthesize call → SSE
            .mockResolvedValueOnce(makeSSEResponse('respuesta final'));

        await store.run('¿Cuál es la presión?', chat);

        expect(mockFetch).toHaveBeenCalledTimes(4);
        expect(mockFetch.mock.calls[0][0]).toContain('/api/crossdoc/decompose');
        expect(mockFetch.mock.calls[1][0]).toContain('/api/crossdoc/subquery');
        expect(mockFetch.mock.calls[3][0]).toContain('/api/crossdoc/synthesize');
    });

    it('run() → chat.messages tiene la respuesta final', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query uno test'] }), { status: 200 }))
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado', success: true }), { status: 200 }))
            .mockResolvedValueOnce(makeSSEResponse('respuesta final'));

        await store.run('¿test?', chat);

        expect(chat.messages).toHaveLength(1);
        expect(chat.messages[0].role).toBe('assistant');
        expect(chat.messages[0].content).toContain('respuesta final');
    });

    it('stop() aborta el pipeline limpiamente', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        // decompose aborts → AbortError
        mockFetch.mockImplementation(() =>
            Promise.reject(Object.assign(new Error('AbortError'), { name: 'AbortError' }))
        );

        const runPromise = store.run('¿test?', chat);
        setTimeout(() => store.stop(), 10);
        await runPromise; // must resolve without throwing

        expect(chat.streaming).toBe(false);
    });
});
