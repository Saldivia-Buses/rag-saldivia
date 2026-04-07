---
name: PR 2.0.3 Review
description: Review of 7-commit 2.0.3 branch -- health checks, MCP tools, OTel metrics, Prometheus alerts, Grafana dashboards, astro audit, extractor tests
type: project
---

Branch 2.0.3 review (2026-04-07): CAMBIOS REQUERIDOS

3 blockers:
- B1: MCP rag_query JSON injection via fmt.Sprintf (tools/mcp/main.go:365)
- B2: MCP/CLI db_query uses keyword filter not read-only transaction -- bypassable SQL injection
- B3: MCP rag_query no response body size limit (OOM)

8 must-fix:
- D1: OTel Insecure:true hardcoded in all 11 services
- D2: ServiceDown alert fires on idle (false positive) AND misses truly down services
- D3: humanize1024 applied to already-divided value in PrometheusStorageHigh
- D4: Astro audit entries missing IP/UserAgent fields
- D5: pkg/health missing tests
- D6: DBQuery false positive on table names containing blocked words (SELECTED->DELETE)
- D7: Grafana hardcoded admin password
- D8: MCP deploy tool has no confirmation gate

**Why:** Security tools (MCP db_query, deploy) are high-privilege and must be hardened. The JSON injection is classic string interpolation bug.
**How to apply:** Block merge until B1-B3 fixed. D1-D8 can be batched.
