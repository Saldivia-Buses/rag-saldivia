<script lang="ts">
    import { enhance } from '$app/forms';
    let { data, form } = $props();
    let showCreate = $state(false);
    let creating = $state(false);
    let deletingId = $state<number | null>(null);

    const formFields: Array<[string, string, string]> = [
        ['email', 'Email', 'email'],
        ['name', 'Nombre', 'text'],
        ['password', 'Contrasena', 'password']
    ];
</script>

<div class="p-6">
    <div class="flex items-center justify-between mb-4">
        <h1 class="text-lg font-semibold text-[var(--text)]">Usuarios</h1>
        <button
            onclick={() => showCreate = true}
            class="bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white text-sm px-3 py-1.5 rounded"
        >
            + Nuevo usuario
        </button>
    </div>

    {#if form?.success && form?.api_key}
        <div class="bg-[#065f46] text-[#6ee7b7] p-3 rounded mb-4 text-sm">
            Usuario creado. API key: <code class="font-mono">{form.api_key}</code>
        </div>
    {/if}

    {#if form?.error}
        <div class="bg-[#7f1d1d] text-[#fca5a5] p-3 rounded mb-4 text-sm">
            {form.error}
        </div>
    {/if}

    <table class="w-full text-sm">
        <thead>
            <tr class="text-[var(--text-faint)] text-xs border-b border-[var(--border)]">
                <th class="text-left pb-2">Email</th>
                <th class="text-left pb-2">Nombre</th>
                <th class="text-left pb-2">Area</th>
                <th class="text-left pb-2">Rol</th>
                <th class="text-left pb-2">Estado</th>
                <th class="pb-2"></th>
            </tr>
        </thead>
        <tbody>
            {#each data.users as user}
                <tr class="border-b border-[var(--border)] text-[var(--text-muted)]">
                    <td class="py-2">{user.email}</td>
                    <td class="py-2">{user.name}</td>
                    <td class="py-2">{data.areas.find(a => a.id === user.area_id)?.name ?? user.area_id}</td>
                    <td class="py-2">{user.role}</td>
                    <td class="py-2">
                        <span class="text-xs {user.active ? 'text-[#4ade80]' : 'text-[#f87171]'}">
                            {user.active ? 'Activo' : 'Inactivo'}
                        </span>
                    </td>
                    <td class="py-2 text-right">
                        <form method="POST" action="?/delete" use:enhance={() => {
                            deletingId = user.id;
                            return async ({ update }) => {
                                deletingId = null;
                                await update();
                            };
                        }} class="inline">
                            <input type="hidden" name="id" value={user.id} />
                            <button type="submit"
                                    disabled={deletingId === user.id}
                                    class="text-xs text-[#f87171] hover:underline
                                           disabled:opacity-50 disabled:cursor-not-allowed">
                                {deletingId === user.id ? 'Desactivando...' : 'Desactivar'}
                            </button>
                        </form>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>

    {#if showCreate}
        <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-6 w-96">
                <h2 class="text-sm font-semibold text-[var(--text)] mb-4">Nuevo usuario</h2>
                <form method="POST" action="?/create" use:enhance={() => {
                    creating = true;
                    return async ({ update }) => {
                        creating = false;
                        showCreate = false;
                        await update();
                    };
                }} class="flex flex-col gap-3">
                    {#each formFields as [fname, label, ftype]}
                        <div>
                            <label class="text-xs text-[var(--text-faint)]">{label}</label>
                            <input name={fname} type={ftype} required
                                   disabled={creating}
                                   class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)] rounded
                                          px-2 py-1 text-sm text-[var(--text)] focus:outline-none focus:border-[var(--accent)]
                                          disabled:opacity-50" />
                        </div>
                    {/each}
                    <div>
                        <label class="text-xs text-[var(--text-faint)]">Area</label>
                        <select name="area_id" disabled={creating}
                                class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)]
                                       rounded px-2 py-1 text-sm text-[var(--text)] disabled:opacity-50">
                            {#each data.areas as area}
                                <option value={area.id}>{area.name}</option>
                            {/each}
                        </select>
                    </div>
                    <div>
                        <label class="text-xs text-[var(--text-faint)]">Rol</label>
                        <select name="role" disabled={creating}
                                class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)]
                                       rounded px-2 py-1 text-sm text-[var(--text)] disabled:opacity-50">
                            <option value="user">Usuario</option>
                            <option value="area_manager">Gestor de Area</option>
                            <option value="admin">Admin</option>
                        </select>
                    </div>
                    <div class="flex gap-2 mt-2">
                        <button type="submit"
                                disabled={creating}
                                class="flex-1 bg-[var(--accent)] text-white text-sm py-1.5 rounded
                                       disabled:opacity-50 disabled:cursor-not-allowed">
                            {creating ? 'Creando...' : 'Crear'}
                        </button>
                        <button type="button" onclick={() => showCreate = false}
                                disabled={creating}
                                class="flex-1 bg-[var(--bg-surface)] text-[var(--text-muted)] text-sm py-1.5 rounded
                                       disabled:opacity-50">
                            Cancelar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    {/if}
</div>
