import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { gatewayListCollections, gatewayCollectionStats, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
    let collections: string[];
    try {
        ({ collections } = await gatewayListCollections());
    } catch (err) {
        if (err instanceof GatewayError) throw err;
        throw error(503, 'No se pudo cargar la lista de colecciones.');
    }

    const statsResults = await Promise.allSettled(
        collections.map(name => gatewayCollectionStats(name))
    );
    const stats = Object.fromEntries(
        collections.map((name, i) => [
            name,
            statsResults[i].status === 'fulfilled' ? (statsResults[i] as PromiseFulfilledResult<any>).value : null
        ])
    );
    return { collections, stats };
};
