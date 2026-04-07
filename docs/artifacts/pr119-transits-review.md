# Gateway Review -- PR #119 Slow Planet Transits

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

Ninguno.

## Debe corregirse

1. **[transits.go:74] Sentinel zero-check may skip legitimate vernal-equinox positions**

   The filter `pos.RA == 0 && pos.Dec == 0 && pos.Lon == 0` is meant to skip synthetic points that lack real coordinates. However, any point with *exactly* `Lon == 0.0` (0 Aries) AND `RA == 0.0` AND `Dec == 0.0` would be falsely excluded. In practice this is astronomically improbable for real planets, but the check is fragile. A safer guard: skip by name (the calculated points with no real equatorial position are known: `AS`, `MC`, `Vertice`), but since those are NOT in `chart.Planets`, this filter only matters for entries like `Fortuna`/`Espiritu` whose RA is computed via `EclToEq` and could theoretically be near zero.

   **Recommendation:** Low probability, but consider adding a comment documenting the sentinel assumption, or use a dedicated `hasEquatorial bool` field on `PlanetPos`.

2. **[transits.go:91] Double ephemeris call per sample (perf)**

   Each of the ~73 samples/year per planet makes TWO `CalcPlanet` calls (ecliptic + equatorial). For 7 slow planets * 73 samples = 1022 ephemeris calls. This works, but the natal chart already uses `CalcPlanetFullLocked` which gets both in one helper. Consider using the same pattern here to halve the call count.

   **Fix:** Replace the two `CalcPlanet` calls with a single `CalcPlanetFull`-style helper:
   ```go
   pos, err := ephemeris.CalcPlanetFull(jd, sp.ID, flags)
   // pos.Lon, pos.RA, pos.Speed all populated
   samples = append(samples, transitSample{jd: jd, lon: pos.Lon, ra: pos.RA, speed: pos.Speed})
   ```
   Note: `CalcPlanetFullLocked` takes the lock internally; you'd need an unlocked variant or just call `calcPlanetFullLocked` after holding the lock yourself. Alternatively, keep the two calls if locking per-sample is acceptable (it is for this workload).

3. **[transits.go:187] groupEpisodes assumes sorted input**

   `groupEpisodes` expects hits sorted by JD. They ARE sorted because the outer loop iterates `jd` in ascending order. But this invariant is implicit. If the code is ever refactored (e.g., parallelized), this breaks silently.

   **Fix:** Add a `sort.Slice(hits, func(i, j int) bool { return hits[i].jd < hits[j].jd })` at the top of `groupEpisodes`, or document the precondition.

4. **[transits_test.go:33] Threshold of 20 is arbitrary and fragile**

   `len(results) < 20` will break as soon as orbs or planet lists change. The golden file comparison below is the real validation. Either tie the threshold to `len(golden.Output)` with a tolerance (e.g., `len(results) < len(golden.Output)*70/100`) or remove the hard floor.

5. **[transits.go:104-106] Self-skip compares Spanish names across two different name sources**

   `sp.Name` comes from `slowPlanetIDs` (hardcoded Spanish) and `np.name` comes from `chart.Planets` map keys (also Spanish, from `natal/chart.go`). They match today, but this is a latent coupling. If either source changes naming (e.g., accent normalization), transits-to-self leak through.

   **Recommendation:** Low risk given both are in this repo, but a comment noting the coupling would help.

## Sugerencias

- **Natal angles as transit targets.** ASC, MC, and Vertex are NOT in `chart.Planets`, so transits to angles are not computed. If the Python reference includes them (common in transit analysis), this is a feature gap. Consider adding them to `natPoints` from `chart.ASC`/`chart.MC` with their RA computed via `EclToEq`.

- **Aspect name consistency.** Transits use English aspect names (`conjunction`, `trine`, etc.) while the golden file uses Spanish (`Conjuncion`, `Trigono`). The test (line 44-48) manually maps them. Consider using `astromath.AspectNames` to normalize, or store both forms in `TransitActivation`.

- **Episode detail month granularity.** With 5-day sampling, `MonthStart` and `MonthEnd` have +/-5 day precision. Document this limitation or interpolate to find the actual aspect-entry date.

- **Convergence weight.** Transits score weight=2 per episode-month in `buildConvergenceMatrix`. For a triple-pass Pluto transit spanning 6 months, that's 12+ points. This might dominate the matrix over primary directions (weight=3 but 1 month). Consider capping per-transit contribution or weighting by orb tightness.

## Lo que esta bien

- **Mundane aspects via RA** are correctly implemented. The `FindAspect` function works in the 0-360 degree domain and RA is stored in degrees, so the aspect geometry is valid.
- **Episode grouping** with 20-day gap threshold is a clean approach to detecting the classic 3-pass (direct-retro-direct) transit pattern.
- **Retrograde detection** correctly uses ecliptic longitude speed (negative = Rx), not RA speed.
- **Transit orb of 3.0 degrees** from `OrbDefaults.Transit` matches the Python reference config.
- **Builder integration** is clean: single call in `builder.go:96`, warning-free (CalcTransits doesn't error), and the brief section at `brief.go:100-114` formats correctly with pass count, retro flag, and nature.
- **Convergence integration** at `brief.go:204-217` correctly iterates episode months and avoids double-labeling (only first month of each episode gets the technique tag).
- **Test structure** is solid: golden file comparison + structural invariant test (fields non-empty, passes == len(ep_details), nature in valid set).
