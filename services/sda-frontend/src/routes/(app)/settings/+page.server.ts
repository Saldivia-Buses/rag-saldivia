import type { PageServerLoad, Actions } from './$types';
import { gatewayRefreshKey } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    return { user: locals.user };
};

export const actions: Actions = {
    refresh_key: async ({ locals }) => {
        const result = await gatewayRefreshKey(locals.user!.id);
        return { api_key: result.api_key };
    }
};
