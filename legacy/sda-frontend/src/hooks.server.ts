// src/hooks.server.ts
import type { Handle, HandleServerError } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { verifySession, clearSessionCookie } from '$lib/server/auth';
import { GatewayError } from '$lib/server/gateway';

export const handle: Handle = async ({ event, resolve }) => {
    event.locals.user = await verifySession(event.cookies) ?? null;

    try {
        return await resolve(event);
    } catch (err) {
        // If any server load/action gets a 401 from gateway, the JWT is likely
        // expired or the API key was rotated. Clear the stale cookie and force
        // re-login so the user doesn't see a cryptic error.
        if (err instanceof GatewayError && err.status === 401) {
            clearSessionCookie(event.cookies);
            throw redirect(302, '/login');
        }
        throw err;
    }
};

export const handleError: HandleServerError = ({ error, event, status }) => {
    console.error(`[server error] ${status} ${event.url.pathname}`, error);
    return {
        message: status === 404 ? 'Página no encontrada' : 'Error interno del servidor',
    };
};
