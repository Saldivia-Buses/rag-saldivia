import { marked } from 'marked';
import { markedHighlight } from 'marked-highlight';
import hljs from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';
import python from 'highlight.js/lib/languages/python';
import javascript from 'highlight.js/lib/languages/javascript';
import typescript from 'highlight.js/lib/languages/typescript';
import json from 'highlight.js/lib/languages/json';
import yaml from 'highlight.js/lib/languages/yaml';
import sql from 'highlight.js/lib/languages/sql';

// Registrar lenguajes para syntax highlighting
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('python', python);
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('json', json);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('sql', sql);

// Configurar marked con highlight.js (una sola vez al cargar el módulo)
marked.use(markedHighlight({
    langPrefix: 'hljs language-',
    highlight(code, lang) {
        if (lang && hljs.getLanguage(lang)) {
            try {
                return hljs.highlight(code, { language: lang }).value;
            } catch {
                // lenguaje registrado pero highlight falló — fallback sin color
            }
        }
        // Lenguaje desconocido o sin lenguaje: devolver code sin modificar.
        // marked-highlight detecta que no cambió y escapa el HTML por defecto.
        return code;
    }
}));

/**
 * Convierte Markdown a HTML. Sanitización XSS se aplica
 * en el componente MarkdownRenderer (solo en browser).
 */
export function parseMarkdown(content: string): string {
    return marked.parse(content) as string;
}
