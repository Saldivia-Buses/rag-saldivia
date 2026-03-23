import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const PATCH: RequestHandler = async ({ params, request, locals }) => {
    if (!locals.user || locals.user.role !== 'admin') throw error(403);

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503);

    const body = await request.json().catch(() => ({}));
    const resp = await fetch(
        `${gatewayUrl}/v1/admin/alerts/${params.alertId}/resolve`,
        {
            method: 'PATCH',
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body),
        },
    ).catch(() => { throw error(502); });

    return json(await resp.json(), { status: resp.status });
};
