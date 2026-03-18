<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6 max-w-2xl">
    <div class="flex items-center gap-3 mb-6">
        <a href="/collections" class="text-[#475569] hover:text-[#94a3b8] text-sm">← Colecciones</a>
        <h1 class="text-lg font-semibold text-[#e2e8f0]">{data.name}</h1>
    </div>

    <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4 mb-4">
        <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
                <div class="text-[#475569] text-xs mb-1">Entidades</div>
                <div class="text-[#e2e8f0] font-medium">
                    {data.stats?.entity_count?.toLocaleString() ?? '—'}
                </div>
            </div>
            <div>
                <div class="text-[#475569] text-xs mb-1">Documentos</div>
                <div class="text-[#e2e8f0] font-medium">
                    {data.stats?.document_count?.toLocaleString() ?? '—'}
                </div>
            </div>
        </div>
    </div>

    <a href="/chat"
       class="inline-block bg-[#6366f1] hover:bg-[#4f46e5] text-white text-sm
              px-4 py-2 rounded transition-colors">
        Consultar esta colección
    </a>
</div>
