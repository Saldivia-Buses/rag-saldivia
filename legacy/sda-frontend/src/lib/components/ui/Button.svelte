<script lang="ts">
    interface Props {
        variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
        size?: 'sm' | 'md' | 'lg';
        disabled?: boolean;
        loading?: boolean;
        type?: 'button' | 'submit' | 'reset';
        onclick?: () => void;
        children: any;
    }

    let {
        variant = 'primary',
        size = 'md',
        disabled = false,
        loading = false,
        type = 'button',
        onclick,
        children,
    }: Props = $props();

    const base = 'inline-flex items-center justify-center gap-2 font-medium rounded-[var(--radius-md)] transition-all duration-150 focus-visible:outline-2 focus-visible:outline-[var(--accent)] disabled:opacity-50 disabled:cursor-not-allowed';

    const variants = {
        primary:   'bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)] active:scale-[0.98]',
        secondary: 'bg-[var(--bg-surface)] border border-[var(--border)] text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--text)]',
        danger:    'bg-[var(--danger-bg)] border border-[var(--danger)] text-[var(--danger)] hover:bg-[var(--danger)] hover:text-white',
        ghost:     'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]',
    };

    const sizes = {
        sm: 'text-xs px-2.5 py-1.5 h-7',
        md: 'text-sm px-3.5 py-2 h-9',
        lg: 'text-sm px-5 py-2.5 h-11',
    };
</script>

<button
    {type}
    {onclick}
    disabled={disabled || loading}
    class="{base} {variants[variant]} {sizes[size]}"
>
    {#if loading}
        <span class="w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full animate-spin"></span>
    {/if}
    {@render children()}
</button>
