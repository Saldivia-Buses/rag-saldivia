<script lang="ts">
    import { enhance } from '$app/forms';
    import { toastStore } from '$lib/stores/toast.svelte';

    let { data, form } = $props();

    let selectedCollection = $state<Record<number, string>>({});
    let selectedPermission = $state<Record<number, string>>({});
    let grantingAreaId = $state<number | null>(null);
    let revokingKey = $state<string | null>(null);

    function getAvailable(areaId: number): string[] {
        const area = data.areas.find(a => a.id === areaId);
        const assigned = new Set(area?.collections.map(c => c.name) ?? []);
        return data.allCollections.filter(c => !assigned.has(c));
    }
</script>

<div class="p-6">
    <div class="mb-6">
        <h1 class="text-lg font-semibold text-[var(--text)]">
            {data.isManager ? 'Permisos de tu área' : 'Gestión de permisos por área'}
        </h1>
        <p class="text-sm text-[var(--text-muted)] mt-1">
            Asigná colecciones a cada área. Los usuarios del área solo acceden a las colecciones asignadas.
        </p>
    </div>

    {#if form?.error}
        <div class="bg-[var(--danger-bg)] text-[var(--danger)] p-3 rounded mb-4 text-sm">
            {form.error}
        </div>
    {/if}

    {#if data.areas.length === 0}
        <div class="border border-[var(--border)] rounded-lg p-8 text-center text-sm text-[var(--text-muted)]">
            No hay áreas disponibles.
            {#if !data.isManager}
                <a href="/admin/areas" class="text-[var(--accent)] hover:underline ml-1">Crear áreas</a>
            {/if}
        </div>
    {:else}
        <div class="flex flex-col gap-4">
            {#each data.areas as area (area.id)}
                <div class="border border-[var(--border)] rounded-lg p-4">
                    <div class="flex items-start justify-between mb-3">
                        <div>
                            <h2 class="text-sm font-semibold text-[var(--text)]">{area.name}</h2>
                            {#if area.description}
                                <p class="text-xs text-[var(--text-faint)] mt-0.5">{area.description}</p>
                            {/if}
                        </div>
                        <span class="text-xs text-[var(--text-faint)]">{area.collections.length} colección(es)</span>
                    </div>

                    <!-- Chips de colecciones asignadas -->
                    <div class="flex flex-wrap gap-2 mb-3 min-h-[28px]">
                        {#if area.collections.length === 0}
                            <span class="text-xs text-[var(--text-faint)] italic">Sin colecciones asignadas</span>
                        {:else}
                            {#each area.collections as col (col.name)}
                                <div class="flex items-center gap-1 bg-[var(--bg-surface)] border border-[var(--border)]
                                            text-xs text-[var(--text-muted)] px-2 py-1 rounded-full">
                                    <span>{col.name}</span>
                                    <span class="text-[var(--text-faint)] px-1">·</span>
                                    <span class="text-[var(--text-faint)]">{col.permission}</span>
                                    <form method="POST" action="?/revoke" use:enhance={() => {
                                        revokingKey = `${area.id}-${col.name}`;
                                        return async ({ result, update }) => {
                                            revokingKey = null;
                                            if (result.type === 'success') toastStore.success('Acceso revocado');
                                            await update();
                                        };
                                    }} class="inline ml-1">
                                        <input type="hidden" name="area_id" value={area.id} />
                                        <input type="hidden" name="collection" value={col.name} />
                                        <button type="submit"
                                                disabled={revokingKey === `${area.id}-${col.name}`}
                                                class="text-[var(--text-faint)] hover:text-[var(--danger)]
                                                       transition-colors disabled:opacity-30 leading-none"
                                                title="Revocar">×</button>
                                    </form>
                                </div>
                            {/each}
                        {/if}
                    </div>

                    <!-- Agregar colección -->
                    {#if getAvailable(area.id).length > 0}
                        <form method="POST" action="?/grant" use:enhance={() => {
                            grantingAreaId = area.id;
                            return async ({ result, update }) => {
                                grantingAreaId = null;
                                if (result.type === 'success') {
                                    toastStore.success('Colección asignada');
                                    selectedCollection[area.id] = '';
                                }
                                await update();
                            };
                        }} class="flex items-center gap-2">
                            <input type="hidden" name="area_id" value={area.id} />
                            <select name="collection"
                                    bind:value={selectedCollection[area.id]}
                                    disabled={grantingAreaId === area.id}
                                    class="bg-[var(--bg-surface)] border border-[var(--border)] rounded
                                           px-2 py-1 text-xs text-[var(--text)] disabled:opacity-50
                                           focus:outline-none focus:border-[var(--accent)]">
                                <option value="" disabled selected>Colección…</option>
                                {#each getAvailable(area.id) as col}
                                    <option value={col}>{col}</option>
                                {/each}
                            </select>
                            <select name="permission"
                                    bind:value={selectedPermission[area.id]}
                                    disabled={grantingAreaId === area.id}
                                    class="bg-[var(--bg-surface)] border border-[var(--border)] rounded
                                           px-2 py-1 text-xs text-[var(--text)] disabled:opacity-50
                                           focus:outline-none focus:border-[var(--accent)]">
                                <option value="read">read</option>
                                <option value="write">write</option>
                            </select>
                            <button type="submit"
                                    disabled={!selectedCollection[area.id] || grantingAreaId === area.id}
                                    class="text-xs px-3 py-1 rounded bg-[var(--accent)] text-white
                                           disabled:opacity-40 disabled:cursor-not-allowed
                                           hover:bg-[var(--accent-hover)] transition-colors">
                                {grantingAreaId === area.id ? 'Asignando…' : 'Asignar'}
                            </button>
                        </form>
                    {:else}
                        <p class="text-xs text-[var(--text-faint)] italic">
                            {data.allCollections.length === 0
                                ? 'No hay colecciones en el sistema.'
                                : 'Todas las colecciones están asignadas.'}
                        </p>
                    {/if}
                </div>
            {/each}
        </div>
    {/if}
</div>
