import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';

const GATEWAY_URL = process.env.GATEWAY_URL ?? 'http://localhost:9000';
const SYSTEM_API_KEY = process.env.SYSTEM_API_KEY;

export const POST: RequestHandler = async ({ params, request, locals }) => {
    if (!locals.user) throw error(401, 'Unauthorized');
    if (!SYSTEM_API_KEY) throw error(500, 'Server misconfiguration');

    const { query, collection_names, crossdoc } = await request.json();

    // Forward to gateway /v1/generate as SSE
    const gatewayResp = await fetch(`${GATEWAY_URL}/v1/generate`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${SYSTEM_API_KEY}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            messages: [{ role: 'user', content: query }],
            collection_names,
            use_knowledge_base: true,
        }),
    });

    if (!gatewayResp.ok) {
        throw error(gatewayResp.status, 'Gateway error');
    }

    // Pipe the SSE stream back to the browser
    return new Response(gatewayResp.body, {
        headers: {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
        },
    });
};
