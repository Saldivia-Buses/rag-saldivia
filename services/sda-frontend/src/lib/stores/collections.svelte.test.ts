import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CollectionsStore } from './collections.svelte.js';

describe('CollectionsStore', () => {
    let store: CollectionsStore;

    beforeEach(() => {
        store = new CollectionsStore();
        vi.resetAllMocks();
    });

    it('load() popula collections desde /api/collections', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ collections: ['col-a', 'col-b'] }),
        }));
        await store.load();
        expect(store.collections).toEqual(['col-a', 'col-b']);
        expect(store.loading).toBe(false);
    });

    it('load() con error fetch no rompe — deja collections vacías', async () => {
        vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network error')));
        await store.load();
        expect(store.collections).toEqual([]);
        expect(store.loading).toBe(false);
    });

    it('create() llama POST /api/collections y agrega a la lista', async () => {
        store.collections = ['existing'];
        const mockFetch = vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ name: 'new-col' }),
        });
        vi.stubGlobal('fetch', mockFetch);
        await store.create('new-col', 'default');
        expect(mockFetch).toHaveBeenCalledWith('/api/collections', expect.objectContaining({
            method: 'POST',
        }));
        expect(store.collections).toContain('new-col');
    });

    it('delete() llama DELETE /api/collections/[name] y quita de la lista', async () => {
        store.collections = ['col-a', 'col-b'];
        const mockFetch = vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ ok: true }),
        });
        vi.stubGlobal('fetch', mockFetch);
        await store.delete('col-a');
        expect(mockFetch).toHaveBeenCalledWith('/api/collections/col-a', expect.objectContaining({
            method: 'DELETE',
        }));
        expect(store.collections).toEqual(['col-b']);
    });

    it('init() hidrata colecciones desde datos del servidor', () => {
        store.init(['a', 'b', 'c']);
        expect(store.collections).toEqual(['a', 'b', 'c']);
    });
});
