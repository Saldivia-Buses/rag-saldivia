<script lang="ts">
    let { children, data } = $props();
    const user = $derived(data.user);
    const allowed = $derived(user.role === 'admin' || user.role === 'area_manager');
</script>

{#if allowed}
    {@render children()}
{:else}
    <div class="p-6 text-[#475569]">Acceso denegado.</div>
{/if}
