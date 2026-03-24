<!-- src/lib/components/admin/ModelSelector.svelte -->
<script lang="ts">
    let { label, value = $bindable(), options, description } = $props<{
        label: string; value: string; options: string[]; description?: string;
    }>();
</script>

<div class="mb-4">
    <label for="model-{label.toLowerCase().replace(/\s+/g, '-')}" class="block text-sm font-medium text-[var(--text)] mb-1">{label}</label>
    <select
        id="model-{label.toLowerCase().replace(/\s+/g, '-')}"
        bind:value
        class="w-full text-sm bg-[var(--bg-surface)] border border-[var(--border)] rounded
               px-2 py-1.5 text-[var(--text)] outline-none focus:border-[var(--accent)]"
    >
        {#each options as opt}
            <option value={opt}>{opt}</option>
        {/each}
    </select>
    {#if description}
        <p class="text-xs text-[var(--text-faint)] mt-1">{description}</p>
    {/if}
</div>
