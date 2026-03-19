import type { RequestHandler } from './$types';
import { gatewayGetSession, gatewayDeleteSession, GatewayError } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    try {
        const session = await gatewayGetSession(params.id, locals.user.id);
        if (!session) throw error(404, 'Session not found');
        return json(session);
    } catch (err) {
        if (err instanceof GatewayError) throw error(err.status, err.detail);
        throw err;
    }
};

export const DELETE: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    try {
        await gatewayDeleteSession(params.id, locals.user.id);
        return json({ ok: true });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo eliminar la sesión.');
    }
};
