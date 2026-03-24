# Sidebar Components

Navigation sidebar components. These build the icon-based navigation menu visible on the left side of the app.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `Sidebar.svelte` | Container component that renders a vertical navigation sidebar with role-based item filtering (admin/manager-only items are hidden for regular users) and logout button. | `SidebarItem.svelte`, lucide-svelte |
| `SidebarItem.svelte` | Single navigation item component that renders an icon link with active state highlighting (border-left accent) and hover tooltip. | `$app/stores` (page store) |

## Design notes

The `Sidebar.svelte` component conditionally renders admin/manager items based on the `role` prop. It uses the lucide-svelte icon library for consistent iconography.

The `SidebarItem.svelte` component derives its active state from SvelteKit's `$page.url.pathname`, highlighting the current route.

**Naming note:** This `sidebar/Sidebar.svelte` is the navigation sidebar component, while `layout/Sidebar.svelte` is the app shell sidebar wrapper. They serve different purposes.
