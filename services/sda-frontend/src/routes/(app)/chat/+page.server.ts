import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayCreateSession, gatewayListCollections, gatewayListSessions } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    const [{ sessions }, { collections }] = await Promise.all([
        gatewayListSessions(locals.user!.id),
        gatewayListCollections(),
    ]);

    // Redirect to most recent session if one exists
    if (sessions.length > 0) {
        throw redirect(302, `/chat/${sessions[0].id}`);
    }

    // Otherwise create a new one
    const defaultCollection = collections[0] ?? '';
    const session = await gatewayCreateSession(locals.user!.id, defaultCollection);
    throw redirect(302, `/chat/${session.id}`);
};
