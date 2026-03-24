<script lang="ts">
	import type { Source } from '$lib/stores/chat.svelte';

	interface Props {
		sources: Source[];
		open: boolean;
	}

	let { sources, open }: Props = $props();
</script>

<div
	class="flex-shrink-0 bg-[var(--bg-surface)] border-l border-[var(--border)] overflow-hidden
           transition-[width] duration-200 ease-in-out
           {open && sources.length > 0 ? 'w-64' : 'w-0'}"
>
	<!-- Contenedor con ancho fijo para que no se comprima durante la transición -->
	<div class="w-64 h-full flex flex-col p-3 overflow-y-auto">
		<div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider mb-3 flex-shrink-0">
			Fuentes ({sources.length})
		</div>

		{#each sources as source, i (source.document + i)}
			<div
				class="mb-2 rounded-[var(--radius-sm)] p-2 border-l-2 bg-[var(--bg-base)]
                       {i === 0
					? 'border-[var(--accent)]'
					: i === 1
						? 'border-[var(--accent-hover)]'
						: 'border-[var(--border)]'}"
			>
				<div class="text-[11px] text-[var(--accent)] font-semibold truncate">
					{source.document}
				</div>
				{#if source.page}
					<div class="text-[10px] text-[var(--text-faint)] mt-0.5">
						p. {source.page}
					</div>
				{/if}
				<div class="text-[10px] text-[var(--text-muted)] mt-1 line-clamp-4 leading-relaxed">
					{source.excerpt}
				</div>
			</div>
		{/each}
	</div>
</div>
