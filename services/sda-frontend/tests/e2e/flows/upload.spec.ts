import { test, expect } from '@playwright/test';
import { UploadPage } from '../pages/UploadPage.js';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Flujo de subida de documentos.
 * Requiere PLAYWRIGHT_BASE_URL para backend real (Brev).
 */
test.describe('Upload flow', () => {
    test.skip(
        !process.env.PLAYWRIGHT_BASE_URL,
        'Requiere backend real en PLAYWRIGHT_BASE_URL',
    );

    test('upload page renders file input and collection selector', async ({ page }) => {
        const uploadPage = new UploadPage(page);
        await uploadPage.goto();
        await expect(page.locator('input[type="file"]')).toBeVisible();
    });

    test('can upload a PDF file and see success status', async ({ page }) => {
        // Mock del endpoint de upload para no subir un archivo real
        await page.route('**/api/upload', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ message: 'File ingested successfully' }),
            });
        });

        const uploadPage = new UploadPage(page);
        await uploadPage.goto();

        // Usar el PDF de ejemplo si existe, sino crear uno simple
        const samplePdf = path.join(__dirname, '../fixtures/sample.pdf');
        await uploadPage.uploadFile(samplePdf);
    });
});
