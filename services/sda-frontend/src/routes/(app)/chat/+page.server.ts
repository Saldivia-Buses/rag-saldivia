import type { PageServerLoad } from './$types';
import { error, redirect } from '@sveltejs/kit';
import { gatewayCreateSession, gatewayListCollections, gatewayListSessions, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    try {
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
    } catch (err) {
        // Re-throw redirects — they're not errors
        if (err instanceof Response || (err as any)?.status === 302 || (err as any)?.location) throw err;
        // Re-throw GatewayErrors so hooks.server.ts can handle 401
        if (err instanceof GatewayError) throw err;
        throw error(503, 'No se pudo conectar con el servidor. Intentá de nuevo en unos segundos.');
    }
};
