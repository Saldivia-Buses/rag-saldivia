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

    it('load() con respuesta 404 → deja collections vacías', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 404,
            json: () => Promise.resolve({ error: 'Not found' }),
        }));
        await store.load();
        expect(store.collections).toEqual([]);
        expect(store.loading).toBe(false);
    });

    it('load() con respuesta sin campo collections → deja vacío', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ data: {} }),
        }));
        await store.load();
        expect(store.collections).toEqual([]);
    });

    it('create() con error del servidor lanza excepción con mensaje', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 400,
            json: () => Promise.resolve({ message: 'Collection already exists' }),
        }));
        await expect(store.create('duplicate', 'default')).rejects.toThrow('Collection already exists');
    });

    it('delete() con error del servidor lanza excepción', async () => {
        store.collections = ['col-a'];
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 500,
            json: () => Promise.resolve({ message: 'Internal server error' }),
        }));
        await expect(store.delete('col-a')).rejects.toThrow('Internal server error');
        // Collection no debe eliminarse si falló
        expect(store.collections).toEqual(['col-a']);
    });

    it('create() cuando res.json() falla usa mensaje fallback "Error {status}"', async () => {
        // Cubre línea 36: (err as any).message ?? `Error ${res.status}`
        // Cuando json() rechaza, catch(() => ({})) retorna {} → .message es undefined → usa fallback
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 503,
            json: () => Promise.reject(new Error('invalid json')),
        }));
        await expect(store.create('mi-col', 'default')).rejects.toThrow('Error 503');
    });

    it('delete() cuando res.json() falla usa mensaje fallback "Error {status}"', async () => {
        // Cubre línea 45: mismo patrón en delete()
        store.collections = ['col-x'];
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false,
            status: 500,
            json: () => Promise.reject(new Error('invalid json')),
        }));
        await expect(store.delete('col-x')).rejects.toThrow('Error 500');
    });

    it('loading se activa durante load() y se desactiva al terminar', async () => {
        let resolvePromise: any;
        const slowFetch = vi.fn().mockReturnValue(new Promise(resolve => {
            resolvePromise = resolve;
        }));
        vi.stubGlobal('fetch', slowFetch);

        const loadPromise = store.load();
        expect(store.loading).toBe(true);

        resolvePromise({
            ok: true,
            json: () => Promise.resolve({ collections: ['test'] }),
        });
        await loadPromise;
        expect(store.loading).toBe(false);
    });
});
