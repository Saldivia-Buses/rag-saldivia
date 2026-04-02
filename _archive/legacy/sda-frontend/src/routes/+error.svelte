<script lang="ts">
    import { page } from '$app/stores';
</script>

<div class="min-h-screen flex items-center justify-center bg-[var(--bg-base)] p-6">
    <div class="text-center max-w-md">
        <div class="text-7xl font-bold text-[var(--border-focus)] mb-4 tabular-nums">
            {$page.status}
        </div>
        <h1 class="text-lg font-semibold text-[var(--text)] mb-2">
            {$page.status === 404 ? 'Página no encontrada' : 'Error del servidor'}
        </h1>
        <p class="text-sm text-[var(--text-muted)] mb-8">
            {$page.error?.message ?? 'Ocurrió un error inesperado.'}
        </p>
        <div class="flex gap-3 justify-center">
            <a href="/chat" class="
                inline-flex items-center gap-2 px-4 py-2 rounded-[var(--radius-md)]
                bg-[var(--accent)] text-white text-sm font-medium
                hover:bg-[var(--accent-hover)] transition-colors
            ">
                Ir al Chat
            </a>
            <button onclick={() => window.location.reload()} class="
                inline-flex items-center gap-2 px-4 py-2 rounded-[var(--radius-md)]
                bg-[var(--bg-surface)] border border-[var(--border)]
                text-[var(--text-muted)] text-sm
                hover:text-[var(--text)] hover:border-[var(--accent)] transition-colors
            ">
                Reintentar
            </button>
        </div>
    </div>
</div>
