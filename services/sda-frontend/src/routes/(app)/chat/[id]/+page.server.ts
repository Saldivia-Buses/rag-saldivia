import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayGetSession, gatewayListSessions, gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params, locals }) => {
    let sessionData;
    try {
        sessionData = await gatewayGetSession(params.id, locals.user!.id);
    } catch (err: any) {
        if (err.status === 404 || err.status === 403) {
            throw redirect(302, '/chat');
        }
        throw err;
    }

    if (!sessionData) throw redirect(302, '/chat');

    const [historyData, collectionsData] = await Promise.all([
        gatewayListSessions(locals.user!.id),
        gatewayListCollections(),
    ]);

    return {
        session: sessionData,
        history: historyData.sessions,
        collections: collectionsData.collections,
    };
};
