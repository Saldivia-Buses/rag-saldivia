import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import ChatInput from './ChatInput.svelte';

const defaultProps = {
    streaming: false,
    crossdoc: false,
    onsubmit: vi.fn(),
    onstop: vi.fn(),
    oncrossdoctoggle: vi.fn(),
};

beforeEach(() => {
    vi.clearAllMocks();
});

describe('ChatInput', () => {
    it('renders a textarea input', () => {
        render(ChatInput, { props: defaultProps });
        expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    it('textarea has placeholder text', () => {
        render(ChatInput, { props: defaultProps });
        expect(screen.getByPlaceholderText(/escribí tu consulta/i)).toBeInTheDocument();
    });

    it('shows send button when not streaming', () => {
        render(ChatInput, { props: defaultProps });
        expect(screen.getByTitle(/enviar/i)).toBeInTheDocument();
    });

    it('shows stop button when streaming=true', () => {
        render(ChatInput, { props: { ...defaultProps, streaming: true } });
        expect(screen.getByTitle(/detener/i)).toBeInTheDocument();
    });

    it('hides send button when streaming=true', () => {
        render(ChatInput, { props: { ...defaultProps, streaming: true } });
        expect(screen.queryByTitle(/enviar/i)).toBeNull();
    });

    it('textarea is disabled when streaming=true', () => {
        render(ChatInput, { props: { ...defaultProps, streaming: true } });
        expect(screen.getByRole('textbox')).toBeDisabled();
    });

    it('calls onsubmit with message text when send button clicked', async () => {
        const user = userEvent.setup();
        const onsubmit = vi.fn();
        render(ChatInput, { props: { ...defaultProps, onsubmit } });

        await user.type(screen.getByRole('textbox'), 'ejemplo de pregunta');
        await user.click(screen.getByTitle(/enviar/i));

        expect(onsubmit).toHaveBeenCalledWith('ejemplo de pregunta');
    });

    it('does not call onsubmit when input is empty', async () => {
        const user = userEvent.setup();
        const onsubmit = vi.fn();
        render(ChatInput, { props: { ...defaultProps, onsubmit } });

        await user.click(screen.getByTitle(/enviar/i));

        expect(onsubmit).not.toHaveBeenCalled();
    });

    it('calls onstop when stop button clicked during streaming', async () => {
        const user = userEvent.setup();
        const onstop = vi.fn();
        render(ChatInput, { props: { ...defaultProps, streaming: true, onstop } });

        await user.click(screen.getByTitle(/detener/i));

        expect(onstop).toHaveBeenCalledOnce();
    });

    it('submits on Enter key (no shift)', async () => {
        const user = userEvent.setup();
        const onsubmit = vi.fn();
        render(ChatInput, { props: { ...defaultProps, onsubmit } });

        const textarea = screen.getByRole('textbox');
        await user.type(textarea, 'pregunta via enter');
        await user.keyboard('{Enter}');

        expect(onsubmit).toHaveBeenCalledWith('pregunta via enter');
    });

    it('clears input after submit', async () => {
        const user = userEvent.setup();
        render(ChatInput, { props: defaultProps });

        const textarea = screen.getByRole('textbox');
        await user.type(textarea, 'mensaje a enviar');
        await user.keyboard('{Enter}');

        expect(textarea).toHaveValue('');
    });
});
