import type { PageServerLoad, Actions } from './$types';
import {
    gatewayRefreshKey, gatewayListCollections,
    gatewayUpdatePreferences, gatewayUpdateProfile,
    gatewayUpdatePassword, GatewayError
} from '$lib/server/gateway';
import { fail, redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals, parent }) => {
    if (!locals.user) throw redirect(302, '/login');
    const [{ preferences }, collectionsRes] = await Promise.all([
        parent(),
        gatewayListCollections(),
    ]);
    return {
        user: locals.user,
        preferences,
        collections: collectionsRes.collections,
    };
};

export const actions: Actions = {
    refresh_key: async ({ locals }) => {
        if (!locals.user) return fail(401);
        try {
            const result = await gatewayRefreshKey();
            return { api_key: result.api_key };
        } catch {
            return fail(503, { error: 'No se pudo regenerar la clave. Intentá de nuevo.' });
        }
    },

    update_profile: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        const name = (data.get('name') as string ?? '').trim();
        const avatar_color = data.get('avatar_color') as string ?? '#6366f1';
        const ui_language = data.get('ui_language') as string ?? 'es';

        if (!name) return fail(400, { error: 'El nombre no puede estar vacío', field: 'name' });

        try {
            await Promise.all([
                gatewayUpdateProfile(locals.user.id, name),
                gatewayUpdatePreferences(locals.user.id, { avatar_color, ui_language }),
            ]);
            return { success: true, section: 'profile' };
        } catch {
            return fail(503, { error: 'No se pudo guardar el perfil. Intentá de nuevo.' });
        }
    },

    update_password: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        const current_password = data.get('current_password') as string;
        const new_password = data.get('new_password') as string;
        const confirm_password = data.get('confirm_password') as string;

        if (new_password !== confirm_password)
            return fail(400, { error: 'Las contraseñas no coinciden', field: 'confirm_password' });
        if (new_password.length < 8)
            return fail(400, { error: 'La contraseña debe tener al menos 8 caracteres', field: 'new_password' });

        try {
            await gatewayUpdatePassword(locals.user.id, current_password, new_password);
            return { success: true, section: 'password' };
        } catch (err) {
            if (err instanceof GatewayError && err.status === 400)
                return fail(400, { error: 'Contraseña actual incorrecta', field: 'current_password' });
            throw err;
        }
    },

    update_preferences: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        try {
            await gatewayUpdatePreferences(locals.user.id, {
                default_collection: data.get('default_collection') as string ?? '',
                default_query_mode: data.get('default_query_mode') === 'crossdoc' ? 'crossdoc' : 'standard',
                vdb_top_k: Math.max(1, parseInt(data.get('vdb_top_k') as string, 10) || 10),
                reranker_top_k: Math.max(1, parseInt(data.get('reranker_top_k') as string, 10) || 5),
                max_sub_queries: Math.max(1, parseInt(data.get('max_sub_queries') as string, 10) || 4),
                follow_up_retries: data.get('follow_up_retries') === 'on',
                show_decomposition: data.get('show_decomposition') === 'on',
            });
            return { success: true, section: 'preferences' };
        } catch {
            return fail(503, { error: 'No se pudo guardar las preferencias. Intentá de nuevo.' });
        }
    },

    update_notifications: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        try {
            await gatewayUpdatePreferences(locals.user.id, {
                notify_ingestion_done: data.get('notify_ingestion_done') === 'on',
                notify_system_alerts: data.get('notify_system_alerts') === 'on',
            });
            return { success: true, section: 'notifications' };
        } catch {
            return fail(503, { error: 'No se pudo guardar las notificaciones. Intentá de nuevo.' });
        }
    },
};
