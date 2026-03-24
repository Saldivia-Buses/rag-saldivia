import type { Page } from '@playwright/test';

export class AdminPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/admin/users');
    }

    async getUsers(): Promise<string[]> {
        const rows = this.page.locator('tbody tr td:first-child');
        return rows.evaluateAll(els => els.map(el => el.textContent?.trim() ?? ''));
    }

    async createUser(email: string, role: string) {
        await this.page.getByRole('button', { name: /nuevo usuario/i }).click();
        await this.page.getByLabel(/email/i).fill(email);
        await this.page.selectOption('select[name="role"]', role);
        await this.page.getByRole('button', { name: /crear/i }).click();
    }

    async deleteUser(email: string) {
        // Busca la fila del usuario y clickea su botón de eliminación
        const row = this.page.locator('tr', { hasText: email });
        await row.getByRole('button', { name: /eliminar|borrar|delete/i }).click();
    }
}
