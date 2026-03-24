import type { RequestHandler } from './$types';
import { error, json } from '@sveltejs/kit';
import { gatewayGenerateText, GatewayError } from '$lib/server/gateway';
import { parseSubQueries, DECOMPOSE_PROMPT } from '$lib/crossdoc/pipeline';

export const POST: RequestHandler = async ({ request, locals }) => {
	if (!locals.user) throw error(401, 'Unauthorized');

	const { question, maxSubQueries = 0 } = await request.json();
	if (!question?.trim()) throw error(400, 'question is required');

	try {
		const text = await gatewayGenerateText({
			messages: [{ role: 'user', content: DECOMPOSE_PROMPT(question) }],
			use_knowledge_base: false,
			max_tokens: 2048,
		});
		const subQueries = parseSubQueries(text, maxSubQueries);
		return json({ subQueries });
	} catch (err) {
		if (err instanceof GatewayError) throw error(err.status, err.detail);
		throw error(502, 'Decompose failed');
	}
};
