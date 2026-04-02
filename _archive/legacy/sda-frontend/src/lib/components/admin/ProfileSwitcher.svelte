<!-- src/lib/components/admin/ProfileSwitcher.svelte -->
<script lang="ts">
    let { currentProfile, onSwitch } = $props<{
        currentProfile: string;
        onSwitch: (profile: string) => void;
    }>();

    const PROFILES = ['workstation-1gpu', 'brev-2gpu', 'full-cloud'];
    let selected = $state(currentProfile);
    let showConfirm = $state(false);
</script>

<div class="mb-4">
    <label for="profile-switcher" class="block text-sm font-medium text-[var(--text)] mb-1">Perfil activo</label>
    <div class="flex gap-2">
        <select
            id="profile-switcher"
            bind:value={selected}
            class="flex-1 text-sm bg-[var(--bg-surface)] border border-[var(--border)] rounded
                   px-2 py-1.5 text-[var(--text)] outline-none focus:border-[var(--accent)]"
        >
            {#each PROFILES as p}
                <option value={p}>{p}</option>
            {/each}
        </select>
        <button
            type="button"
            onclick={() => { if (selected !== currentProfile) showConfirm = true; }}
            disabled={selected === currentProfile}
            class="text-sm px-3 py-1.5 rounded bg-[var(--accent)] text-white disabled:opacity-50"
        >
            Cambiar
        </button>
    </div>
    <p class="text-xs text-[var(--text-faint)] mt-1">
        Algunos cambios requieren reinicio manual del servicio.
    </p>
</div>

{#if showConfirm}
    <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
        <div class="bg-[var(--bg)] border border-[var(--border)] rounded-xl p-6 max-w-sm w-full mx-4">
            <h2 class="font-semibold text-[var(--text)] mb-2">Confirmar cambio de perfil</h2>
            <p class="text-sm text-[var(--text-muted)] mb-4">
                Cambiando a <strong>{selected}</strong>. Los servicios Docker requieren
                reinicio manual para aplicar el perfil completamente.
            </p>
            <div class="flex gap-2 justify-end">
                <button
                    type="button"
                    onclick={() => showConfirm = false}
                    class="text-sm px-3 py-1.5 border border-[var(--border)] rounded text-[var(--text-muted)]"
                >
                    Cancelar
                </button>
                <button
                    type="button"
                    onclick={() => { onSwitch(selected); showConfirm = false; }}
                    class="text-sm px-3 py-1.5 bg-[var(--accent)] text-white rounded"
                >
                    Confirmar
                </button>
            </div>
        </div>
    </div>
{/if}
