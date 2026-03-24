<script lang="ts">
    import { Upload, File as FileIcon, X } from 'lucide-svelte';
    import TierBadge from './TierBadge.svelte';
    import { classifyTierBySize, type Tier } from '$lib/ingestion/types.js';

    let {
        onFile,
        disabled = false,
    }: {
        onFile: (file: File) => void;
        disabled?: boolean;
    } = $props();

    const ACCEPTED_EXTENSIONS = ['.pdf', '.txt', '.md', '.docx'];
    const ACCEPTED_MIME = [
        'application/pdf', 'text/plain', 'text/markdown',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ];

    let selectedFile = $state<File | null>(null);
    let dragging = $state(false);
    let fileError = $state('');
    let estimatedTier = $derived<Tier | null>(
        selectedFile ? classifyTierBySize(selectedFile.size) : null
    );

    function validateFile(file: File): string {
        const ext = '.' + (file.name.split('.').pop()?.toLowerCase() ?? '');
        if (!ACCEPTED_EXTENSIONS.includes(ext))
            return `Tipo no soportado. Acepta: ${ACCEPTED_EXTENSIONS.join(', ')}`;
        if (file.size > 50 * 1024 * 1024)
            return 'El archivo no puede superar 50 MB.';
        return '';
    }

    function handleFileSelect(files: FileList | null) {
        if (!files || files.length === 0) return;
        const file = files[0];
        fileError = validateFile(file);
        if (!fileError) {
            selectedFile = file;
            onFile(file);
        } else {
            selectedFile = null;
        }
    }

    function clearFile(e: MouseEvent) {
        e.stopPropagation();
        selectedFile = null;
        fileError = '';
    }
</script>

<div
    role="button"
    tabindex={disabled ? -1 : 0}
    ondragenter={(e) => { e.preventDefault(); dragging = true; }}
    ondragleave={(e) => { e.preventDefault(); dragging = false; }}
    ondragover={(e) => e.preventDefault()}
    ondrop={(e) => { e.preventDefault(); dragging = false; handleFileSelect(e.dataTransfer?.files ?? null); }}
    onclick={() => !disabled && document.getElementById('file-input-dz')?.click()}
    onkeydown={(e) => e.key === 'Enter' && !disabled && document.getElementById('file-input-dz')?.click()}
    class="border-2 border-dashed rounded-[var(--radius-lg)] p-8 text-center transition-colors
           {disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
           {dragging ? 'border-[var(--accent)] bg-[var(--accent)]/5'
                     : 'border-[var(--border)] hover:border-[var(--text-faint)]'}"
>
    <input
        id="file-input-dz"
        type="file"
        class="hidden"
        accept={ACCEPTED_MIME.join(',')}
        onchange={(e) => handleFileSelect((e.target as HTMLInputElement).files)}
        {disabled}
    />

    {#if selectedFile}
        <div class="flex items-center justify-center gap-3">
            <FileIcon size={20} class="text-[var(--accent)]" />
            <span class="text-sm font-medium text-[var(--text)]">{selectedFile.name}</span>
            {#if estimatedTier}
                <TierBadge tier={estimatedTier} />
            {/if}
            <button onclick={clearFile} class="text-[var(--text-faint)] hover:text-[var(--text)]" aria-label="Quitar archivo">
                <X size={16} />
            </button>
        </div>
        <p class="text-xs text-[var(--text-faint)] mt-1.5">
            {(selectedFile.size / 1024).toFixed(1)} KB
        </p>
    {:else}
        <Upload size={24} class="text-[var(--text-faint)] mx-auto mb-3" />
        <p class="text-sm text-[var(--text-muted)]">
            Arrastrá un archivo o <span class="text-[var(--accent)]">hacé click para elegir</span>
        </p>
        <p class="text-xs text-[var(--text-faint)] mt-1">
            {ACCEPTED_EXTENSIONS.join(', ')} · máx 50 MB
        </p>
    {/if}
</div>

{#if fileError}
    <p class="text-xs text-[var(--danger)] mt-2">{fileError}</p>
{/if}
