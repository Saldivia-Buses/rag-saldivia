import type { RequestHandler } from './$types';
import { gatewayGetSession, gatewayDeleteSession, gatewayRenameSession, GatewayError } from '$lib/server/gateway';
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

export const PATCH: RequestHandler = async ({ params, locals, request }) => {
    if (!locals.user) throw error(401);
    const body = await request.json();
    const title = typeof body?.title === 'string' ? body.title.trim() : '';
    if (!title) throw error(400, 'El título no puede estar vacío.');
    try {
        await gatewayRenameSession(params.id, locals.user.id, title);
        return json({ ok: true });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo renombrar la sesión.');
    }
};
