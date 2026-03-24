import type { Page } from '@playwright/test';

export class CollectionsPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/collections');
    }

    async createCollection(name: string) {
        await this.page.getByRole('button', { name: 'Nueva colección' }).click();
        // El CreateModal tiene un input para el nombre
        await this.page.getByLabel(/nombre/i).fill(name);
        await this.page.getByRole('button', { name: /crear/i }).click();
    }

    async deleteCollection(name: string) {
        // Navegar a la colección y buscar botón de borrar
        await this.page.goto(`/collections/${name}`);
        await this.page.getByRole('button', { name: /eliminar|borrar|delete/i }).click();
        // Confirmar si hay dialog de confirmación
        const confirmBtn = this.page.getByRole('button', { name: /confirmar|sí/i });
        if (await confirmBtn.isVisible()) {
            await confirmBtn.click();
        }
    }

    async getCollectionNames(): Promise<string[]> {
        const links = this.page.locator('a[href^="/collections/"]');
        return links.evaluateAll(els => els.map(el => el.textContent?.trim() ?? ''));
    }

    async openCollection(name: string) {
        await this.page.getByRole('link', { name }).click();
    }
}
