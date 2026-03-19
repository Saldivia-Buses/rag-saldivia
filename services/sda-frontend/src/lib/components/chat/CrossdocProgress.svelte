<script lang="ts">
	import { crossdoc } from '$lib/stores/crossdoc.svelte';

	const STEPS = ['decomposing', 'querying', 'retrying', 'synthesizing'] as const;
	const LABELS: Record<string, string> = {
		decomposing: 'Analizando pregunta',
		querying: 'Consultando documentos',
		retrying: 'Reintentando fallidos',
		synthesizing: 'Sintetizando respuesta',
	};

	const p = $derived(crossdoc.progress);

	function stepState(step: string): 'done' | 'active' | 'pending' {
		if (!p) return 'pending';
		const currentIdx = STEPS.indexOf(p.phase as any);
		const stepIdx = STEPS.indexOf(step as any);
		if (stepIdx < currentIdx) return 'done';
		if (stepIdx === currentIdx) return 'active';
		return 'pending';
	}
</script>

{#if p && p.phase !== 'done' && p.phase !== 'error'}
	<div class="space-y-2 py-2">
		<!-- Phase pills -->
		<div class="flex flex-wrap gap-2">
			{#each STEPS as step}
				{@const state = stepState(step)}
				<span
					class="text-xs px-2 py-0.5 rounded-full border transition-colors
					{state === 'done'
						? 'bg-green-500/10 text-green-400 border-green-500/30'
						: ''}
					{state === 'active'
						? 'bg-[var(--accent)]/10 text-[var(--accent)] border-[var(--accent)]/30 animate-pulse'
						: ''}
					{state === 'pending'
						? 'bg-transparent text-[var(--text-faint)] border-[var(--border)]'
						: ''}"
				>
					{state === 'done' ? '✓' : state === 'active' ? '⟳' : '○'}
					{LABELS[step]}
				</span>
			{/each}
		</div>

		<!-- Numeric progress bar (querying and retrying phases only) -->
		{#if (p.phase === 'querying' || p.phase === 'retrying') && p.total > 0}
			<div class="space-y-1">
				<div class="h-1 bg-[var(--bg-surface)] rounded-full overflow-hidden">
					<div
						class="h-full bg-[var(--accent)] rounded-full transition-all duration-300"
						style="width: {(p.completed / p.total) * 100}%"
					></div>
				</div>
				<p class="text-xs text-[var(--text-faint)]">{p.completed} / {p.total} sub-queries</p>
			</div>
		{/if}
	</div>
{/if}
