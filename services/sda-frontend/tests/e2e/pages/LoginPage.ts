import type { Page } from '@playwright/test';

export class LoginPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/login');
    }

    async login(email: string, password: string) {
        await this.page.getByLabel('Email').fill(email);
        await this.page.getByLabel('Contraseña').fill(password);
        await this.page.getByRole('button', { name: 'Ingresar' }).click();
    }

    async getErrorMessage() {
        // Error se muestra en un div con clase de peligro, no role="alert"
        return this.page.locator('form + * [class*="danger"], form div[class*="danger"]').textContent();
    }
}
