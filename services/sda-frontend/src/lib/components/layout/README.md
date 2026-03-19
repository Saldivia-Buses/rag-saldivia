# Layout Components

Layout-level structural components that provide the app shell and page structure.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `Sidebar.svelte` | App shell sidebar that wraps the entire authenticated layout. Renders navigation items, user info, collapse/expand toggle, theme toggle, and logout button. Uses role-based rendering to show/hide admin routes. | `$app/stores` (page store), lucide-svelte, mode-watcher |

## Design notes

**Naming distinction:** This `layout/Sidebar.svelte` is the app shell sidebar (structural component used in `(app)/+layout.svelte`), while `sidebar/Sidebar.svelte` is the navigation sidebar component (icon-based menu).

The layout sidebar includes:
- Collapsible state (`collapsed` rune)
- Theme toggle (light/dark mode via mode-watcher)
- User profile section
- Logout handler (calls `/api/auth/session` DELETE endpoint)
- Role-based navigation (admin, area manager, user)
