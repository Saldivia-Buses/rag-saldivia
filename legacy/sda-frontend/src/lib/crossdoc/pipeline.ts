const JACCARD_THRESHOLD = 0.65;
const MAX_RESPONSE_CHARS = 15_000;

export function jaccard(a: string, b: string): number {
	const normalize = (s: string) =>
		s
			.toLowerCase()
			.normalize('NFD')
			.replace(/[\u0300-\u036f]/g, '');
	const setA = new Set(normalize(a).split(/\s+/).filter(Boolean));
	const setB = new Set(normalize(b).split(/\s+/).filter(Boolean));
	const intersection = [...setA].filter(x => setB.has(x)).length;
	const union = new Set([...setA, ...setB]).size;
	return union === 0 ? 0 : intersection / union;
}

export function dedup(queries: string[]): string[] {
	const result: string[] = [];
	for (const q of queries) {
		if (!result.some(existing => jaccard(existing, q) >= JACCARD_THRESHOLD)) {
			result.push(q);
		}
	}
	return result;
}

export function parseSubQueries(text: string, maxSubQueries = 0): string[] {
	let queries = text
		.split('\n')
		.map(line => line.replace(/^\d+[\.\)]\s*/, '').trim())
		.filter(line => line.length > 5 && line.length < 200);
	queries = dedup(queries);
	if (maxSubQueries > 0) queries = queries.slice(0, maxSubQueries);
	return queries;
}

export function hasUsefulData(text: string): boolean {
	const trimmed = text.trim();
	if (trimmed.length < 3) return false;
	const emptyPatterns = [
		/^(no|sin)\s+(information|data|results|context)/i,
		/^out of context$/i,
		/^i (cannot|can't|don't)/i,
	];
	return !emptyPatterns.some(p => p.test(trimmed));
}

export function truncateIfRepetitive(text: string): string {
	const WINDOW = 60;
	const THRESHOLD = 3;
	if (text.length <= WINDOW * THRESHOLD) return text;
	const tail = text.slice(-WINDOW);
	const preceding = text.slice(-(WINDOW * (THRESHOLD + 1)), -WINDOW);
	if (preceding.split(tail).length - 1 >= THRESHOLD - 1) {
		const firstIdx = text.indexOf(tail);
		if (firstIdx > 0 && firstIdx < text.length - WINDOW) {
			return text.slice(0, firstIdx + WINDOW);
		}
	}
	if (text.length > MAX_RESPONSE_CHARS) return text.slice(0, MAX_RESPONSE_CHARS);
	return text;
}

export const DECOMPOSE_PROMPT = (question: string) =>
	`You are a search query decomposer for a technical document retrieval system.

Given the user's question, generate multiple retrieval-focused sub-queries. Each sub-query should:
- Target a SPECIFIC product, component, or technical specification
- Use generic catalog/manual terminology (not user-specific context)
- Be at most 15 words
- Be independent — each should retrieve different documents

Return ONLY the sub-queries, one per line. No numbering, no explanations.

User question: ${question}`;

export const FOLLOWUP_PROMPT = (failedQueries: string[]) =>
	`These search queries returned no useful results:
${failedQueries.map(q => `- ${q}`).join('\n')}

Generate alternative queries using synonyms, broader terms, or different technical vocabulary.
One query per line, no numbering.`;

export const SYNTHESIS_PROMPT = (question: string, results: { query: string; content: string }[]) => {
	const context = results
		.map((r, i) => `[Sub-query ${i + 1}: "${r.query}"]\n${r.content}`)
		.join('\n\n---\n\n');
	return `You are a senior engineer writing a comprehensive technical answer.

Based on the following retrieval results from multiple sub-queries, write a single unified answer to the user's original question.

Rules:
- Cite sources when possible (mention which sub-query or document the info came from)
- Include specific numbers, measurements, and technical specifications
- Be thorough but concise — cover all relevant information
- Use professional technical language

Original question: ${question}

Retrieval results:
${context}`;
};
