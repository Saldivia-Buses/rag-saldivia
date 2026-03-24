<script lang="ts">
    import { parseMarkdown } from '$lib/utils/markdown';

    let { content }: { content: string } = $props();

    let sanitizedHtml = $state('');

    $effect(async () => {
        const raw = parseMarkdown(content);
        // DOMPurify requiere browser — import dinámico evita crash en SSR
        try {
            const { default: DOMPurify } = await import('dompurify');
            sanitizedHtml = DOMPurify.sanitize(raw, {
                ALLOWED_TAGS: [
                    'p', 'br', 'strong', 'em', 'del', 'code', 'pre',
                    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
                    'ul', 'ol', 'li', 'blockquote', 'hr',
                    'table', 'thead', 'tbody', 'tr', 'th', 'td',
                    'a', 'span', 'div'
                ],
                ALLOWED_ATTR: ['href', 'class', 'target', 'rel'],
            });
        } catch {
            // Fallback sin sanitización (solo en SSR o si DOMPurify no carga)
            sanitizedHtml = raw;
        }
    });
</script>

<div class="markdown-body text-xs text-[var(--text)] leading-relaxed">
    {@html sanitizedHtml}
</div>

<style>
    .markdown-body :global(h1),
    .markdown-body :global(h2),
    .markdown-body :global(h3) {
        color: var(--text);
        font-weight: 600;
        margin: 0.6em 0 0.2em;
        line-height: 1.3;
    }
    .markdown-body :global(h1) { font-size: 1.1em; }
    .markdown-body :global(h2) { font-size: 1.0em; }
    .markdown-body :global(h3) { font-size: 0.95em; }

    .markdown-body :global(p) { margin: 0.4em 0; line-height: 1.7; }

    .markdown-body :global(ul),
    .markdown-body :global(ol) { padding-left: 1.25em; margin: 0.3em 0; }

    .markdown-body :global(li) { margin: 0.15em 0; }

    .markdown-body :global(strong) { color: var(--text); font-weight: 600; }

    .markdown-body :global(code) {
        background: var(--bg-base);
        border: 1px solid var(--border);
        border-radius: 3px;
        padding: 0.1em 0.35em;
        font-family: 'Cascadia Code', 'Fira Code', monospace;
        font-size: 0.85em;
    }

    .markdown-body :global(pre) {
        background: var(--bg-base);
        border: 1px solid var(--border);
        border-radius: var(--radius-md);
        padding: 0.75em 1em;
        overflow-x: auto;
        margin: 0.5em 0;
    }

    .markdown-body :global(pre code) {
        background: none;
        border: none;
        padding: 0;
        font-size: 0.8em;
        line-height: 1.6;
    }

    .markdown-body :global(table) {
        border-collapse: collapse;
        width: 100%;
        margin: 0.5em 0;
        font-size: 0.85em;
    }

    .markdown-body :global(th),
    .markdown-body :global(td) {
        border: 1px solid var(--border);
        padding: 0.3em 0.6em;
        text-align: left;
    }

    .markdown-body :global(th) {
        background: var(--bg-surface);
        font-weight: 600;
        color: var(--text-muted);
    }

    .markdown-body :global(a) {
        color: var(--accent);
        text-decoration: none;
    }

    .markdown-body :global(a:hover) { text-decoration: underline; }

    .markdown-body :global(blockquote) {
        border-left: 3px solid var(--accent);
        padding-left: 0.75em;
        margin: 0.4em 0;
        color: var(--text-muted);
        font-style: italic;
    }

    .markdown-body :global(hr) {
        border: none;
        border-top: 1px solid var(--border);
        margin: 0.5em 0;
    }

    /* highlight.js colors — tema oscuro custom */
    .markdown-body :global(.hljs) { color: #abb2bf; background: transparent; }
    .markdown-body :global(.hljs-keyword),
    .markdown-body :global(.hljs-selector-tag) { color: #c678dd; }
    .markdown-body :global(.hljs-string) { color: #98c379; }
    .markdown-body :global(.hljs-number) { color: #d19a66; }
    .markdown-body :global(.hljs-comment) { color: #5c6370; font-style: italic; }
    .markdown-body :global(.hljs-title),
    .markdown-body :global(.hljs-function) { color: #61afef; }
    .markdown-body :global(.hljs-attr) { color: #e06c75; }
    .markdown-body :global(.hljs-literal) { color: #56b6c2; }
    .markdown-body :global(.hljs-built_in) { color: #e6c07b; }
    .markdown-body :global(.hljs-variable) { color: #e06c75; }
</style>
