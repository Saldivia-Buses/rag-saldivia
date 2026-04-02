import type { RequestHandler } from './$types';
import { gatewayMessageFeedback, GatewayError } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ params, locals, request }) => {
    if (!locals.user) throw error(401);
    const { rating } = await request.json();
    if (rating !== 'up' && rating !== 'down') throw error(400, 'Rating inválido.');
    try {
        await gatewayMessageFeedback(params.id, Number(params.msgId), locals.user.id, rating);
        return json({ ok: true });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo guardar el feedback.');
    }
};
