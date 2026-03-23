import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayListAlerts } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') {
        throw redirect(302, '/chat');
    }
    try {
        const data = await gatewayListAlerts(false);
        return { alerts: data.alerts.map(({ file_hash, ...rest }) => rest) };
    } catch {
        return { alerts: [] };
    }
};
