import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    if (!/^[\w-]+$/.test(params.jobId)) throw error(400, 'jobId invalido');

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503);

    await fetch(`${gatewayUrl}/v1/jobs/${params.jobId}/alert`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${apiKey}` },
    }).catch(() => {});

    return json({ ok: true });
};
