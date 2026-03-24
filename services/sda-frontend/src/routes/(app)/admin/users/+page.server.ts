import type { PageServerLoad, Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
    gatewayListUsers, gatewayCreateUser, gatewayDeleteUser,
    gatewayListAreas, gatewayAddUserArea, gatewayRemoveUserArea
} from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');
    try {
        const [usersData, areasData] = await Promise.all([
            gatewayListUsers(),
            gatewayListAreas(),
        ]);
        return { users: usersData.users, areas: areasData.areas };
    } catch (err) {
        console.error('[admin/users loader]', err);
        return { users: [], areas: [], error: 'No se pudo cargar la lista de usuarios' };
    }
};

export const actions: Actions = {
    create: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        const email = data.get('email') as string;
        const name = data.get('name') as string;
        const password = data.get('password') as string;
        const role = data.get('role') as string;
        if (!email || !name || !password || !role) {
            return fail(400, { error: 'Todos los campos requeridos deben completarse' });
        }
        const areaIdsRaw = data.getAll('area_ids') as string[];
        const area_ids = areaIdsRaw.map(Number).filter(n => !isNaN(n) && n > 0);
        try {
            const result = await gatewayCreateUser({
                email,
                name,
                area_ids,
                role,
                password,
            });
            return { success: true, api_key: result.api_key };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? e?.message ?? 'Error al crear el usuario' });
        }
    },
    delete: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            await gatewayDeleteUser(Number(data.get('id')));
            return { success: true };
        } catch {
            return fail(503, { error: 'No se pudo desactivar el usuario. Intentá de nuevo.' });
        }
    },
    add_area: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            await gatewayAddUserArea(Number(data.get('user_id')), Number(data.get('area_id')));
            return { success: true };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? 'Error al asignar área' });
        }
    },
    remove_area: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            await gatewayRemoveUserArea(Number(data.get('user_id')), Number(data.get('area_id')));
            return { success: true };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? 'Error al remover área' });
        }
    }
};
