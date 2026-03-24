import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { gatewayLogin, GatewayError } from '$lib/server/gateway';
import { setSessionCookie } from '$lib/server/auth';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user) throw redirect(302, '/chat');
    return {};
};

export const actions: Actions = {
    default: async ({ request, cookies, url }) => {
        const data = await request.formData();
        const email = data.get('email') as string;
        const password = data.get('password') as string;

        if (!email || !password) {
            return fail(400, { error: 'Email y contrasena requeridos' });
        }

        try {
            const result = await gatewayLogin(email, password);
            setSessionCookie(cookies, result.token);
        } catch (e: any) {
            if (e instanceof GatewayError && (e.status === 401 || e.status === 403)) {
                return fail(401, { error: 'Email o contrasena incorrectos' });
            }
            // Gateway unreachable or timeout
            return fail(503, { error: 'El servidor no responde. Intenta de nuevo en unos segundos.' });
        }

        // Validate next is a relative path to prevent open redirects
        const next = url.searchParams.get('next');
        const redirectTo = next?.startsWith('/') ? next : '/chat';
        throw redirect(302, redirectTo);
    }
};
