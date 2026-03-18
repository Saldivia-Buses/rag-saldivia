import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ locals, url }) => {
    if (!locals.user) {
        throw redirect(302, `/login?next=${url.pathname}`);
    }
    return { user: locals.user };
};
