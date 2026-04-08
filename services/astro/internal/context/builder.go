// Package context orchestrates all predictive techniques and produces
// structured text for LLM narration.
package context

import (
	"fmt"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// FullContext holds the complete astrological analysis for a contact+year.
type FullContext struct {
	ContactName  string                           `json:"contact_name"`
	Year         int                              `json:"year"`
	Chart        *natal.Chart                     `json:"-"`

	// --- Phase 1: Existing techniques (Plan 11) ---
	SolarArc     []technique.SolarArcResult       `json:"solar_arc"`
	Transits     []technique.TransitActivation     `json:"transits"`
	Stations     []technique.Station               `json:"stations"`
	Directions   []technique.PrimaryDirection      `json:"directions"`
	Progressions *technique.ProgressionsResult     `json:"progressions"`
	SolarReturn  *technique.SolarReturn            `json:"solar_return"`
	LunarReturns []technique.LunarReturn           `json:"lunar_returns"`
	Profection   *technique.Profection             `json:"profection"`
	Firdaria     *technique.Firdaria               `json:"firdaria"`
	Eclipses     []technique.EclipseActivation     `json:"eclipses"`
	FixedStars   []technique.FixedStarConjunction  `json:"fixed_stars"`
	ZRFortune    *technique.ZRResult               `json:"zr_fortune"`
	ZRSpirit     *technique.ZRResult               `json:"zr_spirit"`

	// --- Plan 12 Fase 1: Pure Arithmetic ---
	Almuten       *astromath.AlmutenResult         `json:"almuten,omitempty"`
	Lots          []astromath.LotResult            `json:"lots,omitempty"`
	LotsActivated []astromath.LotActivation        `json:"lots_activated,omitempty"`
	Disposition   *astromath.DispositionResult      `json:"disposition,omitempty"`
	Sect          *astromath.SectAnalysis           `json:"sect,omitempty"`
	Temperament   *astromath.TemperamentResult      `json:"temperament,omitempty"`
	Melothesia    *astromath.MelothesiaResult       `json:"melothesia,omitempty"`
	Hyleg         *astromath.HylegResult            `json:"hyleg,omitempty"`
	HouseRulers   *astromath.HouseRulerResult       `json:"house_rulers,omitempty"`

	// --- Plan 12 Fase 2: Time-based ---
	TertiaryProg    *technique.TertiaryResult             `json:"tertiary_prog,omitempty"`
	Decennials      *technique.DecennialResult             `json:"decennials,omitempty"`
	FastTransits    []technique.FastTransitActivation       `json:"fast_transits,omitempty"`
	Lunations       *technique.LunationResult              `json:"lunations,omitempty"`
	PrenatalEclipse *technique.PrenatalEclipseResult       `json:"prenatal_eclipse,omitempty"`
	EclipseTriggers []technique.EclipseTrigger             `json:"eclipse_triggers,omitempty"`
	PlanetaryCycles []technique.PlanetaryCycle             `json:"planetary_cycles,omitempty"`
	TimingWindows   []technique.TimingWindow               `json:"timing_windows,omitempty"`
	ActivationChains []technique.ActivationChain           `json:"activation_chains,omitempty"`

	// --- Plan 12 Fase 4: Specialized ---
	Midpoints    *technique.MidpointResult                 `json:"midpoints,omitempty"`
	Declinations *technique.DeclinationResult              `json:"declinations,omitempty"`

	// --- Plan 12: Scoring + Cross-analyses + Natal analysis ---
	Score            int                                    `json:"score"`             // 0-100 activation score
	MonthlyScores    [12]int                                `json:"monthly_scores"`
	Verdicts         []TechniqueVerdict                     `json:"verdicts,omitempty"`
	Contradictions   []Contradiction                        `json:"contradictions,omitempty"`
	AspectPatterns   []astromath.AspectPattern              `json:"aspect_patterns,omitempty"`
	ChartShape       *astromath.ChartShape                  `json:"chart_shape,omitempty"`
	Hemispheres      *astromath.HemisphericDist             `json:"hemispheres,omitempty"`
	FullDignities    []astromath.DignityEntry               `json:"full_dignities,omitempty"`
	PlanetaryAge     *astromath.PlanetaryAgePeriod          `json:"planetary_age,omitempty"`
	Divisor          *DivisorResult                         `json:"divisor,omitempty"`
	TriplicityLords  *TriplicityLordsResult                 `json:"triplicity_lords,omitempty"`
	ChronoCross      *ChronocratorCross                     `json:"chrono_cross,omitempty"`
	RSLRCrossings    []RSLRCrossing                         `json:"rs_lr_crossings,omitempty"`
	PrenatalTransits []PrenatalEclipseActivation            `json:"prenatal_transits,omitempty"`

	Brief        string                                    `json:"brief"`
	Warnings     []string                                  `json:"warnings,omitempty"`
}

// Build runs all techniques and produces a FullContext.
func Build(chart *natal.Chart, contactName string, birthDate time.Time, year int) (*FullContext, error) {
	ctx := &FullContext{
		ContactName: contactName,
		Year:        year,
		Chart:       chart,
	}

	// Consistent mid-year anchor for age and JD
	jdMid := ephemeris.JulDay(year, 7, 1, 12.0)
	midYear := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)
	age := midYear.Sub(birthDate).Hours() / (24 * 365.25)

	// Solar Arc
	ctx.SolarArc = technique.FindSolarArcActivations(chart, jdMid)

	// Primary Directions
	ctx.Directions = technique.FindDirections(chart, age, 2.0)

	// Progressions
	if prog, err := technique.CalcProgressions(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("progressions: %v", err))
	} else {
		ctx.Progressions = prog
	}

	// Solar Return
	if sr, err := technique.CalcSolarReturnAtBirthplace(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("solar_return: %v", err))
	} else {
		ctx.SolarReturn = sr
	}

	// Lunar Returns
	if lr, err := technique.CalcLunarReturns(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("lunar_returns: %v", err))
	} else {
		ctx.LunarReturns = lr
	}

	// Profection
	ctx.Profection = technique.CalcProfection(chart, birthDate, year)

	// Firdaria
	ctx.Firdaria = technique.CalcFirdaria(birthDate, chart.Diurnal, year)

	// Eclipses
	if ecl, err := technique.FindEclipseActivations(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("eclipses: %v", err))
	} else {
		ctx.Eclipses = ecl
	}

	// Fixed Stars
	ctx.FixedStars = technique.FindFixedStarConjunctions(chart)

	// Zodiacal Releasing
	ctx.ZRFortune = technique.CalcZodiacalReleasing(chart, "Fortune", age)
	ctx.ZRSpirit = technique.CalcZodiacalReleasing(chart, "Spirit", age)

	// Slow planet transits (5-day sampling, mundane aspects)
	ctx.Transits = technique.CalcTransits(chart, year)

	// Station detection (D→Rx, Rx→D near natal points)
	ctx.Stations = technique.FindStations(chart, year)

	// ── Plan 12: Pure arithmetic (no ephemeris, no errors) ──

	ctx.Almuten = astromath.CalcAlmuten(chart.Planets, chart.ASC, chart.MC, chart.Diurnal)
	ctx.Lots = astromath.CalcAllLots(chart.Planets, chart.ASC, chart.Diurnal, chart.Cusps)
	if len(ctx.Lots) > 0 {
		ctx.LotsActivated = astromath.CalcLotsActivations(ctx.Lots, chart.Planets, chart.JD, year)
	}
	ctx.Disposition = astromath.CalcDisposition(chart.Planets)
	ctx.Sect = astromath.CalcSect(chart.Planets, chart.Diurnal)

	planetLons := make(map[string]float64)
	for name, pos := range chart.Planets {
		planetLons[name] = pos.Lon
	}
	ctx.Temperament = astromath.CalcTemperament(chart.Planets, chart.ASC, chart.MC)
	ctx.Melothesia = astromath.CalcMelothesia(planetLons)
	ctx.Hyleg = astromath.CalcHyleg(chart.Planets, chart.Cusps, chart.ASC, chart.Diurnal)

	houseRulerLons := make(map[string]float64)
	for name, pos := range chart.Planets {
		houseRulerLons[name] = pos.Lon
	}
	ctx.HouseRulers = astromath.CalcHouseRulers(chart.Cusps, houseRulerLons)

	// ── Plan 12: Time-based techniques ──

	if tp, err := technique.CalcTertiaryProgressions(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("tertiary_prog: %v", err))
	} else {
		ctx.TertiaryProg = tp
	}

	ctx.Decennials = technique.CalcDecennials(chart, birthDate, year)
	ctx.FastTransits = technique.CalcFastTransits(chart, year)

	if lun, err := technique.CalcLunations(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("lunations: %v", err))
	} else {
		ctx.Lunations = lun
	}

	if pe, err := technique.CalcPrenatalEclipses(chart); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("prenatal_eclipse: %v", err))
	} else {
		ctx.PrenatalEclipse = pe
	}

	if et, err := technique.CalcEclipseTriggers(chart, year); err != nil {
		ctx.Warnings = append(ctx.Warnings, fmt.Sprintf("eclipse_triggers: %v", err))
	} else {
		ctx.EclipseTriggers = et
	}

	ctx.PlanetaryCycles = technique.CalcPlanetaryCycles(chart, year)

	// ── Plan 12: Specialized (no ephemeris for these) ──

	ctx.Midpoints = technique.CalcMidpoints(chart)
	ctx.Declinations = technique.CalcDeclinations(chart)

	// ── Post-processing: cross-technique analysis ──

	ctx.TimingWindows = technique.CalcTimingWindows(
		ctx.Transits, ctx.Eclipses, ctx.Stations, ctx.Directions,
	)
	ctx.ActivationChains = technique.CalcActivationChains(
		ctx.SolarArc, ctx.Directions, ctx.Transits, ctx.Eclipses, ctx.Stations,
	)

	// ── Plan 12: Natal sub-analyses ──

	ctx.AspectPatterns = astromath.DetectAspectPatterns(chart.Planets)
	ctx.ChartShape = astromath.DetectChartShape(chart.Planets)
	ctx.Hemispheres = astromath.CalcHemisphericDist(chart.Planets, chart.Cusps)
	ctx.FullDignities = astromath.BuildFullDignityTable(chart.Planets, chart.Diurnal)
	ctx.PlanetaryAge = astromath.CurrentPlanetaryAge(age)

	// ── Plan 12: Cross-technique analyses ──

	ctx.Divisor = CalcDivisor(chart, age)
	ctx.TriplicityLords = CalcTriplicityLords(chart, age)
	ctx.ChronoCross = CalcChronocratorFirdariaCross(ctx.Profection, ctx.Firdaria)
	if ctx.SolarReturn != nil {
		ctx.RSLRCrossings = CalcRSLRCrossings(ctx.SolarReturn, ctx.LunarReturns)
	}
	ctx.PrenatalTransits = CalcPrenatalEclipseTransits(
		chart, ctx.PrenatalEclipse, ctx.SolarArc, ctx.Transits, ctx.Directions, year,
	)

	// ── Plan 12: Scoring + Synthesis ──

	ctx.Score = ActivationScore(ctx)
	ctx.MonthlyScores = MonthScores(ctx)
	ctx.Verdicts = ExtractVerdicts(ctx)
	ctx.Contradictions = ResolveContradictions(ctx.Verdicts)

	// Build intelligence brief
	ctx.Brief = BuildBrief(ctx)

	return ctx, nil
}
