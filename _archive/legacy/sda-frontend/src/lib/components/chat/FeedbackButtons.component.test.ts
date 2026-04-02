// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import FeedbackButtons from './FeedbackButtons.svelte';

describe('FeedbackButtons', () => {
    afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });

    it('renderiza dos botones', () => {
        const { getAllByRole } = render(FeedbackButtons, {
            props: { messageId: 1, sessionId: 'ses-1' }
        });
        expect(getAllByRole('button')).toHaveLength(2);
    });

    it('click en up hace POST al BFF con rating up', async () => {
        const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ ok: true }) });
        vi.stubGlobal('fetch', fetchMock);

        const { getAllByRole } = render(FeedbackButtons, {
            props: { messageId: 1, sessionId: 'ses-1' }
        });
        await fireEvent.click(getAllByRole('button')[0]);

        expect(fetchMock).toHaveBeenCalledWith(
            '/api/chat/sessions/ses-1/messages/1/feedback',
            expect.objectContaining({ method: 'POST' })
        );
    });

    it('click en down hace POST con rating down', async () => {
        const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ ok: true }) });
        vi.stubGlobal('fetch', fetchMock);

        const { getAllByRole } = render(FeedbackButtons, {
            props: { messageId: 2, sessionId: 'ses-2' }
        });
        await fireEvent.click(getAllByRole('button')[1]);

        const body = JSON.parse(fetchMock.mock.calls[0][1].body);
        expect(body.rating).toBe('down');
    });

    it('no hace fetch si messageId es undefined', async () => {
        const fetchMock = vi.fn();
        vi.stubGlobal('fetch', fetchMock);

        const { getAllByRole } = render(FeedbackButtons, {
            props: { messageId: undefined, sessionId: 'ses-1' }
        });
        await fireEvent.click(getAllByRole('button')[0]);

        expect(fetchMock).not.toHaveBeenCalled();
    });
});
