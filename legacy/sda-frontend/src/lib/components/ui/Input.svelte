<script lang="ts">
    interface Props {
        label?: string;
        error?: string;
        type?: string;
        placeholder?: string;
        value?: string;
        name?: string;
        id?: string;
        required?: boolean;
        disabled?: boolean;
        class?: string;
    }

    let {
        label,
        error,
        type = 'text',
        placeholder,
        value = $bindable(''),
        name,
        id,
        required = false,
        disabled = false,
        class: cls = '',
    }: Props = $props();

    const inputId = id ?? name ?? crypto.randomUUID();
</script>

<div class="flex flex-col gap-1 {cls}">
    {#if label}
        <label
            for={inputId}
            class="text-xs font-medium text-[var(--text-muted)]"
        >
            {label}{#if required}<span class="text-[var(--danger)] ml-0.5">*</span>{/if}
        </label>
    {/if}

    <input
        {type}
        {name}
        {placeholder}
        {required}
        {disabled}
        id={inputId}
        bind:value
        class="
            w-full px-3 py-2 text-sm rounded-[var(--radius-md)]
            bg-[var(--bg-surface)] border text-[var(--text)]
            placeholder:text-[var(--text-faint)]
            focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:border-[var(--accent)]
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-colors
            {error ? 'border-[var(--danger)]' : 'border-[var(--border)]'}
        "
    />

    {#if error}
        <p class="text-xs text-[var(--danger)]">{error}</p>
    {/if}
</div>
