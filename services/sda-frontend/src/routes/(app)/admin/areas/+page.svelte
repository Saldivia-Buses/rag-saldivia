<script lang="ts">
    import { enhance } from '$app/forms';
    import { toastStore } from '$lib/stores/toast.svelte';

    let { data, form } = $props();

    let showCreate = $state(false);
    let editingArea = $state<{ id: number; name: string; description: string } | null>(null);
    let confirmDelete = $state<{ id: number; name: string } | null>(null);
    let saving = $state(false);
    let deletingId = $state<number | null>(null);

    function blockingUsers(areaId: number) {
        return data.usersByArea[areaId] ?? [];
    }

    $effect(() => {
        if (form?.success) {
            if (form.created) toastStore.success('Área creada correctamente');
            else toastStore.success('Área actualizada');
        }
    });
</script>

<div class="p-6">
    <div class="flex items-center justify-between mb-4">
        <h1 class="text-lg font-semibold text-[var(--text)]">Áreas</h1>
        <button
            onclick={() => showCreate = true}
            class="bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white text-sm px-3 py-1.5 rounded"
        >
            + Nueva área
        </button>
    </div>

    {#if form?.error}
        <div class="bg-[var(--danger-bg)] text-[var(--danger)] p-3 rounded mb-4 text-sm">
            {form.error}
        </div>
    {/if}

    {#if data.areas.length === 0}
        <div class="border border-[var(--border)] rounded-lg p-8 text-center text-sm text-[var(--text-muted)]">
            No hay áreas configuradas.
        </div>
    {:else}
        <div class="border border-[var(--border)] rounded-lg overflow-hidden">
            <table class="w-full text-sm">
                <thead>
                    <tr class="text-[var(--text-faint)] text-xs border-b border-[var(--border)] bg-[var(--bg-surface)]">
                        <th class="text-left px-4 py-2">Nombre</th>
                        <th class="text-left px-4 py-2">Descripción</th>
                        <th class="text-center px-4 py-2">Usuarios</th>
                        <th class="px-4 py-2"></th>
                    </tr>
                </thead>
                <tbody>
                    {#each data.areas as area (area.id)}
                        <tr class="border-b border-[var(--border)] last:border-0 text-[var(--text-muted)]">
                            <td class="px-4 py-2 font-medium text-[var(--text)]">{area.name}</td>
                            <td class="px-4 py-2 text-xs">{area.description || '—'}</td>
                            <td class="px-4 py-2 text-center">
                                <span class="text-xs {area.userCount > 0 ? 'text-[var(--text)]' : 'text-[var(--text-faint)]'}">
                                    {area.userCount}
                                </span>
                            </td>
                            <td class="px-4 py-2">
                                <div class="flex items-center gap-2 justify-end">
                                    <button
                                        onclick={() => editingArea = { id: area.id, name: area.name, description: area.description }}
                                        class="text-xs text-[var(--text-muted)] hover:text-[var(--accent)] transition-colors"
                                    >
                                        Editar
                                    </button>
                                    <button
                                        onclick={() => confirmDelete = { id: area.id, name: area.name }}
                                        class="text-xs text-[var(--danger)] hover:underline"
                                    >
                                        Eliminar
                                    </button>
                                </div>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    {/if}
</div>

<!-- Modal: Crear área -->
{#if showCreate}
    <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
        <div class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-6 w-96">
            <h2 class="text-sm font-semibold text-[var(--text)] mb-4">Nueva área</h2>
            <form method="POST" action="?/create" use:enhance={() => {
                saving = true;
                return async ({ result, update }) => {
                    saving = false;
                    if (result.type === 'success') showCreate = false;
                    await update();
                };
            }} class="flex flex-col gap-3">
                <div>
                    <label class="text-xs text-[var(--text-faint)]">Nombre *</label>
                    <input name="name" type="text" required disabled={saving}
                           class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)]
                                  rounded px-2 py-1 text-sm text-[var(--text)]
                                  focus:outline-none focus:border-[var(--accent)] disabled:opacity-50" />
                </div>
                <div>
                    <label class="text-xs text-[var(--text-faint)]">Descripción</label>
                    <input name="description" type="text" disabled={saving}
                           class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)]
                                  rounded px-2 py-1 text-sm text-[var(--text)]
                                  focus:outline-none focus:border-[var(--accent)] disabled:opacity-50" />
                </div>
                <div class="flex gap-2 mt-2">
                    <button type="submit" disabled={saving}
                            class="flex-1 bg-[var(--accent)] text-white text-sm py-1.5 rounded disabled:opacity-50">
                        {saving ? 'Creando...' : 'Crear'}
                    </button>
                    <button type="button" onclick={() => showCreate = false} disabled={saving}
                            class="flex-1 bg-[var(--bg-surface)] text-[var(--text-muted)] text-sm py-1.5 rounded disabled:opacity-50">
                        Cancelar
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

<!-- Modal: Editar área -->
{#if editingArea}
    <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
        <div class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-6 w-96">
            <h2 class="text-sm font-semibold text-[var(--text)] mb-4">Editar área</h2>
            <form method="POST" action="?/update" use:enhance={() => {
                saving = true;
                return async ({ result, update }) => {
                    saving = false;
                    if (result.type === 'success') editingArea = null;
                    await update();
                };
            }} class="flex flex-col gap-3">
                <input type="hidden" name="id" value={editingArea.id} />
                <div>
                    <label class="text-xs text-[var(--text-faint)]">Nombre *</label>
                    <input name="name" type="text" required value={editingArea.name} disabled={saving}
                           class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)]
                                  rounded px-2 py-1 text-sm text-[var(--text)]
                                  focus:outline-none focus:border-[var(--accent)] disabled:opacity-50" />
                </div>
                <div>
                    <label class="text-xs text-[var(--text-faint)]">Descripción</label>
                    <input name="description" type="text" value={editingArea.description} disabled={saving}
                           class="w-full mt-0.5 bg-[var(--bg-surface)] border border-[var(--border)]
                                  rounded px-2 py-1 text-sm text-[var(--text)]
                                  focus:outline-none focus:border-[var(--accent)] disabled:opacity-50" />
                </div>
                <div class="flex gap-2 mt-2">
                    <button type="submit" disabled={saving}
                            class="flex-1 bg-[var(--accent)] text-white text-sm py-1.5 rounded disabled:opacity-50">
                        {saving ? 'Guardando...' : 'Guardar'}
                    </button>
                    <button type="button" onclick={() => editingArea = null} disabled={saving}
                            class="flex-1 bg-[var(--bg-surface)] text-[var(--text-muted)] text-sm py-1.5 rounded disabled:opacity-50">
                        Cancelar
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

<!-- Modal: Confirmar eliminar -->
{#if confirmDelete}
    {@const users = blockingUsers(confirmDelete.id)}
    <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
        <div class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-6 w-[420px]">
            <h2 class="text-sm font-semibold text-[var(--text)] mb-2">¿Eliminar área?</h2>
            <p class="text-sm text-[var(--text-muted)] mb-3">
                Área: <strong class="text-[var(--text)]">{confirmDelete.name}</strong>
            </p>

            {#if users.length > 0}
                <div class="bg-[var(--danger-bg)] rounded p-3 mb-4">
                    <p class="text-xs text-[var(--danger)] font-medium mb-2">
                        No se puede eliminar — {users.length} usuario(s) asignado(s):
                    </p>
                    <ul class="text-xs text-[var(--danger)] space-y-0.5">
                        {#each users as u}
                            <li>{u.name} ({u.email})</li>
                        {/each}
                    </ul>
                    <p class="text-xs text-[var(--danger)] mt-2">
                        Reasigná o desactivá estos usuarios primero.
                    </p>
                </div>
                <button onclick={() => confirmDelete = null}
                        class="w-full bg-[var(--bg-surface)] text-[var(--text-muted)] text-sm py-1.5 rounded">
                    Cerrar
                </button>
            {:else}
                <p class="text-sm text-[var(--text-muted)] mb-4">Esta acción no se puede deshacer.</p>
                <form method="POST" action="?/delete" use:enhance={() => {
                    deletingId = confirmDelete?.id ?? null;
                    return async ({ result, update }) => {
                        deletingId = null;
                        if (result.type === 'success') {
                            toastStore.success('Área eliminada');
                            confirmDelete = null;
                        }
                        await update();
                    };
                }} class="flex gap-2">
                    <input type="hidden" name="id" value={confirmDelete.id} />
                    <button type="submit" disabled={deletingId !== null}
                            class="flex-1 bg-[var(--danger)] text-white text-sm py-1.5 rounded disabled:opacity-50">
                        {deletingId !== null ? 'Eliminando...' : 'Eliminar'}
                    </button>
                    <button type="button" onclick={() => confirmDelete = null}
                            class="flex-1 bg-[var(--bg-surface)] text-[var(--text-muted)] text-sm py-1.5 rounded">
                        Cancelar
                    </button>
                </form>
            {/if}
        </div>
    </div>
{/if}
