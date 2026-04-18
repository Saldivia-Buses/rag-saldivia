# Changelog

Release notes per version live as **GitHub Releases**:
<https://github.com/Saldivia-Buses/rag-saldivia/releases>

Each working-branch cycle (`2.0.N`) that merges to `main` gets an
annotated tag `v2.0.N` + a GitHub release with the PR body as notes.
See `.claude/projects/-home-enzo-rag-saldivia/memory/feedback_version_tagging.md`
for the cycle discipline.

## Index

| Version | Date | Highlights |
|---|---|---|
| [v2.0.7](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.7) | 2026-04-18 | Phase 0 closed (5/5) — agent tool RBAC capability gate + workstation drift closed. Phase 1 data-migration roadmap + W-004/W-005/W-006 waivers (31 HTX* infra, 5 `*_OLD`, 225 zero-row). Prod hardening (app in `docker-compose.prod`, distroless healthcheck, redis auth). |
| [v2.0.6](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.6) | 2026-04-18 | Phase 0 items 1–3 — migration integrity (`rows_read = rows_written + rows_skipped + rows_duplicate`), no-op migrators fixed (1,142,427 rows recovered on `saldivia_bench`), orphan tables resolved via read queries + W-002/W-003 waivers. |
| [v2.0.5](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.5) | 2026-04-16 | Plan 26 — spine. NATS as the typed event transport. |
| [v2.0.4](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.4) | 2026-04-09 | Plan 15 — BigBrother network intelligence service. |
| [v2.0.3](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.3) | 2026-04-08 | Astro Super Agent. |
| [v2.0.2](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.2) | 2026-04-07 | Interim polish. |
| [v2.0.1](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.1) | 2026-04-06 | Backend hardening + gRPC + polish. |
| [v2.0.0](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v2.0.0) | 2026-04-05 | 2.x initial cut. |
| [v1.0.0](https://github.com/Saldivia-Buses/rag-saldivia/releases/tag/v1.0.0) | 2026-03-27 | Initial release. |

## ADR 027 scoreboard

Every release lists its ticks against ADR 027
(`docs/decisions/027-mvp-success-criteria.md`) in the release notes.
Aggregated state as of `v2.0.7`:

- **Phase 0 (transversal)**: **5/5** — done.
- **Phase 1 (Histrix parity + shutdown)**: §Data migration is gated
  open (roadmap landed in `v2.0.7`, first migrator is the Phase 1
  row-1 tick — scheduled for `v2.0.8`). §UI parity, §Operational
  parity, §Cutover readiness all sequenced after.
- **Phase 2 (SDA layer)**, **Phase 3 (background agents)**,
  **Phase 4 (differential UX)**: unblocked by Phase 0 item 4 in
  `v2.0.7`; queued after Phase 1.
