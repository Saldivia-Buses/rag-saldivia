import { describe, it, expect, vi } from 'vitest';
import { ChatStore } from './chat.svelte.js';

describe('ChatStore', () => {
    it('startStream crea abortController y activa streaming', () => {
        const chat = new ChatStore();
        chat.startStream();
        expect(chat.abortController).not.toBeNull();
        expect(chat.abortController).toBeInstanceOf(AbortController);
        expect(chat.streaming).toBe(true);
        expect(chat.streamingContent).toBe('');
        expect(chat.sources).toHaveLength(0);
    });

    it('stopStream llama abort y deja streaming en false', () => {
        const chat = new ChatStore();
        chat.startStream();
        const controller = chat.abortController!;
        const abortSpy = vi.spyOn(controller, 'abort');
        chat.stopStream();
        expect(abortSpy).toHaveBeenCalledOnce();
        expect(chat.streaming).toBe(false);
        expect(chat.streamingContent).toBe('');
    });

    it('stopStream con contenido parcial guarda lo que llegó', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('texto parcial');
        chat.stopStream();
        expect(chat.messages).toHaveLength(1);
        expect(chat.messages[0].content).toBe('texto parcial');
        expect(chat.messages[0].role).toBe('assistant');
    });

    it('stopStream sin startStream previo no lanza error', () => {
        const chat = new ChatStore();
        expect(() => chat.stopStream()).not.toThrow();
    });
});
