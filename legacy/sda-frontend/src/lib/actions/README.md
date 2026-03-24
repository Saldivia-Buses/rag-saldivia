# Actions

Svelte actions for DOM-level interactions. Actions are reusable functions that attach behavior to DOM elements via the `use:` directive.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `clickOutside.ts` | Fires a callback when the user clicks outside the element. Used for closing popovers, dropdowns, and modals. | svelte/action |

## Usage

The `clickOutside` action uses a **callback API** (NOT Svelte 4 custom events).

**Correct usage:**
```svelte
<script>
  import { clickOutside } from '$lib/actions/clickOutside';

  let open = $state(false);
  function handleClose() {
    open = false;
  }
</script>

<div use:clickOutside={handleClose}>
  <!-- Popover content -->
</div>
```

**Incorrect usage (Svelte 4 syntax):**
```svelte
<!-- ❌ WRONG — this will not work -->
<div use:clickOutside on:clickoutside={handleClose}>
  <!-- ... -->
</div>
```

The action signature is `Action<HTMLElement, () => void>`, meaning it accepts a callback function as its parameter.

## Design notes

### Capture phase event listener

The `clickOutside` action uses `addEventListener('click', handler, true)` with `capture: true`. This ensures the handler runs **before** other click handlers, including those that call `stopPropagation()`.

Without capture phase, clicks on elements that stop propagation (e.g., a button inside a modal) would prevent the outside click detection from working correctly.
