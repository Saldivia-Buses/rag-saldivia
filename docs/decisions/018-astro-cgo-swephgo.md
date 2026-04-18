# ADR 018: Astro service uses CGO (Swiss Ephemeris)

**Date:** 2026-04-06
**Status:** Superseded 2026-04-17 — astro service removed entirely from scope. Product pivot: Saldivia RAG for bus company, no astrology.
**Plan:** 11, 12

## Decision

The astro service uses CGO via swephgo to wrap the Swiss Ephemeris C library. This is the only SDA service that requires CGO.

## Context

Astrological calculations require the Swiss Ephemeris — the industry standard for planetary position computation. It's a C library with no pure Go equivalent. Options:
1. Pure Go reimplementation (years of work, error-prone)
2. CGO wrapper (swephgo, existing, GPL-3.0)
3. Python subprocess (our old astro-v2 approach)

## Choice

CGO via `github.com/mshafiee/swephgo` v1.1.0. All ephemeris access goes through `internal/ephemeris/sweph.go` — no direct swephgo calls elsewhere.

`CalcMu` mutex protects compound SetTopo + CalcPlanet sequences (topocentric positions require setting observer location before calculating).

## Consequences

- Dockerfile uses `alpine + gcc musl-dev` (not distroless)
- `CGO_ENABLED=1` required for all builds and tests
- `libswe.a` must be compiled from Swiss Ephemeris C sources
- Lock-per-chart (not global lock) for expensive operations (rectification, electional)
- GPL-3.0 license on swephgo (acceptable for current timeline)
