import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import {
    gatewayListAlerts, gatewayListUsers, gatewayListAreas,
    gatewayListCollections, gatewayCollectionStats, gatewayListActiveJobs
} from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');

    const systemApiKey = process.env.SYSTEM_API_KEY;
    if (!systemApiKey) console.error('[admin/system] SYSTEM_API_KEY no está definida — jobs no disponibles');

    const [alertsResult, usersResult, areasResult, collectionsResult, jobsResult] =
        await Promise.allSettled([
            gatewayListAlerts(false),
            gatewayListUsers(),
            gatewayListAreas(),
            gatewayListCollections(),
            systemApiKey
                ? gatewayListActiveJobs(systemApiKey)
                : Promise.reject(new Error('SYSTEM_API_KEY missing')),
        ]);

    const alerts = alertsResult.status === 'fulfilled'
        ? alertsResult.value.alerts.map(({ file_hash, ...rest }) => rest)
        : [];

    // Colecciones con ≥1 documento
    let collectionsWithDocs: number | null = null;
    if (collectionsResult.status === 'fulfilled') {
        const colNames = collectionsResult.value.collections;
        const statsResults = await Promise.allSettled(
            colNames.map(name => gatewayCollectionStats(name))
        );
        collectionsWithDocs = statsResults.filter(
            r => r.status === 'fulfilled' && r.value.entity_count > 0
        ).length;
    }

    const stats = {
        activeUsers: usersResult.status === 'fulfilled'
            ? usersResult.value.users.filter(u => u.active).length
            : null,
        totalAreas: areasResult.status === 'fulfilled'
            ? areasResult.value.areas.length
            : null,
        collectionsWithDocs,
    };

    const activeJobs = jobsResult.status === 'fulfilled' ? jobsResult.value : [];

    return { alerts, stats, activeJobs };
};
