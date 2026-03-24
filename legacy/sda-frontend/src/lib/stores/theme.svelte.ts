// services/sda-frontend/src/lib/stores/theme.svelte.ts
// Wrapper de mode-watcher. El modo dark/light se gestiona por mode-watcher
// con defaultMode="dark" en +layout.svelte (raíz).
// Este módulo re-exporta las utilidades para uso consistente en la app.
export { toggleMode, resetMode } from 'mode-watcher';
