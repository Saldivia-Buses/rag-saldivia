import type { RequestHandler } from './$types';
import { gatewayListSessions, gatewayCreateSession, GatewayError } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ locals }) => {
    if (!locals.user) throw error(401);
    try {
        const data = await gatewayListSessions(locals.user.id);
        return json(data);
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudieron cargar las sesiones.');
    }
};

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);
    try {
        const { collection, crossdoc } = await request.json();
        const session = await gatewayCreateSession(locals.user.id, collection, crossdoc);
        return json(session, { status: 201 });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo crear la sesión.');
    }
};
