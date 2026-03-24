<script lang="ts">
    import { enhance } from '$app/forms';
    import { setMode, mode } from 'mode-watcher';
    import { User, Key, Moon, Sun, LogOut, Copy, Check, Lock, Sliders, Bell } from 'lucide-svelte';
    import Badge from '$lib/components/ui/Badge.svelte';
    import type { PageData, ActionData } from './$types';

    let { data, form }: { data: PageData; form: ActionData } = $props();

    // Avatar
    let nameInitial = $derived(data.user?.name?.charAt(0)?.toUpperCase() ?? '?');

    // Password fields
    let currentPw = $state('');
    let newPw = $state('');
    let confirmPw = $state('');
    let pwMatch = $derived(newPw === confirmPw || confirmPw === '');

    // API Key
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

    <!-- Sección 1: Perfil -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <User size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">Perfil</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <form method="POST" action="?/update_profile" use:enhance>
                <!-- Avatar preview -->
                <div class="flex items-center gap-3 mb-4">
                    <div
                        class="w-10 h-10 rounded-full flex items-center justify-center text-white font-bold text-sm flex-shrink-0"
                        style="background: {data.preferences?.avatar_color ?? '#6366f1'}"
                    >
                        {nameInitial}
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="text-sm font-semibold text-[var(--text)]">{data.user?.name ?? '—'}</div>
                        <div class="text-xs text-[var(--text-muted)] truncate">{data.user?.email}</div>
                    </div>
                    <Badge variant={roleBadge[data.user?.role ?? 'user'] as any}>
                        {roleLabel[data.user?.role ?? 'user'] ?? data.user?.role}
                    </Badge>
                </div>

                <!-- Nombre -->
                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="name">Nombre</label>
                    <input
                        id="name" name="name" type="text"
                        value={data.user?.name ?? ''}
                        class="w-full text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]"
                    />
                    {#if form?.field === 'name'}
                        <p class="text-xs text-[var(--danger)] mt-1">{form.error}</p>
                    {/if}
                </div>

                <!-- Color de avatar -->
                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="avatar_color">Color de avatar</label>
                    <input
                        id="avatar_color" name="avatar_color" type="color"
                        value={data.preferences?.avatar_color ?? '#6366f1'}
                        class="h-8 w-16 rounded border border-[var(--border)] bg-[var(--bg-base)] cursor-pointer"
                    />
                </div>

                <!-- Idioma -->
                <div class="mb-4">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="ui_language">Idioma de interfaz</label>
                    <select
                        id="ui_language" name="ui_language"
                        class="text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]"
                    >
                        <option value="es" selected={data.preferences?.ui_language === 'es'}>Español</option>
                        <option value="en" selected={data.preferences?.ui_language === 'en'}>English</option>
                    </select>
                </div>

                {#if form?.success && form.section === 'profile'}
                    <p class="text-xs text-[var(--success)] mb-3">Perfil actualizado.</p>
                {/if}

                <button type="submit"
                    class="text-xs px-3 py-1.5 rounded-[var(--radius-md)] bg-[var(--accent)] text-white hover:opacity-90 transition-opacity">
                    Guardar perfil
                </button>
            </form>
        </div>
    </section>

    <!-- Sección 2: Contraseña -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <Lock size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">Contraseña</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <form method="POST" action="?/update_password" use:enhance={() => {
                return async ({ result, update }) => {
                    await update();
                    if (result.type === 'success') { currentPw = ''; newPw = ''; confirmPw = ''; }
                };
            }}>
                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="current_password">Contraseña actual</label>
                    <input
                        id="current_password" name="current_password" type="password"
                        bind:value={currentPw}
                        class="w-full text-sm bg-[var(--bg-base)] border rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]
                               {form?.field === 'current_password' ? 'border-[var(--danger)]' : 'border-[var(--border)]'}"
                    />
                    {#if form?.field === 'current_password'}
                        <p class="text-xs text-[var(--danger)] mt-1">{form.error}</p>
                    {/if}
                </div>

                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="new_password">Nueva contraseña</label>
                    <input
                        id="new_password" name="new_password" type="password"
                        bind:value={newPw}
                        class="w-full text-sm bg-[var(--bg-base)] border rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]
                               {form?.field === 'new_password' ? 'border-[var(--danger)]' : 'border-[var(--border)]'}"
                    />
                    {#if form?.field === 'new_password'}
                        <p class="text-xs text-[var(--danger)] mt-1">{form.error}</p>
                    {/if}
                </div>

                <div class="mb-4">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="confirm_password">Confirmar contraseña</label>
                    <input
                        id="confirm_password" name="confirm_password" type="password"
                        bind:value={confirmPw}
                        class="w-full text-sm bg-[var(--bg-base)] border rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]
                               {(!pwMatch || form?.field === 'confirm_password') ? 'border-[var(--danger)]' : 'border-[var(--border)]'}"
                    />
                    {#if !pwMatch}
                        <p class="text-xs text-[var(--danger)] mt-1">Las contraseñas no coinciden</p>
                    {:else if form?.field === 'confirm_password'}
                        <p class="text-xs text-[var(--danger)] mt-1">{form.error}</p>
                    {/if}
                </div>

                {#if form?.success && form.section === 'password'}
                    <p class="text-xs text-[var(--success)] mb-3">Contraseña actualizada.</p>
                {/if}

                <button type="submit" disabled={!pwMatch}
                    class="text-xs px-3 py-1.5 rounded-[var(--radius-md)] bg-[var(--accent)] text-white hover:opacity-90 transition-opacity disabled:opacity-40">
                    Cambiar contraseña
                </button>
            </form>
        </div>
    </section>

    <!-- Sección 3: Preferencias RAG -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <Sliders size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">Preferencias RAG</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <form method="POST" action="?/update_preferences" use:enhance>
                <!-- Colección por defecto -->
                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="default_collection">Colección por defecto</label>
                    <select id="default_collection" name="default_collection"
                        class="text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]">
                        <option value="">— Sin preferencia —</option>
                        {#each data.collections ?? [] as col}
                            <option value={col} selected={data.preferences?.default_collection === col}>{col}</option>
                        {/each}
                    </select>
                </div>

                <!-- Modo de query -->
                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="default_query_mode">Modo de búsqueda</label>
                    <select id="default_query_mode" name="default_query_mode"
                        class="text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]">
                        <option value="standard" selected={data.preferences?.default_query_mode === 'standard'}>Estándar</option>
                        <option value="crossdoc" selected={data.preferences?.default_query_mode === 'crossdoc'}>CrossDoc</option>
                    </select>
                </div>

                <!-- Sliders numéricos -->
                <div class="grid grid-cols-2 gap-3 mb-3">
                    <div>
                        <label class="block text-xs text-[var(--text-muted)] mb-1" for="vdb_top_k">VDB Top-K</label>
                        <input id="vdb_top_k" name="vdb_top_k" type="number" min="1" max="100"
                            value={data.preferences?.vdb_top_k ?? 10}
                            class="w-full text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]" />
                    </div>
                    <div>
                        <label class="block text-xs text-[var(--text-muted)] mb-1" for="reranker_top_k">Reranker Top-K</label>
                        <input id="reranker_top_k" name="reranker_top_k" type="number" min="1" max="20"
                            value={data.preferences?.reranker_top_k ?? 5}
                            class="w-full text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]" />
                    </div>
                </div>

                <div class="mb-3">
                    <label class="block text-xs text-[var(--text-muted)] mb-1" for="max_sub_queries">Max sub-queries (CrossDoc)</label>
                    <input id="max_sub_queries" name="max_sub_queries" type="number" min="1" max="20"
                        value={data.preferences?.max_sub_queries ?? 4}
                        class="w-full text-sm bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-md)] px-3 py-1.5 text-[var(--text)] focus:outline-none focus:border-[var(--accent)]" />
                </div>

                <!-- Toggles -->
                <div class="flex flex-col gap-2 mb-4">
                    <label class="flex items-center gap-2 cursor-pointer">
                        <input name="follow_up_retries" type="checkbox"
                            checked={data.preferences?.follow_up_retries ?? true}
                            class="rounded border-[var(--border)]" />
                        <span class="text-xs text-[var(--text-muted)]">Reintentar follow-ups fallidos</span>
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer">
                        <input name="show_decomposition" type="checkbox"
                            checked={data.preferences?.show_decomposition ?? false}
                            class="rounded border-[var(--border)]" />
                        <span class="text-xs text-[var(--text-muted)]">Mostrar descomposición CrossDoc</span>
                    </label>
                </div>

                {#if form?.success && form.section === 'preferences'}
                    <p class="text-xs text-[var(--success)] mb-3">Preferencias guardadas.</p>
                {/if}

                <button type="submit"
                    class="text-xs px-3 py-1.5 rounded-[var(--radius-md)] bg-[var(--accent)] text-white hover:opacity-90 transition-opacity">
                    Guardar preferencias
                </button>
            </form>
        </div>
    </section>

    <!-- Sección 4: Notificaciones -->
    <section class="mb-4">
        <div class="flex items-center gap-2 mb-3">
            <Bell size={14} class="text-[var(--text-faint)]" />
            <span class="text-xs font-semibold text-[var(--text-faint)] uppercase tracking-wider">Notificaciones</span>
        </div>
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
            <form method="POST" action="?/update_notifications" use:enhance>
                <div class="flex flex-col gap-3 mb-4">
                    <label class="flex items-start gap-3 cursor-pointer">
                        <input name="notify_ingestion_done" type="checkbox"
                            checked={data.preferences?.notify_ingestion_done ?? true}
                            class="mt-0.5 rounded border-[var(--border)]" />
                        <div>
                            <div class="text-sm text-[var(--text)]">Ingesta completada</div>
                            <div class="text-xs text-[var(--text-muted)]">Notificar cuando un documento termine de procesarse</div>
                        </div>
                    </label>
                    <label class="flex items-start gap-3 cursor-pointer">
                        <input name="notify_system_alerts" type="checkbox"
                            checked={data.preferences?.notify_system_alerts ?? true}
                            class="mt-0.5 rounded border-[var(--border)]" />
                        <div>
                            <div class="text-sm text-[var(--text)]">Alertas del sistema</div>
                            <div class="text-xs text-[var(--text-muted)]">Mostrar alertas de errores del sistema al iniciar sesión</div>
                        </div>
                    </label>
                </div>

                {#if form?.success && form.section === 'notifications'}
                    <p class="text-xs text-[var(--success)] mb-3">Notificaciones guardadas.</p>
                {/if}

                <button type="submit"
                    class="text-xs px-3 py-1.5 rounded-[var(--radius-md)] bg-[var(--accent)] text-white hover:opacity-90 transition-opacity">
                    Guardar notificaciones
                </button>
            </form>
        </div>
    </section>

    <!-- Sección 5: API Key + Apariencia + Sesión -->
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

            {#if form?.error && !form.section}
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
