import type { RequestHandler } from './$types';
import { gatewayGetSession, gatewayDeleteSession } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    const session = await gatewayGetSession(params.id, locals.user.id);
    if (!session) throw error(404, 'Session not found');
    return json(session);
};

export const DELETE: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    await gatewayDeleteSession(params.id, locals.user.id);
    return json({ ok: true });
};
