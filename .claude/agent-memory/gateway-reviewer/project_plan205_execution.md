---
name: 2.0.5 Execution Plan Review
description: Pre-implementation review of 2.0.5 execution plan (Plans 18+21+22 partial) -- 4 blockers found, cross-service wiring is the main structural gap
type: project
---

Reviewed 2.0.5 execution plan on 2026-04-12. Plans 18 (ERP Workflows), 21 (Data Migration), 22 partial (Frontend).

Key findings:
- 4 BLOCKERS: sqlc `CreateFiscalYear` name collision, missing `GetFiscalYear` query, cross-service struct wiring (Invoicing/Treasury/Purchasing need Accounting/Stock/CurrentAccounts references), ~20 sqlc queries in pseudocode not defined
- 6 HIGH: trigger side-effects (032 allows journal `cancelled`), ListFiscalYears missing new columns, no DOWN migrations, permission grants missing for admin role, treasury query column gaps, result_account_id nullable without DB guard
- Migration order 030-035 is correct for FK dependencies
- Shared packages (pkg/crypto, pkg/audit, pkg/storage, pkg/middleware/rbac) all exist and match plan assumptions
- NATS pattern matches existing `publisher.Broadcast(slug, "erp_{domain}", data)`

**Why:** First major execution plan review for ERP. Cross-service dependency injection is a recurring pattern that needs solving -- will affect all future cross-domain workflows.
**How to apply:** When reviewing the implementation PR, verify the 4 blockers are addressed. The cross-service wiring pattern chosen (interfaces vs concrete types) will set the precedent for all future ERP workflows.
