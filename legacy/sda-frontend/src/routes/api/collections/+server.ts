import type { RequestHandler } from './$types';
import { gatewayListCollections, gatewayCreateCollection, GatewayError } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ locals }) => {
    if (!locals.user) throw error(401);
    try {
        const data = await gatewayListCollections();
        return json(data);
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudieron cargar las colecciones.');
    }
};

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);
    const body = await request.json().catch(() => ({}));
    const { name, schema = 'default' } = body as { name?: string; schema?: string };
    if (!name?.trim()) throw error(400, 'El nombre de la colección es requerido.');
    try {
        const result = await gatewayCreateCollection(name.trim(), schema);
        return json(result, { status: 201 });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo crear la colección.');
    }
};
