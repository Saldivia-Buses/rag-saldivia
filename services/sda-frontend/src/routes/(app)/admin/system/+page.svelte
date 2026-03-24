<script lang="ts">
    import { invalidateAll } from '$app/navigation';
    import { toastStore } from '$lib/stores/toast.svelte';

    let { data } = $props();

    let alerts = $state([...data.alerts]);
    let resolvingId = $state<string | null>(null);
    let refreshing = $state(false);

    function formatDate(iso: string | null | undefined): string {
        if (!iso) return '—';
        const d = new Date(iso);
        const diff = Math.floor((Date.now() - d.getTime()) / 1000);
        if (diff < 3600) return `hace ${Math.floor(diff / 60)}m`;
        if (diff < 86400) return `hace ${Math.floor(diff / 3600)}h`;
        return `hace ${Math.floor(diff / 86400)}d`;
    }

    function jobStateBadge(state: string): string {
        const map: Record<string, string> = {
            running: 'bg-blue-500/15 text-blue-600',
            pending: 'bg-[var(--bg-surface)] text-[var(--text-faint)]',
            queued:  'bg-[var(--bg-surface)] text-[var(--text-faint)]',
            done:    'bg-green-500/15 text-green-600',
            failed:  'bg-red-500/15 text-[var(--danger)]',
        };
        return map[state] ?? 'bg-[var(--bg-surface)] text-[var(--text-faint)]';
    }

    async function resolveAlert(id: string) {
        resolvingId = id;
        try {
            const resp = await fetch(`/api/admin/alerts/${id}/resolve`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({}),
            });
            if (resp.ok) {
                alerts = alerts.filter(a => a.id !== id);
            } else {
                toastStore.error('No se pudo resolver la alerta. Intentá de nuevo.');
            }
        } catch {
            toastStore.error('Error al contactar el servidor.');
        } finally {
            resolvingId = null;
        }
    }

    async function refresh() {
        refreshing = true;
        try {
            await invalidateAll();
            alerts = [...data.alerts];
        } finally {
            refreshing = false;
        }
    }
</script>

<div class="p-6">
    <div class="flex items-center justify-between mb-6">
        <h1 class="text-lg font-semibold text-[var(--text)]">Sistema</h1>
        <button
            onclick={refresh}
            disabled={refreshing}
            class="text-xs text-[var(--text-muted)] border border-[var(--border)] px-3 py-1.5
                   rounded hover:border-[var(--accent)] hover:text-[var(--accent)] transition-colors
                   disabled:opacity-50"
        >
            {refreshing ? 'Actualizando…' : '↻ Actualizar'}
        </button>
    </div>

    <!-- Stats cards -->
    <div class="grid grid-cols-3 gap-4 mb-8">
        <div class="border border-[var(--border)] rounded-lg p-4">
            <p class="text-xs text-[var(--text-faint)] mb-1">Usuarios activos</p>
            <p class="text-2xl font-semibold text-[var(--text)]">
                {data.stats.activeUsers ?? '—'}
            </p>
        </div>
        <div class="border border-[var(--border)] rounded-lg p-4">
            <p class="text-xs text-[var(--text-faint)] mb-1">Áreas</p>
            <p class="text-2xl font-semibold text-[var(--text)]">
                {data.stats.totalAreas ?? '—'}
            </p>
        </div>
        <div class="border border-[var(--border)] rounded-lg p-4">
            <p class="text-xs text-[var(--text-faint)] mb-1">Colecciones con documentos</p>
            <p class="text-2xl font-semibold text-[var(--text)]">
                {data.stats.collectionsWithDocs ?? '—'}
            </p>
        </div>
    </div>

    <!-- Jobs activos -->
    <section class="mb-8">
        <h2 class="text-sm font-semibold text-[var(--text)] mb-3">Jobs de ingesta activos</h2>

        {#if data.activeJobs.length === 0}
            <div class="border border-[var(--border)] rounded-lg p-6 text-center text-sm text-[var(--text-muted)]">
                Sin jobs activos
            </div>
        {:else}
            <div class="border border-[var(--border)] rounded-lg overflow-hidden">
                <table class="w-full text-sm">
                    <thead>
                        <tr class="text-[var(--text-faint)] text-xs border-b border-[var(--border)] bg-[var(--bg-surface)]">
                            <th class="text-left px-3 py-2">Archivo</th>
                            <th class="text-left px-3 py-2">Colección</th>
                            <th class="text-left px-3 py-2">Estado</th>
                            <th class="text-left px-3 py-2 w-40">Progreso</th>
                            <th class="text-left px-3 py-2">Iniciado</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each data.activeJobs as job (job.id)}
                            <tr class="border-b border-[var(--border)] last:border-0 text-[var(--text-muted)]">
                                <td class="px-3 py-2 max-w-[200px]">
                                    <span class="truncate block" title={job.filename}>{job.filename}</span>
                                </td>
                                <td class="px-3 py-2">{job.collection}</td>
                                <td class="px-3 py-2">
                                    <span class="text-xs px-2 py-0.5 rounded-full {jobStateBadge(job.state)}">
                                        {job.state}
                                    </span>
                                </td>
                                <td class="px-3 py-2">
                                    <div class="flex items-center gap-2">
                                        <div class="flex-1 bg-[var(--bg-surface)] rounded-full h-1.5 overflow-hidden">
                                            <div
                                                class="h-full bg-[var(--accent)] rounded-full transition-all"
                                                style="width: {Math.min(job.progress, 100)}%"
                                            ></div>
                                        </div>
                                        <span class="text-xs text-[var(--text-faint)] w-8 text-right">
                                            {Math.round(job.progress)}%
                                        </span>
                                    </div>
                                </td>
                                <td class="px-3 py-2 text-xs">{formatDate(job.created_at)}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </section>

    <!-- Alertas de ingesta -->
    <section>
        <div class="flex items-center gap-2 mb-3">
            <h2 class="text-sm font-semibold text-[var(--text)]">Alertas de ingesta</h2>
            {#if alerts.length > 0}
                <span class="bg-amber-500/15 text-amber-600 text-xs font-medium px-2 py-0.5 rounded-full">
                    {alerts.length}
                </span>
            {/if}
        </div>

        {#if alerts.length === 0}
            <div class="border border-[var(--border)] rounded-lg p-6 text-center text-sm text-[var(--text-muted)]">
                Sin alertas pendientes
            </div>
        {:else}
            <div class="border border-[var(--border)] rounded-lg overflow-hidden">
                <table class="w-full text-sm">
                    <thead>
                        <tr class="text-[var(--text-faint)] text-xs border-b border-[var(--border)] bg-[var(--bg-surface)]">
                            <th class="text-left px-3 py-2">Archivo</th>
                            <th class="text-left px-3 py-2">Colección</th>
                            <th class="text-left px-3 py-2">Error</th>
                            <th class="text-center px-3 py-2">Reintentos</th>
                            <th class="text-center px-3 py-2">Progreso</th>
                            <th class="text-left px-3 py-2">Fecha</th>
                            <th class="px-3 py-2"></th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each alerts as alert (alert.id)}
                            <tr class="border-b border-[var(--border)] last:border-0 text-[var(--text-muted)]">
                                <td class="px-3 py-2 max-w-[180px]">
                                    <span class="truncate block" title={alert.filename}>{alert.filename}</span>
                                </td>
                                <td class="px-3 py-2">{alert.collection}</td>
                                <td class="px-3 py-2 max-w-[200px]">
                                    <span class="truncate block text-[var(--danger)] text-xs" title={alert.error ?? ''}>
                                        {alert.error ?? '—'}
                                    </span>
                                </td>
                                <td class="px-3 py-2 text-center">{alert.retry_count}</td>
                                <td class="px-3 py-2 text-center">{alert.progress_at_failure}%</td>
                                <td class="px-3 py-2 text-xs">{formatDate(alert.created_at)}</td>
                                <td class="px-3 py-2">
                                    <button
                                        onclick={() => resolveAlert(alert.id)}
                                        disabled={resolvingId === alert.id}
                                        class="text-xs px-2 py-1 rounded border border-[var(--border)]
                                               text-[var(--text-muted)] hover:border-[var(--accent)]
                                               hover:text-[var(--accent)] transition-colors disabled:opacity-50"
                                    >
                                        {resolvingId === alert.id ? '…' : 'Resolver'}
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </section>
</div>
