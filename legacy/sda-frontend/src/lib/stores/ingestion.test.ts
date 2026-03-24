import { describe, it, expect, vi, beforeEach } from 'vitest';

async function freshStore() {
    vi.resetModules();
    const mod = await import('./ingestion.svelte.js');
    return mod;
}

describe('ingestionStore', () => {
    it('addJob agrega un job al store', async () => {
        const { ingestionStore } = await freshStore();
        const job = {
            jobId: 'j1', filename: 'a.pdf', collection: 'col',
            tier: 'tiny' as const, pageCount: 10, state: 'pending' as const,
            progress: 0, eta: null, startedAt: Date.now(), lastProgressAt: Date.now(),
        };
        ingestionStore.addJob(job);
        expect(ingestionStore.jobs.find(j => j.jobId === 'j1')).toBeDefined();
    });

    it('updateJob actualiza solo el job correcto', async () => {
        const { ingestionStore } = await freshStore();
        ingestionStore.addJob({
            jobId: 'j2', filename: 'b.pdf', collection: 'col',
            tier: 'small' as const, pageCount: 50, state: 'pending' as const,
            progress: 0, eta: null, startedAt: Date.now(), lastProgressAt: Date.now(),
        });
        ingestionStore.updateJob('j2', { progress: 42, state: 'running' });
        const job = ingestionStore.jobs.find(j => j.jobId === 'j2');
        expect(job?.progress).toBe(42);
        expect(job?.state).toBe('running');
    });

    it('removeJob elimina el job', async () => {
        const { ingestionStore } = await freshStore();
        ingestionStore.addJob({
            jobId: 'j3', filename: 'c.pdf', collection: 'col',
            tier: 'medium' as const, pageCount: 100, state: 'completed' as const,
            progress: 100, eta: 0, startedAt: Date.now(), lastProgressAt: Date.now(),
        });
        ingestionStore.removeJob('j3');
        expect(ingestionStore.jobs.find(j => j.jobId === 'j3')).toBeUndefined();
    });

    it('hydrateFromServer no duplica jobs existentes', async () => {
        const { ingestionStore } = await freshStore();
        const existing = {
            jobId: 'j4', filename: 'd.pdf', collection: 'col',
            tier: 'tiny' as const, pageCount: 5, state: 'running' as const,
            progress: 30, eta: null, startedAt: Date.now(), lastProgressAt: Date.now(),
        };
        ingestionStore.addJob(existing);
        ingestionStore.hydrateFromServer([
            { id: 'j4', filename: 'd.pdf', collection: 'col', tier: 'tiny',
              page_count: 5, state: 'running', progress: 30, created_at: '' }
        ]);
        expect(ingestionStore.jobs.filter(j => j.jobId === 'j4').length).toBe(1);
    });
});
