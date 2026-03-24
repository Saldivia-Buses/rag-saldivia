<!-- src/lib/components/admin/GuardrailsToggle.svelte -->
<script lang="ts">
    let { value = $bindable(), description } = $props<{
        value: boolean; description?: string;
    }>();
</script>

<div class="flex items-start gap-3 mb-4">
    <button
        type="button"
        role="switch"
        aria-label="Guardrails"
        aria-checked={value}
        onclick={() => value = !value}
        class="mt-0.5 w-9 h-5 rounded-full transition-colors flex-shrink-0
               {value ? 'bg-[var(--accent)]' : 'bg-[var(--border)]'}"
    >
        <span class="block w-4 h-4 rounded-full bg-white shadow transition-transform mx-0.5
                     {value ? 'translate-x-4' : 'translate-x-0'}"></span>
    </button>
    <div>
        <p class="text-sm font-medium text-[var(--text)]">Guardrails</p>
        {#if description}
            <p class="text-xs text-[var(--text-faint)]">{description}</p>
        {/if}
    </div>
</div>
