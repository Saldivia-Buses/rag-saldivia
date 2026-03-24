import type { PageServerLoad } from './$types';
import { gatewayListCollections, gatewayListActiveJobs } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ cookies }) => {
	const token = cookies.get('sda_session') ?? '';

	const [collectionsResult, activeJobs] = await Promise.all([
		gatewayListCollections().catch(() => ({ collections: [] as string[] })),
		token ? gatewayListActiveJobs(token).catch(() => []) : Promise.resolve([]),
	]);

	return {
		collections: collectionsResult.collections ?? [],
		activeJobs,
	};
};
