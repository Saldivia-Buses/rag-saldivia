import type { RequestHandler } from './$types';
import { gatewayDeleteCollection, GatewayError } from '$lib/server/gateway';
import { error } from '@sveltejs/kit';

export const DELETE: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    try {
        await gatewayDeleteCollection(params.name);
        return new Response(null, { status: 204 });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, `No se pudo eliminar la colección "${params.name}".`);
    }
};
