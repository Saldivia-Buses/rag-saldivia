<script lang="ts">
    import { enhance } from '$app/forms';
    import { setMode, mode } from 'mode-watcher';
    import { User, Key, Moon, Sun, LogOut, Copy, Check } from 'lucide-svelte';
    import Badge from '$lib/components/ui/Badge.svelte';

    let { data, form } = $props();

    let refreshing = $state(false);
    let loggingOut = $state(false);
    let copied = $state(false);

    async function handleLogout() {
        loggingOut = true;
        try {
            await fetch('/api/auth/session', { method: 'DELETE' });
            window.location.href = '/login';
        } catch {
            window.location.href = '/login';
        }
    }

    async function copyApiKey() {
        if (!form?.api_key) return;
        await navigator.clipboard.writeText(form.api_key);
        copied = true;
        setTimeout(() => { copied = false; }, 2000);
    }

    const roleLabel: Record<string, string> = {
        admin: 'Administrador',
        area_manager: 'Gestor de área',
        user: 'Usuario',
    };
    const roleBadge: Record<string, string> = {
        admin: 'blue',
        area_manager: 'orange',
        user: 'gray',
    };
</script>

<div class="p-6 max-w-xl">
    <h1 class="text-lg font-semibold text-[var(--text)] mb-6">Configuración</h1>

    <!-- Perfil -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <User size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">Perfil</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-[var(--accent)] rounded-full flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                    {data.user?.name?.charAt(0)?.toUpperCase() ?? data.user?.email?.charAt(0)?.toUpperCase() ?? '?'}
                </div>
                <div class="flex-1 min-w-0">
                    <div class="text-sm font-semibold text-[var(--text)]">{data.user?.name ?? '—'}</div>
                    <div class="text-xs text-[var(--text-muted)] truncate">{data.user?.email}</div>
                </div>
                <Badge variant={roleBadge[data.user?.role ?? 'user'] as any}>
                    {roleLabel[data.user?.role ?? 'user'] ?? data.user?.role}
                </Badge>
            </div>
        </div>
    </section>

    <!-- Apariencia -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <Sun size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">Apariencia</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <div class="flex items-center justify-between">
                <div>
                    <div class="text-sm text-[var(--text)]">Tema de la interfaz</div>
                    <div class="text-xs text-[var(--text-muted)] mt-0.5">
                        {#if $mode === 'dark'}Modo oscuro activo{:else if $mode === 'light'}Modo claro activo{:else}Según el sistema{/if}
                    </div>
                </div>
                <div class="flex items-center gap-1 bg-[var(--bg-hover)] rounded-[var(--radius-md)] p-1">
                    <button
                        onclick={() => setMode('light')}
                        class="flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] text-xs transition-colors
                               {$mode === 'light' ? 'bg-[var(--bg-surface)] text-[var(--text)] shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text)]'}"
                    >
                        <Sun size={13} />
                        <span>Claro</span>
                    </button>
                    <button
                        onclick={() => setMode('dark')}
                        class="flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] text-xs transition-colors
                               {$mode === 'dark' ? 'bg-[var(--bg-surface)] text-[var(--text)] shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text)]'}"
                    >
                        <Moon size={13} />
                        <span>Oscuro</span>
                    </button>
                </div>
            </div>
        </div>
    </section>

    <!-- API Key -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <Key size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">API Key personal</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <p class="text-xs text-[var(--text-muted)] mb-3">
                Usá esta clave para acceder a la API de SDA desde scripts o integraciones externas.
            </p>

            {#if form?.api_key}
                <div class="flex items-center gap-2 bg-[var(--success-bg)] border border-[var(--success)] rounded-[var(--radius-md)] px-3 py-2 mb-3">
                    <code class="text-xs text-[var(--success)] font-mono flex-1 break-all">{form.api_key}</code>
                    <button onclick={copyApiKey} class="text-[var(--success)] hover:opacity-80 flex-shrink-0" title="Copiar">
                        {#if copied}<Check size={14} />{:else}<Copy size={14} />{/if}
                    </button>
                </div>
                <p class="text-xs text-[var(--warning)] mb-3">Guardala ahora — no se vuelve a mostrar.</p>
            {/if}

            {#if form?.error}
                <div class="bg-[var(--danger-bg)] text-[var(--danger)] px-3 py-2 rounded-[var(--radius-md)] text-xs mb-3">
                    {form.error}
                </div>
            {/if}

            <form method="POST" action="?/refresh_key" use:enhance={() => {
                refreshing = true;
                return async ({ update }) => { refreshing = false; await update(); };
            }}>
                <button type="submit" disabled={refreshing}
                    class="text-xs px-3 py-1.5 rounded-[var(--radius-md)]
                           bg-[var(--bg-hover)] hover:bg-[var(--border)] text-[var(--text-muted)] hover:text-[var(--text)]
                           border border-[var(--border)] transition-colors disabled:opacity-50">
                    {refreshing ? 'Regenerando...' : 'Regenerar API key'}
                </button>
            </form>
        </div>
    </section>

    <!-- Sesión -->
    <div class="pt-2 border-t border-[var(--border)]">
        <button
            onclick={handleLogout}
            disabled={loggingOut}
            class="flex items-center gap-2 text-xs px-3 py-2 rounded-[var(--radius-md)]
                   text-[var(--danger)] hover:bg-[var(--danger-bg)] border border-transparent
                   hover:border-[var(--danger)] transition-colors disabled:opacity-50"
        >
            <LogOut size={13} />
            {loggingOut ? 'Cerrando sesión...' : 'Cerrar sesión'}
        </button>
    </div>
</div>
