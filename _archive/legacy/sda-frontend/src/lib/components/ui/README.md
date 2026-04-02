# UI Components

Generic, reusable UI primitives that form the design system foundation for the SDA Frontend.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `Badge.svelte` | Colored badge/pill for status indicators. Supports variants: blue, green, red, yellow, gray, orange. | None |
| `Button.svelte` | Primary button component with variants (primary, secondary, danger, ghost), sizes (sm, md, lg), loading state, and disabled state. | None |
| `Card.svelte` | Generic card container with optional header, footer, and body slots. | None |
| `Input.svelte` | Styled text input with error state support and label. | None |
| `Modal.svelte` | Modal dialog with overlay, close button, configurable size (sm, md, lg), and optional title/footer. Uses `$bindable` for open state. | None |
| `Skeleton.svelte` | Loading skeleton placeholder with configurable width, height, and border radius. Animated gradient effect. | None |
| `Toast.svelte` | Single toast notification component with variants (success, error, warning, info) and auto-dismiss. | `$lib/stores/toast.svelte` |
| `ToastContainer.svelte` | Fixed-position container that renders all active toasts from the toast store. | `$lib/stores/toast.svelte`, `Toast.svelte` |

## Design notes

All components use Svelte 5 runes (`$state`, `$props`, `$derived`, `$bindable`) and are typed with TypeScript interfaces.

The color system is based on CSS custom properties defined in `app.css` (e.g., `var(--accent)`, `var(--bg-surface)`, `var(--text)`), making theme switching seamless.
