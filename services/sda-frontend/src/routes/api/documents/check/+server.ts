import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ url, locals }) => {
    if (!locals.user) throw error(401);

    const hash = url.searchParams.get('hash');
    const collection = url.searchParams.get('collection');
    if (!hash || !collection) throw error(400, 'hash y collection requeridos');

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    let resp: Response;
    try {
        resp = await fetch(
            `${gatewayUrl}/v1/documents/check?hash=${encodeURIComponent(hash)}&collection=${encodeURIComponent(collection)}`,
            { headers: { 'Authorization': `Bearer ${apiKey}` } },
        );
    } catch {
        throw error(502, 'Gateway inalcanzable.');
    }

    const body = await resp.json().catch(() => ({}));
    return json(body, { status: resp.status });
};
