<script lang="ts">
	import type { SubResult } from '$lib/crossdoc/types';

	interface Props {
		results: SubResult[];
	}

	let { results }: Props = $props();
	let open = $state(false);
</script>

{#if results?.length}
	<div class="border-t border-[var(--border)] mt-3 pt-2">
		<button
			onclick={() => (open = !open)}
			class="flex items-center gap-1 text-xs text-[var(--text-faint)] hover:text-[var(--text)] transition-colors"
		>
			⚡ Sub-queries usadas ({results.length}) <span>{open ? '▴' : '▾'}</span>
		</button>

		{#if open}
			<ul class="mt-2 space-y-1">
				{#each results as r}
					<li class="flex items-start gap-2 text-xs">
						<span
							class="flex-shrink-0 mt-0.5
							{r.success ? 'text-green-400' : 'text-red-400'}"
						>
							{r.success ? '✓' : '✗'}
						</span>
						<span class="text-[var(--text-faint)] leading-relaxed">{r.query}</span>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
{/if}
