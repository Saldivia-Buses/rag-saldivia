<!-- services/sda-frontend/src/lib/components/ui/Select.svelte -->
<script lang="ts">
    interface Props {
        value?: string;
        options: Array<{ value: string; label: string }>;
        placeholder?: string;
        disabled?: boolean;
        label?: string;
        error?: string;
        class?: string;
        onchange?: (value: string) => void;
    }

    let {
        value = $bindable(''),
        options,
        placeholder = 'Seleccionar...',
        disabled = false,
        label,
        error,
        class: className = '',
        onchange,
    }: Props = $props();

    const selectId = `select-${Math.random().toString(36).slice(2, 9)}`;

    function handleChange(e: Event) {
        const target = e.target as HTMLSelectElement;
        value = target.value;
        onchange?.(target.value);
    }
</script>

<div class="flex flex-col gap-1 {className}">
    {#if label}
        <label for={selectId} class="text-xs font-medium text-[var(--text-muted)]">{label}</label>
    {/if}
    <select
        id={selectId}
        bind:value
        {disabled}
        onchange={handleChange}
        class="h-9 px-3 text-sm rounded-[var(--radius-md)] bg-[var(--bg-surface)] border
               {error ? 'border-[var(--danger)]' : 'border-[var(--border)]'}
               text-[var(--text)] focus:outline-none focus:border-[var(--accent)]
               disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
    >
        {#if placeholder}
            <option value="" disabled selected={!value}>{placeholder}</option>
        {/if}
        {#each options as opt}
            <option value={opt.value}>{opt.label}</option>
        {/each}
    </select>
    {#if error}
        <p class="text-xs text-[var(--danger)]">{error}</p>
    {/if}
</div>
