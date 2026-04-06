// Package context orchestrates all predictive techniques and produces
// structured text for LLM narration.
package context

import (
	"fmt"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// FullContext holds the complete astrological analysis for a contact+year.
type FullContext struct {
	ContactName  string                           `json:"contact_name"`
	Year         int                              `json:"year"`
	Chart        *natal.Chart                     `json:"-"`
	SolarArc     []technique.SolarArcResult       `json:"solar_arc"`
	Transits     []technique.TransitActivation   `json:"transits"`
	Stations     []technique.Station            `json:"stations"`
	Directions   []technique.PrimaryDirection     `json:"directions"`
	Progressions *technique.ProgressionsResult    `json:"progressions"`
	SolarReturn  *technique.SolarReturn           `json:"solar_return"`
	LunarReturns []technique.LunarReturn          `json:"lunar_returns"`
	Profection   *technique.Profection            `json:"profection"`
	Firdaria     *technique.Firdaria              `json:"firdaria"`
	Eclipses     []technique.EclipseActivation    `json:"eclipses"`
	FixedStars   []technique.FixedStarConjunction `json:"fixed_stars"`
	ZRFortune    *technique.ZRResult              `json:"zr_fortune"`
	ZRSpirit     *technique.ZRResult              `json:"zr_spirit"`
	Brief        string                           `json:"brief"`
	Warnings     []string                         `json:"warnings,omitempty"`
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

	// Build intelligence brief
	ctx.Brief = BuildBrief(ctx)

	return ctx, nil
}
