# Phase 4 follow-up — E2E specs need selector rewrite per cluster

**Detected:** 2026-04-20 during cycle 2.0.21 Phase 4 (E2E reform).

## Status

The structural refactor of `e2e/erp-mutations.spec.ts` is committed:
- `MUTATIONS_ENABLED` skip flag removed (all tests run every run).
- Credentials moved to env-driven defaults (`TEST_EMAIL` /
  `TEST_PASSWORD` in `helpers/auth.ts`, default `e2e@sda.local`).
- 3 create tests gain `page.waitForRequest` network assertions
  + readback of the unique row from the list.
- URLs updated for the Phase 2 IA changes
  (`/administracion/tesoreria` → `/tesoreria`,
  `/administracion/calidad` → `/calidad/no-conformidades`).
- Login flow waits for `/inicio`, not `/dashboard`.

When run against the local `sda_tenant_test` mirror, **all 7 tests
fail** for two structural reasons:

1. **Form labels are not associated with their inputs.** The forms
   render `<Label>Tipo</Label>` next to `<Input value=… />` with no
   `htmlFor` / `id` pair. Playwright's `getByLabel('Tipo')` finds
   nothing; screen readers do the same. Examples:
   `CreateCatalogForm`, the tesorería movement dialog, the NC dialog.
   This is also the Phase 8 (a11y) audit's job — fixing it once per
   form fixes the test AND the screen-reader experience.
2. **Some buttons have duplicates** at the same accessible name
   (e.g. `/administracion/almacen` has two `Nuevo artículo` buttons —
   one in the page header and one inside the "Artículos" tab). Spec
   needs `.first()` or scope the locator to the dialog/section.

## What the next E2E session must do

For each cluster the spec covers (catálogos, almacén, tesorería,
contabilidad, calidad, manufactura, seguridad), rewrite the test:

1. Read the actual form component used by that page.
2. Add `htmlFor` / `id` to each `<Label>` / `<Input>` pair (this is
   a Phase 8 a11y fix that the test forces). Commit as a separate
   "a11y(<area>): asociar labels e inputs" change.
3. Update the spec's `getByLabel(...)` calls to match the now-real
   accessible names; add `.first()` or scoped locators for any
   ambiguity.
4. Convert the 4 cancel-only tests (contabilidad asiento, calidad
   NC, manufactura unidad, seguridad accidente) into create tests
   with side-effect readback.
5. Add the per-cluster specs (1 spec per cluster shipped 1–16) the
   plan calls for.

This work is mechanical but volumous (~15–30 min per cluster). It
belongs to a dedicated cycle 2.0.22 E2E session, not a frontend-IA
session like 2.0.21.

## Decision recorded for 2.0.21

The structural refactor stays merged. The selector rewrites do not
ship in 2.0.21. The cycle's "no cluster counts as shipped without
green E2E" rule (memory `feedback_quality_over_quantity`) implies
that NO cluster from 1.x is officially "shipped" until 2.0.22 closes
the E2E gap — accepted, with the understanding that 2.0.21 fixes
the foundations (DB local, IA, dead code) the E2E reform depends on.
