import type { PageServerLoad } from './$types';
import { gatewayListCollections, gatewayCollectionStats } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
    const { collections } = await gatewayListCollections();
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
