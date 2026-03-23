import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';
import { GATEWAY_URL } from '$lib/server/gateway';

export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);

    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    let resp: Response;
    try {
        resp = await fetch(`${GATEWAY_URL}/v1/jobs/${params.jobId}/status`, {
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
