import { describe, it, expect } from 'vitest';
import { parseMarkdown } from './markdown.js';

describe('parseMarkdown', () => {
    it('convierte texto en negrita', () => {
        const result = parseMarkdown('**negrita**');
        expect(result).toContain('<strong>negrita</strong>');
    });

    it('convierte cursiva', () => {
        const result = parseMarkdown('*cursiva*');
        expect(result).toContain('<em>cursiva</em>');
    });

    it('convierte header h1', () => {
        const result = parseMarkdown('# Título');
        expect(result).toContain('<h1>');
        expect(result).toContain('Título');
    });

    it('convierte lista desordenada', () => {
        const result = parseMarkdown('- item uno\n- item dos');
        expect(result).toContain('<li>item uno</li>');
        expect(result).toContain('<li>item dos</li>');
    });

    it('convierte bloque de código con tag pre y code', () => {
        const result = parseMarkdown('```python\nprint("hola")\n```');
        expect(result).toContain('<pre>');
        expect(result).toContain('<code');
    });

    it('retorna string, no promesa', () => {
        const result = parseMarkdown('**texto**');
        expect(typeof result).toBe('string');
    });

    it('maneja string vacío sin error', () => {
        expect(() => parseMarkdown('')).not.toThrow();
    });
});
