import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { createRawSnippet } from 'svelte';
import Modal from './Modal.svelte';

// Snippet de contenido mínimo para Modal
const contentSnippet = createRawSnippet(() => ({
    render: () => `<p>Contenido del modal</p>`,
}));

describe('Modal', () => {
    it('no renderiza el dialog cuando open=false', () => {
        render(Modal, { props: { open: false, children: contentSnippet } });
        expect(screen.queryByRole('dialog')).toBeNull();
    });

    it('renderiza el dialog cuando open=true', () => {
        render(Modal, { props: { open: true, children: contentSnippet } });
        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('muestra el título cuando se pasa title', () => {
        render(Modal, {
            props: { open: true, title: 'Mi Modal', children: contentSnippet },
        });
        expect(screen.getByText('Mi Modal')).toBeInTheDocument();
    });

    it('no renderiza heading cuando no hay title', () => {
        render(Modal, { props: { open: true, children: contentSnippet } });
        expect(screen.queryByRole('heading')).toBeNull();
    });

    it('renderiza el contenido children', () => {
        render(Modal, { props: { open: true, children: contentSnippet } });
        expect(screen.getByText('Contenido del modal')).toBeInTheDocument();
    });

    it('botón Cerrar llama a onclose', async () => {
        const user = userEvent.setup();
        const onclose = vi.fn();
        render(Modal, {
            props: { open: true, title: 'Test', onclose, children: contentSnippet },
        });

        await user.click(screen.getByRole('button', { name: /cerrar/i }));

        expect(onclose).toHaveBeenCalledOnce();
    });

    it('presionar Escape llama a onclose cuando open=true', async () => {
        const user = userEvent.setup();
        const onclose = vi.fn();
        render(Modal, {
            props: { open: true, title: 'Test', onclose, children: contentSnippet },
        });

        await user.keyboard('{Escape}');

        expect(onclose).toHaveBeenCalledOnce();
    });
});
