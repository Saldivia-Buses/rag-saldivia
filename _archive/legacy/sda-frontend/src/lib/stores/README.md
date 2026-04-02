# Stores

Reactive state stores using Svelte 5 runes. These are **class-based stores** (not Svelte 4 `writable()`), leveraging `$state` runes for reactivity.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `chat.svelte.ts` | ChatStore: manages chat messages, sources, streaming state, collection selection, crossdoc toggle, and abort controller. Includes helpers to add messages, start/stop streaming, and append streaming content. | `$lib/crossdoc/types` |
| `collections.svelte.ts` | CollectionsStore: manages collection list state. Provides `init()` (hydrate from server data), `load()` (fetch from BFF), `create()`, and `delete()` methods. | None |
| `crossdoc.svelte.ts` | CrossdocStore: orchestrates the 4-phase crossdoc pipeline (decompose → parallel subquery → dedup/rerank → synthesize). Tracks progress, options, and results. Includes abort support. | `$lib/crossdoc/types`, `chat.svelte` |
| `toast.svelte.ts` | ToastStore: manages toast notifications. Provides `success()`, `error()`, `info()`, `warning()` methods with auto-dismiss timers. Each toast has a unique ID. | None |

## Design notes

### Svelte 5 runes pattern

Unlike Svelte 4 stores (`writable()`, `derived()`, `get()`), these stores are **classes with `$state` runes**. This means:

- No `$store` subscriptions in components — just read the store object directly: `chat.messages`, `crossdoc.progress`
- Components reactively update when store state changes (Svelte 5 fine-grained reactivity)
- Store instances are exported as singletons: `export const chat = new ChatStore();`

Example usage:
```svelte
<script>
  import { chat } from '$lib/stores/chat.svelte';
</script>

{#each chat.messages as message}
  <div>{message.content}</div>
{/each}
```

### CrossdocStore orchestration

The `CrossdocStore.run()` method implements the full crossdoc pipeline as a sequence of fetch calls to `/api/crossdoc/*` BFF endpoints. It updates `progress` state throughout the pipeline, enabling real-time UI feedback via `CrossdocProgress.svelte`.
