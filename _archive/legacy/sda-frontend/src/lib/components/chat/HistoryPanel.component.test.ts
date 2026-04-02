// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import HistoryPanel from './HistoryPanel.svelte';

const sessions = [
    { id: 's1', title: 'Sesión uno', updated_at: '2026-03-23T10:00:00Z' },
    { id: 's2', title: 'Sesión dos', updated_at: '2026-03-22T10:00:00Z' },
];

describe('HistoryPanel', () => {
    beforeEach(() => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({ ok: true }) }));
        const store: Record<string, string> = {};
        vi.stubGlobal('localStorage', {
            getItem: (k: string) => store[k] ?? null,
            setItem: (k: string, v: string) => { store[k] = v; },
            removeItem: (k: string) => { delete store[k]; },
        });
    });
    afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });

    it('renderiza la lista de sesiones', () => {
        const { getByText } = render(HistoryPanel, { props: { sessions, currentId: 's1' } });
        expect(getByText('Sesión uno')).toBeTruthy();
        expect(getByText('Sesión dos')).toBeTruthy();
    });

    it('doble click en sesión activa activa inline edit con input visible', async () => {
        const { getByText, container } = render(HistoryPanel, { props: { sessions, currentId: 's1' } });
        await fireEvent.dblClick(getByText('Sesión uno'));
        const input = container.querySelector('input[type="text"], input:not([type])');
        expect(input).not.toBeNull();
    });

    it('click en botón eliminar muestra modal de confirmación', async () => {
        const { getAllByTitle, getByText } = render(HistoryPanel, { props: { sessions, currentId: 's1' } });
        const deleteButtons = getAllByTitle('Eliminar');
        await fireEvent.click(deleteButtons[0]);
        expect(getByText(/Eliminar esta sesión/i)).toBeTruthy();
    });
});
