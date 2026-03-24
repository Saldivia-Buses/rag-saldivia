// DEV ONLY — auto-login como admin para testear el UI sin gateway
// Solo funciona si NODE_ENV !== 'production'
import { redirect } from '@sveltejs/kit';
import { SignJWT } from 'jose';
import { setSessionCookie } from '$lib/server/auth';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ cookies }) => {
    if (process.env.NODE_ENV === 'production') {
        return new Response('Not available in production', { status: 403 });
    }

    const secret = new TextEncoder().encode(process.env.JWT_SECRET ?? 'dev-secret-local');

    const token = await new SignJWT({
        user_id: 1,
        email: 'admin@saldivia.com',
        name: 'Admin Dev',
        role: 'admin',
        area_id: 1,
    })
        .setProtectedHeader({ alg: 'HS256' })
        .setExpirationTime('8h')
        .sign(secret);

    setSessionCookie(cookies, token);
    throw redirect(302, '/chat');
};
