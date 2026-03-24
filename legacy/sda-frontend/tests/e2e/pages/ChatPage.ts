import type { Page } from '@playwright/test';

export class ChatPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/chat');
    }

    async sendMessage(text: string) {
        await this.page.getByRole('textbox').fill(text);
        // El botón de envío tiene title="Enviar (Enter)"
        await this.page.getByTitle('Enviar (Enter)').click();
    }

    async waitForResponse() {
        // Espera a que aparezca un mensaje de asistente (el streaming termina)
        await this.page.waitForSelector('[data-testid="assistant-message"]');
    }

    async enableCrossdoc() {
        // El toggle de crossdoc está en CrossdocSettingsPopover
        await this.page.getByRole('button', { name: /crossdoc/i }).click();
    }

    async getLastMessage() {
        const messages = this.page.locator('[data-testid="message"]');
        return messages.last().textContent();
    }

    async getInputValue() {
        return this.page.getByRole('textbox').inputValue();
    }

    async isStreamingVisible() {
        return this.page.locator('.animate-pulse').isVisible();
    }
}
