import type { PageServerLoad, Actions } from './$types';
import { error, redirect } from '@sveltejs/kit';
import { gatewayCollectionStats, gatewayDeleteCollection, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params }) => {
    try {
        const stats = await gatewayCollectionStats(params.name);
        return { name: params.name, stats };
    } catch (err) {
        if (err instanceof GatewayError && err.status === 404) {
            throw error(404, `Colección "${params.name}" no encontrada.`);
        }
        console.error('[collection detail loader]', err);
        return { name: params.name, stats: null, error: 'No se pudo cargar las estadísticas de la colección' };
    }
};

export const actions: Actions = {
    delete: async ({ params, locals }) => {
        if (!locals.user) throw error(401);
        try {
            await gatewayDeleteCollection(params.name);
        } catch (err) {
            const status = err instanceof GatewayError ? err.status : 503;
            throw error(status, `No se pudo eliminar la colección "${params.name}".`);
        }
        throw redirect(303, '/collections');
    },
};
