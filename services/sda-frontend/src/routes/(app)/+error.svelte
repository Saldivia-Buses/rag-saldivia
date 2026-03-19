<script lang="ts">
    import { page } from '$app/stores';
</script>

<div class="flex items-center justify-center flex-1 p-6">
    <div class="text-center max-w-md">
        <div class="text-5xl font-bold text-[#334155] mb-4">{$page.status}</div>
        <p class="text-[#94a3b8] text-sm mb-6">
            {$page.error?.message ?? 'Ocurrio un error inesperado.'}
        </p>
        <div class="flex gap-3 justify-center">
            <a href="/chat"
               class="bg-[#6366f1] hover:bg-[#4f46e5] text-white text-sm px-4 py-2 rounded transition-colors">
                Ir al Chat
            </a>
            <button onclick={() => window.location.reload()}
                    class="bg-[#1e293b] hover:bg-[#334155] text-[#94a3b8] text-sm px-4 py-2 rounded transition-colors">
                Reintentar
            </button>
        </div>
    </div>
</div>
