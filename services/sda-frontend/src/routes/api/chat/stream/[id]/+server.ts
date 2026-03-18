import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';

const GATEWAY_URL = process.env.GATEWAY_URL ?? 'http://localhost:9000';
const SYSTEM_API_KEY = process.env.SYSTEM_API_KEY;

async function persistMessage(sessionId: string, userId: number, role: string, content: string) {
    if (!SYSTEM_API_KEY || !content) return;
    try {
        await fetch(`${GATEWAY_URL}/chat/sessions/${sessionId}/messages?user_id=${userId}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${SYSTEM_API_KEY}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ role, content }),
        });
    } catch {
        // Non-fatal: persistence failure doesn't break the stream
    }
}

export const POST: RequestHandler = async ({ params, request, locals }) => {
    if (!locals.user) throw error(401, 'Unauthorized');
    if (!SYSTEM_API_KEY) throw error(500, 'Server misconfiguration');

    const { query, collection_names, crossdoc } = await request.json();
    const sessionId = params.id;
    const userId = locals.user.id;

    // Persist user message (non-blocking)
    persistMessage(sessionId, userId, 'user', query).catch(() => {});

    // Forward to gateway /v1/generate as SSE
    const gatewayResp = await fetch(`${GATEWAY_URL}/v1/generate`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${SYSTEM_API_KEY}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            messages: [{ role: 'user', content: query }],
            collection_names,
            use_knowledge_base: true,
        }),
    });

    if (!gatewayResp.ok) {
        throw error(gatewayResp.status, 'Gateway error');
    }

    if (!gatewayResp.body) {
        throw error(502, 'No response body from gateway');
    }

    // Tee the stream: one branch to browser, one to accumulate assistant response
    const [browserStream, accStream] = gatewayResp.body.tee();

    // Accumulate and persist assistant message in background
    (async () => {
        const decoder = new TextDecoder();
        const reader = accStream.getReader();
        let assistantContent = '';
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                const text = decoder.decode(value, { stream: true });
                for (const line of text.split('\n')) {
                    if (!line.startsWith('data: ')) continue;
                    const data = line.slice(6).trim();
                    if (data === '[DONE]') continue;
                    try {
                        const obj = JSON.parse(data);
                        assistantContent += obj?.choices?.[0]?.delta?.content ?? '';
                    } catch { /* ignore malformed chunks */ }
                }
            }
        } finally {
            reader.releaseLock();
        }
        await persistMessage(sessionId, userId, 'assistant', assistantContent).catch(() => {});
    })();

    // Return browser branch immediately
    return new Response(browserStream, {
        headers: {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
        },
    });
};
