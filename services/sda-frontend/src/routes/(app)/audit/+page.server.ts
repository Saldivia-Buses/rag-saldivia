import type { PageServerLoad } from './$types';
import { error, redirect } from '@sveltejs/kit';
import { gatewayGetAudit, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals, url }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');
    const params = {
        user_id: url.searchParams.get('user_id') ? Number(url.searchParams.get('user_id')) : undefined,
        action: url.searchParams.get('action') ?? undefined,
        collection: url.searchParams.get('collection') ?? undefined,
        limit: 100,
    };
    try {
        const data = await gatewayGetAudit(params);
        return { entries: data.entries };
    } catch (err) {
        if (err instanceof GatewayError) throw err;
        throw error(503, 'No se pudo cargar el log de auditoría.');
    }
};
