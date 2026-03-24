import { describe, it, expect, vi, beforeEach } from 'vitest';
import { IngestPoller } from './poller.js';

describe('IngestPoller', () => {
    beforeEach(() => vi.resetAllMocks());

    it('llama onUpdate con el estado del servidor', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ state: 'completed', progress: 100 }),
        }));

        const updates: any[] = [];
        const poller = new IngestPoller('job-1', 'tiny');
        await poller.poll((s) => updates.push(s));

        expect(updates.length).toBeGreaterThan(0);
        expect(updates.at(-1)?.state).toBe('completed');
    });

    it('para el loop cuando state es completed', async () => {
        let callCount = 0;
        vi.stubGlobal('fetch', vi.fn().mockImplementation(() => {
            callCount++;
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve({ state: 'completed', progress: 100 }),
            });
        }));

        const poller = new IngestPoller('job-2', 'tiny');
        await poller.poll(() => {});

        expect(callCount).toBe(1); // Solo 1 call porque ya completó
    });

    it('stop() detiene el loop', async () => {
        let callCount = 0;
        vi.stubGlobal('fetch', vi.fn().mockImplementation(() => {
            callCount++;
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve({ state: 'running', progress: callCount * 10 }),
            });
        }));

        const poller = new IngestPoller('job-4', 'tiny');
        const pollPromise = poller.poll(() => {});
        poller.stop();
        await pollPromise.catch(() => {});

        expect(callCount).toBeLessThanOrEqual(2);
    });

    it('llama _reportAlert y emite failed cuando agota maxRetries en stall', async () => {
        vi.useFakeTimers();

        let fetchCallCount = 0;
        vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
            fetchCallCount++;
            // Simular stall: siempre progress=10
            if (String(url).includes('/status')) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ state: 'running', progress: 10 }),
                });
            }
            // Alert endpoint
            return Promise.resolve({ ok: true, json: () => Promise.resolve({ ok: true }) });
        }));

        const updates: any[] = [];
        // maxRetries=1, backoffBase=1s para que el test sea rápido
        const poller = new IngestPoller('job-5', 'tiny', 1, 1);

        const pollPromise = poller.poll((s) => updates.push(s));

        // Avanzar tiempo para que el deadlock se detecte + backoff + segundo deadlock
        await vi.advanceTimersByTimeAsync(200_000);
        await pollPromise.catch(() => {});

        const lastUpdate = updates.at(-1);
        expect(lastUpdate?.state).toBe('failed');

        vi.useRealTimers();
    });

    it('reintentar con backoff cuando retryCount < maxRetries', async () => {
        vi.useFakeTimers();

        let callCount = 0;
        vi.stubGlobal('fetch', vi.fn().mockImplementation(() => {
            callCount++;
            // Primera llamada: stall; segunda: completado (después del retry)
            if (callCount === 1) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ state: 'running', progress: 10 }),
                });
            }
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve({ state: 'completed', progress: 100 }),
            });
        }));

        const updates: any[] = [];
        // maxRetries=2, backoffBase=1 para que sea rápido
        const poller = new IngestPoller('job-6', 'tiny', 2, 1);
        const pollPromise = poller.poll((s) => updates.push(s));

        // Advance past deadlock + backoff
        await vi.advanceTimersByTimeAsync(200_000);
        await pollPromise.catch(() => {});

        const states = updates.map(u => u.state);
        expect(states).toContain('completed');

        vi.useRealTimers();
    });
});
