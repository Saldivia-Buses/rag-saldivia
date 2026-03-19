import type { PageServerLoad, Actions } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import { gatewayListUsers, gatewayCreateUser, gatewayDeleteUser,
         gatewayListAreas, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') {
        throw redirect(302, '/chat');
    }
    try {
        const [usersData, areasData] = await Promise.all([
            gatewayListUsers(),
            gatewayListAreas(),
        ]);
        return { users: usersData.users, areas: areasData.areas };
    } catch (err) {
        if (err instanceof GatewayError) throw err;
        throw error(503, 'No se pudo cargar la lista de usuarios.');
    }
};

export const actions: Actions = {
    create: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            const result = await gatewayCreateUser({
                email: data.get('email') as string,
                name: data.get('name') as string,
                area_id: Number(data.get('area_id')),
                role: data.get('role') as string,
                password: data.get('password') as string,
            });
            return { success: true, api_key: result.api_key };
        } catch (e: any) {
            const msg = e?.detail ?? e?.message ?? 'Error al crear el usuario';
            return fail(400, { error: msg });
        }
    },
    delete: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            await gatewayDeleteUser(Number(data.get('id')));
            return { success: true };
        } catch (e: any) {
            return fail(503, { error: 'No se pudo desactivar el usuario. Intentá de nuevo.' });
        }
    }
};
