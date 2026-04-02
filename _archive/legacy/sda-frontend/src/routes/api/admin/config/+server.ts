// src/routes/api/admin/config/+server.ts
import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import {
    gatewayGetRagConfig,
    gatewayUpdateRagConfig,
    gatewayResetRagConfig,
    GatewayError,
} from '$lib/server/gateway';

export const GET: RequestHandler = async ({ locals }) => {
    if (locals.user?.role !== 'admin') error(403, 'Admin only');
    try {
        const config = await gatewayGetRagConfig();
        return json(config);
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo obtener la configuración RAG.');
    }
};

export const PATCH: RequestHandler = async ({ locals, request }) => {
    if (locals.user?.role !== 'admin') error(403, 'Admin only');
    try {
        const params = await request.json();
        await gatewayUpdateRagConfig(params);
        return json({ ok: true });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo actualizar la configuración RAG.');
    }
};

export const POST: RequestHandler = async ({ locals, url }) => {
    if (locals.user?.role !== 'admin') error(403, 'Admin only');
    if (url.searchParams.get('action') === 'reset') {
        try {
            await gatewayResetRagConfig();
            return json({ ok: true });
        } catch (err) {
            const status = err instanceof GatewayError ? err.status : 503;
            throw error(status, 'No se pudo resetear la configuración RAG.');
        }
    }
    error(400, 'Unknown action');
};
