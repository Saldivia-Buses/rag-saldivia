# Astro Service

## What it does

Professional astrological intelligence engine. Computes 55+ astrological
techniques deterministically via Swiss Ephemeris (CGO), routes queries through
a 20-domain intelligence layer, and narrates results through an LLM via SSE
streaming. Includes business intelligence module, quality audit system,
session management, prediction tracking, and SVG chart rendering.

Version: 0.2.0 | Port: 8011

## Architecture

```
ephemeris (swephgo CGO wrapper — Swiss Ephemeris C library)
  → astromath (21 files: angles, aspects, dignities, lots, sabian, disposition, etc.)
    → natal (chart builder + SVG wheel + aspect grid)
      → technique (35 files: 55+ predictive techniques)
        → context (orchestrator: 55 techniques phased, 25+ brief sections, scoring)
          → intelligence (18 files: domain routing, gating, cross-refs, interpretations)
            → quality (4 files: audit, validator, certainty, benchmark)
              → business (11 files: timing, cashflow, risk, forecast, hiring, etc.)
                → handler (6 files: 64 HTTP endpoints, SSE streaming)
                  → cache (LRU charts + Redis context)
```

Key design decisions:
- All Swiss Ephemeris access through `internal/ephemeris` — never swephgo directly
- `ephemeris.CalcMu` mutex protects compound SetTopo + CalcPlanet sequences
- Lock-per-chart (not per-scan) for expensive techniques (rectification, electional)
- Planet names in Spanish (`Sol`, `Luna`, `Mercurio`, etc.)
- House system: Polich-Page Topocentric (`'T'`)
- Topocentric planetary positions (observer-location-corrected)
- Intelligence layer is stateless (Engine created once at startup)
- Quality audit is deterministic (no LLM, <10ms)
- Single-tenant deployment (Saldivia Buses), tenant_id for defense-in-depth

## Packages (111 Go files)

| Package | Files | Purpose |
|---------|-------|---------|
| `technique/` | 35 | 55+ astrological techniques (transits, directions, synastry, electional, etc.) |
| `astromath/` | 21 | Math utilities, dignity tables, 360 Sabian symbols, 20 hellenistic lots, geocoding |
| `intelligence/` | 18 | Domain registry, intent parser, technique gate, cross-references, interpretations, memory |
| `business/` | 11 | Negotiation timing, cashflow, risk heatmap, forecast, hiring, succession, vocational |
| `handler/` | 6 | 64 HTTP endpoints + SSE streaming |
| `context/` | 5 | Orchestrates 55 techniques, builds brief (25+ sections), scoring, cross-analyses |
| `quality/` | 4 | Deterministic audit, response validation, certainty scoring, benchmark |
| `natal/` | 3 | Chart builder + SVG wheel (800×800) + aspect grids |
| `cache/` | 2 | In-memory LRU chart registry (500 entries) + Redis context cache |
| `repository/` | 5 | sqlc generated — contacts, sessions, messages, predictions, feedback, usage |
| `ephemeris/` | 2 | Swiss Ephemeris CGO wrapper |

## Techniques (55+)

### Core Predictive (Plan 11 — original 11)
| Technique | File |
|-----------|------|
| Natal Chart | `natal/chart.go` |
| Slow Transits (5-day sampling) | `technique/transits.go` |
| Solar Arc Directions + Antiscia | `technique/solar_arc.go` |
| Primary Directions (Polich-Page Mundane) | `technique/primary_dir.go` |
| Secondary Progressions | `technique/progressions.go` |
| Solar/Lunar Return | `technique/solar_return.go` |
| Profections | `technique/profections.go` |
| Firdaria | `technique/firdaria.go` |
| Eclipses | `technique/eclipses.go` |
| Fixed Stars | `technique/fixed_stars.go` |
| Zodiacal Releasing | `technique/zodiacal_releasing.go` |

### Plan 12 — Expanded (44 new)
| Technique | File |
|-----------|------|
| Tertiary Progressions | `technique/tertiary_prog.go` |
| Decennials (Vettius Valens) | `technique/decennials.go` |
| Fast Transits (inner planets) | `technique/fast_transits.go` |
| Lunations (New/Full Moon) | `technique/lunations.go` |
| Prenatal Eclipse | `technique/prenatal_eclipse.go` |
| Eclipse Triggers | `technique/eclipse_triggers.go` |
| Planetary Cycles (returns) | `technique/planetary_cycles.go` |
| Planetary Returns (exact) | `technique/planetary_returns.go` |
| Timing Windows | `technique/timing_window.go` |
| Weekly Transits | `technique/weekly_transits.go` |
| Activation Chains | `technique/activation_chains.go` |
| Activation Timeline | `technique/activation_timeline.go` |
| Synastry | `technique/synastry.go` |
| Composite Chart | `technique/composite.go` |
| Davison Chart | `technique/davison.go` |
| Predictive Synastry | `technique/predictive_synastry.go` |
| Electional Astrology | `technique/electional.go` |
| Horary | `technique/horary.go` |
| Midpoints (Ebertin) | `technique/midpoints.go` |
| Declinations | `technique/declinations.go` |
| Astrocartography | `technique/astrocartography.go` |
| Rectification | `technique/rectification.go` |
| Relocation | `technique/relocation.go` |
| Lilith/Vertex | `technique/lilith_vertex.go` |
| Time Lords (unified) | `technique/time_lords.go` |
| Aspect Patterns (T-Square, Grand Trine, Yod) | `astromath/natal_analysis.go` |
| Chart Shape (Jones patterns) | `astromath/natal_analysis.go` |
| Hemispheric Distribution | `astromath/natal_analysis.go` |
| Full Dignity Table | `astromath/natal_analysis.go` |
| Planetary Age (Valens) | `astromath/natal_analysis.go` |
| Almuten Figuris | `astromath/almuten.go` |
| 360 Sabian Symbols | `astromath/sabian.go` |
| 20 Hellenistic Lots | `astromath/lots.go` |
| Dispositor Chains | `astromath/disposition.go` |
| Sect Analysis | `astromath/sect.go` |
| House Rulership | `astromath/house_rulership.go` |
| Classical Temperament | `astromath/temperament.go` |
| Melothesia (medical) | `astromath/melothesia.go` |
| Hyleg/Alcochoden | `astromath/hyleg.go` |
| Decumbitura | `astromath/decumbitura.go` |
| Argentine Calendar | `astromath/calendar.go` |
| Geocoding (20 AR cities) | `astromath/geocoding.go` |

### Scoring & Synthesis
| Function | File |
|----------|------|
| Activation Score (0-100) | `context/scoring.go` |
| Monthly Scores | `context/scoring.go` |
| Technique Verdicts | `context/scoring.go` |
| Contradiction Resolution | `context/scoring.go` |
| Synthesis Brief | `context/scoring.go` |
| Dominant Themes | `context/scoring.go` |
| Convergence Score by Point | `context/scoring.go` |

### Cross-Technique Analyses
| Analysis | File |
|----------|------|
| RS × LR Crossings | `context/cross_analyses.go` |
| Prenatal Eclipse Transits | `context/cross_analyses.go` |
| Divisor (Hellenistic) | `context/cross_analyses.go` |
| Triplicity Lords | `context/cross_analyses.go` |
| Chronocrator × Firdaria Cross | `context/cross_analyses.go` |
| Multi-entity Comparison Table | `context/cross_analyses.go` |
| Antiscia Context | `context/transit_contexts.go` |
| Fixed Stars Transit | `context/transit_contexts.go` |
| Cazimi/Combustion Transit | `context/transit_contexts.go` |
| Davison Transits | `context/transit_contexts.go` |
| VOC Moon Periods | `context/transit_contexts.go` |

## Intelligence Layer (18 files)

| Module | File | Purpose |
|--------|------|---------|
| Engine | `intelligence.go` | Orchestrator: Analyze() → single entry point |
| Domain Registry | `domain.go` + `domain_data.go` | 20 domains (12 root + 8 sub) with inheritance |
| Intent Parser | `intent.go` | Spanish keyword → domain routing |
| Technique Gate | `gate.go` | Struct inspection → validated/ghost classification |
| Cross-References | `crossref.go` | 3 algorithms: ruler, point, temporal convergence |
| Intelligence Brief | `brief.go` | Domain-weighted, ghost-free brief |
| System Prompt | `prompt.go` | Domain-aware LLM prompt builder |
| Interpretations | `interpretation.go` + `interpretations_full.go` | 14 technique-specific interpretation functions |
| Narrative Arc | `narrative.go` | Response structure: opening/convergences/closing |
| Response Skeleton | `skeleton.go` | Template per domain for consistent output |
| Compressor | `compressor.go` | Brief compression by removing low-weight sections |
| Chronocrator | `chronocrator.go` | Time lord filter: boost/demote activations |
| Contraindications | `contraindications.go` | Detects misleading readings |
| Contact Resolver | `contact_resolver.go` | Fuzzy name matching |
| Memory | `memory.go` | Inter-session wakeup context + event extraction |
| Adaptive | `adaptive.go` | LLM params by query complexity |

## Business Module (11 files)

| Module | File | Purpose |
|--------|------|---------|
| Dashboard | `business.go` | Orchestrator: full dashboard + team compatibility |
| Timing | `timing.go` | Negotiation windows per day, Mercury Rx penalty |
| Cash Flow | `cashflow.go` | 12-month H2/H8 transit forecast |
| Risk | `risk.go` | 5 categories × 12 months heatmap |
| Forecast | `forecast.go` | Quarterly Jupiter/Saturn outlook |
| Mercury Rx | `mercury.go` | Retrograde calendar with impact by element |
| Agenda | `agenda.go` | Daily action items from timing + risk |
| Employee | `employee.go` | Candidate screening via synastry |
| Succession | `succession.go` | Leadership transition planning |
| Vocational | `vocational.go` | Career analysis from MC/H10 |
| Types | `types.go` | Shared business types |

## Endpoints (64)

### Technique endpoints (POST, `astro.read`)
`/v1/astro/natal`, `/transits`, `/solar-arc`, `/directions`, `/progressions`,
`/returns`, `/profections`, `/firdaria`, `/fixed-stars`, `/brief`,
`/eclipses`, `/zodiacal-releasing`, `/lunations`, `/lots`, `/dignities`,
`/midpoints`, `/declinations`, `/fast-transits`, `/wheel` (SVG),
`/synastry`, `/composite`, `/tertiary-progressions`, `/decennials`,
`/planetary-cycles`, `/planetary-returns`, `/lilith-vertex`, `/time-lords`,
`/electional`, `/horary`, `/astrocartography`, `/rectification`,
`/weekly-transits`, `/activation-timeline`, `/score`, `/voc-moon`,
`/tabla`, `/vocational`, `/employee-screening`

### Session endpoints (`astro.read` / `astro.write`)
`GET/POST /v1/astro/sessions`, `GET/DELETE/PATCH /v1/astro/sessions/{id}`,
`GET /v1/astro/sessions/{id}/messages`

### Chat SSE (`astro.read`, 5/min)
`POST /v1/astro/query` — full pipeline: context → intelligence → LLM → audit → SSE

### Business endpoints (`astro.business`)
`GET /v1/astro/business/dashboard`, `/cashflow`, `/risk`, `/forecast`,
`/team`, `/hiring`, `/mercury-rx`

### Quality/Tracking (`astro.read` / `astro.write`)
`POST/GET /v1/astro/predictions`, `PATCH /v1/astro/predictions/{id}/verify`,
`GET /v1/astro/predictions/stats`, `POST /v1/astro/feedback`,
`GET /v1/astro/usage`

### Contact CRUD (`astro.read` / `astro.write`)
`GET/POST /v1/astro/contacts`, `GET /v1/astro/contacts/search`,
`PUT/DELETE /v1/astro/contacts/{id}`

## Agent Runtime Tools (54)

Manifest: `modules/astro/tools.yaml` — 54 tools callable by the Agent Runtime.

## NATS Events

| Subject | Event | Trigger |
|---------|-------|---------|
| `tenant.{slug}.astro.sessions` | created/deleted | Session CRUD |
| `tenant.{slug}.feedback.astro_quality` | quality metrics | After audit |

## SSE Event Protocol (`/v1/astro/query`)

```
contact_recognized → calc_context → think_start → think_delta → think_done
→ token → natal_wheel → aspect_grid → tabla → usage → audit → done | error
```

## Env vars

| Var | Required | Default | Description |
|-----|----------|---------|-------------|
| `ASTRO_PORT` | No | `8011` | HTTP listen port |
| `EPHE_PATH` | No | `/ephe` | Swiss Ephemeris data files |
| `POSTGRES_TENANT_URL` | No | `""` | Tenant DB (contacts/sessions disabled if empty) |
| `SGLANG_LLM_URL` | No | `""` | LLM endpoint for narration |
| `SGLANG_LLM_MODEL` | No | `""` | Model name |
| `LLM_API_KEY` | No | `""` | API key for LLM |
| `JWT_PUBLIC_KEY` | Yes | — | Ed25519/RSA public key |
| `REDIS_URL` | No | `localhost:6379` | Redis for blacklist + cache |
| `NATS_URL` | No | `nats://localhost:4222` | NATS for events |
| `TENANT_SLUG` | No | `saldivia` | Tenant slug for NATS subjects |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry |

## DB Schema (7 tables)

- `contacts` — birth data with tenant+user isolation
- `astro_sessions` — conversation threads
- `astro_messages` — chat messages (role: user/assistant only)
- `astro_predictions` — prediction tracking with verification
- `astro_feedback` — thumbs up/down per message
- `astro_followups` — business follow-up items
- `astro_usage` — daily token usage

Migrations: `db/tenant/migrations/010-012`

## Build

Requires **CGO_ENABLED=1** (swephgo wraps Swiss Ephemeris C library).

```bash
# Local build
cd services/astro && CGO_ENABLED=1 go build -o astro ./cmd/...

# Run tests
CGO_ENABLED=1 CGO_LDFLAGS="-L$(pwd) -lm" go test ./internal/... -count=1

# Docker
docker build -f services/astro/Dockerfile .
```

## Testing

20 test files across 7 packages. All pass.
Golden file tests against Python astro-v2 reference data.
