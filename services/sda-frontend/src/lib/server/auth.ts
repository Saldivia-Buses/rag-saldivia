// src/lib/server/auth.ts
// JWT decode (server-side, reads JWT_SECRET from env).
import { jwtVerify } from 'jose';
import type { SessionUser } from './gateway.js';
import type { Cookies } from '@sveltejs/kit';

const JWT_SECRET_RAW = process.env.JWT_SECRET ?? '';
const secret = new TextEncoder().encode(JWT_SECRET_RAW);
const COOKIE_NAME = 'sda_session';

export async function verifySession(cookies: Cookies): Promise<SessionUser | null> {
    const token = cookies.get(COOKIE_NAME);
    if (!token) return null;
    try {
        const { payload } = await jwtVerify(token, secret);
        return {
            id: payload['user_id'] as number,
            email: payload['email'] as string,
            name: (payload['name'] as string) ?? '',
            role: payload['role'] as string,
            area_id: payload['area_id'] as number,
        };
    } catch {
        return null;
    }
}

export function setSessionCookie(cookies: Cookies, token: string) {
    cookies.set(COOKIE_NAME, token, {
        path: '/',
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 60 * 60 * 8   // 8 hours
    });
}

export function clearSessionCookie(cookies: Cookies) {
    cookies.delete(COOKIE_NAME, { path: '/' });
}
