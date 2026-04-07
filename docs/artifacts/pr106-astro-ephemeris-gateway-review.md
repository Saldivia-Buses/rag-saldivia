# Gateway Review -- PR #106 Astro Ephemeris Layer (Phase 2, Plan 11)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

**Scope:** `services/astro/internal/ephemeris/sweph.go`, `solcross.go`, `sweph_test.go`

---

## Bloqueantes

### B1. `FixstarUT` buffer too small -- swephgo writes back into `star` [CRITICAL]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:169`

```go
star := make([]byte, len(name)+1)
copy(star, name)
```

Swiss Ephemeris `swe_fixstar_ut()` **writes the resolved star name back** into the `star` buffer (e.g., input `"Regulus"` becomes `"Regulus,alLeo"`). The C API requires `char star[SE_MAX_STNAME*2+1]` = 41 bytes minimum. swephgo passes this buffer through to C via cgo.

If the input name is short (e.g., 7 bytes for "Regulus"), the buffer is only 8 bytes. The C function will write past the buffer boundary, causing **memory corruption or a segfault**.

**Fix:**
```go
star := make([]byte, 41) // SE_MAX_STNAME*2+1
copy(star, name)
```

---

### B2. `CalcPlanetFull` is NOT atomic under CalcMu -- race between two `CalcPlanet` calls [CRITICAL]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:121-137`

The doc comment says "Caller must hold CalcMu if using topocentric positions." But the function itself makes **two** separate `CalcPlanet` calls (ecliptic then equatorial). Between these two calls, another goroutine can call `SetTopo()` + `CalcPlanet()`, changing the global observer position.

The compound operation of "CalcPlanet(ecliptic) + CalcPlanet(equatorial) for the same observer" is not atomic. This is the exact scenario `CalcMu` was designed to protect.

The comment pushes the responsibility to the caller, which is correct in principle, but the function should either:
- (a) **Accept a `sync.Locker` and lock internally** for the two-call sequence, or
- (b) **Document explicitly** that the caller must hold `CalcMu` for the ENTIRE duration of `CalcPlanetFull` when topocentric is involved -- not just "if using topocentric."

Currently, the most likely call site (BuildNatal) will call `CalcPlanetFull` inside a loop for each planet. If the caller locks before the loop and unlocks after, it's fine. If the caller locks per-call, it's broken.

**Severity note:** This is CRITICAL because concurrent natal chart builds for different locations WILL produce incorrect positions. It won't crash -- it'll silently return wrong astronomical data.

**Fix (recommended):** Add explicit locking in the function when topocentric flag is present:

```go
func CalcPlanetFull(jdUT float64, planet, baseFlags int) (*PlanetPos, error) {
    if baseFlags&FlagTopoctr != 0 {
        CalcMu.Lock()
        defer CalcMu.Unlock()
    }
    // ... rest unchanged
}
```

Or document that callers must hold CalcMu across the entire batch. Either way, the current comment is insufficient.

---

## Debe corregirse

### M1. `CalcPlanet` contaminates `Lon`/`Lat` fields when `FlagEquatorial` is set [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:106-116`

```go
pos := &PlanetPos{
    Lon:   xx[0],
    Lat:   xx[1],
    Dist:  xx[2],
    Speed: xx[3],
}
if flags&FlagEquatorial != 0 {
    pos.RA = xx[0]
    pos.Dec = xx[1]
}
```

When `FlagEquatorial` is set, Swiss Ephemeris returns RA in `xx[0]` and Dec in `xx[1]`. The code unconditionally assigns these to `Lon`/`Lat` AND then also to `RA`/`Dec`. This means `pos.Lon` contains right ascension, not ecliptic longitude.

Within `CalcPlanetFull` this doesn't cause bugs (the equatorial result's `Lon`/`Lat` are discarded). But if anyone calls `CalcPlanet` directly with `FlagEquatorial`, they'll get misleading data in `Lon`/`Lat`.

**Fix:**
```go
pos := &PlanetPos{
    Dist:  xx[2],
    Speed: xx[3],
}
if flags&FlagEquatorial != 0 {
    pos.RA = xx[0]
    pos.Dec = xx[1]
} else {
    pos.Lon = xx[0]
    pos.Lat = xx[1]
}
```

---

### M2. `CalcPlanetFull` strips topocentric from equatorial call -- equatorial RA/Dec will be geocentric [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:128`

```go
eqFlags := (baseFlags &^ FlagTopoctr) | FlagEquatorial | FlagSwieph | FlagSpeed
```

This explicitly removes `FlagTopoctr` from the equatorial calculation. So when a caller requests topocentric-equatorial data (e.g., for rise/set refinement or parallax-corrected RA/Dec), the ecliptic positions are topocentric but the RA/Dec values are geocentric. This mismatch could produce errors up to ~1 degree for the Moon.

If stripping topocentric is intentional (e.g., because `SetTopo` isn't set for the equatorial call), this needs a comment explaining why. If it's a copy-paste from a reference implementation that didn't need topocentric RA/Dec, it should be reviewed against actual use cases.

**Fix:** Either keep `FlagTopoctr` in the equatorial flags:
```go
eqFlags := baseFlags | FlagEquatorial | FlagSwieph | FlagSpeed
```
Or add a comment: `// Equatorial call is geocentric by design — topocentric parallax applied only to ecliptic.`

---

### M3. Error string from `serr` includes trailing null bytes [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:103`

```go
return nil, fmt.Errorf("swe.CalcUt(planet=%d): %s", planet, string(serr))
```

`serr` is a 256-byte buffer. `string(serr)` converts ALL 256 bytes, including null padding after the C string terminator. This produces error messages with trailing `\x00` garbage bytes.

Same issue in `FixstarUT` (line 174), `SolEclipseWhenGlob` (line 187), `LunEclipseWhen` (line 199), `EclNut` (line 215), and `RiseTrans` (line 227).

**Fix:** Trim null bytes before conversion. Add a helper:
```go
func cstr(b []byte) string {
    if i := bytes.IndexByte(b, 0); i >= 0 {
        return string(b[:i])
    }
    return string(b)
}
```

Then: `return nil, fmt.Errorf("swe.CalcUt(planet=%d): %s", planet, cstr(serr))`

---

### M4. Hardcoded magic numbers for swephgo constants [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:35-40`

```go
FlagSwieph     = 2
FlagSpeed      = 256
FlagEquatorial = 2048
FlagTopoctr    = 32768
```

These duplicate constants from `swephgo` (`swe.SEFLG_SWIEPH`, etc.). If swephgo ever changes values (unlikely but possible with a major version bump), these will silently diverge. More importantly, reviewing correctness requires cross-referencing two codebases.

**Fix:** Use swephgo's constants directly:
```go
const (
    FlagSwieph     = swe.SEFLG_SWIEPH
    FlagSpeed      = swe.SEFLG_SPEED
    FlagEquatorial = swe.SEFLG_EQUATORIAL
    FlagTopoctr    = swe.SEFLG_TOPOCTR
)
```

If this was intentional to avoid swephgo leaking into the public API, then at minimum add a `func init()` with assertions:
```go
func init() {
    if FlagSwieph != swe.SEFLG_SWIEPH { panic("FlagSwieph drift") }
    // ...
}
```

---

### M5. `SolcrossUT` Newton iteration can overshoot across year boundary [MEDIUM]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/solcross.go:33-34`

```go
if speed == 0 {
    speed = 1.0
}
jd += diff / speed
```

The Sun's ecliptic speed is always ~0.95-1.02 deg/day, so `speed == 0` guard is fine as a safety net. However, there is no **step clamping**. If `diff` is close to 180 (maximum after normalization) and speed is ~1, the step is +180 days. This can overshoot into the next year's crossing or oscillate.

For the Sun, convergence is practically guaranteed because: (a) the Sun moves monotonically in longitude (no retrograde), (b) speed is nearly constant. But for a generic Newton solver, the lack of step clamping is fragile.

**Fix (defensive):** Clamp the step to a reasonable maximum:
```go
step := diff / speed
if step > 200 {
    step = 200
} else if step < -200 {
    step = -200
}
jd += step
```

---

## Sugerencias

### S1. Tests don't call `Init`/`Close` consistently -- shared global state [LOW]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph_test.go`

Tests `TestCalcPlanet_SunPosition`, `TestCalcPlanetFull_Equatorial`, `TestCalcHouses_TopocentricoRosario`, `TestEclNut`, and `TestSolcrossUT` each independently call `Init()`/`defer Close()`. But `TestJulDay_KnownDate` and `TestRevJul_RoundTrip` don't (they don't need ephemeris data since JulDay is pure math).

Go runs tests in a single binary. If test order changes, a test calling `Close()` could affect a subsequent test. Consider using `TestMain` to init/close once:

```go
func TestMain(m *testing.M) {
    Init(os.Getenv("EPHE_PATH"))
    code := m.Run()
    Close()
    os.Exit(code)
}
```

---

### S2. Missing test: `CalcPlanetFull` equatorial call with topocentric flag [LOW]

The test `TestCalcPlanetFull_Equatorial` uses `FlagSwieph|FlagSpeed` (no topocentric). The most complex code path -- `CalcPlanetFull` with `FlagTopoctr` -- is untested. This is the path where the CalcMu race (B2) and the topocentric stripping (M2) both manifest.

---

### S3. Missing test: `SolcrossUT` non-convergence [LOW]

No test verifies the error path when `SolcrossUT` fails to converge. A test with a planet other than Sun (if the function is ever generalized) or with deliberately wrong flags would exercise this.

---

### S4. Missing test: `FixstarUT` [LOW]

No test coverage for `FixstarUT`. This is also the function with the buffer overflow bug (B1). A test with a known star (e.g., "Regulus") at J2000.0 would catch the crash.

---

### S5. Missing test: `RiseTrans`, `SolEclipseWhenGlob`, `LunEclipseWhen` [INFO]

Zero coverage for eclipse and rise/set functions. These are important for transit calculations. Consider at minimum one golden-data test per function.

---

### S6. `CalcHouses` error message is vague [INFO]

**Archivo:** `/Users/enzo/rag-saldivia/services/astro/internal/ephemeris/sweph.go:149`

```go
return nil, nil, fmt.Errorf("swe.Houses failed")
```

Unlike `CalcPlanet`, this doesn't include the `serr` buffer because `swe.Houses` doesn't take one. But the error message should include the inputs for debugging:

```go
return nil, nil, fmt.Errorf("swe.Houses(jd=%.4f, lat=%.4f, lon=%.4f, sys=%c) failed", jdUT, geolat, geolon, rune(hsys))
```

Same for `CalcHousesEx` (line 160).

---

## Lo que esta bien

1. **Clean abstraction boundary.** No other package imports swephgo directly -- this is the single point of contact. Good for testability and future swephgo upgrades.

2. **CalcMu design is sound in principle.** Using an application-level mutex for compound operations (SetTopo+Calc) while relying on swephgo's internal mutex for individual calls is the correct approach. The issue is only in documentation/enforcement (B2).

3. **PlanetPos struct is well-designed.** Having both ecliptic (Lon/Lat) and equatorial (RA/Dec) in one struct with Speed and Dist avoids the need for multiple return types.

4. **SolcrossUT is a good replacement for the missing swephgo binding.** The Newton iteration is mathematically correct for the Sun's case (monotonic longitude, near-constant speed). 50 iterations with 1e-7 tolerance is generous -- convergence typically happens in 3-5 iterations.

5. **Tests use known astronomical constants.** J2000.0 epoch (JD 2451545.0), Sun at ~280.46 longitude, obliquity ~23.44 -- these are well-established reference values that make the tests self-documenting.

6. **Moshier fallback for tests.** `ephePath` returns empty string when `EPHE_PATH` is unset, falling back to swephgo's built-in Moshier algorithm. This means tests run without external ephemeris files -- good for CI.

7. **RevJul round-trip test.** Testing JulDay->RevJul round-trip with a real birth date (not just J2000.0) catches edge cases in hour precision.
