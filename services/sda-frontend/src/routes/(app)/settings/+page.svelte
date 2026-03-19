<script lang="ts">
    import { enhance } from '$app/forms';

    let { data, form } = $props();

    let refreshing = $state(false);
    let loggingOut = $state(false);

    async function handleLogout() {
        loggingOut = true;
        try {
            await fetch('/api/auth/session', { method: 'DELETE' });
            window.location.href = '/login';
        } catch {
            // Even on error, force redirect to clear state
            window.location.href = '/login';
        }
    }
</script>

<div class="p-6 max-w-lg">
    <h1 class="text-lg font-semibold text-[var(--text)] mb-4">Configuración</h1>

    <div class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-4 mb-4">
        <div class="text-xs text-[var(--text-faint)] mb-3">Perfil</div>
        <div class="text-sm text-[var(--text)]">{data.user?.name}</div>
        <div class="text-xs text-[var(--text-muted)]">{data.user?.email}</div>
        <div class="text-xs text-[var(--accent)] mt-1">{data.user?.role}</div>
    </div>

    <div class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-4">
        <div class="text-xs text-[var(--text-faint)] mb-3">API Key personal</div>
        {#if form?.api_key}
            <div class="bg-[#065f46] text-[#6ee7b7] p-2 rounded text-xs font-mono mb-3 break-all">
                {form.api_key}
            </div>
        {/if}
        {#if form?.error}
            <div class="bg-[#7f1d1d] text-[#fca5a5] p-2 rounded text-xs mb-3">
                {form.error}
            </div>
        {/if}
        <form method="POST" action="?/refresh_key" use:enhance={() => {
            refreshing = true;
            return async ({ update }) => {
                refreshing = false;
                await update();
            };
        }}>
            <button type="submit"
                    disabled={refreshing}
                    class="bg-[var(--bg-surface)] hover:bg-[var(--border)] text-[var(--text-muted)] text-xs px-3 py-1.5 rounded
                           disabled:opacity-50 disabled:cursor-not-allowed">
                {refreshing ? 'Regenerando...' : 'Regenerar API key'}
            </button>
        </form>
    </div>

    <!-- Logout section -->
    <div class="mt-6 pt-6 border-t border-[var(--border)]">
        <button onclick={handleLogout}
                disabled={loggingOut}
                class="bg-[var(--bg-surface)] hover:bg-[var(--border)] text-[var(--text-muted)] text-xs px-3 py-1.5 rounded
                       disabled:opacity-50 disabled:cursor-not-allowed">
            {loggingOut ? 'Cerrando sesión...' : 'Cerrar sesión'}
        </button>
    </div>
</div>
