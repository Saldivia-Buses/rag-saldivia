<script lang="ts">
    interface Props {
        variant?: 'blue' | 'green' | 'red' | 'yellow' | 'gray' | 'orange';
        children: any;
    }
    let { variant = 'gray', children }: Props = $props();

    const variants = {
        blue:   'bg-[var(--info-bg)] text-[var(--info)]',
        green:  'bg-[var(--success-bg)] text-[var(--success)]',
        red:    'bg-[var(--danger-bg)] text-[var(--danger)]',
        yellow: 'bg-[var(--warning-bg)] text-[var(--warning)]',
        gray:   'bg-[var(--bg-hover)] text-[var(--text-muted)]',
        orange: 'bg-orange-950/40 text-orange-400',
    };
</script>

<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold {variants[variant]}">
    {@render children()}
</span>
