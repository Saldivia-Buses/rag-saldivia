import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    let resp: Response;
    try {
        resp = await fetch(`${gatewayUrl}/v1/jobs/${params.jobId}/status`, {
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'X-User-Id': String(locals.user.id),
            },
        });
    } catch {
        throw error(502, 'Gateway inalcanzable.');
    }

    const body = await resp.json().catch(() => ({}));
    return json(body, { status: resp.status });
};
