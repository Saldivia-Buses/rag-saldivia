import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import MessageList from './MessageList.svelte';

let msgCounter = 0;
function makeMsg(role: 'user' | 'assistant', content: string) {
    // timestamps únicos para evitar each_key_duplicate en Svelte {#each msg (msg.timestamp)}
    return { role, content, timestamp: new Date(Date.now() + msgCounter++).toISOString() };
}

const baseProps = {
    messages: [],
    streaming: false,
    streamingContent: '',
    crossdoc: false,
};

describe('MessageList', () => {
    it('renders without errors when messages is empty', () => {
        const { container } = render(MessageList, { props: baseProps });
        expect(container).toBeTruthy();
    });

    it('renders user message content directly', () => {
        render(MessageList, {
            props: { ...baseProps, messages: [makeMsg('user', 'hola mundo')] },
        });
        // User messages are rendered as plain text, not via MarkdownRenderer
        expect(screen.getByText('hola mundo')).toBeInTheDocument();
    });

    it('renders multiple user messages', () => {
        render(MessageList, {
            props: {
                ...baseProps,
                messages: [makeMsg('user', 'primera'), makeMsg('user', 'segunda')],
            },
        });
        expect(screen.getByText('primera')).toBeInTheDocument();
        expect(screen.getByText('segunda')).toBeInTheDocument();
    });

    it('renders assistant message via MarkdownRenderer (async)', async () => {
        render(MessageList, {
            props: { ...baseProps, messages: [makeMsg('assistant', 'respuesta del asistente')] },
        });
        // MarkdownRenderer uses DOMPurify (dynamic import) → findByText waits for async render
        expect(await screen.findByText(/respuesta del asistente/)).toBeInTheDocument();
    });

    it('shows streaming cursor ▋ when streaming with content', () => {
        render(MessageList, {
            props: { ...baseProps, streaming: true, streamingContent: 'texto parcial' },
        });
        expect(screen.getByText('▋')).toBeInTheDocument();
    });

    it('shows animated pulse avatar when streaming without content', () => {
        const { container } = render(MessageList, {
            props: { ...baseProps, streaming: true, streamingContent: '' },
        });
        // Avatar div has animate-pulse class during streaming
        expect(container.querySelector('.animate-pulse')).toBeInTheDocument();
    });

    it('does not show cursor ▋ when not streaming', () => {
        render(MessageList, {
            props: baseProps,
        });
        expect(screen.queryByText('▋')).toBeNull();
    });

    it('renderiza FeedbackButtons para mensajes assistant', async () => {
        const msgs = [
            { id: 1, role: 'assistant' as const, content: 'Respuesta', timestamp: '2026-01-01T00:00:00Z' },
        ];
        const { container } = render(MessageList, {
            props: {
                messages: msgs,
                streaming: false,
                streamingContent: '',
                crossdoc: false,
                sessionId: 'ses-1',
                userId: 7,
            }
        });
        // FeedbackButtons renderiza 2 botones (👍 👎)
        const buttons = container.querySelectorAll('button');
        expect(buttons.length).toBeGreaterThanOrEqual(2);
    });

    it('no renderiza FeedbackButtons para mensajes sin id', () => {
        const msgs = [
            { role: 'assistant' as const, content: 'Respuesta sin id', timestamp: '2026-01-01T00:00:00Z' },
            // sin campo id
        ];
        const { container } = render(MessageList, {
            props: {
                messages: msgs,
                streaming: false,
                streamingContent: '',
                crossdoc: false,
                sessionId: 'ses-1',
                userId: 7,
            }
        });
        // Si no hay id, no debe haber botones de feedback (👍/👎)
        const feedbackButtons = container.querySelectorAll('button[title="Útil"], button[title="No útil"]');
        expect(feedbackButtons).toHaveLength(0);
    });

    it('no muestra follow-ups al cargar historial inicial (sin streaming)', () => {
        const msgs = [
            { id: 1, role: 'user' as const, content: '¿Qué dice el doc?', timestamp: '2026-01-01T00:00:00Z' },
            { id: 2, role: 'assistant' as const, content: 'El documento dice algo. También dice otra cosa. Y una tercera cosa importante aquí.', timestamp: '2026-01-01T00:00:01Z' },
        ];
        const { container } = render(MessageList, {
            props: {
                messages: msgs,
                streaming: false,
                streamingContent: '',
                crossdoc: false,
                sessionId: 'ses-1',
                userId: 7,
            }
        });
        // En carga inicial (streaming nunca fue true), no debe haber chips de follow-up
        const chips = container.querySelectorAll('button.rounded-full');
        expect(chips).toHaveLength(0);
    });
});
