# Gateway Review — PR #113 Phase 10: Eclipses, Zodiacal Releasing, Fixed Stars

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

### B1. `eclipseSubType` uses tret timing array, not return flags [MEDIUM-HIGH]

**File:** `services/astro/internal/technique/eclipses.go:126-136`

The function infers eclipse subtype by checking `tret[4] > 0` (totality begin time) and `tret[5] > 0` (totality end time). This is unreliable:

- For **total** eclipses, both `tret[4]` and `tret[5]` are nonzero, so `tret[4] > 0` returns "total" -- happens to work.
- For **annular** eclipses, the annularity begin/end are in `tret[6]`/`tret[7]` (if available), but `tret[4]` (totality begin) is 0. However, `tret[5]` (totality end) is also 0, so it falls through to "partial". **Annular eclipses are misclassified as partial.**
- The Swiss Ephemeris convention is that `swe_sol_eclipse_when_glob()` and `swe_lun_eclipse_when()` return the eclipse type as a **bitmask in the return value** (`SE_ECL_TOTAL=4`, `SE_ECL_ANNULAR=8`, `SE_ECL_PARTIAL=16`).

**Fix:** The `ephemeris.SolEclipseWhenGlob` and `ephemeris.LunEclipseWhen` wrappers must return `ret` (the bitmask) alongside `tret`. Then `eclipseSubType` should check the bitmask, not the timing array:

```go
// In ephemeris/sweph.go — return (retFlags int, tret []float64, err error)
func SolEclipseWhenGlob(...) (int, []float64, error) {
    // ...
    return ret, tret, nil
}

// In eclipses.go
func eclipseSubType(retFlags int) string {
    if retFlags&ephemeris.EclTotal != 0 { return "total" }
    if retFlags&ephemeris.EclAnnular != 0 { return "annular" }
    return "partial"
}
```

### B2. `SolEclipseWhenGlob` return bitmask discarded [MEDIUM-HIGH]

**File:** `services/astro/internal/ephemeris/sweph.go:219-228`

The wrapper stores `ret` from `swe.SolEclipseWhenGlob()` but only checks `ret < 0`. The positive value carries eclipse type flags that are needed by eclipseSubType. Same issue in `LunEclipseWhen` at line 231-240.

**Fix:** Change both wrappers to return `(int, []float64, error)` where the int is the return flags.

---

## Debe corregirse

### D1. Loosing of the Bond: `i == 6` only catches the first occurrence [MEDIUM]

**File:** `services/astro/internal/technique/zodiacal_releasing.go:82-83`

```go
loosing := i == 6
```

This is correct for the **first cycle** (i goes 0-11), but on cycle 2+ (when `cycle > 0`), `i` resets to 0 and the 7th sign from the start is again at `i == 6`. This works because the inner loop always starts from 0. So the logic is correct per iteration -- every time the person cycles back through the same sequence, sign 6 from the start is the opposite.

However, this is **only correct if loosing means "opposite to the starting sign"**. In Hellenistic tradition, loosing of the bond also occurs when the sub-period reaches the 7th sign from the *sub-period's* starting sign, not just from the lot sign. This is relevant for Level 2 (`findSubPeriod` line 133: `loosing: i == 6`). Since the sub-period starts from the major period's sign, `i == 6` correctly identifies the 7th sign from that starting point. **This is actually correct for all 3 levels.**

After re-analysis: **no change needed on loosing logic.** Removing from "must fix".

### D2. `findSubPeriod` computed `ageInMajor` is unused [LOW]

**File:** `services/astro/internal/technique/zodiacal_releasing.go:113, 137`

```go
ageInMajor := targetAge - major.StartAge
// ...
_ = ageInMajor  // line 137
```

Dead variable. Appears to be a leftover from debugging. Remove it.

### D3. No CalcMu protection in `FindEclipses` [LOW-MEDIUM]

**File:** `services/astro/internal/technique/eclipses.go:47`

`CalcPlanet` is called without holding `CalcMu`. Since eclipses.go uses `FlagSwieph|FlagSpeed` (no topocentric flag), and `CalcPlanet` does not call `SetTopo`, this is safe in the current code. However, if another goroutine calls `BuildNatal` concurrently (which does `SetTopo` + `CalcPlanet`), the global Swiss Ephemeris state could be inconsistent.

The plan doc says "CalcMu is an application-level mutex for compound operations where SetTopo + CalcPlanet must be atomic." Since eclipses don't use `SetTopo`, the contract is technically satisfied. But `SolEclipseWhenGlob` and `LunEclipseWhen` also modify global Swiss Ephemeris state internally. If concurrent calls are planned (e.g., HTTP handler serving multiple requests), this needs auditing.

**Recommendation:** Add a comment documenting why `CalcMu` is not needed here, or protect the entire `FindEclipses` call with `CalcMu` if concurrent usage is expected.

### D4. Eclipse loop advances by fixed +30 days -- can miss closely-spaced eclipses [LOW]

**File:** `services/astro/internal/technique/eclipses.go:49, 61, 75, 87`

After finding an eclipse, `jd = eclJD + 30`. Solar eclipses are typically ~6 months apart (minimum ~1 month in rare cases), so +30 is safe for solar. Lunar eclipses can also be ~6 months apart. The +30 day advance is acceptable and matches the Python implementation pattern.

No infinite loop risk: `eclJD` is always `>= jd` (Swiss Ephemeris returns the *next* eclipse), and adding 30 ensures forward progress. **This is correct.**

### D5. `FindFixedStarConjunctions` does not hold CalcMu [LOW]

**File:** `services/astro/internal/technique/fixed_stars.go:33`

`FixstarUT` calls `swe.FixstarUt` which uses global Swiss Ephemeris state. Same consideration as D3. Since there's no `SetTopo` involved, this is safe under current contract, but document the assumption.

---

## Sugerencias

1. **Test coverage for ZR nesting sum.** Add a test that verifies all Level 2 sub-period durations within a major period sum to exactly the major duration (within float64 epsilon). This would catch rounding drift:

```go
func TestZR_SubPeriodsSumToMajor(t *testing.T) {
    // For each sign as starting point, verify sub-periods sum correctly
    for startIdx := 0; startIdx < 12; startIdx++ {
        major := findActivePeriod(startIdx, 0.0, 1) // age 0 = first period
        var sum float64
        cumAge := major.StartAge
        totalYears := 0.0
        for i := 0; i < 12; i++ {
            totalYears += astromath.ZRSignYears[i]
        }
        for i := 0; i < 12; i++ {
            signIdx := (major.SignIndex + i) % 12
            sub := (astromath.ZRSignYears[signIdx] / totalYears) * major.Duration
            sum += sub
        }
        if math.Abs(sum - major.Duration) > 1e-10 {
            t.Errorf("sign %d: sub-periods sum %.15f != major %.15f", startIdx, sum, major.Duration)
        }
    }
}
```

2. **Golden file test for eclipses.** The test at `phase10_test.go:9-43` only validates count and ranges. A golden file with known 2026 eclipses (dates, types, longitudes) would catch regressions -- especially important given B1 above.

3. **`eclipseSubType` for lunar eclipses.** The `tret` array for `swe_lun_eclipse_when` has a different layout than `swe_sol_eclipse_when_glob`. Using the same `eclipseSubType` for both is fragile even if you fix it to use return flags. The return-flags approach (B1 fix) handles both correctly.

4. **Fixed stars test is weak.** `TestFindFixedStarConjunctions` only checks non-empty fields and orb bounds. Add at least one known conjunction for Adrian's chart (e.g., Regulus at ~29 Leo should be near specific natal points).

5. **`ZRSignYears` sum.** The total is 15+8+20+25+19+20+8+15+12+27+30+12 = 211 years. This is the correct Hellenistic value (sum of minor years). Good.

---

## Lo que esta bien

1. **FixstarUT buffer size (Phase 2 blocker).** Correctly handled in `ephemeris/sweph.go:196-209`. Buffer is `max(41, len(name)+1)`, matching `SE_MAX_STNAME*2+1`. The Phase 2 blocker is resolved.

2. **Eclipse loop termination.** Both solar and lunar loops have clear exit conditions: `tret[0] >= jdEnd` breaks, and `jd = eclJD + 30` ensures forward progress. No infinite loop risk.

3. **ZR 3-level nesting is mathematically sound.** Major period uses `ZRSignYears[signIdx]` directly. Sub-periods use proportional scaling `(ZRSignYears[signIdx] / totalYears) * major.Duration`, which guarantees the sum equals `major.Duration` (since proportions sum to 1.0 by construction). Bound periods use `((b.ToDeg - b.FromDeg) / 30.0) * sub.Duration`, and Egyptian bounds always sum to 30 per sign. No drift.

4. **Fixed star catalog.** 14 stars with correct Swiss Ephemeris names (`SweName` field). The separation of display name (Spanish) from SweName is clean.

5. **Tenant isolation.** Not applicable to this service -- pure computational, no DB, no tenant context. Correct for Phase 10.

6. **`SignLord` as `map[int]string`.** Traditional rulerships correctly mapped. All 12 indices present.
