<script lang="ts">
    import { FileText, CheckCircle, XCircle, RefreshCw } from 'lucide-svelte';
    import TierBadge from './TierBadge.svelte';
    import type { IngestionJob } from '$lib/stores/ingestion.svelte.js';

    let { job, onRetry }: {
        job: IngestionJob;
        onRetry?: (jobId: string) => void;
    } = $props();

    const stateLabel: Record<string, string> = {
        pending:   'En cola',
        running:   'Procesando',
        completed: 'Completado',
        failed:    'Error',
        stalled:   'Sin progreso',
    };

    const isFinished = $derived(job.state === 'completed');
    const isError = $derived(job.state === 'failed' || job.state === 'stalled');
    const isActive = $derived(job.state === 'pending' || job.state === 'running');
</script>

<div class="flex items-center gap-3 py-3 border-b border-[var(--border)] last:border-0">
    <div class="shrink-0">
        {#if isFinished}
            <CheckCircle size={18} class="text-green-500" />
        {:else if isError}
            <XCircle size={18} class="text-[var(--danger)]" />
        {:else}
            <FileText size={18} class="text-[var(--text-faint)]" />
        {/if}
    </div>

    <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-0.5">
            <span class="text-sm font-medium text-[var(--text)] truncate">{job.filename}</span>
            <TierBadge tier={job.tier} />
        </div>

        <div class="flex items-center gap-2 text-xs text-[var(--text-muted)]">
            <span>{stateLabel[job.state] ?? job.state}</span>
            {#if job.pageCount}
                <span>·</span>
                <span>{job.pageCount} págs</span>
            {/if}
            {#if isActive && job.eta !== null && job.eta > 0}
                <span>·</span>
                <span>~{job.eta}s</span>
            {/if}
        </div>

        {#if isActive || isError}
            <div class="mt-1.5 h-1.5 bg-[var(--bg-hover)] rounded-full overflow-hidden">
                <div
                    class="h-full transition-all duration-500 rounded-full {isError ? 'bg-[var(--danger)]' : 'bg-[var(--accent)]'}"
                    style="width: {job.progress}%"
                ></div>
            </div>
        {/if}
    </div>

    <div class="shrink-0 text-right">
        {#if isActive}
            <span class="text-sm font-medium text-[var(--text-muted)]">{job.progress}%</span>
        {:else if isError && onRetry}
            <button
                onclick={() => onRetry?.(job.jobId)}
                class="flex items-center gap-1 text-xs text-[var(--accent)] hover:underline"
            >
                <RefreshCw size={12} />
                Reintentar
            </button>
        {/if}
    </div>
</div>
