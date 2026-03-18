import type { RequestHandler } from './$types';
import { gatewayLogin } from '$lib/server/gateway';
import { setSessionCookie, clearSessionCookie } from '$lib/server/auth';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request, cookies }) => {
    const { email, password } = await request.json();
    try {
        const result = await gatewayLogin(email, password);
        setSessionCookie(cookies, result.token);
        return json({ user: result.user });
    } catch (e: any) {
        throw error(401, e.detail ?? 'Invalid credentials');
    }
};

export const DELETE: RequestHandler = async ({ cookies }) => {
    clearSessionCookie(cookies);
    return json({ ok: true });
};
