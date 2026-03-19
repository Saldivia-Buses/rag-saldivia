import { test, expect } from '@playwright/test';
import { CollectionsPage } from '../pages/CollectionsPage.js';

/**
 * Flujo de gestión de colecciones.
 * Requiere PLAYWRIGHT_BASE_URL para backend real (Brev).
 */
test.describe('Collections flow', () => {
    test.skip(
        !process.env.PLAYWRIGHT_BASE_URL,
        'Requiere backend real en PLAYWRIGHT_BASE_URL',
    );

    test('collections page shows list of collections', async ({ page }) => {
        const collectionsPage = new CollectionsPage(page);
        await collectionsPage.goto();
        await expect(page.getByRole('heading', { name: /colecciones/i })).toBeVisible();
    });

    test('can create and delete a collection', async ({ page }) => {
        const collectionsPage = new CollectionsPage(page);
        const testColName = `e2e-test-${Date.now()}`;

        await collectionsPage.goto();
        await collectionsPage.createCollection(testColName);

        // Verificar que la colección aparece en la lista
        await expect(page.getByText(testColName)).toBeVisible();

        // Borrar la colección de prueba
        await collectionsPage.deleteCollection(testColName);
        await collectionsPage.goto();
        await expect(page.getByText(testColName)).not.toBeVisible();
    });

    test('clicking a collection card navigates to detail page', async ({ page }) => {
        const collectionsPage = new CollectionsPage(page);
        await collectionsPage.goto();

        const names = await collectionsPage.getCollectionNames();
        if (names.length === 0) {
            test.skip(true, 'No hay colecciones para navegar');
        }

        await collectionsPage.openCollection(names[0]);
        await expect(page).toHaveURL(new RegExp(`/collections/${names[0]}`));
    });
});
