import type { RequestHandler } from './$types';
import { error, json } from '@sveltejs/kit';
import { gatewayGenerateText, GatewayError } from '$lib/server/gateway';
import { hasUsefulData, truncateIfRepetitive } from '$lib/crossdoc/pipeline';

export const POST: RequestHandler = async ({ request, locals }) => {
	if (!locals.user) throw error(401, 'Unauthorized');

	const { query, collection_names, vdbTopK = 10, rerankerTopK = 5 } = await request.json();
	if (!query?.trim()) throw error(400, 'query is required');

	try {
		const raw = await gatewayGenerateText({
			messages: [{ role: 'user', content: query }],
			use_knowledge_base: true,
			collection_names,
			vdb_top_k: vdbTopK,
			reranker_top_k: rerankerTopK,
			enable_reranker: true,
			max_tokens: 2048,
		});
		const content = truncateIfRepetitive(raw);
		return json({ content, success: hasUsefulData(content) });
	} catch (err) {
		if (err instanceof GatewayError) throw error(err.status, err.detail);
		throw error(502, 'Subquery failed');
	}
};
