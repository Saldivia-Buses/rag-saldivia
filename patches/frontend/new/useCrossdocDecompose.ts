/**
 * useCrossdocDecompose — LLM-based query decomposition for cross-document retrieval.
 *
 * Sends the user's question to the LLM (via RAG API with use_knowledge_base=false)
 * to decompose it into multiple retrieval-focused sub-queries.
 * Deduplicates using Jaccard similarity. Generates follow-up alternatives for failed queries.
 */

import { useCallback } from "react";
import { useSettingsStore } from "../store/useSettingsStore";

const JACCARD_THRESHOLD = 0.65;

/** Compute Jaccard similarity between two strings (word-level). */
function jaccard(a: string, b: string): number {
  const setA = new Set(a.toLowerCase().split(/\s+/));
  const setB = new Set(b.toLowerCase().split(/\s+/));
  const intersection = new Set([...setA].filter((x) => setB.has(x)));
  const union = new Set([...setA, ...setB]);
  return union.size === 0 ? 0 : intersection.size / union.size;
}

/** Remove sub-queries that are too similar to ones already in the list. */
function dedup(queries: string[]): string[] {
  const result: string[] = [];
  for (const q of queries) {
    if (!result.some((existing) => jaccard(existing, q) >= JACCARD_THRESHOLD)) {
      result.push(q);
    }
  }
  return result;
}

interface DecomposeOptions {
  ragUrl?: string;
  maxSubQueries?: number;
}

export function useCrossdocDecompose() {
  const settings = useSettingsStore();

  const decompose = useCallback(
    async (question: string, opts: DecomposeOptions = {}): Promise<string[]> => {
      const ragUrl = opts.ragUrl ?? "/api/v1/generate";
      const maxSub = opts.maxSubQueries ?? 0; // 0 = unlimited

      const prompt = `You are a search query decomposer for a technical document retrieval system.

Given the user's question, generate multiple retrieval-focused sub-queries. Each sub-query should:
- Target a SPECIFIC product, component, or technical specification
- Use generic catalog/manual terminology (not user-specific context)
- Be at most 15 words
- Be independent — each should retrieve different documents

Return ONLY the sub-queries, one per line. No numbering, no explanations.

User question: ${question}`;

      const response = await fetch(ragUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: [{ role: "user", content: prompt }],
          use_knowledge_base: false,
          model: settings.model,
          llm_endpoint: settings.llmEndpoint,
          max_tokens: 2048,
        }),
      });

      if (!response.ok) throw new Error(`Decompose failed: ${response.status}`);

      // Collect the full streamed response
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
                  if (token) text += token;
                } catch { /* skip non-JSON lines */ }
              }
            }
          }
        }
      } else {
        const json = await response.json();
        text = json?.choices?.[0]?.message?.content ?? "";
      }

      // Parse sub-queries (one per line, skip empty/numbered)
      let subQueries = text
        .split("\n")
        .map((line: string) => line.replace(/^\d+[\.\)]\s*/, "").trim())
        .filter((line: string) => line.length > 5 && line.length < 200);

      // Dedup
      subQueries = dedup(subQueries);

      // Cap if maxSubQueries is set
      if (maxSub > 0 && subQueries.length > maxSub) {
        subQueries = subQueries.slice(0, maxSub);
      }

      return subQueries;
    },
    [settings.model, settings.llmEndpoint]
  );

  /** Generate follow-up alternatives for failed sub-queries. */
  const generateFollowUps = useCallback(
    async (
      failedQueries: string[],
      ragUrl?: string
    ): Promise<string[]> => {
      const url = ragUrl ?? "/api/v1/generate";
      const prompt = `These search queries returned no useful results:
${failedQueries.map((q) => `- ${q}`).join("\n")}

Generate alternative queries using synonyms, broader terms, or different technical vocabulary.
One query per line, no numbering.`;

      const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: [{ role: "user", content: prompt }],
          use_knowledge_base: false,
          model: settings.model,
          llm_endpoint: settings.llmEndpoint,
          max_tokens: 1024,
        }),
      });

      if (!response.ok) return [];

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
                  if (token) text += token;
                } catch { /* skip */ }
              }
            }
          }
        }
      } else {
        const json = await response.json();
        text = json?.choices?.[0]?.message?.content ?? "";
      }

      return text
        .split("\n")
        .map((line: string) => line.replace(/^\d+[\.\)]\s*/, "").trim())
        .filter((line: string) => line.length > 5 && line.length < 200);
    },
    [settings.model, settings.llmEndpoint]
  );

  return { decompose, generateFollowUps, dedup };
}
