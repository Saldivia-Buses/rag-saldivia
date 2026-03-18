// src/hooks.server.ts
import type { Handle } from '@sveltejs/kit';
import { verifySession } from '$lib/server/auth';

export const handle: Handle = async ({ event, resolve }) => {
    event.locals.user = await verifySession(event.cookies) ?? null;
    return resolve(event);
};
