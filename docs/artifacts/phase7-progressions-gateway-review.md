# Gateway Review -- Phase 7 Secondary Progressions

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

**Archivos revisados:**
- `/Users/enzo/rag-saldivia/services/astro/internal/technique/progressions.go`
- `/Users/enzo/rag-saldivia/services/astro/internal/technique/progressions_test.go`
- Python reference: `/Users/enzo/astro-v2/secondary_progressions.py`

---

## Bloqueantes

### 1. [progressions.go:52] Missing topocentric flag -- positions will differ from Python

The Go code uses `FlagSwieph | FlagSpeed` (geocentric). The Python uses `FLG_SWIEPH | FLG_TOPOCTR` (topocentric) and calls `swe.set_topo()` before calculating.

This is a correctness bug: progressed positions will be slightly different from the Python golden data and from the natal chart (which uses topocentric). Secondary progressions in the Polich-Page system should use topocentric positions to be consistent.

**Fix:**

```go
// Use topocentric flag + lock CalcMu for SetTopo atomicity
flag := ephemeris.FlagSwieph | ephemeris.FlagSpeed | ephemeris.FlagTopoctr

ephemeris.CalcMu.Lock()
defer ephemeris.CalcMu.Unlock()
ephemeris.SetTopo(chart.Lon, chart.Lat, chart.Alt)

// ... then use CalcPlanet (or CalcPlanetFullLocked) inside the lock
```

This also addresses the thread safety question -- currently the code does NOT hold `CalcMu` and does NOT call `SetTopo`, so it's technically thread-safe but **wrong**. With the fix, CalcMu must be held for the compound SetTopo + CalcPlanet block, same pattern as `BuildNatal`.

### 2. [progressions.go:43-44] Age calculation uses mid-year anchor (July 1) -- imprecise for ingress timing

```go
jdMid := ephemeris.JulDay(targetYear, 7, 1, 12.0)
ageYears := (jdMid - chart.JD) / 365.25
```

The Python caller passes an exact `age_years` computed from the actual query date. Using July 1 as anchor means:
- A January ingress and a December ingress both map to July 1 of the target year
- The progressed JD is always the same for a given year, losing ~6 months of precision

For the initial implementation this is acceptable (the plan says "ingress detection compares current year vs previous year"), but the function signature should accept a precise JD or date rather than just `targetYear int`, to support the future Context Builder (Phase 11) which will need exact timing.

**Fix (minimal, for follow-up):** Add a `CalcProgressionsForDate(chart, jd)` variant, or change the signature to accept `targetJD float64` instead of `targetYear int`. Keep `CalcProgressions(chart, year)` as a convenience wrapper.

---

## Debe corregirse

### 3. [progressions.go:50-51] prevProgressedJD subtracts 1.0 from ageYears, not 1.0/365.25

```go
prevProgressedJD := chart.JD + (ageYears - 1.0)
```

The intent is "previous year." Since `ageYears` is in years and the formula is `progressedJD = natal_jd + ageYears`, subtracting 1.0 from ageYears means `prevProgressedJD` is 1.0 JD days earlier (= 1 progressed year earlier). This is **correct** for ingress detection -- it compares the position "one year of life ago" which is exactly the right semantics.

Confirmed correct. No fix needed. (Keeping this note because it's the most confusing line in the file and easy to misread.)

### 4. [progressions.go:76-80] Sign ingress hides house ingress

```go
if prevSign != pp.Sign {
    pp.Ingress = "sign"
    pp.PrevSign = prevSign
} else if prevHouse != pp.House {
    pp.Ingress = "house"
    pp.PrevHouse = prevHouse
}
```

If a planet changes both sign AND house in the same year (common -- sign boundaries often coincide with house boundaries), only the sign ingress is reported. The house ingress is silently dropped.

**Fix:** Use a slice or report both:

```go
type ProgressedPosition struct {
    // ...
    Ingresses []string `json:"ingresses,omitempty"` // ["sign", "house"]
    PrevSign  string   `json:"prev_sign,omitempty"`
    PrevHouse int      `json:"prev_house,omitempty"`
}
```

Or at minimum, always populate both `PrevSign` and `PrevHouse` and set `Ingress` to `"sign+house"` when both change.

### 5. [progressions.go:31-36] Map iteration order is non-deterministic

```go
var progressedPlanets = map[string]int{ ... }
```

`for name, pid := range progressedPlanets` iterates in random order. The `Positions` slice in `ProgressionsResult` will have planets in different order on each call. This makes tests flaky (the test iterates to find Sol, so it works, but consumers comparing JSON output will get inconsistent results).

**Fix:** Use a slice of structs (same pattern as `SlowPlanets` in constants.go):

```go
var progressedPlanets = []struct {
    Name string
    ID   int
}{
    {"Sol", ephemeris.Sun},
    {"Luna", ephemeris.Moon},
    {"Mercurio", ephemeris.Mercury},
    {"Venus", ephemeris.Venus},
    {"Marte", ephemeris.Mars},
    {"Jupiter", ephemeris.Jupiter},
    {"Saturno", ephemeris.Saturn},
}
```

### 6. [progressions.go:57-59] Errors silently skipped

```go
pos, err := ephemeris.CalcPlanet(progressedJD, pid, flag)
if err != nil {
    continue
}
```

If Sol or Luna fails (critical planets in progressions), the result silently omits them. The caller has no way to know. At minimum, Sol and Luna errors should be propagated.

**Fix:** Return error for Sol/Luna; `continue` is acceptable for the others.

### 7. [progressions.go:28] Field name `AgeDays` is misleading

```go
AgeDays float64 `json:"age_days"`
```

The value stored is `ageYears` (line 44), not age in days. The JSON tag says `age_days` which will confuse every consumer.

**Fix:** Rename to `AgeYears float64 \`json:"age_years"\`` and fix the JSON tag.

---

## Missing from Python (scope for future phases)

These are features present in the Python `secondary_progressions.py` that are **not** in the Go implementation. Not all are needed now, but they should be tracked:

| Feature | Python function | Priority |
|---------|----------------|----------|
| Progressed aspects to natal (orb 1.5) | `prog_aspects_to_natal()` | **High** -- core functionality |
| Progressed MC/ASC angles | included in `calculate_secondary_progressions()` | **High** -- most important progressed points |
| Progressed Moon block (speed, time-to-ingress, aspects) | `_luna_prog_block()` | **High** -- Moon is the main SP timer |
| Progressed New/Full Moon cycle | `_prog_nlll_block()` | Medium |
| Stationary planet detection (D->R, R->D) | `_prog_stations_block()` | Medium |
| Progressed cusps vs natal | `_prog_cusps_block()` | Medium |
| Natal planets in progressed houses | `_natal_planets_in_prog_houses()` | Low |
| Cusp ingress scanning (bisection) | `_prog_cusp_ingresses()` | Low |
| Topocentric houses for progressed chart | `swe.houses(prog_jd, lat, lon, b'T')` | **High** -- missing entirely |
| Chiron in progressions | included in Python | Low (slow, but some use it) |

The biggest gap is that the Python computes a full progressed chart (houses, angles, aspects to natal) while the Go version only computes planet positions with sign/house ingress detection. The Go version is roughly `calculate_secondary_progressions()` minus houses/angles/aspects.

---

## Sugerencias

1. **Golden file test.** The solar arc and profections tests verify against Python golden files. Progressions should too. Generate `testdata/golden/progressions_adrian_2026.json` from the Python and add a golden comparison test.

2. **Progressed Moon deserves a dedicated function.** The Python has `_luna_prog_block` because progressed Moon is the most important timing indicator in SP. Consider `CalcProgressedMoon(chart, targetYear)` that returns Moon position, speed, time-to-next-sign-ingress, and aspects to natal.

3. **Handler still a stub.** `handler/astro.go:36` has `Progressions` as a stub returning 501. This is expected since the handler integration is Phase 13, but worth noting.

---

## Lo que esta bien

1. **Day-for-year formula is correct.** `progressedJD = natal_jd + ageYears` where ageYears is in years and each year maps to 1 JD day -- textbook secondary progressions.

2. **Planet selection is correct.** Sol through Saturno only. Outer planets (Urano, Neptuno, Pluton) barely move in progressed charts and are correctly excluded.

3. **Ingress detection approach is sound.** Comparing this year vs. previous year will catch all ingresses that happen within a 1-year window. The 1.0 JD-day offset correctly maps to 1 progressed year.

4. **Reuses established patterns.** `astromath.SignName`, `astromath.HouseForLon`, `ephemeris.CalcPlanet` -- consistent with the rest of the codebase.

5. **Tests verify the core math.** The Sun-moved-~50-degrees check is a good sanity test for the day-for-year formula.

6. **No thread safety bugs (currently).** Since the code does not use `FlagTopoctr` and does not call `SetTopo`, it doesn't need `CalcMu`. This is accidentally safe for the wrong reason (see Bloqueante #1).
