import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ url, locals }) => {
    if (!locals.user || locals.user.role !== 'admin') throw error(403);

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503);

    const resolved = url.searchParams.get('resolved');
    const qs = resolved !== null ? `?resolved=${resolved}` : '';

    const resp = await fetch(`${gatewayUrl}/v1/admin/alerts${qs}`, {
        headers: { 'Authorization': `Bearer ${apiKey}` },
    }).catch(() => { throw error(502); });

    return json(await resp.json(), { status: resp.status });
};
