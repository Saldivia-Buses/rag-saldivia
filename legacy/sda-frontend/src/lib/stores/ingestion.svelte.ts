import type { Tier } from '$lib/ingestion/types.js';

export type JobState = 'pending' | 'running' | 'completed' | 'failed' | 'stalled';

export interface IngestionJob {
    jobId: string;
    filename: string;
    collection: string;
    tier: Tier;
    pageCount: number | null;
    state: JobState;
    progress: number;
    eta: number | null;
    startedAt: number;
    lastProgressAt: number;
}

export interface ServerJob {
    id: string;
    filename: string;
    collection: string;
    tier: string;
    page_count: number | null;
    state: string;
    progress: number;
    created_at: string;
}

let jobs = $state<IngestionJob[]>([]);

export const ingestionStore = {
    get jobs(): IngestionJob[] {
        return jobs;
    },

    addJob(job: IngestionJob): void {
        jobs = [...jobs, job];
    },

    updateJob(jobId: string, updates: Partial<IngestionJob>): void {
        jobs = jobs.map(j => j.jobId === jobId ? { ...j, ...updates } : j);
    },

    removeJob(jobId: string): void {
        jobs = jobs.filter(j => j.jobId !== jobId);
    },

    hydrateFromServer(serverJobs: ServerJob[]): void {
        for (const sj of serverJobs) {
            if (jobs.find(j => j.jobId === sj.id)) continue;
            jobs = [...jobs, {
                jobId: sj.id,
                filename: sj.filename,
                collection: sj.collection,
                tier: sj.tier as Tier,
                pageCount: sj.page_count,
                state: sj.state as JobState,
                progress: sj.progress,
                eta: null,
                startedAt: new Date(sj.created_at).getTime() || Date.now(),
                lastProgressAt: Date.now(),
            }];
        }
    },
};
