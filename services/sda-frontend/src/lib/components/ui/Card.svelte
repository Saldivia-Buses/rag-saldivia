<script lang="ts">
    interface Props {
        padding?: boolean;
        class?: string;
        children: any;
    }
    let { padding = true, class: cls = '', children }: Props = $props();
</script>

<div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] {padding ? 'p-4' : ''} {cls}">
    {@render children()}
</div>
