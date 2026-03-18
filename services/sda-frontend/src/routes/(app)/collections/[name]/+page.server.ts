import type { PageServerLoad } from './$types';
import { gatewayCollectionStats } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params }) => {
    const stats = await gatewayCollectionStats(params.name);
    return { name: params.name, stats };
};
