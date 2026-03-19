/**
 * E2E tests run against the built app in preview mode.
 *
 * Before running tests:
 *   npm run build && npm run preview
 *   # App starts on http://localhost:4173
 *
 * Or set PLAYWRIGHT_BASE_URL to point to Brev:
 *   PLAYWRIGHT_BASE_URL=https://brev-instance-url npx playwright test
 *
 * Flows tagged @slow only run when PLAYWRIGHT_BASE_URL is set (Brev only).
 */
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests/e2e/flows',
    fullyParallel: false,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,
    workers: 1,
    reporter: 'html',
    use: {
        baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:4173',
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
    },
    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
});
