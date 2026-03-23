<script lang="ts">
    interface Props {
        messageId: number | undefined;
        sessionId: string;
    }
    let { messageId, sessionId }: Props = $props();

    let rating = $state<'up' | 'down' | null>(null);

    async function vote(r: 'up' | 'down') {
        if (!messageId) return;
        const next = rating === r ? null : r;
        rating = next;
        if (next) {
            await fetch(`/api/chat/sessions/${sessionId}/messages/${messageId}/feedback`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ rating: next }),
            });
        }
    }
</script>

<div class="flex gap-1 mt-1">
    <button
        onclick={() => vote('up')}
        title="Útil"
        class="p-1 rounded text-xs transition-colors
               {rating === 'up'
                   ? 'text-[var(--accent)]'
                   : 'text-[var(--text-faint)] hover:text-[var(--text-muted)]'}"
        aria-label="Útil"
    >👍</button>
    <button
        onclick={() => vote('down')}
        title="No útil"
        class="p-1 rounded text-xs transition-colors
               {rating === 'down'
                   ? 'text-[var(--danger,#ef4444)]'
                   : 'text-[var(--text-faint)] hover:text-[var(--text-muted)]'}"
        aria-label="No útil"
    >👎</button>
</div>
