import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/svelte';
import { afterEach } from 'vitest';

// @testing-library/svelte no llama cleanup automáticamente con sveltekit() plugin
afterEach(() => cleanup());

// jsdom no implementa scrollTo en HTMLElement → polyfill para que los efectos
// de scroll del componente no rompan los tests de renderizado.
// La guarda evita ReferenceError en entornos 'node' donde HTMLElement no existe.
if (typeof HTMLElement !== 'undefined') {
    HTMLElement.prototype.scrollTo = () => {};
}
