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
});
