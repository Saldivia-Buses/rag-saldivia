<script lang="ts">
    import { page } from '$app/stores';
</script>

<div class="flex items-center justify-center flex-1 p-8 min-h-[400px]">
    <div class="text-center max-w-md">
        <div class="text-6xl font-bold text-[var(--border-focus)] mb-4">{$page.status}</div>
        <h1 class="text-base font-semibold text-[var(--text)] mb-2">
            {$page.status === 404 ? 'Página no encontrada' : 'Error del servidor'}
        </h1>
        <p class="text-sm text-[var(--text-muted)] mb-6">
            {$page.error?.message ?? 'Ocurrió un error inesperado.'}
        </p>
        <div class="flex gap-3 justify-center">
            <a href="/chat" class="
                px-4 py-2 rounded-[var(--radius-md)] bg-[var(--accent)] text-white text-sm font-medium
                hover:bg-[var(--accent-hover)] transition-colors
            ">Ir al Chat</a>
            <button onclick={() => window.location.reload()} class="
                px-4 py-2 rounded-[var(--radius-md)] bg-[var(--bg-surface)] border border-[var(--border)]
                text-[var(--text-muted)] text-sm hover:border-[var(--accent)] hover:text-[var(--text)] transition-colors
            ">Reintentar</button>
        </div>
    </div>
</div>
