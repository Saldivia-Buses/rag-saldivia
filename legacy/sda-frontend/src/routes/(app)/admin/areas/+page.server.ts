import type { PageServerLoad, Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
    gatewayListAreas, gatewayCreateArea,
    gatewayUpdateArea, gatewayDeleteArea, gatewayListUsers
} from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');
    try {
        const [areasData, usersData] = await Promise.all([
            gatewayListAreas(),
            gatewayListUsers(),
        ]);
        // Calcular conteo de usuarios activos por área
        const userCountByArea: Record<number, number> = {};
        for (const user of usersData.users) {
            if (!user.active) continue;
            for (const area of user.areas) {
                userCountByArea[area.id] = (userCountByArea[area.id] ?? 0) + 1;
            }
        }
        const areas = areasData.areas.map(a => ({
            ...a,
            userCount: userCountByArea[a.id] ?? 0,
        }));
        // Para el modal de eliminar: usuarios por área
        const usersByArea: Record<number, { id: number; email: string; name: string }[]> = {};
        for (const user of usersData.users) {
            if (!user.active) continue;
            for (const area of user.areas) {
                if (!usersByArea[area.id]) usersByArea[area.id] = [];
                usersByArea[area.id].push({ id: user.id, email: user.email, name: user.name });
            }
        }
        return { areas, usersByArea };
    } catch {
        return { areas: [], usersByArea: {}, error: 'No se pudo cargar las áreas' };
    }
};

export const actions: Actions = {
    create: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            const result = await gatewayCreateArea(
                data.get('name') as string,
                (data.get('description') as string) ?? ''
            );
            return { success: true, created: result };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? 'Error al crear el área' });
        }
    },
    update: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            await gatewayUpdateArea(
                Number(data.get('id')),
                data.get('name') as string,
                (data.get('description') as string) ?? ''
            );
            return { success: true };
        } catch (e: any) {
            return fail(400, { error: e?.detail ?? 'Error al actualizar el área' });
        }
    },
    delete: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        try {
            await gatewayDeleteArea(Number(data.get('id')));
            return { success: true };
        } catch (e: any) {
            const status = e?.status === 409 ? 409 : 503;
            const error = e?.status === 409
                ? (e?.detail ?? 'No se puede eliminar el área. Tiene usuarios asignados.')
                : 'Error al eliminar el área. Intentá de nuevo.';
            return fail(status, { error });
        }
    }
};
