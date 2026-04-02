import type { Page } from '@playwright/test';
import path from 'path';

export class UploadPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/upload');
    }

    async uploadFile(filePath: string) {
        const input = this.page.locator('input[type="file"]');
        await input.setInputFiles(path.resolve(filePath));
    }

    async waitForUploadComplete() {
        // Espera a que aparezca un indicador de éxito o error
        await this.page.waitForSelector('[data-testid="upload-status"], .upload-success, .upload-error', {
            timeout: 30000,
        });
    }

    async getUploadStatus(): Promise<string> {
        const status = this.page.locator('[data-testid="upload-status"]');
        return (await status.textContent()) ?? '';
    }
}
