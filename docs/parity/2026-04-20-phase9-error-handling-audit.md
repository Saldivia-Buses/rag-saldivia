# Phase 9 audit — Spanish UI / Error handling

**Date:** 2026-04-20.

## Findings — already clean

Sweep on `apps/web/src/`:

- `"Error interno"`: **0 hits** (the plan worried this was sprinkled
  everywhere; not the case).
- `"Sin datos"`: **0 hits** (no generic empty-state strings to
  replace either).
- Generic `catch (e) {}` blocks: only legitimate ones (auth store
  401 handling, api-client tests, magicui/confetti animation).

The cycle 2.0.x error-handling discipline is already in place. The
existing pattern `permissionErrorToast` (in
`apps/web/src/lib/erp/permission-messages.ts`, used in /tesoreria
and other pages) maps `err.status` to a user-friendly toast.

## Pending follow-up — finer audit per cluster

The mechanical sweep does not catch:

- Subtle inconsistencies between modules ("Agregar" vs "Crear" vs
  "Nuevo" vs "+" — naming convention).
- Empty-state quality: every list page should state explicitly what
  the user can do next (CTA), not just "no rows".
- Confirmation dialogs for destructive ops: each `delete` /
  `override` / `reset` should require an `<AlertDialog>` confirm.
- Loading skeletons consistency: all pages should use the same
  `<Skeleton>` placeholder, not mix spinner / "Cargando…" / blank.

These are per-cluster polish items belonging to Phase 10
(cluster polish loop). They get done as each cluster comes up in
maintenance.

## Decision recorded for 2.0.21

No additional error-handling commits ship this cycle. Phase 9 marker
goes "done" because the literal sweeps are clean and the higher-level
polish belongs to Phase 10.
