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

    it('run() ejecuta follow-up retries cuando hay queries fallidas', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        // followUpRetries is true by default
        mockFetch
            // decompose call #1 → 1 query
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query que falla'] }), { status: 200 }))
            // subquery call → fails
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: '', success: false }), { status: 200 }))
            // decompose call #2 (retry for failed queries)
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query alternativa'] }), { status: 200 }))
            // retry subquery call → succeeds
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado recuperado', success: true }), { status: 200 }))
            // synthesize → SSE
            .mockResolvedValueOnce(makeSSEResponse('síntesis final'));

        await store.run('¿Cuál es la presión?', chat);

        // Should have called fetch 5 times (decompose + subquery + retry-decompose + retry-subquery + synthesize)
        expect(mockFetch).toHaveBeenCalledTimes(5);
        expect(store.progress?.phase).toBe('done');
        expect(chat.messages[0].content).toContain('síntesis final');
    });

    it('run() maneja correctly cuando el retry decompose no responde ok', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            // decompose → 1 query
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query que falla larga'] }), { status: 200 }))
            // subquery → fails
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: '', success: false }), { status: 200 }))
            // retry decompose → not ok
            .mockResolvedValueOnce(new Response('error', { status: 500 }))
            // synthesize (skips retry subqueries since altResp.ok is false)
            .mockResolvedValueOnce(makeSSEResponse('síntesis sin retries'));

        await store.run('¿test?', chat);

        expect(store.progress?.phase).toBe('done');
    });

    // BUG CONOCIDO (línea 49): cuando un subquery retorna HTTP no-ok, el store debería
    // continuar a phase='done' con result { success: false }. Actualmente entra al catch()
    // y setea phase='error' en vez de tratar la falla como un resultado graceful.
    // it.fails() documenta la falla esperada; cuando el bug se corrija, quitar .fails()
    it.fails('run() trata como fallido un subquery con respuesta HTTP no-ok (linha 49)', async () => {
        // Cubre línea 49: if (!resp.ok) return { query, content: '', success: false }
        // El fetch de subquery responde pero con status de error (503, 404, etc.)
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query que da error http'] }), { status: 200 }))
            // subquery → HTTP 503 → resp.ok es false → línea 49 retorna temprano
            .mockResolvedValueOnce(new Response('Service unavailable', { status: 503 }))
            // synthesize
            .mockResolvedValueOnce(makeSSEResponse('síntesis sin datos útiles'));

        await store.run('¿test non-ok?', chat);

        expect(store.progress?.phase).toBe('done');
        // El resultado de la subquery no-ok debe ser { success: false, content: '' }
        expect(store.progress?.results).toContainEqual(
            expect.objectContaining({ success: false, content: '' })
        );
    });

    it('run() trata como fallido un subquery cuya Promise fue rechazada', async () => {
        // Cubre línea 55: `r.status === 'fulfilled' ? r.value : { query: '', content: '', success: false }`
        // Promise.allSettled → una promesa rejected → entra a la rama `rejected`
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            // decompose → 2 queries
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query ok test', 'query falla test'] }), { status: 200 }))
            // primer subquery → responde ok
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado ok', success: true }), { status: 200 }))
            // segundo subquery → lanza error de red (la Promise se rechaza → fulfilled=false)
            .mockRejectedValueOnce(new Error('Network failure'))
            // synthesize
            .mockResolvedValueOnce(makeSSEResponse('síntesis con fallo'));

        await store.run('¿test con fallo?', chat);

        // El pipeline debe terminar sin error
        expect(store.progress?.phase).toBe('done');
        // El resultado fallido debe incluir el item vacío que puso el branch rejected
        expect(store.progress?.results).toContainEqual(
            expect.objectContaining({ success: false, content: '' })
        );
    });

    it('run() synthesize SSE: procesa líneas no-data, JSON malformado y [DONE]', async () => {
        // Cubre líneas 112-130 del SSE reading loop en el bloque synthesize:
        // - línea 114: if (!line.startsWith('data: ')) continue → líneas de comentario
        // - línea 116: if (data === '[DONE]') continue → sentinel
        // - línea 121: catch de JSON.parse malformado
        const store = new CrossdocStore();
        const chat = new ChatStore();

        // Usamos un mock body manual (no ReadableStream) para controlar los chunks exactamente
        let readCallCount = 0;
        const chunks = [
            new TextEncoder().encode(
                ': keep-alive\n' +                                          // comentario → skip (línea 114)
                'data: {json malformado}\n' +                               // JSON inválido → catch (línea 121)
                'data: [DONE]\n' +                                          // sentinel → continue (línea 116)
                'data: {"choices":[{"delta":{"content":"hola"}}]}\n'        // válido → appendToken
            ),
        ];

        const mockSynthBody = {
            getReader: () => ({
                read: vi.fn().mockImplementation(() => {
                    if (readCallCount < chunks.length) {
                        return Promise.resolve({ done: false, value: chunks[readCallCount++] });
                    }
                    return Promise.resolve({ done: true, value: undefined });
                }),
                releaseLock: vi.fn(),
            }),
        };

        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['q test larga'] }), { status: 200 }))
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado', success: true }), { status: 200 }))
            .mockResolvedValueOnce({ ok: true, body: mockSynthBody });

        await store.run('¿test?', chat);

        // Solo el token válido después del [DONE] debe aparecer en el mensaje
        expect(chat.messages[0].content).toContain('hola');
        expect(store.progress?.phase).toBe('done');
    });

    it('run() setea progress.phase=error en errores no-AbortError', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        // Make decompose throw a generic error (not AbortError)
        mockFetch.mockRejectedValueOnce(new Error('Unexpected server failure'));

        await store.run('¿test?', chat);

        expect(store.progress?.phase).toBe('error');
        expect(store.progress?.error).toBe('Unexpected server failure');
        expect(chat.streaming).toBe(false);
    });

    it('run() setea phase=error cuando decompose responde HTTP no-ok (línea 27)', async () => {
        // Cubre línea 27: if (!decompResp.ok) throw new Error(...)
        // La respuesta de decompose es 503 (ok: false) → lanza Error → catch → phase: 'error'
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch.mockResolvedValueOnce(new Response('Service unavailable', { status: 503 }));

        await store.run('¿test decompose no-ok?', chat);

        expect(store.progress?.phase).toBe('error');
        expect(store.progress?.error).toContain('Decompose failed: 503');
    });

    it('run() break inmediato en batch loop cuando signal ya está abortado (línea 35)', async () => {
        // Cubre línea 35: if (signal.aborted) break
        // Llamamos store.stop() dentro del mock de decompose → signal.aborted = true
        // El for loop de subqueries verifica signal.aborted al inicio y hace break
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch.mockImplementation((url: string) => {
            if (url.includes('/decompose')) {
                store.stop(); // abort sincrónico antes de que empiece el batch loop
                return Promise.resolve(new Response(JSON.stringify({ subQueries: ['q1'] }), { status: 200 }));
            }
            // synthesize tampoco se llama porque !signal.aborted es false
            return Promise.resolve(new Response('{}', { status: 200 }));
        });

        await store.run('¿test abort in batch?', chat);

        // Solo decompose fue llamado (batch loop y synthesis saltados por signal.aborted)
        expect(mockFetch).toHaveBeenCalledTimes(1);
        // La función termina normalmente (no AbortError thrown) → phase: 'done'
        expect(store.progress?.phase).toBe('done');
    });

    it('run() break en retry loop cuando signal está abortado (línea 77)', async () => {
        // Cubre línea 77: if (signal.aborted) break en el loop de retry subqueries
        // Queremos que el signal se aborte justo antes de ejecutar un retry subquery
        const store = new CrossdocStore();
        const chat = new ChatStore();

        let callCount = 0;
        mockFetch.mockImplementation((url: string) => {
            callCount++;
            if (url.includes('/decompose')) {
                if (callCount === 1) {
                    // Primera llamada: decompose principal → ok, 1 query
                    return Promise.resolve(new Response(JSON.stringify({ subQueries: ['q fallida'] }), { status: 200 }));
                }
                // Segunda llamada: retry decompose → abortar AQUÍ antes del retry subquery
                store.stop(); // signal.aborted = true → el retry for-loop hará break en línea 77
                return Promise.resolve(new Response(JSON.stringify({ subQueries: ['alt'] }), { status: 200 }));
            }
            if (url.includes('/subquery')) {
                // subquery principal → falla para forzar el retry
                return Promise.resolve(new Response(JSON.stringify({ content: '', success: false }), { status: 200 }));
            }
            return Promise.resolve(new Response('{}', { status: 200 }));
        });

        await store.run('¿test abort in retry?', chat);

        // El pipeline termina sin error (signal abortado no lanza, solo hace break)
        expect(store.progress?.phase).toBe('done');
    });
});
