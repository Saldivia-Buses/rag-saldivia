<script lang="ts">
    import { onMount } from 'svelte';
    import { ChatStore } from '$lib/stores/chat.svelte';
    import { crossdoc } from '$lib/stores/crossdoc.svelte';
    import HistoryPanel from '$lib/components/chat/HistoryPanel.svelte';
    import MessageList from '$lib/components/chat/MessageList.svelte';
    import SourcesPanel from '$lib/components/chat/SourcesPanel.svelte';
    import ChatInput from '$lib/components/chat/ChatInput.svelte';

    let { data } = $props();
    const chat = new ChatStore();

    let selectedCollection = $state(data.session.collection ?? '');
    let sourcesOpen = $state(false);

    onMount(() => {
        chat.collection = data.session.collection ?? '';
        chat.crossdoc = data.session.crossdoc ?? false;
        if (data.session.messages?.length) {
            chat.loadMessages(data.session.messages);
        }
    });

    async function streamNormal(query: string) {
        chat.startStream();
        try {
            const resp = await fetch(`/api/chat/stream/${data.session.id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    query,
                    collection_names: [selectedCollection],
                }),
                signal: chat.abortController!.signal,
            });

            if (!resp.ok) {
                chat.appendToken('\n[Error: no se pudo conectar con el servidor]');
                return;
            }

            const reader = resp.body!.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                buffer += decoder.decode(value, { stream: true });
                const lines = buffer.split('\n');
                buffer = lines.pop() ?? '';
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const dataStr = line.slice(6).trim();
                        if (dataStr === '[DONE]') continue;
                        try {
                            const event = JSON.parse(dataStr);
                            if (event.choices?.[0]?.delta?.content) {
                                chat.appendToken(event.choices[0].delta.content);
                            }
                            if (event.citations) {
                                chat.setSources(event.citations.map((c: any) => ({
                                    document: c.source_name ?? c.document ?? '',
                                    page: c.page,
                                    excerpt: c.content ?? c.excerpt ?? '',
                                })));
                            }
                        } catch { /* ignorar errores de parse */ }
                    }
                }
            }
        } catch (err: any) {
            if (err?.name !== 'AbortError') {
                chat.appendToken('\n[Error: conexión interrumpida]');
            }
        } finally {
            chat.finalizeStream();
        }
    }

    async function sendMessage(query: string) {
        chat.addUserMessage(query);
        if (chat.crossdoc) {
            await crossdoc.run(query, chat);
        } else {
            await streamNormal(query);
        }
    }
</script>

<div class="flex h-screen overflow-hidden">
    <!-- Panel izquierdo: historial con búsqueda -->
    <HistoryPanel
        sessions={data.history}
        currentId={data.session.id}
    />

    <!-- Centro: header + mensajes + input -->
    <div class="flex-1 flex flex-col border-r border-[var(--border)] min-w-0">

        <!-- Header: colección + crossdoc + toggle fuentes -->
        <div class="flex items-center gap-2 px-3 py-2 border-b border-[var(--border)] text-xs flex-shrink-0">
            <select
                bind:value={selectedCollection}
                class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-sm)]
                       px-2 py-0.5 text-[var(--text)] outline-none focus:border-[var(--accent)]"
            >
                {#each data.collections as col}
                    <option value={col}>{col}</option>
                {/each}
            </select>

            <label class="flex items-center gap-1 text-[var(--text-muted)] cursor-pointer select-none">
                <input type="checkbox" bind:checked={chat.crossdoc} class="accent-[var(--accent)]" />
                Crossdoc
            </label>

            <button
                onclick={() => sourcesOpen = !sourcesOpen}
                disabled={chat.sources.length === 0}
                class="ml-auto px-2 py-1 rounded-[var(--radius-sm)] text-xs transition-colors
                       disabled:opacity-40
                       {sourcesOpen && chat.sources.length > 0
                           ? 'bg-[var(--accent)] text-white'
                           : 'bg-[var(--bg-surface)] border border-[var(--border)] text-[var(--text-muted)] hover:text-[var(--text)]'}"
            >
                Fuentes ({chat.sources.length})
            </button>
        </div>

        <!-- Lista de mensajes con auto-scroll -->
        <MessageList
            messages={chat.messages}
            streaming={chat.streaming}
            streamingContent={chat.streamingContent}
            crossdoc={chat.crossdoc}
        />

        <!-- Input con stop button -->
        <ChatInput
            streaming={chat.streaming}
            crossdoc={chat.crossdoc}
            onsubmit={sendMessage}
            onstop={() => { chat.stopStream(); crossdoc.stop(); }}
            oncrossdoctoggle={() => { chat.crossdoc = !chat.crossdoc; }}
        />
    </div>

    <!-- Panel derecho: fuentes (toggleable) -->
    <SourcesPanel
        sources={chat.sources}
        open={sourcesOpen}
    />
</div>
