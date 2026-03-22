<script lang="ts">
    import { TIER_CONFIG, type Tier } from '$lib/ingestion/types.js';

    let { tier }: { tier: Tier } = $props();

    const config = $derived(TIER_CONFIG[tier]);

    const colorClasses: Record<string, string> = {
        green: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
        blue:  'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
        amber: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
        red:   'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    };
</script>

<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {colorClasses[config.color]}">
    {config.label}
</span>
