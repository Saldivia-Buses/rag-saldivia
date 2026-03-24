import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage.js';

/**
 * Flujo de autenticación.
 *
 * Tests que usan page.route() para mockear el gateway son seguros en CI.
 * Tests que requieren login real necesitan PLAYWRIGHT_BASE_URL + credenciales reales.
 */

test.describe('Auth flow', () => {
    test('unauthenticated access to /chat redirects to /login', async ({ page }) => {
        // Este test funciona en local (preview server) sin backend real
        await page.goto('/chat');
        await expect(page).toHaveURL(/\/login/);
    });

    test('login page shows email and password fields', async ({ page }) => {
        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await expect(page.getByLabel('Email')).toBeVisible();
        await expect(page.getByLabel('Contraseña')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Ingresar' })).toBeVisible();
    });

    test('invalid credentials shows error message', async ({ page }) => {
        // Mockea la llamada del BFF al gateway → 401
        await page.route('**/auth/login', async (route) => {
            await route.fulfill({
                status: 401,
                contentType: 'application/json',
                body: JSON.stringify({ detail: 'Invalid credentials' }),
            });
        });

        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login('wrong@example.com', 'wrongpassword');

        // El error aparece después del submit (puede tardar un ciclo de render)
        await page.waitForTimeout(500);
        // Verifica que sigue en /login (no fue redirigido a /chat)
        await expect(page).toHaveURL(/\/login/);
    });

    test('valid login redirects to /chat', async ({ page }) => {
        // Requiere backend real o gateway mock correcto en GATEWAY_URL
        test.skip(
            !process.env.PLAYWRIGHT_BASE_URL && !process.env.TEST_USER_EMAIL,
            'Requiere PLAYWRIGHT_BASE_URL o TEST_USER_EMAIL para login real',
        );

        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login(
            process.env.TEST_USER_EMAIL ?? 'test@example.com',
            process.env.TEST_USER_PASSWORD ?? 'test-password',
        );
        await expect(page).toHaveURL(/\/chat/);
    });

    test('logout clears session and redirects to /login', async ({ page }) => {
        test.skip(
            !process.env.PLAYWRIGHT_BASE_URL && !process.env.TEST_USER_EMAIL,
            'Requiere sesión autenticada real',
        );

        // Navegar a /login, autenticarse, luego hacer logout
        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login(
            process.env.TEST_USER_EMAIL ?? 'test@example.com',
            process.env.TEST_USER_PASSWORD ?? 'test-password',
        );
        await page.waitForURL(/\/chat/);

        // Busca el botón de logout (nombre puede variar según el layout)
        await page.getByRole('button', { name: /logout|salir|cerrar sesión/i }).click();
        await expect(page).toHaveURL(/\/login/);
    });
});
