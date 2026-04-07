# Astro Service

## What it does

Astrological calculation engine. Computes natal charts, predictive techniques, and
structured intelligence briefs via Swiss Ephemeris (CGO). Optionally narrates results
through an LLM via SSE streaming.

Version: 0.1.0 | Port: 8011

## Architecture

```
ephemeris (swephgo CGO wrapper)
  -> astromath (angles, aspects, houses, bounds, stars, conversions)
    -> natal (chart builder: planets, cusps, lots, combustion, retrograde)
      -> technique (11 predictive techniques)
        -> context (orchestrator: runs all techniques, builds brief)
          -> handler (HTTP endpoints, SSE streaming, contact CRUD)
```

Key design decisions:
- All Swiss Ephemeris access goes through `internal/ephemeris` -- never swephgo directly
- `ephemeris.CalcMu` mutex protects compound SetTopo + CalcPlanet sequences
- Planet names in Spanish (`Sol`, `Luna`, `Mercurio`, etc.)
- House system: Polich-Page Topocentric (`'T'`)
- Topocentric planetary positions (observer-location-corrected)

## Techniques (11)

| # | Technique | File | Function |
|---|-----------|------|----------|
| 1 | Natal chart | `natal/chart.go` | `BuildNatal()` |
| 2 | Transits (slow planets) | `technique/transits.go` | `CalcTransits()` |
| 3 | Solar Arc directions | `technique/solar_arc.go` | `CalcSolarArcForYear()`, `FindSolarArcActivations()` |
| 4 | Primary Directions (Naibod RA) | `technique/primary_dir.go` | `FindDirections()` |
| 5 | Secondary Progressions | `technique/progressions.go` | `CalcProgressions()` |
| 6 | Solar Return | `technique/solar_return.go` | `CalcSolarReturn()`, `CalcSolarReturnAtBirthplace()` |
| 7 | Profections | `technique/profections.go` | `CalcProfection()` |
| 8 | Firdaria | `technique/firdaria.go` | `CalcFirdaria()` |
| 9 | Eclipses | `technique/eclipses.go` | `FindEclipses()`, `FindEclipseActivations()` |
| 10 | Fixed Stars | `technique/fixed_stars.go` | `FindFixedStarConjunctions()` |
| 11 | Zodiacal Releasing | `technique/zodiacal_releasing.go` | `CalcZodiacalReleasing()` |

Additional computations inside context builder: planetary stations (`FindStations()`),
lunar returns (`CalcLunarReturns()`), monthly convergence matrix.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/v1/astro/natal` | JWT | Natal chart for a contact |
| POST | `/v1/astro/transits` | JWT | Slow-planet transits for a year |
| POST | `/v1/astro/solar-arc` | JWT | Solar arc directions for a year |
| POST | `/v1/astro/directions` | JWT | Primary directions (Naibod RA) |
| POST | `/v1/astro/progressions` | JWT | Secondary progressions for a year |
| POST | `/v1/astro/returns` | JWT | Solar return chart for a year |
| POST | `/v1/astro/profections` | JWT | Annual profection for a year |
| POST | `/v1/astro/firdaria` | JWT | Firdaria periods for a year |
| POST | `/v1/astro/fixed-stars` | JWT | Fixed star conjunctions to natal points |
| POST | `/v1/astro/brief` | JWT | Full context: all techniques + intelligence brief |
| POST | `/v1/astro/query` | JWT | SSE stream: context + LLM narration |
| GET | `/v1/astro/contacts` | JWT | List contacts for current user/tenant |
| POST | `/v1/astro/contacts` | JWT | Create a contact |

All POST technique endpoints accept:
```json
{"contact_name": "string", "year": 2026}
```
- `year` defaults to current year if omitted, range: -5000 to 5000
- `contact_name` is resolved per-tenant per-user from the contacts table

The `/v1/astro/query` SSE endpoint accepts:
```json
{"contact_name": "string", "query": "string", "year": 2026}
```
SSE events: `contact_recognized`, `calc_context`, `token` (chunked LLM response), `brief` (if no LLM), `error`, `done`.

## NATS Events

None. This service is stateless compute -- no pub/sub.

## Env vars

| Var | Required | Default | Description |
|-----|----------|---------|-------------|
| `ASTRO_PORT` | No | `8011` | HTTP listen port |
| `EPHE_PATH` | No | `/ephe` | Path to Swiss Ephemeris data files |
| `POSTGRES_TENANT_URL` | No | `""` | Tenant DB connection string (contacts CRUD disabled if empty) |
| `SGLANG_LLM_URL` | No | `""` | LLM endpoint for narration (SSE brief sent if empty) |
| `SGLANG_LLM_MODEL` | No | `""` | Model name for LLM requests |
| `LLM_API_KEY` | No | `""` | API key for LLM endpoint |
| `JWT_PUBLIC_KEY` | Yes | -- | Ed25519/RSA public key for JWT verification |
| `REDIS_URL` | No | `localhost:6379` | Redis for JWT blacklist |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `localhost:4317` | OpenTelemetry collector |

## Dependencies

- **PostgreSQL:** tenant DB (contacts table, optional -- pure calculation works without DB)
- **Redis:** JWT blacklist via `pkg/security`
- **Swiss Ephemeris:** ephemeris data files mounted at `EPHE_PATH`
- **LLM (optional):** SGLang or any OpenAI-compatible endpoint for query narration
- No NATS dependency

## DB schema

- Migrations: uses shared tenant migrations (`db/tenant/migrations/`) -- contacts table
- sqlc config: `db/sqlc.yaml`
- sqlc queries: `internal/repository/queries.sql`
- Generated code: `internal/repository/` (queries.sql.go, models.go, db.go)

## Build

Requires **CGO_ENABLED=1** because swephgo wraps the Swiss Ephemeris C library.

```bash
# Local build (macOS/Linux with gcc)
cd services/astro && CGO_ENABLED=1 go build -o astro ./cmd/...

# Docker (uses alpine + musl-dev)
docker build -f services/astro/Dockerfile .
```

The Dockerfile installs `gcc` and `musl-dev` in the builder stage, produces a minimal
alpine image with `/ephe` as a volume mount point for ephemeris data.

## Testing

14 test files, uses golden file pattern for cross-validation with Python astro-v2.

```bash
# Run all astro tests (needs EPHE_PATH set or ephemeris files in default location)
make test-astro

# With explicit ephemeris path
cd services/astro && EPHE_PATH=/path/to/ephe go test ./... -v -count=1
```

Test patterns:
- `TestMain` in handler and technique tests calls `ephemeris.Init()` / `ephemeris.Close()`
- Golden files in `testdata/golden/` generated by `testdata/generate_golden.py` from Python astro-v2
- `adrianChart(t)` helper builds a reference natal chart for technique tests
- Tests validate against known astronomical events (eclipse counts, station dates, etc.)
