import type { PageServerLoad, Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
    gatewayListAreas, gatewayGetAreaCollections,
    gatewayListCollections, gatewayGrantCollection, gatewayRevokeCollection
} from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    const role = locals.user?.role;
    if (role !== 'admin' && role !== 'area_manager') throw redirect(302, '/chat');

    try {
        const [allAreasData, collectionsData] = await Promise.all([
            gatewayListAreas(),
            gatewayListCollections(),
        ]);

        const areasWithCollections = await Promise.all(
            allAreasData.areas.map(async (area) => {
                try {
                    const colData = await gatewayGetAreaCollections(area.id);
                    return { ...area, collections: colData.collections };
                } catch {
                    return { ...area, collections: [] };
                }
            })
        );

        return {
            areas: areasWithCollections,
            allCollections: collectionsData.collections,
            isManager: role === 'area_manager',
        };
    } catch {
        return { areas: [], allCollections: [], isManager: false, error: 'No se pudo cargar los permisos' };
    }
};

export const actions: Actions = {
    grant: async ({ request, locals }) => {
        const role = locals.user?.role;
        if (role !== 'admin' && role !== 'area_manager') return fail(403, { error: 'Sin permisos' });
        const data = await request.formData();
        try {
            await gatewayGrantCollection(
                Number(data.get('area_id')),
                data.get('collection') as string,
                data.get('permission') as string || 'read'
            );
            return { success: true };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? 'Error al asignar colección' });
        }
    },
    revoke: async ({ request, locals }) => {
        const role = locals.user?.role;
        if (role !== 'admin' && role !== 'area_manager') return fail(403, { error: 'Sin permisos' });
        const data = await request.formData();
        try {
            await gatewayRevokeCollection(
                Number(data.get('area_id')),
                data.get('collection') as string
            );
            return { success: true };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? 'Error al revocar colección' });
        }
    }
};
