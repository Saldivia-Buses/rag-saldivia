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

    it('finalizeStream guarda crossdocResults en el mensaje', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('respuesta de síntesis');
        const results = [
            { query: 'presión bomba', content: 'La presión es 12 bar', success: true },
        ];
        chat.finalizeStream({ crossdocResults: results });
        expect(chat.messages[0].crossdocResults).toEqual(results);
    });

    it('finalizeStream sin opts funciona igual que antes (backwards compat)', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('respuesta normal');
        chat.finalizeStream();
        expect(chat.messages[0].crossdocResults).toBeUndefined();
        expect(chat.messages[0].content).toBe('respuesta normal');
    });

    it('addUserMessage agrega mensaje con timestamp', () => {
        const chat = new ChatStore();
        chat.addUserMessage('¿Cuál es la presión máxima?');
        expect(chat.messages).toHaveLength(1);
        expect(chat.messages[0].role).toBe('user');
        expect(chat.messages[0].content).toBe('¿Cuál es la presión máxima?');
        expect(chat.messages[0].timestamp).toBeTruthy();
    });

    it('multi-turn conversation preserva orden de mensajes', () => {
        const chat = new ChatStore();
        chat.addUserMessage('Primera pregunta');
        chat.startStream();
        chat.appendToken('Primera respuesta');
        chat.finalizeStream();
        chat.addUserMessage('Segunda pregunta');
        chat.startStream();
        chat.appendToken('Segunda respuesta');
        chat.finalizeStream();

        expect(chat.messages).toHaveLength(4);
        expect(chat.messages[0].content).toBe('Primera pregunta');
        expect(chat.messages[1].content).toBe('Primera respuesta');
        expect(chat.messages[2].content).toBe('Segunda pregunta');
        expect(chat.messages[3].content).toBe('Segunda respuesta');
    });

    it('appendToken acumula tokens con markdown', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('**Título en bold**\n\n');
        chat.appendToken('Párrafo normal.\n\n');
        chat.appendToken('- Item 1\n- Item 2');
        expect(chat.streamingContent).toBe('**Título en bold**\n\nPárrafo normal.\n\n- Item 1\n- Item 2');
    });

    it('setSources actualiza sources reactivamente', () => {
        const chat = new ChatStore();
        const sources = [
            { document: 'manual.pdf', page: 5, excerpt: 'Presión máxima 12 bar' },
            { document: 'spec.pdf', page: 10, excerpt: 'Temperatura nominal 80°C' },
        ];
        chat.setSources(sources);
        expect(chat.sources).toEqual(sources);
        expect(chat.sources).toHaveLength(2);
    });

    it('loadMessages reemplaza mensajes existentes', () => {
        const chat = new ChatStore();
        chat.addUserMessage('mensaje inicial');

        const loadedMessages = [
            { role: 'user' as const, content: 'cargado 1', timestamp: '2026-03-19T10:00:00Z' },
            { role: 'assistant' as const, content: 'cargado 2', timestamp: '2026-03-19T10:01:00Z' },
        ];
        chat.loadMessages(loadedMessages);

        expect(chat.messages).toHaveLength(2);
        expect(chat.messages[0].content).toBe('cargado 1');
        expect(chat.messages[1].content).toBe('cargado 2');
    });

    it('finalizeStream guarda mensaje con id opcional', () => {
        const store = new ChatStore();
        store.startStream();
        store.appendToken('Respuesta');
        store.finalizeStream({ messageId: 42 });
        expect(store.messages[0].id).toBe(42);
    });

    it('finalizeStream sin messageId deja id como undefined', () => {
        const store = new ChatStore();
        store.startStream();
        store.appendToken('Respuesta');
        store.finalizeStream();
        expect(store.messages[0].id).toBeUndefined();
    });
});
