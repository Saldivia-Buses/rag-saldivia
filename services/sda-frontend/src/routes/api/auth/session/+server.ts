import type { RequestHandler } from './$types';
import { gatewayLogin, GatewayError } from '$lib/server/gateway';
import { setSessionCookie, clearSessionCookie } from '$lib/server/auth';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request, cookies }) => {
    const { email, password } = await request.json();
    try {
        const result = await gatewayLogin(email, password);
        setSessionCookie(cookies, result.token);
        return json({ user: result.user });
    } catch (err) {
        if (err instanceof GatewayError && (err.status === 401 || err.status === 403)) {
            throw error(401, 'Email o contrasena incorrectos');
        }
        throw error(503, 'No se pudo conectar con el servidor de autenticacion.');
    }
};

export const DELETE: RequestHandler = async ({ cookies }) => {
    clearSessionCookie(cookies);
    return json({ ok: true });
};
