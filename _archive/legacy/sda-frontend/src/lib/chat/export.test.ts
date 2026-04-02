// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { buildMarkdown, buildJSON } from './export.js';

const session = { id: 'ses-1', title: 'Mi chat', created_at: '2026-03-23T10:00:00Z' };
const messages = [
    { role: 'user' as const, content: '¿Qué dice el manual?', timestamp: '2026-03-23T10:01:00Z' },
    { role: 'assistant' as const, content: 'El manual indica...', timestamp: '2026-03-23T10:01:05Z' },
];

describe('buildMarkdown', () => {
    it('incluye el título de la sesión', () => {
        const md = buildMarkdown(session, messages);
        expect(md).toContain('# Mi chat');
    });

    it('incluye ambos mensajes con rol correcto', () => {
        const md = buildMarkdown(session, messages);
        expect(md).toContain('¿Qué dice el manual?');
        expect(md).toContain('El manual indica...');
        expect(md).toContain('**Vos:**');
        expect(md).toContain('**SDA:**');
    });
});

describe('buildJSON', () => {
    it('retorna JSON parseable con session y messages', () => {
        const raw = buildJSON(session, messages);
        const parsed = JSON.parse(raw);
        expect(parsed.session.title).toBe('Mi chat');
        expect(parsed.messages).toHaveLength(2);
    });

    it('el JSON incluye el role de cada mensaje', () => {
        const raw = buildJSON(session, messages);
        const parsed = JSON.parse(raw);
        expect(parsed.messages[0].role).toBe('user');
        expect(parsed.messages[1].role).toBe('assistant');
    });
});

describe('downloadFile', () => {
    it('invoca createObjectURL con el contenido', async () => {
        const createObjectURL = vi.fn(() => 'blob:fake');
        const revokeObjectURL = vi.fn();
        vi.stubGlobal('URL', { createObjectURL, revokeObjectURL });
        const a = { href: '', download: '', click: vi.fn() };
        vi.spyOn(document, 'createElement').mockReturnValueOnce(a as any);
        vi.spyOn(document.body, 'appendChild').mockImplementationOnce(() => a as any);
        vi.spyOn(document.body, 'removeChild').mockImplementationOnce(() => a as any);

        const { downloadFile } = await import('./export.js');
        downloadFile('test content', 'file.md', 'text/markdown');

        expect(createObjectURL).toHaveBeenCalledWith(expect.any(Blob));
        expect(a.click).toHaveBeenCalled();
        expect(revokeObjectURL).toHaveBeenCalledWith('blob:fake');
        vi.restoreAllMocks();
        vi.unstubAllGlobals();
    });
});
