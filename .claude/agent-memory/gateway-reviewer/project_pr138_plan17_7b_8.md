---
name: PR #138 Plan 17 Phases 7b+8
description: ERP current accounts + production + C2 tax entry fix, blockers: trigger crashes Allocate (no status col), UpdateMovementBalance silent on insufficient balance
type: project
---

PR #138 reviewed 2026-04-09. Plan 17 phases 7b (current accounts) + 8 (production) + C2 fix (PostInvoice tax entries).

**Why:** ERP migration from Histrix legacy. Current accounts = bridge between invoicing and treasury. Production = bus manufacturing tracking.

**How to apply:** Two critical blockers must be fixed before merge:
1. `erp_prevent_financial_mutation()` trigger on `erp_account_movements` crashes on `OLD.status` (table has no status column) -- Allocate is completely broken
2. `UpdateMovementBalance` uses `:exec` not `:execrows` -- insufficient balance silently succeeds, creating financial integrity violation

Additional high findings: RequireModule("erp") not applied, GetOrder swallows errors, allocation uses Write not WriteStrict, no status validation on step/order transitions.
