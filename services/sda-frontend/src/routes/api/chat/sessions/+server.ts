import type { RequestHandler } from './$types';
import { gatewayListSessions, gatewayCreateSession } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ locals }) => {
    if (!locals.user) throw error(401);
    const data = await gatewayListSessions(locals.user.id);
    return json(data);
};

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);
    const { collection, crossdoc } = await request.json();
    const session = await gatewayCreateSession(locals.user.id, collection, crossdoc);
    return json(session, { status: 201 });
};
