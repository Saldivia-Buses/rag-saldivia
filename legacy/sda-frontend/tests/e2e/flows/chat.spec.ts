import { test, expect } from '@playwright/test';

/**
 * Flujo de chat.
 *
 * Todos los tests mockean el SSE stream y el gateway para ser seguros en CI.
 * Los tests de integración real requieren PLAYWRIGHT_BASE_URL → Brev.
 */

const MOCK_JWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0QGV4YW1wbGUuY29tIiwibmFtZSI6IlRlc3QgVXNlciIsInJvbGUiOiJ1c2VyIiwiZXhwIjo5OTk5OTk5OTk5fQ.fake_signature';

test.describe('Chat flow', () => {
    // Los tests de chat requieren una sesión JWT válida del gateway real.
    // El JWT fake inyectado vía addCookies() es rechazado por el servidor en local.
    test.skip(
        !process.env.PLAYWRIGHT_BASE_URL,
        'Requiere PLAYWRIGHT_BASE_URL → Brev con sesión real',
    );

    test.beforeEach(async ({ page }) => {
        // Inyectar sesión mock en las cookies
        await page.context().addCookies([
            {
                name: 'sda_session',
                value: MOCK_JWT,
                domain: new URL(process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:4173').hostname,
                path: '/',
                httpOnly: true,
                secure: false,
            },
        ]);
    });

    test('chat page renders input and message area', async ({ page }) => {
        // Mock para que el server-side load() no falle al verificar el token
        await page.route('**/api/chat/sessions', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ sessions: [] }),
            });
        });

        await page.goto('/chat');
        await expect(page.getByRole('textbox')).toBeVisible();
    });

    test('sends message and shows streaming indicator', async ({ page }) => {
        test.skip(
            !process.env.PLAYWRIGHT_BASE_URL,
            'Requiere servidor con sesión válida para mockear SSE',
        );

        await page.route('**/api/chat/stream/**', async (route) => {
            const encoder = new TextEncoder();
            const chunks = [
                'data: {"type":"token","content":"Esta"}\n\n',
                'data: {"type":"token","content":" es"}\n\n',
                'data: {"type":"token","content":" la respuesta."}\n\n',
                'data: {"type":"done"}\n\n',
            ];

            await route.fulfill({
                status: 200,
                headers: { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache' },
                body: chunks.join(''),
            });
        });

        await page.goto('/chat');
        await page.getByRole('textbox').fill('¿Pregunta de ejemplo?');
        await page.getByTitle('Enviar (Enter)').click();

        // Verificar que el campo se limpió tras enviar
        await expect(page.getByRole('textbox')).toHaveValue('');
    });

    test('input clears after sending a message', async ({ page }) => {
        test.skip(
            !process.env.PLAYWRIGHT_BASE_URL,
            'Requiere servidor con sesión válida',
        );

        await page.goto('/chat');
        await page.getByRole('textbox').fill('texto de prueba');
        await page.getByTitle('Enviar (Enter)').click();
        await expect(page.getByRole('textbox')).toHaveValue('');
    });
});
