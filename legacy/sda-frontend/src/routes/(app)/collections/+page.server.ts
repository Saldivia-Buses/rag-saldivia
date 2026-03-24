import type { PageServerLoad } from './$types';
import { gatewayListCollections, gatewayCollectionStats } from '$lib/server/gateway';
import type { CollectionStats } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
    let collections: string[] = [];
    try {
        ({ collections } = await gatewayListCollections());
    } catch (err) {
        console.error('[collections loader]', err);
        return { collections: [], stats: {}, error: 'No se pudo cargar la lista de colecciones' };
    }

    const statsResults = await Promise.allSettled(
        collections.map(name => gatewayCollectionStats(name))
    );
    const stats: Record<string, CollectionStats | null> = Object.fromEntries(
        collections.map((name, i) => [
            name,
            statsResults[i].status === 'fulfilled' ? (statsResults[i] as PromiseFulfilledResult<any>).value : null
        ])
    );
    return { collections, stats };
};
