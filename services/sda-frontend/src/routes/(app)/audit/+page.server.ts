import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayGetAudit } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals, url }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');
    const params = {
        user_id: url.searchParams.get('user_id') ? Number(url.searchParams.get('user_id')) : undefined,
        action: url.searchParams.get('action') ?? undefined,
        collection: url.searchParams.get('collection') ?? undefined,
        limit: 100,
    };
    const data = await gatewayGetAudit(params);
    return { entries: data.entries };
};
