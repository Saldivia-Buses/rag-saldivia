<script lang="ts">
	import { toastStore } from '$lib/stores/toast.svelte';
	import { Upload, File as FileIcon, X } from 'lucide-svelte';

	let { data } = $props();

	const ACCEPTED_EXTENSIONS = ['.pdf', '.txt', '.md', '.docx'];
	const ACCEPTED_MIME = [
		'application/pdf',
		'text/plain',
		'text/markdown',
		'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
	];

	let selectedFile = $state<File | null>(null);
	let selectedCollection = $state(data.collections[0] ?? '');
	let dragging = $state(false);
	let uploading = $state(false);
	let progress = $state(0);
	let fileError = $state('');

	function validateFile(file: File): string {
		const ext = '.' + (file.name.split('.').pop()?.toLowerCase() ?? '');
		if (!ACCEPTED_EXTENSIONS.includes(ext)) {
			return `Tipo no soportado. Acepta: ${ACCEPTED_EXTENSIONS.join(', ')}`;
		}
		if (file.size > 50 * 1024 * 1024) return 'El archivo no puede superar 50 MB.';
		return '';
	}

	function handleFileSelect(files: FileList | null) {
		if (!files || files.length === 0) return;
		const file = files[0];
		fileError = validateFile(file);
		selectedFile = fileError ? null : file;
	}

	function onDragEnter(e: DragEvent) {
		e.preventDefault();
		dragging = true;
	}
	function onDragLeave(e: DragEvent) {
		e.preventDefault();
		dragging = false;
	}
	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		handleFileSelect(e.dataTransfer?.files ?? null);
	}

	async function handleUpload() {
		if (!selectedFile || !selectedCollection) return;
		uploading = true;
		progress = 0;

		// Simulated progress bar (ingestion is async)
		const interval = setInterval(() => {
			progress = Math.min(progress + 10, 85);
		}, 200);

		try {
			const formData = new FormData();
			formData.append('file', selectedFile);
			formData.append('collection', selectedCollection);

			const res = await fetch('/api/upload', { method: 'POST', body: formData });
			clearInterval(interval);
			progress = 100;

			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				throw new Error((body as any).message ?? `Error ${res.status}`);
			}

			toastStore.success(
				`"${selectedFile.name}" enviado a ingesta en "${selectedCollection}".`
			);
			selectedFile = null;
			progress = 0;
		} catch (e: any) {
			clearInterval(interval);
			progress = 0;
			toastStore.error(e.message ?? 'Error al subir el archivo.');
		} finally {
			uploading = false;
		}
	}
</script>

<div class="p-6 max-w-xl">
	<h1 class="text-lg font-semibold text-[var(--text)] mb-6">Subir documentos</h1>

	<!-- Drop zone -->
	<div
		role="button"
		tabindex="0"
		ondragenter={onDragEnter}
		ondragleave={onDragLeave}
		ondragover={(e) => e.preventDefault()}
		ondrop={onDrop}
		onclick={() => document.getElementById('file-input')?.click()}
		onkeydown={(e) => e.key === 'Enter' && document.getElementById('file-input')?.click()}
		class="border-2 border-dashed rounded-[var(--radius-lg)] p-8 text-center cursor-pointer
		       transition-colors {dragging
			   ? 'border-[var(--accent)] bg-[var(--accent)]/5'
			   : 'border-[var(--border)] hover:border-[var(--text-faint)]'}"
	>
		<input
			id="file-input"
			type="file"
			class="hidden"
			accept={ACCEPTED_MIME.join(',')}
			onchange={(e) => handleFileSelect((e.target as HTMLInputElement).files)}
		/>

		{#if selectedFile}
			<div class="flex items-center justify-center gap-3">
				<FileIcon size={20} class="text-[var(--accent)]" />
				<span class="text-sm text-[var(--text)] font-medium">{selectedFile.name}</span>
				<button
					onclick={(e) => {
						e.stopPropagation();
						selectedFile = null;
						fileError = '';
					}}
					class="text-[var(--text-faint)] hover:text-[var(--text)] transition-colors"
					aria-label="Quitar archivo"
				>
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

	<!-- Collection selector -->
	<div class="mt-5">
		<label for="collection-select" class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
			Colección destino
		</label>
		{#if data.collections.length === 0}
			<p class="text-sm text-[var(--text-faint)]">
				No hay colecciones disponibles.
				<a href="/collections" class="text-[var(--accent)] hover:underline">Creá una primero.</a>
			</p>
		{:else}
			<select
				id="collection-select"
				bind:value={selectedCollection}
				class="w-full bg-[var(--bg-base)] border border-[var(--border)] rounded-lg
				       px-3 py-2 text-sm text-[var(--text)] focus:outline-none
				       focus:border-[var(--accent)] transition-colors"
			>
				{#each data.collections as col (col)}
					<option value={col}>{col}</option>
				{/each}
			</select>
		{/if}
	</div>

	<!-- Progress bar -->
	{#if uploading && progress > 0}
		<div class="mt-4">
			<div class="h-1.5 bg-[var(--bg-hover)] rounded-full overflow-hidden">
				<div
					class="h-full bg-[var(--accent)] transition-all duration-200 ease-out rounded-full"
					style="width: {progress}%"
				></div>
			</div>
			<p class="text-xs text-[var(--text-faint)] mt-1">
				{progress < 100 ? 'Subiendo...' : 'Procesando...'}
			</p>
		</div>
	{/if}

	<!-- Upload button -->
	<button
		onclick={handleUpload}
		disabled={!selectedFile || !selectedCollection || uploading}
		class="mt-5 w-full py-2.5 px-4 text-sm font-medium text-white bg-[var(--accent)]
		       rounded-lg hover:opacity-90 transition-opacity
		       disabled:opacity-40 disabled:cursor-not-allowed"
	>
		{uploading ? 'Subiendo...' : 'Subir documento'}
	</button>
</div>
