import { TIER_CONFIG, estimateETA, type Tier } from './types.js';

function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export class IngestPoller {
    private jobId: string;
    private tier: Tier;
    private lastProgress: number = -1;
    private lastProgressAt: number = Date.now();
    private startedAt: number = Date.now();
    private stopped = false;

    constructor(jobId: string, tier: Tier) {
        this.jobId = jobId;
        this.tier = tier;
    }

    stop(): void {
        this.stopped = true;
    }

    async poll(onUpdate: (update: {
        state: string;
        progress: number;
        eta: number | null;
    }) => void): Promise<void> {
        const config = TIER_CONFIG[this.tier];
        this.startedAt = Date.now();

        while (!this.stopped) {
            let data: { state: string; progress: number };

            try {
                const resp = await fetch(`/api/ingestion/${this.jobId}/status`);
                if (!resp.ok) {
                    onUpdate({ state: 'failed', progress: 0, eta: null });
                    break;
                }
                data = await resp.json();
            } catch {
                onUpdate({ state: 'failed', progress: 0, eta: null });
                break;
            }

            const now = Date.now();
            const elapsedSinceStart = (now - this.startedAt) / 1000;
            const elapsedSinceProgress = (now - this.lastProgressAt) / 1000;
            const eta = estimateETA(this.tier, data.progress, elapsedSinceStart);

            // Deadlock detection
            if (data.progress === this.lastProgress) {
                if (elapsedSinceProgress > config.deadlockThreshold) {
                    onUpdate({ state: 'stalled', progress: data.progress, eta: null });
                    break;
                }
            } else {
                this.lastProgress = data.progress;
                this.lastProgressAt = now;
            }

            onUpdate({ state: data.state, progress: data.progress, eta });

            if (data.state === 'completed' || data.state === 'failed') break;

            await sleep(config.pollInterval * 1000);
        }
    }
}
