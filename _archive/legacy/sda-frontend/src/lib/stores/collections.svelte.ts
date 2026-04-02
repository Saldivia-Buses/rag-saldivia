// Svelte 5 runes-based store for collections state.
// Used by CommandPalette (Fase 17) for client-side collection search.

export class CollectionsStore {
	collections = $state<string[]>([]);
	loading = $state(false);

	/** Hydrate from server-loaded data (call from +page.svelte) */
	init(collections: string[]) {
		this.collections = collections;
	}

	/** Client-side refresh from BFF */
	async load(): Promise<void> {
		this.loading = true;
		try {
			const res = await fetch('/api/collections');
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data = await res.json();
			this.collections = data.collections ?? [];
		} catch {
			this.collections = [];
		} finally {
			this.loading = false;
		}
	}

	async create(name: string, schema = 'default'): Promise<void> {
		const res = await fetch('/api/collections', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name, schema }),
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error((err as any).message ?? `Error ${res.status}`);
		}
		this.collections = [...this.collections, name];
	}

	async delete(name: string): Promise<void> {
		const res = await fetch(`/api/collections/${name}`, { method: 'DELETE' });
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			throw new Error((err as any).message ?? `Error ${res.status}`);
		}
		this.collections = this.collections.filter(c => c !== name);
	}
}

export const collectionsStore = new CollectionsStore();
