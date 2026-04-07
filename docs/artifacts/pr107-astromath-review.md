# Gateway Review -- PR #107 Astromath Utilities

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

### [B1] `FindAspectWithMotion` applying/separating logic is broken
**File:** `services/astro/internal/astromath/aspects.go:38-43`

The raw `diff` calculation does not mirror what `AngDiff` computes. When `lon1 > lon2` by the long way around (e.g. lon1=310, lon2=10), `diff` becomes 300 while `asp.Angle` is 60. The applying/separating comparison against `asp.Angle` produces wrong results for roughly half of all aspect configurations.

**Fix:** Use the signed shortest-path difference instead:

```go
func FindAspectWithMotion(lon1, lon2, speed1, speed2, maxOrb float64) *Aspect {
    asp := FindAspect(lon1, lon2, maxOrb)
    if asp == nil {
        return nil
    }
    // Use shortest-path diff to determine direction
    diff := AngDiff(lon1, lon2) // always [0, 180]
    relSpeed := speed2 - speed1 // positive = gap increasing
    // Applying = the gap is closing toward exact
    if asp.Angle == 0 {
        asp.Applying = diff > 0 && relSpeed < 0 // closing toward conjunction
    } else {
        // For non-zero aspects, compare current diff vs exact angle
        asp.Applying = (diff < asp.Angle && relSpeed > 0) || (diff > asp.Angle && relSpeed < 0)
    }
    return asp
}
```

Note: applying/separating is notoriously tricky. The above is a starting point -- consider adding table-driven tests with known ephemeris data (e.g. Moon applying to square Saturn).

## Debe corregirse

### [D1] `stars.go` comment contradicts data
**File:** `services/astro/internal/astromath/stars.go:13`

Comment says "magnitude <= 1.5" but Algol has magnitude 2.1 and Castor has 1.58. Either:
- Change the comment to "magnitude <= 2.5" or "major fixed stars with astrological significance" (drop the magnitude claim), or
- Remove Algol/Castor from the list.

Algol (Caput Algol) is astrologically critical despite being magnitude 2.1, so updating the comment is the right call.

### [D2] `HouseForLon` has no bounds/length check on `cusps` slice
**File:** `services/astro/internal/astromath/houses.go:4`

The function accesses `cusps[1]` through `cusps[12]`, requiring a 13-element slice (Swiss Ephemeris convention: index 0 unused). A 12-element slice causes an index-out-of-range panic. Add a guard:

```go
func HouseForLon(lon float64, cusps []float64) int {
    if len(cusps) < 13 {
        return 1 // or return an error
    }
    // ...
}
```

And add a doc comment specifying the expected format: "cusps must be a 13-element slice where cusps[1]-cusps[12] are house cusps (Swiss Ephemeris convention)."

### [D3] Missing test coverage for `FindAspectWithMotion`
**File:** `services/astro/internal/astromath/angles_test.go`

There is no test for `FindAspectWithMotion`, `HouseForLon`, `SignedDiff`, `PartOfSpirit`, `SouthNode`, `ContraAntiscion`, or `Antiscion` (the antiscion test only checks one value). Given that B1 above is a logic bug, these need tests before merge.

At minimum add:
- `TestFindAspectWithMotion` with applying/separating cases in both directions
- `TestHouseForLon` including wrap-around (e.g. cusp 12 at 350, cusp 1 at 15)
- `TestSignedDiff` with negative inputs

## Sugerencias

### [S1] `Antiscion` test comment is wrong
**File:** `services/astro/internal/astromath/angles_test.go:115`

Comment says "0 Aries -> 180 Virgo (antiscion)" but 0 Aries antiscion is 0 Libra (180 degrees). Virgo ends at 180. The test value (180) is correct, but the comment confuses the sign.

### [S2] Consider returning `(string, bool)` from `BoundLord`
**File:** `services/astro/internal/astromath/bounds.go:40`

The fallback on line 49 (`return EgyptianBounds[sign][4].Lord`) silently handles the edge case of exactly 30.0 degrees (which should never happen after normalization, but could from float precision). Returning `(string, bool)` makes callers handle the impossible case explicitly.

### [S3] `AspectAngles` keys are English, `AspectNames` values are Spanish
This is fine for internal use but document the convention: English keys for programmatic matching, Spanish values for display. A comment above `AspectAngles` would help.

### [S4] `NaibodRate` precision
The Naibod rate is typically quoted as 0.9856473... (59'08.33"/year). The current value 0.985626 differs in the 5th decimal. Not a blocker since actual primary directions use this as an approximation, but verify against the source (De Revolutionibus).

### [S5] `Cástor` magnitude
Castor is a visual double. The combined magnitude is ~1.58 but individual components are fainter. If the intent is "stars visible to naked eye for mundane interpretation," 1.58 is fine. Just noting for completeness.

## Lo que esta bien

- **Egyptian bounds are correct.** All 12 signs verified to sum to exactly 30 degrees. Values match Ptolemy's Tetrabiblos I.21.
- **ZR minor years match Valens.** `{15, 8, 20, 25, 19, 20, 8, 15, 12, 27, 30, 12}` is correct.
- **SignLord uses traditional rulership.** All 12 assignments verified against Dorothean/Ptolemaic scheme.
- **`Normalize360` handles negatives correctly.** `math.Mod` + conditional add is the right approach.
- **`AngDiff` is correct** and symmetric (commutative). Good.
- **`EclToEq` formula is the standard spherical trig conversion.** Verified with (0,0,eps) -> (0,0).
- **Part of Fortune/Spirit correctly swap for diurnal/nocturnal.** Matches Hellenistic tradition.
- **`BoundLord` fallback** handles the 30.0-exactly edge case gracefully.
- **Spanish names consistent** across all files: Signs, PlanetNames, PlanetIDByName, SlowPlanets, SignLord, EgyptianBounds all use the same accented forms.
- **Pure Go, zero dependencies.** Clean package boundary.
- **Test table-driven style** follows project conventions.
