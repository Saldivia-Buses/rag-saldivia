<script lang="ts">
    import { filterSessions, type SessionSummary } from '$lib/utils/chat-utils';
    import { goto } from '$app/navigation';
    import { browser } from '$app/environment';

    interface Props {
        sessions: SessionSummary[];
        currentId: string;
    }

    let { sessions, currentId }: Props = $props();

    let searchQuery = $state('');
    let editingId = $state<string | null>(null);
    let editTitle = $state('');
    let confirmDeleteId = $state<string | null>(null);
    let pinnedIds = $state<Set<string>>(new Set());

    $effect(() => {
        if (browser) {
            pinnedIds = loadPins();
        }
    });

    function loadPins(): Set<string> {
        try {
            const raw = localStorage.getItem('sda_pinned_sessions');
            return new Set(raw ? JSON.parse(raw) : []);
        } catch { return new Set(); }
    }

    function savePins() {
        localStorage.setItem('sda_pinned_sessions', JSON.stringify([...pinnedIds]));
    }

    function togglePin(id: string) {
        if (pinnedIds.has(id)) pinnedIds.delete(id);
        else pinnedIds.add(id);
        pinnedIds = new Set(pinnedIds);
        savePins();
    }

    let filtered = $derived(filterSessions(sessions, searchQuery));
    let sorted = $derived([
        ...filtered.filter(s => pinnedIds.has(s.id)),
        ...filtered.filter(s => !pinnedIds.has(s.id)),
    ]);

    async function commitRename(id: string) {
        if (!editTitle.trim()) { editingId = null; return; }
        const resp = await fetch(`/api/chat/sessions/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title: editTitle.trim() }),
        });
        if (!resp.ok) return; // mantener input abierto si falla
        editingId = null;
    }

    async function confirmDelete(id: string) {
        await fetch(`/api/chat/sessions/${id}`, { method: 'DELETE' });
        confirmDeleteId = null;
        if (id === currentId) goto('/chat');
    }
</script>

<div class="w-[200px] flex-shrink-0 bg-[var(--bg-surface)] border-r border-[var(--border)] flex flex-col overflow-hidden">
    <!-- Header con búsqueda -->
    <div class="px-2 pt-3 pb-2 flex-shrink-0">
        <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider mb-2">
            Historial
        </div>
        <input
            bind:value={searchQuery}
            type="search"
            placeholder="Buscar..."
            class="w-full bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-sm)]
                   px-2 py-1 text-xs text-[var(--text)] placeholder-[var(--text-faint)]
                   outline-none focus:border-[var(--accent)] transition-colors"
        />
    </div>

    <!-- Nueva consulta -->
    <div class="px-2 pb-1 flex-shrink-0">
        <a href="/chat" data-sveltekit-preload-data="false"
           class="flex items-center gap-1 text-xs text-[var(--accent)] hover:underline">
            + Nueva consulta
        </a>
    </div>

    <!-- Lista de sesiones -->
    <div class="flex-1 overflow-y-auto px-2 pb-2 flex flex-col gap-0.5">
        {#each sorted as session (session.id)}
            <div class="group relative rounded-[var(--radius-sm)] transition-colors
                        {session.id === currentId
                            ? 'bg-[var(--bg-hover)] border-l-2 border-[var(--accent)]'
                            : 'hover:bg-[var(--bg-hover)]'}">

                {#if editingId === session.id}
                    <input
                        bind:value={editTitle}
                        onblur={() => commitRename(session.id)}
                        onkeydown={(e) => {
                            if (e.key === 'Enter') commitRename(session.id);
                            if (e.key === 'Escape') editingId = null;
                        }}
                        aria-label="Renombrar sesión"
                        class="w-full bg-transparent border-b border-[var(--accent)] text-xs
                               text-[var(--text)] outline-none py-1 px-2"
                        autofocus
                    />
                {:else}
                    <a
                        href="/chat/{session.id}"
                        data-sveltekit-preload-data="false"
                        ondblclick={(e) => {
                            e.preventDefault();
                            editingId = session.id;
                            editTitle = session.title;
                        }}
                        class="block py-1.5 px-2 pr-12"
                    >
                        <div class="text-xs text-[var(--text-muted)] font-medium truncate leading-tight">
                            {#if pinnedIds.has(session.id)}<span class="mr-1 opacity-60">📌</span>{/if}{session.title}
                        </div>
                        <div class="text-[10px] text-[var(--text-faint)] mt-0.5">
                            {session.updated_at.slice(0, 10)}
                        </div>
                    </a>
                {/if}

                <!-- Acciones visibles en hover -->
                <div class="absolute right-1 top-1 hidden group-hover:flex gap-0.5">
                    <button
                        onclick={() => togglePin(session.id)}
                        title={pinnedIds.has(session.id) ? 'Desfijar' : 'Fijar'}
                        class="p-0.5 rounded text-[10px] text-[var(--text-faint)] hover:text-[var(--accent)]"
                    >📌</button>
                    <button
                        onclick={() => confirmDeleteId = session.id}
                        title="Eliminar"
                        class="p-0.5 rounded text-[10px] text-[var(--text-faint)] hover:text-[var(--danger,#dc2626)]"
                    >🗑</button>
                </div>
            </div>
        {/each}

        {#if filtered.length === 0 && searchQuery.trim()}
            <p class="text-xs text-[var(--text-faint)] px-2 py-4 text-center">Sin resultados</p>
        {/if}
    </div>
</div>

<!-- Modal confirm delete -->
{#if confirmDeleteId}
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-lg p-4 max-w-xs w-full mx-4">
            <p class="text-sm text-[var(--text)] mb-4">¿Eliminar esta sesión? Esta acción no se puede deshacer.</p>
            <div class="flex gap-2 justify-end">
                <button
                    onclick={() => confirmDeleteId = null}
                    class="px-3 py-1 text-xs text-[var(--text-muted)] border border-[var(--border)] rounded hover:bg-[var(--bg-hover)]"
                >Cancelar</button>
                <button
                    onclick={() => confirmDelete(confirmDeleteId!)}
                    class="px-3 py-1 text-xs text-white bg-red-500 rounded hover:bg-red-600"
                >Eliminar</button>
            </div>
        </div>
    </div>
{/if}
