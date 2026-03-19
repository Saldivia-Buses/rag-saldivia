# Chat Components

Components specific to the chat interface, including message rendering, input controls, crossdoc pipeline visualization, and document sources display.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `ChatInput.svelte` | Chat input textarea with submit/stop buttons and crossdoc toggle. Handles Enter-to-submit (Shift+Enter for newline). | `CrossdocSettingsPopover.svelte`, lucide-svelte |
| `CrossdocProgress.svelte` | Displays crossdoc pipeline progress as phase pills (decomposing, querying, retrying, synthesizing) with visual states (done, active, pending). | `$lib/stores/crossdoc.svelte` |
| `CrossdocSettingsPopover.svelte` | Popover chip button for toggling crossdoc mode and configuring crossdoc options (max sub-queries, synthesis model, VDB/reranker top-K, etc.). Uses clickOutside action. | `$lib/stores/crossdoc.svelte`, `$lib/actions/clickOutside` |
| `DecompositionView.svelte` | Collapsible view showing the sub-queries generated during crossdoc decomposition phase, with success/failure icons. | `$lib/crossdoc/types` |
| `HistoryPanel.svelte` | Left sidebar panel showing chat session history with search/filter functionality. Highlights the current session. | `$lib/utils/chat-utils` |
| `MarkdownRenderer.svelte` | Renders markdown content with syntax highlighting (highlight.js) and XSS sanitization (DOMPurify). Lazy-loads DOMPurify to avoid SSR crash. | `$lib/utils/markdown`, dompurify |
| `MessageList.svelte` | Scrollable message container that renders user/assistant messages, displays streaming content, shows crossdoc progress, and includes auto-scroll and "scroll to bottom" button. | `MarkdownRenderer.svelte`, `CrossdocProgress.svelte`, `DecompositionView.svelte`, `$lib/utils/scroll` |
| `SourcesPanel.svelte` | Right sidebar panel showing document sources returned by the RAG system. Top 2 sources are highlighted with accent colors. | `$lib/stores/chat.svelte` |

## Design notes

- `MarkdownRenderer.svelte` uses dynamic import for DOMPurify to prevent SSR crashes (DOMPurify requires browser APIs).
- `ChatInput.svelte` integrates the crossdoc toggle and settings popover as an inline chip, allowing users to enable/disable crossdoc mode without opening a modal.
- `MessageList.svelte` uses `isNearBottom` from scroll utils to decide when to auto-scroll vs. showing the scroll-to-bottom button.
