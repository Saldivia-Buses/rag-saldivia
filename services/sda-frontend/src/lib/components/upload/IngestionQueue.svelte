<script lang="ts">
    import JobCard from './JobCard.svelte';
    import type { IngestionJob } from '$lib/stores/ingestion.svelte.js';

    let { jobs, onRetry }: {
        jobs: IngestionJob[];
        onRetry?: (jobId: string) => void;
    } = $props();

    // Activos primero, luego completados (últimos 5), luego fallados
    const sortedJobs = $derived([
        ...jobs
            .filter(j => j.state === 'pending' || j.state === 'running')
            .sort((a, b) => b.progress - a.progress),
        ...jobs
            .filter(j => j.state === 'completed')
            .slice(0, 5),
        ...jobs
            .filter(j => j.state === 'failed' || j.state === 'stalled'),
    ]);
</script>

{#if sortedJobs.length > 0}
    <div class="mt-6">
        <h2 class="text-xs font-medium text-[var(--text-muted)] uppercase tracking-wide mb-2">
            Cola de ingesta
        </h2>
        <div class="border border-[var(--border)] rounded-[var(--radius-lg)] divide-y divide-[var(--border)] px-3">
            {#each sortedJobs as job (job.jobId)}
                <JobCard {job} {onRetry} />
            {/each}
        </div>
    </div>
{/if}
