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

    it('adds an info toast', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.info('Información disponible');
        const last = toastStore.toasts.at(-1)!;
        expect(last.type).toBe('info');
        expect(last.message).toBe('Información disponible');
        expect(last.duration).toBe(4000);
    });

    it('adds a warning toast', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.warning('Advertencia importante');
        const last = toastStore.toasts.at(-1)!;
        expect(last.type).toBe('warning');
        expect(last.message).toBe('Advertencia importante');
        expect(last.duration).toBe(5000);
    });

    it('con duration=0 no registra auto-dismiss (timeout no se llama)', async () => {
        // Cubre la rama `if (duration > 0)` → false cuando duration es 0
        vi.useFakeTimers();
        const setTimeoutSpy = vi.spyOn(global, 'setTimeout');

        const { toastStore } = await import('./toast.svelte.js');
        toastStore.success('permanente', 0); // duration=0 → no auto-dismiss

        const last = toastStore.toasts.at(-1)!;
        expect(last.duration).toBe(0);
        // setTimeout no debe haber sido llamado para el dismiss de este toast
        expect(setTimeoutSpy).not.toHaveBeenCalled();

        vi.useRealTimers();
    });

    it('removes a toast by id', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.success('test');
        const id = toastStore.toasts.at(-1)!.id;
        toastStore.dismiss(id);
        expect(toastStore.toasts.find(t => t.id === id)).toBeUndefined();
    });
});
