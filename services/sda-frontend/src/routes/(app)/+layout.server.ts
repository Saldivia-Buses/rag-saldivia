import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ locals, url }) => {
    if (!locals.user) {
        // Only pass relative paths to prevent open redirects
        const next = url.pathname.startsWith('/') ? url.pathname : '/chat';
        throw redirect(302, `/login?next=${encodeURIComponent(next)}`);
    }
    return { user: locals.user };
};
