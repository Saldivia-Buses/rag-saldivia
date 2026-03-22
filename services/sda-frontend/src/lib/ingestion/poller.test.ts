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
});
