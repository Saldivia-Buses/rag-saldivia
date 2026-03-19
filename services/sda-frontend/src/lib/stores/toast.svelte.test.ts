import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('ToastStore', () => {
    beforeEach(() => {
        // Clear all toasts before each test
        vi.resetModules();
    });

    it('adds a success toast', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.success('Operación exitosa');
        expect(toastStore.toasts).toHaveLength(1);
        expect(toastStore.toasts[0].type).toBe('success');
        expect(toastStore.toasts[0].message).toBe('Operación exitosa');
    });

    it('adds an error toast', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.error('Error de conexión');
        const last = toastStore.toasts.at(-1)!;
        expect(last.type).toBe('error');
    });

    it('removes a toast by id', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.success('test');
        const id = toastStore.toasts.at(-1)!.id;
        toastStore.dismiss(id);
        expect(toastStore.toasts.find(t => t.id === id)).toBeUndefined();
    });
});
