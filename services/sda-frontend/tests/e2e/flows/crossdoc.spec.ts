import { test, expect } from '@playwright/test';

/**
 * Flujo Crossdoc (búsqueda cross-document en 4 fases).
 * Mockea el pipeline completo para ser seguro en CI.
 */
test.describe('Crossdoc flow', () => {
    test.skip(
        !process.env.PLAYWRIGHT_BASE_URL,
        'Requiere PLAYWRIGHT_BASE_URL con backend + crossdoc habilitado',
    );

    test.beforeEach(async ({ page }) => {
        // Mock decompose endpoint
        await page.route('**/api/crossdoc/decompose', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    subQueries: ['¿Cuándo fue fundada?', '¿Cuántos empleados tiene?'],
                }),
            });
        });

        // Mock subquery endpoint
        await page.route('**/api/crossdoc/subquery', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    content: 'Respuesta de ejemplo para la sub-consulta.',
                    success: true,
                }),
            });
        });

        // Mock synthesize SSE
        await page.route('**/api/crossdoc/synthesize', async (route) => {
            await route.fulfill({
                status: 200,
                headers: { 'Content-Type': 'text/event-stream' },
                body: [
                    'data: {"type":"token","content":"Síntesis"}\n\n',
                    'data: {"type":"token","content":" final."}\n\n',
                    'data: {"type":"done"}\n\n',
                ].join(''),
            });
        });
    });

    test('crossdoc toggle is visible in chat page', async ({ page }) => {
        await page.goto('/chat');
        // El botón de crossdoc debe ser visible en el chat input
        const crossdocBtn = page.locator('[title*="crossdoc"], [title*="Cross"], [aria-label*="crossdoc"]');
        await expect(crossdocBtn).toBeVisible();
    });

    test('crossdoc progress shows phase pills during pipeline', async ({ page }) => {
        test.skip(true, 'Requiere integración real con chat page y estado de crossdoc activo');
    });
});
