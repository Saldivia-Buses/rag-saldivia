# Gateway Review -- PR #115 Context Builder + Intelligence Brief (Phase 11)

**Fecha:** 2026-04-06
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. Build() silently swallows ALL errors -- no diagnostics [builder.go:50-73]

Every technique that returns an error is handled with `if err == nil` and silently discarded on failure. This means if the ephemeris path is misconfigured, or a SwissEph call panics inside CGO, or any computation fails, the caller gets a nil error and a `FullContext` with nil fields -- and no way to know what failed.

At minimum, errors should be collected and either:
- Returned as a multi-error (`errors.Join`)
- Or attached to FullContext in a `Warnings []string` field so the brief can annotate "N/A" for missing techniques

**Fix:** Add `Warnings []string` to `FullContext`. Log each technique error and append to warnings. The brief should render warnings so the LLM knows what data is absent vs. genuinely empty.

```go
if prog, err := technique.CalcProgressions(chart, year); err != nil {
    ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("progressions: %v", err))
} else {
    ctx.Progressions = prog
}
```

### B2. Transits technique missing from Build() [builder.go]

The plan lists Phase 9b as "Slow planet transits" with `transits.go`, and the golden file `transits_adrian_2026.json` exists. However:
- `transits.go` does not exist in `services/astro/internal/technique/`
- `Build()` does not call any transits function
- `FullContext` has no `Transits` field
- The convergence matrix does not score transits

Slow-planet transits (Saturn, Jupiter, etc. over natal points) are a core predictive technique. The brief without transits is materially incomplete for LLM narration.

**Fix:** Either implement Phase 9b before this PR, or add a `TODO` field/comment and a Warning so it is tracked. If intentionally deferred, document it in the plan progress table.

---

## Debe corregirse

### D1. `age` calculation uses approximate year length [builder.go:41]

```go
age := midYear.Sub(birthDate).Hours() / (24 * 365.25)
```

This is consistent with CalcProfection and CalcFirdaria (both use the same formula), so it is internally coherent. However, `midYear` is July 1 but `jdMid` is June 15 (line 42). These two dates should match. The age is computed at July 1 but the JulDay for solar arc is June 15, creating a 16-day discrepancy.

**Fix:** Either use `time.Date(year, 6, 15, 12, 0, 0, 0, time.UTC)` for midYear, or use `ephemeris.JulDay(year, 7, 1, 0.0)` for jdMid. Pick one and be consistent.

### D2. PrimaryDirection month derivation is wrong for birthday != Jan 1 [brief.go:148-156]

```go
fractionalYear := d.AgeExact - float64(int(d.AgeExact))
month := int(fractionalYear*12) + 1
```

`AgeExact` is age from birth, not from January 1. If someone is born December 27 (like Adrian), `fractionalYear=0.0` maps to late December, not January. The month computed here is the month **of the age-year** (birthday to birthday), not the calendar month. For a December birthday, a fractional year of 0.5 means June-ish of the calendar year, which happens to be close -- but for a June birthday, fractional 0.0 maps to calendar June, not January.

The convergence matrix is labeled "monthly" and other techniques (eclipses, lunar returns) use calendar months. Mixing calendar months with age-year months produces incorrect convergence.

**Fix:** Convert age to calendar month:
```go
birthMonth := ctx.Chart.BirthDate.Month() // or pass birthDate into FullContext
calendarMonth := (int(birthMonth) - 1 + int(fractionalYear*12)) % 12 + 1
```

### D3. Solar Arc convergence scoring gives +1 to ALL 12 months but annotates only month 1 [brief.go:159-165]

When solar arcs are active, every month gets +1 score, which is conceptually correct (year-long effect). But the annotation `{N}_SA_activas` only appears on month 1. For the LLM, the other 11 months have an inflated score with no explanation of why.

**Fix:** Either annotate all 12 months (verbose but accurate), or use a separate "baseline" concept in the brief instead of inflating individual month scores.

### D4. Progression ingress scoring inflates all months by +2 per ingress, stacking [brief.go:177-189]

If a chart has 3 progressed sign ingresses (e.g., Sun, Moon, Mercury all change signs), every month gets +6 total. This drowns out targeted monthly events. Primary Directions give +3 to a specific month, but 3 sign ingresses give +6 to ALL months.

**Fix:** Reduce weight to +1 per ingress, or cap the total progression bonus. Alternatively, add ingress points only to the months near the ingress date (progressed speed gives approximate timing).

### D5. Lunar Return always scores +1 regardless of significance [brief.go:169-174]

Lunar returns happen ~13 times/year. Scoring every single one at +1 adds +13 spread across months -- noise that dilutes the signal from rare events like eclipses (+2) or primary directions (+3).

**Fix:** Either remove lunar returns from the convergence matrix (they are not predictive on their own), or only score lunar returns that activate natal points.

### D6. Build() signature does not accept `context.Context` [builder.go:33]

Per Go convention and SDA patterns, the first parameter should be `context.Context`. Even though this is a compute function (no DB/network), it will be called from HTTP handlers. Without `ctx`, there is no way to cancel a long-running Build (ephemeris calculations can be slow for full-year eclipses).

**Fix:** `func Build(ctx context.Context, chart *natal.Chart, contactName string, birthDate time.Time, year int) (*FullContext, error)`

---

## Sugerencias

### S1. `FullContext.Chart` is `json:"-"` but could be useful for handler

The full natal chart is excluded from JSON serialization. This is fine if the handler separately provides natal data, but if the API response needs to include chart positions alongside the brief, it will need to be re-fetched.

### S2. Fixed stars have no time dimension

`FindFixedStarConjunctions(chart)` checks natal-epoch star positions. These are static (natal data), not predictive for the target year. Consider noting this in the brief section or removing from convergence scoring entirely (currently not scored, which is correct).

### S3. Brief has hardcoded Spanish section headers

The brief uses Spanish headers ("SENORES DEL TIEMPO", "DIRECCIONES PRIMARIAS", etc.). This is correct per `bible.md` (UI language = Spanish). But if the LLM is prompted in English, these headers may confuse it. Consider making this configurable or documenting that the brief is always in Spanish.

### S4. Convergence matrix sort is a no-op [brief.go:191-193]

```go
sort.Slice(scores, func(i, j int) bool {
    return scores[i].Month < scores[j].Month
})
```

`scores` is already initialized in month order (1-12) by the `for i := range scores` loop. The sort does nothing. Either remove it or sort by score descending (which would actually be useful for identifying peak months).

### S5. Consider adding `ProfectionCascade` monthly lords to convergence

`CalcProfectionCascade` exists and provides monthly lords. The convergence matrix could score months where the monthly lord activates a tense natal configuration. Currently only the annual profection is used.

### S6. Test coverage is smoke-only

`TestBuild` and `TestBuildBrief_Sections` verify non-nil results and section headers, but do not test:
- Error path (what happens if chart is nil?)
- Convergence matrix scores (are they reasonable for Adrian 2026?)
- Empty case (a chart with zero activations)

---

## Lo que esta bien

- **Clean orchestration pattern:** Build() is a straightforward sequence. Easy to read, easy to add techniques.
- **Struct design:** `FullContext` captures all techniques in a single serializable struct. Clean separation between data and rendering.
- **Brief structure:** The 7-section layout (time lords, PD, SA, progressions, eclipses, solar return, convergence) follows standard astrological interpretation order. Good for LLM consumption.
- **Defensive nil checks in brief:** Every section checks for nil before accessing fields. No nil pointer panics.
- **Convergence matrix concept:** The idea of scoring months by technique overlap is solid and gives the LLM quantitative guidance on timing.
- **Test infra:** `TestMain` properly initializes/closes ephemeris. `adrianChart` helper is reusable.
- **Type safety:** All techniques return typed structs, not `interface{}`. The compiler catches misuse.
- **Correct `min()` usage for top-10 PD:** `top := min(10, len(ctx.Directions))` prevents out-of-bounds on small result sets.
