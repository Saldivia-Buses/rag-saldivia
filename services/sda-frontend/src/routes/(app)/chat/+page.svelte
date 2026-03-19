<script lang="ts">
    const suggestions = [
        '¿Cuáles son las especificaciones del Aries 365?',
        'Normativas de homologación vigentes',
        '¿Qué documentos hay disponibles?',
    ];
</script>

<div class="flex-1 flex items-center justify-center h-full p-8 min-h-[500px]">
    <div class="text-center max-w-lg">
        <div class="w-14 h-14 bg-[var(--accent)] rounded-2xl flex items-center justify-center mx-auto mb-5">
            <span class="text-white font-bold text-xl">S</span>
        </div>

        <h1 class="text-xl font-semibold text-[var(--text)] mb-2">
            ¿En qué puedo ayudarte?
        </h1>
        <p class="text-sm text-[var(--text-muted)] mb-8">
            Consultá sobre los documentos de Saldivia Buses
        </p>

        <div class="flex flex-wrap gap-2 justify-center">
            {#each suggestions as suggestion}
                <a
                    href="/chat/new"
                    class="
                        px-3.5 py-2 rounded-full text-sm
                        bg-[var(--bg-surface)] border border-[var(--border)]
                        text-[var(--text-muted)] hover:text-[var(--text)] hover:border-[var(--accent)]
                        transition-colors
                    "
                >
                    {suggestion}
                </a>
            {/each}
        </div>
    </div>
</div>
