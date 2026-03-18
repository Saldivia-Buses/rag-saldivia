import type { PageServerLoad, Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { gatewayListUsers, gatewayCreateUser, gatewayDeleteUser,
         gatewayListAreas } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') {
        throw redirect(302, '/chat');
    }
    const [usersData, areasData] = await Promise.all([
        gatewayListUsers(),
        gatewayListAreas(),
    ]);
    return { users: usersData.users, areas: areasData.areas };
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
            return fail(400, { error: e.detail ?? 'Error creating user' });
        }
    },
    delete: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        await gatewayDeleteUser(Number(data.get('id')));
        return { success: true };
    }
};
