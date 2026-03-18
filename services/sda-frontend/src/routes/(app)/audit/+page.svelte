<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6">
    <h1 class="text-lg font-semibold text-[#e2e8f0] mb-4">Auditoría</h1>
    <table class="w-full text-xs">
        <thead>
            <tr class="text-[#475569] border-b border-[#1e293b]">
                <th class="text-left pb-2">Timestamp</th>
                <th class="text-left pb-2">User ID</th>
                <th class="text-left pb-2">Acción</th>
                <th class="text-left pb-2">Colección</th>
                <th class="text-left pb-2">Preview</th>
            </tr>
        </thead>
        <tbody>
            {#each data.entries as entry}
                <tr class="border-b border-[#1e293b] text-[#64748b]">
                    <td class="py-1.5">{entry.timestamp}</td>
                    <td class="py-1.5">{entry.user_id}</td>
                    <td class="py-1.5">{entry.action}</td>
                    <td class="py-1.5">{entry.collection ?? '—'}</td>
                    <td class="py-1.5 max-w-xs truncate">{entry.query_preview ?? '—'}</td>
                </tr>
            {/each}
        </tbody>
    </table>
</div>
