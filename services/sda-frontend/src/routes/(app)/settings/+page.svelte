<script lang="ts">
    import { enhance } from '$app/forms';
    let { data, form } = $props();
</script>

<div class="p-6 max-w-lg">
    <h1 class="text-lg font-semibold text-[#e2e8f0] mb-4">Configuración</h1>

    <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4 mb-4">
        <div class="text-xs text-[#475569] mb-3">Perfil</div>
        <div class="text-sm text-[#e2e8f0]">{data.user?.name}</div>
        <div class="text-xs text-[#64748b]">{data.user?.email}</div>
        <div class="text-xs text-[#6366f1] mt-1">{data.user?.role}</div>
    </div>

    <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4">
        <div class="text-xs text-[#475569] mb-3">API Key personal</div>
        {#if form?.api_key}
            <div class="bg-[#065f46] text-[#6ee7b7] p-2 rounded text-xs font-mono mb-3 break-all">
                {form.api_key}
            </div>
        {/if}
        <form method="POST" action="?/refresh_key" use:enhance>
            <button type="submit"
                    class="bg-[#1e293b] hover:bg-[#334155] text-[#94a3b8] text-xs px-3 py-1.5 rounded">
                Regenerar API key
            </button>
        </form>
    </div>
</div>
