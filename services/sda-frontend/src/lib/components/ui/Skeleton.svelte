<script lang="ts">
    interface Props {
        width?: string;
        height?: string;
        rounded?: string;
        class?: string;
    }
    let { width = '100%', height = '1rem', rounded = 'var(--radius-sm)', class: cls = '' }: Props = $props();
</script>

<div
    style="width:{width}; height:{height}; border-radius:{rounded}"
    class="bg-[var(--bg-hover)] animate-pulse {cls}"
    style:background="linear-gradient(90deg, var(--bg-hover) 25%, var(--border-focus) 50%, var(--bg-hover) 75%)"
    style:background-size="200% 100%"
></div>

<style>
    div { animation: shimmer 1.5s infinite; }
    @keyframes shimmer {
        0%   { background-position: 200% 0; }
        100% { background-position: -200% 0; }
    }
</style>
