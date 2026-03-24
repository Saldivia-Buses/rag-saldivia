<script lang="ts">
    import { enhance } from '$app/forms';
    import { toastStore } from '$lib/stores/toast.svelte';
    import ConfigSlider from '$lib/components/admin/ConfigSlider.svelte';
    import ModelSelector from '$lib/components/admin/ModelSelector.svelte';
    import GuardrailsToggle from '$lib/components/admin/GuardrailsToggle.svelte';
    import ProfileSwitcher from '$lib/components/admin/ProfileSwitcher.svelte';

    let { data, form } = $props();

    const cfg = data.config as Record<string, unknown>;

    // Generación
    let temperature = $state((cfg.temperature as number) ?? 0.7);
    let max_tokens = $state((cfg.max_tokens as number) ?? 2048);
    let top_p = $state((cfg.top_p as number) ?? 0.9);
    let top_k = $state((cfg.top_k as number) ?? 50);

    // Vector DB
    let vdb_top_k = $state((cfg.vdb_top_k as number) ?? 10);
    let reranker_top_k = $state((cfg.reranker_top_k as number) ?? 5);

    // Modelos
    const LLM_MODELS = [
        'nvidia/llama-3.3-nemotron-super-49b-v1.5',
        'nvidia/nemotron',
        'meta/llama-3.1-70b-instruct',
    ];
    const EMBED_MODELS = [
        'nvidia/llama-nemotron-embed-1b-v2',
        'nvidia/nv-embedqa-e5-v5',
    ];
    const RERANKER_MODELS = [
        'nvidia/llama-nemotron-rerank-1b-v2',
        'nvidia/nv-rerankqa-mistral-4b-v3',
    ];
    let llm_model = $state((cfg.llm_model as string) ?? LLM_MODELS[0]);
    let embedding_model = $state((cfg.embedding_model as string) ?? EMBED_MODELS[0]);
    let reranker_model = $state((cfg.reranker_model as string) ?? RERANKER_MODELS[0]);

    // Guardrails
    let guardrails_enabled = $state((cfg.guardrails_enabled as boolean) ?? false);

    let saving = $state(false);
    let resetting = $state(false);
    let currentProfile = $state('workstation-1gpu');

    async function handleProfileSwitch(profile: string) {
        try {
            const res = await fetch('/api/admin/profile', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ profile }),
            });
            if (res.ok) {
                currentProfile = profile;
                toastStore.success(`Perfil cambiado a ${profile} — restart requerido para efecto completo`);
            } else {
                toastStore.error('Error al cambiar el perfil');
            }
        } catch {
            toastStore.error('Error al cambiar el perfil');
        }
    }

    $effect(() => {
        if (form?.success && form.action === 'update') toastStore.success('Configuración actualizada');
        if (form?.success && form.action === 'reset') toastStore.success('Defaults restaurados');
        if (form?.error) toastStore.error(form.error);
    });
</script>

<div class="p-6 max-w-2xl">
    <div class="flex items-center justify-between mb-6">
        <h1 class="text-lg font-semibold text-[var(--text)]">RAG Config</h1>
        <form method="POST" action="?/resetConfig" use:enhance={() => {
            resetting = true;
            return async ({ update }) => { resetting = false; await update(); };
        }}>
            <button
                type="submit"
                disabled={resetting}
                class="text-sm px-3 py-1.5 border border-[var(--border)] rounded text-[var(--text-muted)]
                       hover:border-[var(--accent)] disabled:opacity-50"
            >
                {resetting ? 'Restaurando...' : 'Restaurar defaults'}
            </button>
        </form>
    </div>

    {#if data.error}
        <div class="bg-[var(--danger-bg)] text-[var(--danger)] p-3 rounded mb-4 text-sm">
            {data.error}
        </div>
    {/if}

    <form method="POST" action="?/updateConfig" use:enhance={() => {
        saving = true;
        return async ({ update }) => { saving = false; await update(); };
    }}>
        <!-- Generación -->
        <section class="mb-6">
            <h2 class="text-sm font-semibold text-[var(--text-faint)] uppercase tracking-wide mb-3">Generación</h2>
            <input type="hidden" name="temperature" value={temperature} />
            <input type="hidden" name="max_tokens" value={max_tokens} />
            <input type="hidden" name="top_p" value={top_p} />
            <input type="hidden" name="top_k" value={top_k} />
            <ConfigSlider bind:value={temperature} label="Temperature" min={0} max={2} step={0.05}
                description="Controla la aleatoriedad de las respuestas (0 = determinístico, 2 = más creativo)" />
            <ConfigSlider bind:value={max_tokens} label="Max Tokens" min={256} max={8192} step={256}
                description="Longitud máxima de la respuesta generada" />
            <ConfigSlider bind:value={top_p} label="Top P" min={0} max={1} step={0.05}
                description="Nucleus sampling — tokens acumulados a considerar" />
            <ConfigSlider bind:value={top_k} label="Top K" min={1} max={100} step={1}
                description="Número de tokens candidatos en cada paso de generación" />
        </section>

        <!-- Vector DB -->
        <section class="mb-6">
            <h2 class="text-sm font-semibold text-[var(--text-faint)] uppercase tracking-wide mb-3">Vector DB</h2>
            <input type="hidden" name="vdb_top_k" value={vdb_top_k} />
            <input type="hidden" name="reranker_top_k" value={reranker_top_k} />
            <ConfigSlider bind:value={vdb_top_k} label="VDB Top K" min={1} max={50} step={1}
                description="Chunks recuperados de Milvus por query" />
            <ConfigSlider bind:value={reranker_top_k} label="Reranker Top K" min={1} max={20} step={1}
                description="Chunks enviados al LLM después del reranking" />
        </section>

        <!-- Modelos -->
        <section class="mb-6">
            <h2 class="text-sm font-semibold text-[var(--text-faint)] uppercase tracking-wide mb-3">Modelos</h2>
            <input type="hidden" name="llm_model" value={llm_model} />
            <input type="hidden" name="embedding_model" value={embedding_model} />
            <input type="hidden" name="reranker_model" value={reranker_model} />
            <ModelSelector bind:value={llm_model} label="LLM Model" options={LLM_MODELS} />
            <ModelSelector bind:value={embedding_model} label="Embedding Model" options={EMBED_MODELS} />
            <ModelSelector bind:value={reranker_model} label="Reranker Model" options={RERANKER_MODELS} />
        </section>

        <!-- Guardrails -->
        <section class="mb-6">
            <h2 class="text-sm font-semibold text-[var(--text-faint)] uppercase tracking-wide mb-3">Guardrails</h2>
            <input type="hidden" name="guardrails_enabled" value={guardrails_enabled} />
            <GuardrailsToggle bind:value={guardrails_enabled}
                description="Activa filtros de contenido en las respuestas del LLM" />
        </section>

        <div class="border-t border-[var(--border)] pt-4 flex justify-end">
            <button
                type="submit"
                disabled={saving}
                class="bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white text-sm
                       px-4 py-2 rounded disabled:opacity-50"
            >
                {saving ? 'Guardando...' : 'Guardar cambios'}
            </button>
        </div>
    </form>

    <!-- Perfil (fuera del form principal) -->
    <section class="mt-6 border-t border-[var(--border)] pt-6">
        <h2 class="text-sm font-semibold text-[var(--text-faint)] uppercase tracking-wide mb-3">Perfil de deployment</h2>
        <ProfileSwitcher {currentProfile} onSwitch={handleProfileSwitch} />
    </section>
</div>
