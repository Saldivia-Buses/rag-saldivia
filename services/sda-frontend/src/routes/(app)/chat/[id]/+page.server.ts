import type { PageServerLoad } from './$types';
import { error, redirect } from '@sveltejs/kit';
import { gatewayGetSession, gatewayListSessions, gatewayListCollections, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params, locals }) => {
    let sessionData;
    try {
        sessionData = await gatewayGetSession(params.id, locals.user!.id);
    } catch (err: any) {
        if (err.status === 404 || err.status === 403) {
            throw redirect(302, '/chat');
        }
        if (err instanceof GatewayError) throw err;
        throw error(503, 'No se pudo cargar la sesión. El servidor no responde.');
    }

    if (!sessionData) throw redirect(302, '/chat');

    // Fetch history and collections in parallel. If these fail, still show the
    // session but with empty sidebar/collection data so the page isn't broken.
    let history: any[] = [];
    let collections: string[] = [];
    try {
        const [historyData, collectionsData] = await Promise.all([
            gatewayListSessions(locals.user!.id),
            gatewayListCollections(),
        ]);
        history = historyData.sessions;
        collections = collectionsData.collections;
    } catch {
        // Non-critical: page renders with empty sidebar/collections
    }

    return {
        session: sessionData,
        history,
        collections,
    };
};
