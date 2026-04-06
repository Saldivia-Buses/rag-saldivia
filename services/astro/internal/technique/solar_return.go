package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// SolarReturn holds a Solar Return chart for a given year.
type SolarReturn struct {
	JD      float64                          `json:"jd"`
	Year    int                              `json:"year"`
	HourUT  float64                          `json:"hour_ut"`
	ASC     float64                          `json:"asc_lon"`
	MC      float64                          `json:"mc_lon"`
	Cusps   []float64                        `json:"cusps"`
	Planets map[string]*ephemeris.PlanetPos  `json:"planets"`
}

// CalcSolarReturn finds the exact moment the transiting Sun returns to its
// natal longitude in the target year, then builds a chart for that moment.
func CalcSolarReturn(natalChart *natal.Chart, targetYear int, lat, lon, alt float64) (*SolarReturn, error) {
	natalSunLon := natalChart.Planets["Sol"].Lon

	// Start searching from ~10 days before the birthday in the target year
	_, birthMonth, birthDay, _ := ephemeris.RevJul(natalChart.JD)
	jdStart := ephemeris.JulDay(targetYear, birthMonth, birthDay-10, 0.0)

	// Find exact JD when Sun crosses natal Sun longitude
	jdReturn, err := ephemeris.SolcrossUT(natalSunLon, jdStart, ephemeris.FlagSwieph|ephemeris.FlagSpeed)
	if err != nil {
		return nil, err
	}

	_, _, _, hourUT := ephemeris.RevJul(jdReturn)

	// Build chart at the return moment for the specified location
	ephemeris.CalcMu.Lock()
	defer ephemeris.CalcMu.Unlock()

	ephemeris.SetTopo(lon, lat, alt)

	cusps, ascmc, err := ephemeris.CalcHousesEx(jdReturn, ephemeris.FlagSwieph|ephemeris.FlagTopoctr, lat, lon, ephemeris.HouseTopocentric)
	if err != nil {
		return nil, err
	}

	flag := ephemeris.FlagSwieph | ephemeris.FlagTopoctr | ephemeris.FlagSpeed
	planets := make(map[string]*ephemeris.PlanetPos)

	srPlanets := map[string]int{
		"Sol": ephemeris.Sun, "Luna": ephemeris.Moon,
		"Mercurio": ephemeris.Mercury, "Venus": ephemeris.Venus,
		"Marte": ephemeris.Mars, "Júpiter": ephemeris.Jupiter,
		"Saturno": ephemeris.Saturn, "Urano": ephemeris.Uranus,
		"Neptuno": ephemeris.Neptune, "Plutón": ephemeris.Pluto,
	}

	for name, pid := range srPlanets {
		pos, err := ephemeris.CalcPlanetFullLocked(jdReturn, pid, flag)
		if err != nil {
			return nil, err
		}
		planets[name] = pos
	}

	return &SolarReturn{
		JD:      jdReturn,
		Year:    targetYear,
		HourUT:  hourUT,
		ASC:     ascmc[0],
		MC:      ascmc[1],
		Cusps:   cusps,
		Planets: planets,
	}, nil
}

// CalcSolarReturnAtBirthplace is a convenience that uses the natal chart's location.
func CalcSolarReturnAtBirthplace(natalChart *natal.Chart, targetYear int) (*SolarReturn, error) {
	return CalcSolarReturn(natalChart, targetYear, natalChart.Lat, natalChart.Lon, natalChart.Alt)
}

// LunarReturn holds the moment the Moon returns to its natal position.
type LunarReturn struct {
	JD     float64 `json:"jd"`
	Month  int     `json:"month"`
	HourUT float64 `json:"hour_ut"`
}

// CalcLunarReturns finds all Lunar Returns in a given year.
// The Moon completes ~13 cycles per year (~27.3 days each).
func CalcLunarReturns(natalChart *natal.Chart, targetYear int) ([]LunarReturn, error) {
	natalMoonLon := natalChart.Planets["Luna"].Lon
	jdStart := ephemeris.JulDay(targetYear, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(targetYear+1, 1, 1, 0.0)

	var returns []LunarReturn
	jd := jdStart

	for jd < jdEnd {
		// Find when Moon crosses natal Moon longitude
		returnJD, err := moonCrossUT(natalMoonLon, jd, ephemeris.FlagSwieph|ephemeris.FlagSpeed)
		if err != nil || returnJD >= jdEnd {
			break
		}
		y, m, _, h := ephemeris.RevJul(returnJD)
		if y == targetYear {
			returns = append(returns, LunarReturn{
				JD:     returnJD,
				Month:  m,
				HourUT: h,
			})
		}
		// Advance ~25 days to find next return (Moon period ~27.3 days)
		jd = returnJD + 25.0
	}

	return returns, nil
}

// moonCrossUT finds when the Moon's longitude crosses targetLon after jdStart.
// Newton iteration similar to SolcrossUT but for the Moon (~13°/day).
func moonCrossUT(targetLon, jdStart float64, flags int) (float64, error) {
	const maxIter = 50
	const tolerance = 1e-6
	const maxStep = 15.0 // Moon moves ~13°/day, clamp to half a cycle

	jd := jdStart
	for iter := 0; iter < maxIter; iter++ {
		pos, err := ephemeris.CalcPlanet(jd, ephemeris.Moon, flags)
		if err != nil {
			return 0, err
		}
		diff := astromath.SignedDiff(pos.Lon, targetLon)
		if abs(diff) < tolerance {
			return jd, nil
		}
		speed := pos.Speed
		if speed == 0 {
			speed = 13.0 // Moon average ~13°/day
		}
		step := diff / speed
		if step > maxStep {
			step = maxStep
		} else if step < -maxStep {
			step = -maxStep
		}
		jd += step
	}
	// Non-convergence: skip this cycle
	return jdStart + 30, nil
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
