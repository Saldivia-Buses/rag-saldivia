---
name: Plan 34 SSO security audit
description: Security audit of SSO implementation (SAML/OIDC via arctic + node-saml) found 2 critical, 3 high issues. NOT APTO.
type: project
---

Plan 34 SSO security audit completed 2026-04-01. Verdict: NOT APTO.

2 CRITICAL: SAML state token validation bypass (optional `if (stateToken)` check), SAML cert/entryPoint not persisted to DB (createSsoProvider omits samlCert/samlEntryPoint fields).

3 HIGH: State comparison not timing-safe, account takeover via SSO auto-linking to existing email without verification, admin can set defaultRole=admin for auto-provisioned SSO users.

**Why:** SSO is a high-value attack surface. SAML flow specifically has compounding issues (state bypass + empty cert = potential assertion forgery).

**How to apply:** When reviewing SSO fixes, verify all 5 critical/high issues are addressed. The OIDC flow is solid -- focus remediation on SAML path and account-linking logic.
