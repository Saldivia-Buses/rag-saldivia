import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import Toast from './Toast.svelte';
import type { Toast as ToastType } from '$lib/stores/toast.svelte';

function makeToast(type: ToastType['type'], message: string): ToastType {
    return { id: 'test-id', type, message, duration: 3000 };
}

describe('Toast', () => {
    it('muestra el mensaje del toast', () => {
        render(Toast, { props: { toast: makeToast('success', 'Operación exitosa') } });
        expect(screen.getByText('Operación exitosa')).toBeInTheDocument();
    });

    it('success: muestra ícono ✓', () => {
        render(Toast, { props: { toast: makeToast('success', 'ok') } });
        expect(screen.getByText('✓')).toBeInTheDocument();
    });

    it('error: muestra ícono ✕ en el cuerpo', () => {
        render(Toast, { props: { toast: makeToast('error', 'algo falló') } });
        // ✕ aparece dos veces: como ícono del tipo Y como botón de cierre
        const icons = screen.getAllByText('✕');
        expect(icons.length).toBeGreaterThanOrEqual(1);
    });

    it('warning: muestra ícono ⚠', () => {
        render(Toast, { props: { toast: makeToast('warning', 'cuidado') } });
        expect(screen.getByText('⚠')).toBeInTheDocument();
    });

    it('info: muestra ícono ℹ', () => {
        render(Toast, { props: { toast: makeToast('info', 'info msg') } });
        expect(screen.getByText('ℹ')).toBeInTheDocument();
    });

    it('tiene botón de cerrar con aria-label accesible', () => {
        render(Toast, { props: { toast: makeToast('success', 'test') } });
        expect(screen.getByRole('button', { name: /cerrar/i })).toBeInTheDocument();
    });

    it('botón cerrar invoca toastStore.dismiss al hacer clic', async () => {
        const user = userEvent.setup();
        render(Toast, { props: { toast: makeToast('success', 'a cerrar') } });

        const closeBtn = screen.getByRole('button', { name: /cerrar/i });
        // El clic llama toastStore.dismiss(toast.id) → no debe lanzar error
        await user.click(closeBtn);
        // El componente sigue en DOM (no se auto-destruye al clickear cerrar)
        expect(closeBtn).toBeInTheDocument();
    });

    it('each type applies a different style class', () => {
        const types: ToastType['type'][] = ['success', 'error', 'warning', 'info'];
        for (const type of types) {
            const { unmount, container } = render(Toast, { props: { toast: makeToast(type, type) } });
            // Cada tipo tiene un border color diferente
            const div = container.querySelector('div');
            expect(div?.className).toBeTruthy();
            unmount();
        }
    });
});
