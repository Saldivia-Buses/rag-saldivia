# Gateway Review -- PR #114 Primary Directions (Polich-Page Mundane)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. Fortuna/Espiritu used with RA=0, Dec=0 [CORRECTNESS BUG]

**File:** `services/astro/internal/technique/primary_dir.go:182-186`
**Severity:** High

`BuildNatal` creates Fortuna and Espiritu with only `Lon` set:
```go
planets["Fortuna"] = &ephemeris.PlanetPos{Lon: astromath.PartOfFortune(...)}
```

Their `RA` and `Dec` default to `0.0`. The zero-check filter at line 183 only skips entries where `RA == 0 && Dec == 0 && Lon == 0`, so Fortuna/Espiritu pass through with garbage equatorial coordinates. Every direction involving them is mathematically wrong.

**Fix (choose one):**
- **(A)** In `BuildNatal`, compute RA/Dec for Fortuna/Espiritu via `astromath.EclToEq(lon, 0.0, epsilon)`.
- **(B)** In `FindDirections`, skip points that have `RA == 0 && Dec == 0` regardless of `Lon` (simpler, but loses Fortuna/Espiritu from results entirely).

Option A is preferred -- the Python source includes Fortuna as a promissor and it works because `ecl_to_eq()` is called for it.

---

## Debe corregirse

### D1. MC, AS, Vertex missing as promissors/significators

**File:** `services/astro/internal/technique/primary_dir.go:181-187`
**Severity:** Medium-High (functional gap, explains 31 vs 57 results)

`FindDirections` iterates only `chart.Planets`. The Python source adds MC, AS, and Vertice into its `planets` dict (via `ecl_to_eq` at build time), making them available as both promissors and significators. The golden file confirms: 2 results have MC as promissor, 4 AS, 7 Vertice, plus significator appearances.

**Fix:** Add MC, AS, Vertex as synthetic points in `FindDirections`:
```go
// After collecting planets:
mcRA, mcDec := astromath.EclToEq(chart.MC, 0.0, chart.Epsilon)
points = append(points, point{"MC", mcRA, mcDec})
ascRA, ascDec := astromath.EclToEq(chart.ASC, 0.0, chart.Epsilon)
points = append(points, point{"AS", ascRA, ascDec})
vtxRA, vtxDec := astromath.EclToEq(chart.Vertex, 0.0, chart.Epsilon)
points = append(points, point{"Vértice", vtxRA, vtxDec})
```

This alone closes ~16 of the 26 missing results. The remaining gap is Ecl.Prenatal (Python-only feature not yet ported) and Nodo Sur / Lilith significator combos.

### D2. Redundant `min()` function shadows Go 1.21+ builtin

**File:** `services/astro/internal/technique/primary_dir_test.go:134-139`
**Severity:** Low

Go 1.25 has a builtin `min()`. The local definition shadows it. Not a bug, but unnecessary boilerplate.

**Fix:** Delete the `min` function; the builtin works for `int`.

---

## Sugerencias

### S1. Golden test tolerance is too loose

The golden test (`TestFindDirections_Golden`) only checks that the tightest Python result exists with `arcDiff < 1.0` degree. A 1-degree tolerance on a PD arc means ~1 year of timing error. After fixing B1 and D1, tighten to `< 0.1` (or even `< 0.01` for planet-only pairs where ephemeris data matches exactly).

### S2. Consider deterministic iteration order for `points`

`chart.Planets` is a `map[string]*PlanetPos` -- iteration order is random. This means `results` before sorting can vary across runs. Sorting by orb at the end makes the final order deterministic, but if two results have identical orb, their relative order is unstable. Consider sorting `points` by name before iterating, or adding a secondary sort key (e.g., promissor name).

### S3. No test for circumpolar edge case

At extreme latitudes (> ~67 degrees), `diurnalSemiArc` will clamp to 0 or 180, and `semiarcPole` hits the `sa < 0.0001` guard. Worth adding a unit test with `lat = 70` to verify no NaN/Inf leak.

### S4. Minor aspects not yet supported

The Python source has `ASPECTS_MINOR` (semisextile, semisquare, quintile, biquintile, quincunx) with a `minor_only_angles` restriction. The Go code only has major aspects. This is fine for Phase 6 but should be tracked.

---

## Lo que esta bien

- **Core spherical geometry is correct.** `diurnalSemiArc`, `meridianDistance`, `isAboveHorizon`, `semiarcPole`, `obliqueAscension` all match the Python formulas exactly.
- **Mundane position 4-quadrant logic is a faithful port.** Q1-Q4 formulas, division-by-zero guards, and fallback values all match.
- **Arc subtraction direction is correct.** Direct = OA_P - OA_Q, Converse = OA_Q - OA_P, with proper modular arithmetic handling Go's signed `math.Mod`.
- **`clamp()` is used consistently** for all `asin`/`acos` inputs, preventing NaN from floating-point drift.
- **Conjunction/opposition sign=-1 skip** correctly avoids duplicate results (0 and 180 are symmetric).
- **`mundaneAspectPoint` correctly returns Dec=0** for the equatorial point Q, matching the Polich-Page mundane method.
- **NaibodRate constant matches** Python's `NAIBOD_RATE = 0.985626`.
- **No thread safety concern** -- pure math on natal data, no shared mutable state, no CalcMu needed.
- **Sanity test is solid** -- validates field presence, sistema tag, tipo tag, arc bounds, and sort order.
- **Result struct is clean** with proper JSON tags and all the fields needed by the frontend.
