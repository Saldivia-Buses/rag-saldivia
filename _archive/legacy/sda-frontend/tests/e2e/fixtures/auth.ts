import { test as base } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage.js';

/**
 * Fixture de autenticación para E2E.
 *
 * Uso:
 *   import { test, expect } from '../fixtures/auth.js';
 *   test('mi test', async ({ loginPage }) => { ... });
 *
 * Requiere backend real (Brev) o mocks de page.route() en el test individual.
 * Variables de entorno:
 *   TEST_USER_EMAIL    — email del usuario de prueba
 *   TEST_USER_PASSWORD — contraseña del usuario de prueba
 */
export const test = base.extend<{ loginPage: LoginPage }>({
    loginPage: async ({ page }, use) => {
        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login(
            process.env.TEST_USER_EMAIL ?? 'test@example.com',
            process.env.TEST_USER_PASSWORD ?? 'test-password-example',
        );
        // Espera redirect a /chat tras login exitoso
        await page.waitForURL('**/chat');
        await use(loginPage);
    },
});

export { expect } from '@playwright/test';
