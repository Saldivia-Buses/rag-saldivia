/**
 * useCrossdocStream — Orchestrates the crossdoc query flow:
 *   1. Decompose user question into sub-queries (LLM call)
 *   2. Execute each sub-query as an independent RAG call (parallel)
 *   3. Optionally retry failed queries with follow-up alternatives
 *   4. Synthesize all results into a final answer (LLM call)
 *
 * Uses the RAG API with use_knowledge_base=true for retrieval
 * and use_knowledge_base=false for LLM-only calls (decomposition, synthesis).
 */

import { useCallback, useRef } from "react";
import { useSettingsStore } from "../store/useSettingsStore";
import { useCollectionsStore } from "../store/useCollectionsStore";
import { useCrossdocDecompose } from "./useCrossdocDecompose";

const MAX_PARALLEL = 6;
const REPETITION_WINDOW = 60;
const REPETITION_THRESHOLD = 3;
const MAX_RESPONSE_CHARS = 15000;

interface CrossdocResult {
  query: string;
  response: string;
  sources: string[];
  success: boolean;
}

interface CrossdocProgress {
  phase: "decomposing" | "querying" | "following-up" | "synthesizing" | "done" | "error";
  subQueries?: string[];
  completed?: number;
  total?: number;
  results?: CrossdocResult[];
  synthesis?: string;
  error?: string;
}

/** Detect repetition in streamed text. Returns truncation index or -1. */
function detectRepetition(text: string): number {
  if (text.length <= REPETITION_WINDOW * REPETITION_THRESHOLD) return -1;
  const tail = text.slice(-REPETITION_WINDOW);
  const preceding = text.slice(
    -(REPETITION_WINDOW * (REPETITION_THRESHOLD + 1)),
    -REPETITION_WINDOW
  );
  if (
    preceding.split(tail).length - 1 >= REPETITION_THRESHOLD - 1
  ) {
    const firstIdx = text.indexOf(tail);
    if (firstIdx > 0 && firstIdx < text.length - REPETITION_WINDOW) {
      return firstIdx + REPETITION_WINDOW;
    }
  }
  return -1;
}

/** Collect full streamed response from an SSE endpoint. */
async function collectStream(response: Response): Promise<string> {
  let text = "";
  if (response.headers.get("content-type")?.includes("text/event-stream")) {
    const reader = response.body?.getReader();
    const decoder = new TextDecoder();
    if (reader) {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        const chunk = decoder.decode(value, { stream: true });
        for (const line of chunk.split("\n")) {
          if (line.startsWith("data: ")) {
            try {
              const data = JSON.parse(line.slice(6));
              const token = data?.choices?.[0]?.delta?.content;
              if (token && token.length < 500) text += token;
            } catch { /* skip */ }
          }
        }
        // Repetition guard
        const truncIdx = detectRepetition(text);
        if (truncIdx > 0) { text = text.slice(0, truncIdx); break; }
        if (text.length > MAX_RESPONSE_CHARS) break;
      }
    }
  } else {
    const json = await response.json();
    text = json?.choices?.[0]?.message?.content ?? "";
  }
  return text;
}

/** Check if a response contains useful data (not empty / error). */
function hasUsefulData(text: string): boolean {
  const trimmed = text.trim();
  if (trimmed.length < 3) return false;
  const emptyPatterns = [
    /^(no|sin)\s+(information|data|results|context)/i,
    /^out of context$/i,
    /^i (cannot|can't|don't)/i,
    /^$/,
  ];
  return !emptyPatterns.some((p) => p.test(trimmed));
}

export function useCrossdocStream() {
  const settings = useSettingsStore();
  const { selectedCollections } = useCollectionsStore();
  const { decompose, generateFollowUps } = useCrossdocDecompose();
  const abortRef = useRef<AbortController | null>(null);

  const execute = useCallback(
    async (
      question: string,
      onProgress: (progress: CrossdocProgress) => void
    ): Promise<void> => {
      const ragUrl = "/api/v1/generate";
      abortRef.current = new AbortController();

      try {
        // Phase 1: Decompose
        onProgress({ phase: "decomposing" });
        const subQueries = await decompose(question, {
          ragUrl,
          maxSubQueries: settings.crossdocMaxSubQueries ?? 0,
        });
        onProgress({ phase: "querying", subQueries, completed: 0, total: subQueries.length });

        // Phase 2: Parallel RAG queries
        const results: CrossdocResult[] = [];
        const batches: string[][] = [];
        for (let i = 0; i < subQueries.length; i += MAX_PARALLEL) {
          batches.push(subQueries.slice(i, i + MAX_PARALLEL));
        }

        let completed = 0;
        for (const batch of batches) {
          if (abortRef.current?.signal.aborted) break;

          const batchResults = await Promise.allSettled(
            batch.map(async (query) => {
              const resp = await fetch(ragUrl, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                signal: abortRef.current?.signal,
                body: JSON.stringify({
                  messages: [{ role: "user", content: query }],
                  use_knowledge_base: true,
                  collection_names: selectedCollections.length > 0 ? selectedCollections : undefined,
                  vdb_top_k: settings.vdbTopK ?? 100,
                  reranker_top_k: settings.rerankerTopK ?? 25,
                  enable_reranker: true,
                  enable_citations: true,
                  max_tokens: 2048,
                }),
              });
              const text = await collectStream(resp);
              return { query, response: text, sources: [], success: hasUsefulData(text) };
            })
          );

          for (const result of batchResults) {
            if (result.status === "fulfilled") results.push(result.value);
            else results.push({ query: "", response: "", sources: [], success: false });
            completed++;
          }
          onProgress({ phase: "querying", subQueries, completed, total: subQueries.length, results });
        }

        // Phase 3: Follow-up retries for failed queries
        const enableFollowUp = settings.crossdocFollowUpRetries ?? true;
        if (enableFollowUp) {
          const failedQueries = results.filter((r) => !r.success).map((r) => r.query).filter(Boolean);
          if (failedQueries.length > 0) {
            onProgress({ phase: "following-up", subQueries, completed, total: subQueries.length, results });
            const alternatives = await generateFollowUps(failedQueries, ragUrl);
            for (const alt of alternatives.slice(0, 10)) {
              if (abortRef.current?.signal.aborted) break;
              try {
                const resp = await fetch(ragUrl, {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  signal: abortRef.current?.signal,
                  body: JSON.stringify({
                    messages: [{ role: "user", content: alt }],
                    use_knowledge_base: true,
                    collection_names: selectedCollections.length > 0 ? selectedCollections : undefined,
                    vdb_top_k: settings.vdbTopK ?? 100,
                    reranker_top_k: settings.rerankerTopK ?? 25,
                    enable_reranker: true,
                    enable_citations: true,
                    max_tokens: 2048,
                  }),
                });
                const text = await collectStream(resp);
                if (hasUsefulData(text)) {
                  results.push({ query: alt, response: text, sources: [], success: true });
                }
              } catch { /* skip failed follow-up */ }
            }
          }
        }

        // Phase 4: Synthesize
        onProgress({ phase: "synthesizing", results });
        const successResults = results.filter((r) => r.success);
        const context = successResults
          .map((r, i) => `[Sub-query ${i + 1}: "${r.query}"]\n${r.response}`)
          .join("\n\n---\n\n");

        const synthesisPrompt = `You are a senior engineer writing a comprehensive technical answer.

Based on the following retrieval results from multiple sub-queries, write a single unified answer to the user's original question.

Rules:
- Cite sources when possible (mention which sub-query or document the info came from)
- Include specific numbers, measurements, and technical specifications
- Be thorough but concise — cover all relevant information
- If results contain calculations or formulas, show them
- Use professional technical language

Original question: ${question}

Retrieval results:
${context}`;

        const synthModel = settings.crossdocSynthesisModel || settings.model;
        const synthResp = await fetch(ragUrl, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            messages: [{ role: "user", content: synthesisPrompt }],
            use_knowledge_base: false,
            model: synthModel,
            llm_endpoint: settings.llmEndpoint,
            max_tokens: 4096,
          }),
        });

        const synthesis = await collectStream(synthResp);
        onProgress({ phase: "done", results, synthesis });
      } catch (error) {
        if ((error as Error).name === "AbortError") return;
        onProgress({
          phase: "error",
          error: (error as Error).message ?? "Crossdoc failed",
        });
      }
    },
    [decompose, generateFollowUps, settings, selectedCollections]
  );

  const abort = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  return { execute, abort };
}
