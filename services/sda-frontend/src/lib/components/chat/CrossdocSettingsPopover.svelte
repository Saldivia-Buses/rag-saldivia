<script lang="ts">
	import { crossdoc } from '$lib/stores/crossdoc.svelte';
	import { clickOutside } from '$lib/actions/clickOutside';

	interface Props {
		active: boolean;
		ontoggle: () => void;
	}

	let { active, ontoggle }: Props = $props();
	let open = $state(false);
</script>

<div class="relative" use:clickOutside={() => (open = false)}>
	<!-- Chip toggle button -->
	<button
		onclick={() => {
			ontoggle();
		}}
		class="flex items-center gap-1 text-xs px-2 py-1 rounded-full border transition-colors
               {active
			? 'bg-[var(--accent)]/10 text-[var(--accent)] border-[var(--accent)]/40'
			: 'bg-transparent text-[var(--text-faint)] border-[var(--border)] hover:border-[var(--accent)]/40'}"
		title="Click: toggle Crossdoc"
	>
		⚡ Crossdoc
		<span
			class="opacity-50 text-[10px]"
			role="button"
			tabindex="0"
			onclick|stopPropagation={() => (open = !open)}
			onkeydown={(e) => e.key === 'Enter' && (open = !open)}
		>▾</span>
	</button>

	<!-- Settings popover -->
	{#if open}
		<div
			class="absolute bottom-full mb-2 right-0 w-56
                    bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-md)]
                    shadow-lg p-3 space-y-3 z-50"
		>
			<p class="text-xs font-semibold text-[var(--text)]">⚡ Crossdoc Settings</p>

			<label class="flex items-center justify-between gap-2">
				<span class="text-xs text-[var(--text-faint)]">Max sub-queries</span>
				<input
					type="number"
					min="0"
					max="20"
					bind:value={crossdoc.options.maxSubQueries}
					class="w-14 text-xs text-center bg-[var(--bg)] border border-[var(--border)]
                           rounded px-1 py-0.5 text-[var(--text)]"
				/>
			</label>

			<label class="flex items-center justify-between gap-2">
				<span class="text-xs text-[var(--text-faint)]">Synthesis model</span>
				<input
					type="text"
					bind:value={crossdoc.options.synthesisModel}
					placeholder="(default)"
					class="w-24 text-xs bg-[var(--bg)] border border-[var(--border)]
                           rounded px-1 py-0.5 text-[var(--text)] placeholder-[var(--text-faint)]"
				/>
			</label>

			<label class="flex items-center justify-between gap-2">
				<span class="text-xs text-[var(--text-faint)]">Follow-up retries</span>
				<input
					type="checkbox"
					bind:checked={crossdoc.options.followUpRetries}
					class="accent-[var(--accent)]"
				/>
			</label>

			<label class="flex items-center justify-between gap-2">
				<span class="text-xs text-[var(--text-faint)]">Show decomposition</span>
				<input
					type="checkbox"
					bind:checked={crossdoc.options.showDecomposition}
					class="accent-[var(--accent)]"
				/>
			</label>
		</div>
	{/if}
</div>
