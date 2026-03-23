import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';
import { GATEWAY_URL } from '$lib/server/gateway';

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);

    const formData = await request.formData();
    const file = formData.get('file');
    const collection = formData.get('collection');

    if (!file || !(file instanceof File)) throw error(400, 'Se requiere un archivo.');
    if (!collection || typeof collection !== 'string' || !collection.trim()) {
        throw error(400, 'Se requiere seleccionar una colección.');
    }

    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    const gw = new FormData();
    gw.append('file', file);
    gw.append('data', JSON.stringify({ collection_name: collection.trim() }));

    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 30_000); // 30s (non-blocking)

    let resp: Response;
    try {
        resp = await fetch(`${GATEWAY_URL}/v1/documents`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'X-User-Id': String(locals.user.id),
            },
            body: gw,
            signal: controller.signal,
        });
    } catch (err) {
        if ((err as any)?.name === 'AbortError') throw error(504, 'Gateway timeout al subir el documento.');
        throw error(502, 'Gateway inalcanzable.');
    } finally {
        clearTimeout(timer);
    }

    const body = await resp.json().catch(() => ({}));
    return json(body, { status: resp.status });
};
