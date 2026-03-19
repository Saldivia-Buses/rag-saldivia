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

    it('convierte tabla markdown a elemento table', () => {
        const markdown = `| Col1 | Col2 |
|------|------|
| A    | B    |`;
        const result = parseMarkdown(markdown);
        expect(result).toContain('<table');
        expect(result).toContain('<th>');
        expect(result).toContain('<td>');
    });

    it('aplica syntax highlighting a código Python', () => {
        const result = parseMarkdown('```python\ndef test():\n    pass\n```');
        expect(result).toContain('hljs');
        expect(result).toContain('language-python');
    });

    it('aplica syntax highlighting a código JavaScript', () => {
        const result = parseMarkdown('```javascript\nconst x = 42;\n```');
        expect(result).toContain('hljs');
        expect(result).toContain('language-javascript');
    });

    it('convierte links', () => {
        const result = parseMarkdown('[enlace](https://example.com)');
        expect(result).toContain('<a href="https://example.com">');
        expect(result).toContain('enlace</a>');
    });

    it('NO sanitiza HTML por defecto (esto se hace en componente)', () => {
        // La función parseMarkdown NO hace sanitización — eso ocurre en el componente browser
        const result = parseMarkdown('<script>alert("xss")</script>');
        // marked.parse preserva el script (sanitización se aplica en MarkdownRenderer)
        expect(result).toContain('script');
    });

    // BUG CONOCIDO: hljs.getLanguage('unknownlang') → false → usa 'plaintext' como fallback
    // pero 'plaintext' no está registrado en hljs/core → hljs.highlight() lanza error.
    // it.fails() marca la falla como esperada; cuando el bug se corrija el test empezará
    // a fallar (porque ya no lanza), recordando quitar el .fails()
    it.fails('código con lenguaje desconocido no debe lanzar error', () => {
        expect(() => parseMarkdown('```unknownlang\nconst x = 1;\n```')).not.toThrow();
    });

    // BUG CONOCIDO: mismo fallback a 'plaintext' cuando no hay lang
    it.fails('bloque de código sin lenguaje especificado no lanza error', () => {
        expect(() => parseMarkdown('```\ncontenido sin lenguaje\n```')).not.toThrow();
    });

});
