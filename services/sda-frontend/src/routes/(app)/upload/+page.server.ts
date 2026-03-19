import type { PageServerLoad } from './$types';
import { gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
	try {
		const { collections } = await gatewayListCollections();
		return { collections };
	} catch (err) {
		console.error('[upload loader]', err);
		return { collections: [] };
	}
};
