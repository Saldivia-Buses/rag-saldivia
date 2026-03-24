<!-- src/lib/components/admin/ConfigSlider.svelte -->
<script lang="ts">
    let { label, value = $bindable(), min, max, step = 0.1, description } = $props<{
        label: string; value: number; min: number; max: number;
        step?: number; description?: string;
    }>();
</script>

<div class="mb-4">
    <div class="flex justify-between mb-1">
        <label for="slider-{label.toLowerCase().replace(/\s+/g, '-')}" class="text-sm font-medium text-[var(--text)]">{label}</label>
        <input
            type="number"
            bind:value
            {min} {max} {step}
            aria-label="{label} valor"
            class="w-20 text-right text-sm bg-transparent border-b border-[var(--border)]
                   text-[var(--text)] outline-none focus:border-[var(--accent)]"
        />
    </div>
    <input type="range" id="slider-{label.toLowerCase().replace(/\s+/g, '-')}" bind:value {min} {max} {step} class="w-full accent-[var(--accent)]" />
    {#if description}
        <p class="text-xs text-[var(--text-faint)] mt-1">{description}</p>
    {/if}
</div>
