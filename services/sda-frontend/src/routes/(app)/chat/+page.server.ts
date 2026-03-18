import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayCreateSession, gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    const { collections } = await gatewayListCollections();
    const defaultCollection = collections[0] ?? '';
    const session = await gatewayCreateSession(locals.user!.id, defaultCollection);
    throw redirect(302, `/chat/${session.id}`);
};
