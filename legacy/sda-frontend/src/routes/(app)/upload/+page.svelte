<script lang="ts">
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast.svelte';
	import { ingestionStore } from '$lib/stores/ingestion.svelte.js';
	import { IngestPoller } from '$lib/ingestion/poller.js';
	import { computeSHA256 } from '$lib/ingestion/hash.js';
	import { uploadQueue } from '$lib/upload/queue.svelte.js';
	import DropZone from '$lib/components/upload/DropZone.svelte';
	import IngestionQueue from '$lib/components/upload/IngestionQueue.svelte';
	import DuplicateModal from '$lib/components/upload/DuplicateModal.svelte';
	import type { Tier } from '$lib/ingestion/types.js';
	import type { JobState } from '$lib/stores/ingestion.svelte.js';

	let { data } = $props();

	let selectedFile = $state<File | null>(null);
	let selectedCollection = $state(data.collections[0] ?? '');
	let uploading = $state(false);

	// Duplicate detection state
	let duplicateInfo = $state<{
		filename: string;
		collection: string;
		state: 'completed' | 'failed' | 'stalled';
		indexedAt?: string | null;
		pages?: number | null;
		pendingFile: File;
		pendingHash: string;
	} | null>(null);

	onMount(() => {
		// Hydrate active jobs from server (page recovery on reload)
		if (data.activeJobs?.length) {
			ingestionStore.hydrateFromServer(data.activeJobs);
			for (const job of ingestionStore.jobs) {
				if (job.state === 'pending' || job.state === 'running') {
					startPoller(job.jobId, job.tier);
				}
			}
		}

		// Listen for queue advancement events
		function onUploadStart(e: Event) {
			const itemId = (e as CustomEvent<string>).detail;
			const item = uploadQueue.items.find(i => i.id === itemId);
			if (item) executeUpload(item.id, item.file, item.collection, item.tier);
		}
		window.addEventListener('upload:start', onUploadStart);
		return () => window.removeEventListener('upload:start', onUploadStart);
	});

	function startPoller(jobId: string, tier: Tier) {
		const poller = new IngestPoller(jobId, tier);
		poller.poll(({ state, progress, eta }) => {
			ingestionStore.updateJob(jobId, { state: state as JobState, progress, eta });
			if (state === 'completed') {
				const filename = ingestionStore.jobs.find(j => j.jobId === jobId)?.filename;
				toastStore.success(`"${filename}" indexado correctamente.`);
				setTimeout(() => ingestionStore.removeJob(jobId), 5_000);
			}
			if (state === 'failed') {
				const filename = ingestionStore.jobs.find(j => j.jobId === jobId)?.filename;
				toastStore.error(`Error al ingestar "${filename}".`);
			}
		});
	}

	async function executeUpload(itemId: string, file: File, collection: string, tier: Tier | null) {
		uploadQueue.update(itemId, { state: 'uploading' });
		try {
			const formData = new FormData();
			formData.append('file', file);
			formData.append('collection', collection);

			const res = await fetch('/api/upload', { method: 'POST', body: formData });
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				throw new Error((body as any).message ?? `Error ${res.status}`);
			}

			const { job_id, tier: actualTier, page_count, filename } = await res.json();

			uploadQueue.update(itemId, { state: 'processing', jobId: job_id, tier: actualTier });
			ingestionStore.addJob({
				jobId: job_id,
				filename,
				collection,
				tier: actualTier as Tier,
				pageCount: page_count,
				state: 'pending',
				progress: 0,
				eta: null,
				startedAt: Date.now(),
				lastProgressAt: Date.now(),
			});

			startPoller(job_id, actualTier as Tier);
			toastStore.success(`"${filename}" enviado a ingesta.`);
		} catch (e: any) {
			uploadQueue.update(itemId, { state: 'failed', error: e.message ?? 'Error al subir' });
			toastStore.error(e.message ?? 'Error al subir el archivo.');
		}
	}

	async function handleUpload() {
		if (!selectedFile || !selectedCollection || uploading) return;
		uploading = true;

		try {
			const hash = await computeSHA256(selectedFile);

			// Check for duplicate server-side
			const checkRes = await fetch(
				`/api/documents/check?hash=${encodeURIComponent(hash)}&collection=${encodeURIComponent(selectedCollection)}`
			);
			if (checkRes.ok) {
				const dupData = await checkRes.json();
				if (dupData.exists && dupData.state !== 'failed') {
					duplicateInfo = {
						filename: selectedFile.name,
						collection: selectedCollection,
						state: dupData.state as 'completed' | 'failed' | 'stalled',
						indexedAt: dupData.indexed_at,
						pages: dupData.pages,
						pendingFile: selectedFile,
						pendingHash: hash,
					};
					uploading = false;
					return;
				}
			} else {
				toastStore.warning('No se pudo verificar duplicados. Continuando con la subida...');
			}

			enqueueFile(selectedFile, hash);
			selectedFile = null;
		} catch (e: any) {
			toastStore.error(e.message ?? 'Error al subir el archivo.');
		} finally {
			uploading = false;
		}
	}

	function enqueueFile(file: File, hash: string) {
		const id = crypto.randomUUID();
		uploadQueue.add({ id, file, collection: selectedCollection, hash, tier: null });
	}

	function handleDuplicateConfirm() {
		if (!duplicateInfo) return;
		enqueueFile(duplicateInfo.pendingFile, duplicateInfo.pendingHash);
		selectedFile = null;
		duplicateInfo = null;
	}

	function handleDuplicateCancel() {
		duplicateInfo = null;
	}

	function handleRetry(jobId: string) {
		const job = ingestionStore.jobs.find(j => j.jobId === jobId);
		if (!job) return;
		ingestionStore.updateJob(jobId, { state: 'pending', progress: 0, eta: null });
		startPoller(jobId, job.tier);
	}
</script>

{#if duplicateInfo}
	<DuplicateModal
		filename={duplicateInfo.filename}
		collection={duplicateInfo.collection}
		state={duplicateInfo.state}
		indexedAt={duplicateInfo.indexedAt}
		pages={duplicateInfo.pages}
		onConfirm={handleDuplicateConfirm}
		onCancel={handleDuplicateCancel}
	/>
{/if}

<div class="p-6 max-w-xl">
	<h1 class="text-lg font-semibold text-[var(--text)] mb-6">Subir documentos</h1>

	<DropZone onFile={(f) => (selectedFile = f)} disabled={uploading} />

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

	<button
		onclick={handleUpload}
		disabled={!selectedFile || !selectedCollection || uploading}
		class="mt-5 w-full py-2.5 px-4 text-sm font-medium text-white bg-[var(--accent)]
		       rounded-lg hover:opacity-90 transition-opacity
		       disabled:opacity-40 disabled:cursor-not-allowed"
	>
		{uploading ? 'Verificando...' : 'Subir documento'}
	</button>

	<IngestionQueue jobs={ingestionStore.jobs} onRetry={handleRetry} />
</div>
