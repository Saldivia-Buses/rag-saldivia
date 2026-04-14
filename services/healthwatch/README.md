# HealthWatch Service

Health monitoring and triage aggregator for the SDA platform.

## What it does

- Collects health data from all SDA services via `/health` + `/v1/info` endpoints
- Queries Prometheus for error rates, latency, and active alerts
- Monitors Docker containers via docker-socket-proxy (never raw socket)
- Persists health snapshots to platform DB for trend analysis
- Provides scrubbed executive summaries (no raw errors, IPs, credentials)
- Supports AI-powered triage via GitHub Actions daily cron

## Endpoints

All endpoints except `/health` require platform admin JWT.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Self health check (no auth) |
| GET | `/v1/info` | Build info |
| GET | `/v1/healthwatch/summary` | Executive health summary (scrubbed) |
| GET | `/v1/healthwatch/services` | Per-service health status |
| GET | `/v1/healthwatch/alerts` | Active Prometheus alerts |
| POST | `/v1/healthwatch/check` | Manual health check trigger |
| GET | `/v1/healthwatch/triage` | AI triage records |

## Port

`8014` (env: `HEALTHWATCH_PORT`)

## Dependencies

- **Platform PostgreSQL** — health_snapshots + triage_records tables
- **Prometheus** — metrics and alerts API
- **Docker socket proxy** — container status (never raw socket)
- **All SDA services** — `/health` + `/v1/info` HTTP calls

## Security

- All data endpoints require `requirePlatformAdmin` JWT middleware
- Prometheus queries use hardcoded service whitelist (no user input in PromQL)
- Docker access via socket proxy only (DS5)
- Summary output is scrubbed: no raw errors, IPs, credentials, stack traces (M3)

## Data Retention

Health snapshots are cleaned up automatically every 24 hours.
Retention period: 7 days (~141K rows max at steady state).
