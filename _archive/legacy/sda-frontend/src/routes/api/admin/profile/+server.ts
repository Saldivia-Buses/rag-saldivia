// src/routes/api/admin/profile/+server.ts
import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { gatewaySwitchProfile, GatewayError } from '$lib/server/gateway';

export const POST: RequestHandler = async ({ locals, request }) => {
    if (locals.user?.role !== 'admin') error(403, 'Admin only');
    const { profile } = await request.json();
    if (!profile) error(400, 'profile required');
    try {
        await gatewaySwitchProfile(profile as string);
        return json({ ok: true, profile });
    } catch (err) {
        const status = err instanceof GatewayError ? (err as GatewayError).status : 503;
        throw error(status, 'No se pudo cambiar el perfil.');
    }
};
