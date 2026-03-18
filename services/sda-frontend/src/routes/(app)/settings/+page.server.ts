import type { PageServerLoad, Actions } from './$types';
import { fail } from '@sveltejs/kit';
import { gatewayRefreshKey } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    return { user: locals.user };
};

export const actions: Actions = {
    refresh_key: async ({ locals }) => {
        try {
            const result = await gatewayRefreshKey(locals.user!.id);
            return { api_key: result.api_key };
        } catch {
            return fail(503, { error: 'No se pudo regenerar la clave. Intentá de nuevo.' });
        }
    }
};
