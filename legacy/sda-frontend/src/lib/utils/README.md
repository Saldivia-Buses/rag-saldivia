# Utils

Utility functions used across the frontend.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `markdown.ts` | Markdown parsing and rendering. Configures `marked` with `highlight.js` for syntax highlighting. Exports `parseMarkdown(text)` function that converts markdown to HTML. | marked, marked-highlight, highlight.js |
| `scroll.ts` | Scroll utilities. Exports `isNearBottom(el, threshold)` function to detect if a scrollable element is near the bottom (used for auto-scroll logic in MessageList). | None |
| `chat-utils.ts` | Chat-specific utilities. Exports `filterSessions(sessions, query)` for case-insensitive substring search on session titles. | None |

## Design notes

### XSS protection in markdown rendering

`markdown.ts` parses markdown to HTML but does NOT sanitize it. Sanitization is handled by `MarkdownRenderer.svelte` using **DOMPurify** before rendering the HTML to the DOM. This separation ensures:

1. `markdown.ts` is a pure utility (no browser-only dependencies)
2. Sanitization happens only in the browser (DOMPurify requires `window`)
3. SSR-safe (DOMPurify is dynamically imported in the component)

The sanitization in `MarkdownRenderer.svelte` allows only safe HTML tags and attributes, preventing XSS attacks.
