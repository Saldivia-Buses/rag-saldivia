import type { CrossdocOptions, CrossdocProgress, SubResult } from '$lib/crossdoc/types';
import { DEFAULT_CROSSDOC_OPTIONS } from '$lib/crossdoc/types';
import type { ChatStore } from './chat.svelte';

const MAX_PARALLEL = 6;

export class CrossdocStore {
    progress = $state<CrossdocProgress | null>(null);
    options  = $state<CrossdocOptions>({ ...DEFAULT_CROSSDOC_OPTIONS });

    private abortCtrl: AbortController | null = null;

    async run(question: string, chat: ChatStore, collectionNames?: string[]): Promise<void> {
        this.abortCtrl = new AbortController();
        const signal = this.abortCtrl.signal;
        chat.startStream();

        try {
            // Phase 1: Decompose
            this.progress = { phase: 'decomposing', subQueries: [], completed: 0, total: 0, results: [] };
            const decompResp = await fetch('/api/crossdoc/decompose', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question, maxSubQueries: this.options.maxSubQueries }),
                signal,
            });
            if (!decompResp.ok) throw new Error(`Decompose failed: ${decompResp.status}`);
            const { subQueries } = await decompResp.json();

            // Phase 2: Parallel sub-queries in batches
            const results: SubResult[] = [];
            this.progress = { phase: 'querying', subQueries, completed: 0, total: subQueries.length, results: [] };

            for (let i = 0; i < subQueries.length; i += MAX_PARALLEL) {
                if (signal.aborted) break;
                const batch = subQueries.slice(i, i + MAX_PARALLEL);
                const batchResults = await Promise.allSettled(
                    batch.map(async (query: string) => {
                        const resp = await fetch('/api/crossdoc/subquery', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({
                                query,
                                vdbTopK: this.options.vdbTopK,
                                rerankerTopK: this.options.rerankerTopK,
                                ...(collectionNames && collectionNames.length > 0 && { collection_names: collectionNames }),
                            }),
                            signal,
                        });
                        if (!resp.ok) return { query, content: '', success: false } as SubResult;
                        const data = await resp.json();
                        return { query, content: data.content, success: data.success } as SubResult;
                    })
                );
                for (const r of batchResults) {
                    results.push(r.status === 'fulfilled' ? r.value : { query: '', content: '', success: false });
                }
                this.progress = { ...this.progress!, completed: results.length, results: [...results] };
            }

            // Phase 3: Follow-up retries for failed queries
            if (this.options.followUpRetries && !signal.aborted) {
                const failed = results.filter(r => !r.success && r.query).map(r => r.query);
                if (failed.length > 0) {
                    this.progress = { ...this.progress!, phase: 'retrying' };
                    const altResp = await fetch('/api/crossdoc/decompose', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            question: `Failed queries:\n${failed.join('\n')}\nGenerate alternatives.`,
                            maxSubQueries: failed.length,
                        }),
                        signal,
                    });
                    if (altResp.ok) {
                        const { subQueries: alternatives } = await altResp.json();
                        for (const alt of alternatives.slice(0, 10)) {
                            if (signal.aborted) break;
                            const resp = await fetch('/api/crossdoc/subquery', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({
                                    query: alt,
                                    vdbTopK: this.options.vdbTopK,
                                    rerankerTopK: this.options.rerankerTopK,
                                    ...(collectionNames && collectionNames.length > 0 && { collection_names: collectionNames }),
                                }),
                                signal,
                            });
                            if (resp.ok) {
                                const data = await resp.json();
                                if (data.success) results.push({ query: alt, content: data.content, success: true });
                            }
                        }
                    }
                }
            }

            // Phase 4: Synthesis → SSE stream
            if (!signal.aborted) {
                this.progress = { ...this.progress!, phase: 'synthesizing', results: [...results] };
                const synthResp = await fetch('/api/crossdoc/synthesize', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ question, results }),
                    signal,
                });
                if (synthResp.ok && synthResp.body) {
                    const reader = synthResp.body.getReader();
                    const decoder = new TextDecoder();
                    let buffer = '';
                    try {
                        while (true) {
                            const { done, value } = await reader.read();
                            if (done) break;
                            buffer += decoder.decode(value, { stream: true });
                            const lines = buffer.split('\n');
                            buffer = lines.pop() ?? '';
                            for (const line of lines) {
                                if (!line.startsWith('data: ')) continue;
                                const data = line.slice(6).trim();
                                if (data === '[DONE]') continue;
                                try {
                                    const obj = JSON.parse(data);
                                    const token = obj?.choices?.[0]?.delta?.content;
                                    if (token) chat.appendToken(token);
                                } catch { /* skip malformed */ }
                            }
                        }
                    } finally {
                        reader.releaseLock();
                    }
                }
            }

            const subQueriesUsed = this.progress?.subQueries ?? [];
            this.progress = { phase: 'done', subQueries: subQueriesUsed, completed: results.length, total: results.length, results };
            chat.finalizeStream({ crossdocResults: results });

        } catch (err) {
            if ((err as Error).name === 'AbortError') {
                chat.finalizeStream();
                return;
            }
            this.progress = {
                phase: 'error',
                subQueries: [],
                completed: 0,
                total: 0,
                results: [],
                error: (err as Error).message,
            };
            chat.finalizeStream();
        }
    }

    stop() { this.abortCtrl?.abort(); }
    reset() { this.progress = null; }
}

export const crossdoc = new CrossdocStore();
