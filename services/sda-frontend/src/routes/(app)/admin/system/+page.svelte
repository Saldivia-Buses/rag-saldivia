<script lang="ts">
    import { toastStore } from '$lib/stores/toast.svelte';

    let { data } = $props();

    let alerts = $state([...data.alerts]);
    let resolvingId = $state<string | null>(null);

    function formatDate(iso: string | null | undefined): string {
        if (!iso) return '—';
        const d = new Date(iso);
        const diff = Math.floor((Date.now() - d.getTime()) / 1000);
        if (diff < 3600) return `hace ${Math.floor(diff / 60)}m`;
        if (diff < 86400) return `hace ${Math.floor(diff / 3600)}h`;
        return `hace ${Math.floor(diff / 86400)}d`;
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
</script>

<div class="p-6">
    <h1 class="text-lg font-semibold text-[var(--text)] mb-6">Sistema</h1>

    <!-- Alertas de ingesta -->
    <section class="mb-8">
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
                            <th class="text-left px-3 py-2">Reintentos</th>
                            <th class="text-left px-3 py-2">Progreso</th>
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

    <!-- Otras secciones — próximamente -->
    <div class="border border-dashed border-[var(--border)] rounded-lg p-6 text-center text-sm text-[var(--text-faint)]">
        Métricas del sistema y configuración avanzada — próximamente
    </div>
</div>
