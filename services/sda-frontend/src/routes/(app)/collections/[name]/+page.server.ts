import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { gatewayCollectionStats, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params }) => {
    try {
        const stats = await gatewayCollectionStats(params.name);
        return { name: params.name, stats };
    } catch (err) {
        if (err instanceof GatewayError && err.status === 404) {
            throw error(404, `Colección "${params.name}" no encontrada.`);
        }
        if (err instanceof GatewayError) throw err;
        throw error(503, 'No se pudo cargar las estadísticas de la colección.');
    }
};
