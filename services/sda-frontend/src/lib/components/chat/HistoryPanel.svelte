<script lang="ts">
    import { filterSessions, type SessionSummary } from '$lib/utils/chat-utils';

    interface Props {
        sessions: SessionSummary[];
        currentId: string;
    }

    let { sessions, currentId }: Props = $props();

    let searchQuery = $state('');
    let filtered = $derived(filterSessions(sessions, searchQuery));
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
        <a
            href="/chat"
            data-sveltekit-preload-data="false"
            class="flex items-center gap-1 text-xs text-[var(--accent)] hover:underline"
        >
            + Nueva consulta
        </a>
    </div>

    <!-- Lista de sesiones -->
    <div class="flex-1 overflow-y-auto px-2 pb-2 flex flex-col gap-0.5">
        {#each filtered as session (session.id)}
            <a
                href="/chat/{session.id}"
                data-sveltekit-preload-data="false"
                class="block rounded-[var(--radius-sm)] px-2 py-1.5 transition-colors
                       {session.id === currentId
                           ? 'bg-[var(--bg-hover)] border-l-2 border-[var(--accent)] pl-[7px]'
                           : 'hover:bg-[var(--bg-hover)]'}"
            >
                <div class="text-xs text-[var(--text-muted)] font-medium truncate leading-tight">
                    {session.title}
                </div>
                <div class="text-[10px] text-[var(--text-faint)] mt-0.5">
                    {session.updated_at.slice(0, 10)}
                </div>
            </a>
        {/each}

        {#if filtered.length === 0 && searchQuery.trim()}
            <p class="text-xs text-[var(--text-faint)] px-2 py-4 text-center">
                Sin resultados
            </p>
        {/if}
    </div>
</div>
