import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayGetPreferences, gatewayListAlerts } from '$lib/server/gateway';
import { DEFAULT_PREFERENCES } from '$lib/types/preferences';

export const load: LayoutServerLoad = async ({ locals, url }) => {
    if (!locals.user) {
        // Only pass relative paths to prevent open redirects
        const next = url.pathname.startsWith('/') ? url.pathname : '/chat';
        throw redirect(302, `/login?next=${encodeURIComponent(next)}`);
    }

    const [preferences, alerts] = await Promise.all([
        gatewayGetPreferences(locals.user.id).catch((err) => {
            console.error('[layout] Failed to load preferences:', err);
            return DEFAULT_PREFERENCES;
        }),
        locals.user.role === 'admin'
            ? gatewayListAlerts(false).then(r => r.alerts).catch(() => [])
            : Promise.resolve([]),
    ]);

    // Notificaciones pendientes (in-app toasts)
    const notifications: { type: 'ingestion' | 'alert'; message: string }[] = [];
    if (preferences.notify_system_alerts && alerts.length > 0) {
        notifications.push({
            type: 'alert',
            message: `${alerts.length} alerta${alerts.length > 1 ? 's' : ''} de sistema sin resolver`
        });
    }

    return { user: locals.user, preferences, notifications };
};
