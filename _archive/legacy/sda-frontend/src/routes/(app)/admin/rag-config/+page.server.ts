import { redirect, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import {
    gatewayGetRagConfig,
    gatewayUpdateRagConfig,
    gatewayResetRagConfig,
    gatewaySwitchProfile,
} from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');
    try {
        const config = await gatewayGetRagConfig();
        return { config };
    } catch {
        return { config: {} as Record<string, unknown>, error: 'No se pudo cargar la configuración' };
    }
};

export const actions: Actions = {
    updateConfig: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        const params: Record<string, unknown> = {};
        for (const [key, val] of data.entries()) {
            const str = val as string;
            if (str === 'true') params[key] = true;
            else if (str === 'false') params[key] = false;
            else {
                const num = Number(str);
                params[key] = isNaN(num) ? str : num;
            }
        }
        try {
            await gatewayUpdateRagConfig(params);
            return { success: true, action: 'update' };
        } catch (e: any) {
            return fail(500, { error: e?.detail ?? 'Error al guardar' });
        }
    },

    resetConfig: async ({ locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        try {
            await gatewayResetRagConfig();
            return { success: true, action: 'reset' };
        } catch (e: any) {
            return fail(500, { error: e?.detail ?? 'Error al restaurar defaults' });
        }
    },

    switchProfile: async ({ request, locals }) => {
        if (locals.user?.role !== 'admin') return fail(403, { error: 'Admin only' });
        const data = await request.formData();
        const profile = data.get('profile') as string;
        try {
            await gatewaySwitchProfile(profile);
            return { success: true, action: 'profile', profile };
        } catch (e: any) {
            return fail(500, { error: e?.detail ?? 'Error al cambiar perfil' });
        }
    },
};
