<script lang="ts">
    interface Props {
        suggestions: string[];
        onselect: (q: string) => void;
    }
    let { suggestions, onselect }: Props = $props();
</script>

{#if suggestions.length > 0}
    <div class="flex flex-wrap gap-1 mt-2">
        {#each suggestions as s}
            <button
                onclick={() => onselect(s)}
                class="text-[10px] px-2 py-0.5 rounded-full border border-[var(--border)]
                       text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--accent)]
                       transition-colors bg-[var(--bg-base)]"
            >{s}</button>
        {/each}
    </div>
{/if}
