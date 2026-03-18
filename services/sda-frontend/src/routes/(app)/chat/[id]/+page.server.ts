import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { gatewayGetSession, gatewayListSessions, gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params, locals }) => {
    const [sessionData, historyData, collectionsData] = await Promise.all([
        gatewayGetSession(params.id, locals.user!.id),
        gatewayListSessions(locals.user!.id),
        gatewayListCollections(),
    ]);

    if (!sessionData) throw error(404, 'Session not found');

    return {
        session: sessionData,
        history: historyData.sessions,
        collections: collectionsData.collections,
    };
};
