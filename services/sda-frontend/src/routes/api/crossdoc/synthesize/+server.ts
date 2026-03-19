import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';
import { gatewayGenerateStream, GatewayError } from '$lib/server/gateway';
import { SYNTHESIS_PROMPT } from '$lib/crossdoc/pipeline';
import type { SubResult } from '$lib/crossdoc/types';

export const POST: RequestHandler = async ({ request, locals }) => {
	if (!locals.user) throw error(401, 'Unauthorized');

	const { question, results }: { question: string; results: SubResult[] } = await request.json();
	if (!question?.trim()) throw error(400, 'question is required');

	const successResults = (results ?? []).filter(r => r.success);

	try {
		const resp = await gatewayGenerateStream({
			messages: [{ role: 'user', content: SYNTHESIS_PROMPT(question, successResults) }],
			use_knowledge_base: false,
			max_tokens: 4096,
		});

		return new Response(resp.body, {
			status: resp.status,
			headers: {
				'Content-Type': 'text/event-stream',
				'Cache-Control': 'no-cache',
				'Connection': 'keep-alive',
			},
		});
	} catch (err) {
		if (err instanceof GatewayError) throw error(err.status, err.detail);
		throw error(502, 'Synthesis failed');
	}
};
