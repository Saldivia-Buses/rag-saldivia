import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
    plugins: [sveltekit()],
    resolve: {
        // Necesario para que Svelte resuelva el bundle de cliente (no el de servidor)
        // cuando @testing-library/svelte llama a mount() en jsdom
        conditions: ['browser'],
    },
    test: {
        environment: 'node',
        setupFiles: ['./src/test-setup.ts'],
        // Excluir E2E de Playwright — corren con `npx playwright test`, no con vitest
        exclude: ['**/node_modules/**', '**/tests/e2e/**'],
        environmentMatchGlobs: [
            ['src/**/*.svelte.test.ts', 'jsdom'],
            ['src/**/components/**/*.test.ts', 'jsdom'],
            ['src/**/components/**/*.component.test.ts', 'jsdom'],
            ['src/**/_components/**/*.test.ts', 'jsdom'],
            ['src/routes/**/*.component.test.ts', 'jsdom'],
        ],
        coverage: {
            provider: 'v8',
            include: ['src/lib/**/*.ts', 'src/lib/**/*.svelte.ts', 'src/routes/api/**/*.ts'],
            exclude: ['src/**/*.test.ts', 'src/**/*.spec.ts'],
            thresholds: {
                lines: 80,
                functions: 80,
                branches: 80,
                statements: 80,
            },
            reporter: ['text', 'html', 'lcov'],
            reportOnFailure: true, // reportar coverage incluso cuando hay tests fallidos (bugs documentados)
        },
    },
});
