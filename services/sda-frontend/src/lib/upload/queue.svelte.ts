import type { Tier } from '$lib/ingestion/types.js';

export type QueuedFile = {
    id: string;
    file: File;
    collection: string;
    hash: string;
    tier: Tier | null;
    state: 'queued' | 'uploading' | 'processing' | 'completed' | 'failed' | 'stalled';
    progress: number;
    eta: number | null;
    jobId: string | null;
    error: string | null;
};

// Tier group determines which slot pool to use
function tierGroup(tier: Tier | null): 'small' | 'large' {
    if (tier === 'medium' || tier === 'large') return 'large';
    return 'small';
}

class UploadQueue {
    items = $state<QueuedFile[]>([]);
    private activeSmall = $state(0);
    private activeLarge = $state(0);

    // Configurable — defaults match workstation-1gpu.yaml
    maxSmallSlots = $state(2);
    maxLargeSlots = $state(1);

    add(item: Omit<QueuedFile, 'state' | 'progress' | 'eta' | 'jobId' | 'error'>): void {
        this.items = [...this.items, {
            ...item, state: 'queued', progress: 0, eta: null, jobId: null, error: null,
        }];
        this._advance();
    }

    update(id: string, updates: Partial<QueuedFile>): void {
        const item = this.items.find(i => i.id === id);
        this.items = this.items.map(i => i.id === id ? { ...i, ...updates } : i);
        if (updates.state === 'completed' || updates.state === 'failed') {
            if (item) {
                const group = tierGroup(item.tier);
                if (group === 'small') this.activeSmall = Math.max(0, this.activeSmall - 1);
                else this.activeLarge = Math.max(0, this.activeLarge - 1);
            }
            this._advance();
        }
    }

    remove(id: string): void {
        this.items = this.items.filter(i => i.id !== id);
    }

    private _advance(): void {
        for (const item of this.items) {
            if (item.state !== 'queued') continue;
            const group = tierGroup(item.tier);
            const [active, max] = group === 'small'
                ? [this.activeSmall, this.maxSmallSlots]
                : [this.activeLarge, this.maxLargeSlots];
            if (active < max) {
                if (group === 'small') this.activeSmall++;
                else this.activeLarge++;
                if (typeof window !== 'undefined') {
                    window.dispatchEvent(new CustomEvent('upload:start', { detail: item.id }));
                }
                break;
            }
        }
    }
}

export const uploadQueue = new UploadQueue();
