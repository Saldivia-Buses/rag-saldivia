<script lang="ts">
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast.svelte';
	import { ingestionStore } from '$lib/stores/ingestion.svelte.js';
	import { IngestPoller } from '$lib/ingestion/poller.js';
	import DropZone from '$lib/components/upload/DropZone.svelte';
	import IngestionQueue from '$lib/components/upload/IngestionQueue.svelte';
	import type { Tier } from '$lib/ingestion/types.js';

	let { data } = $props();

	let selectedFile = $state<File | null>(null);
	let selectedCollection = $state(data.collections[0] ?? '');
	let uploading = $state(false);

	onMount(() => {
		if (data.activeJobs?.length) {
			ingestionStore.hydrateFromServer(data.activeJobs);
			for (const job of ingestionStore.jobs) {
				if (job.state === 'pending' || job.state === 'running') {
					startPoller(job.jobId, job.tier);
				}
			}
		}
	});

	function startPoller(jobId: string, tier: Tier) {
		const poller = new IngestPoller(jobId, tier);
		poller.poll(({ state, progress, eta }) => {
			ingestionStore.updateJob(jobId, {
				state: state as any,
				progress,
				eta,
			});
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

	async function handleUpload() {
		if (!selectedFile || !selectedCollection || uploading) return;
		uploading = true;

		try {
			const formData = new FormData();
			formData.append('file', selectedFile);
			formData.append('collection', selectedCollection);

			const res = await fetch('/api/upload', { method: 'POST', body: formData });
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				throw new Error((body as any).message ?? `Error ${res.status}`);
			}

			const { job_id, tier, page_count, filename } = await res.json();

			ingestionStore.addJob({
				jobId: job_id,
				filename,
				collection: selectedCollection,
				tier: tier as Tier,
				pageCount: page_count,
				state: 'pending',
				progress: 0,
				eta: null,
				startedAt: Date.now(),
				lastProgressAt: Date.now(),
			});

			startPoller(job_id, tier as Tier);
			toastStore.success(`"${filename}" enviado a ingesta.`);
			selectedFile = null;
		} catch (e: any) {
			toastStore.error(e.message ?? 'Error al subir el archivo.');
		} finally {
			uploading = false;
		}
	}

	function handleRetry(jobId: string) {
		const job = ingestionStore.jobs.find(j => j.jobId === jobId);
		if (!job) return;
		ingestionStore.updateJob(jobId, { state: 'pending', progress: 0, eta: null });
		startPoller(jobId, job.tier);
	}
</script>

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
		{uploading ? 'Enviando...' : 'Subir documento'}
	</button>

	<IngestionQueue jobs={ingestionStore.jobs} onRetry={handleRetry} />
</div>
