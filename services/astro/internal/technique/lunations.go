package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// Lunation represents a New Moon or Full Moon event.
type Lunation struct {
	Type      string  `json:"type"`       // "nueva" (new) or "llena" (full)
	JD        float64 `json:"jd"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Lon       float64 `json:"lon"`        // longitude of the lunation
	Sign      string  `json:"sign"`
	NatPoint  string  `json:"natal_point,omitempty"` // nearest natal point within orb
	NatOrb    float64 `json:"natal_orb,omitempty"`
	NatAspect string  `json:"natal_aspect,omitempty"`
	House     int     `json:"house"`      // natal house where the lunation falls
}

// VOCPeriod represents a Void-of-Course Moon period.
type VOCPeriod struct {
	StartJD    float64 `json:"start_jd"`
	EndJD      float64 `json:"end_jd"`
	StartMonth int     `json:"start_month"`
	StartDay   int     `json:"start_day"`
	EndMonth   int     `json:"end_month"`
	EndDay     int     `json:"end_day"`
	DurationH  float64 `json:"duration_hours"`
	FromSign   string  `json:"from_sign"`
	ToSign     string  `json:"to_sign"`
}

// LunationResult holds all lunations and VOC periods for a year.
type LunationResult struct {
	Lunations []Lunation  `json:"lunations"`
	VOC       []VOCPeriod `json:"voc_periods,omitempty"`
}

const lunationOrb = 3.0

// CalcLunations finds all New and Full Moons in a year and checks natal aspects.
func CalcLunations(chart *natal.Chart, year int) (*LunationResult, error) {
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(year+1, 1, 1, 0.0)
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// Build natal points for aspect checking
	natalPoints := make(map[string]float64)
	for name, pos := range chart.Planets {
		natalPoints[name] = pos.Lon
	}
	natalPoints["ASC"] = chart.ASC
	natalPoints["MC"] = chart.MC

	var lunations []Lunation

	// Scan for lunations by checking Sun-Moon angular separation daily
	prevDiff := 0.0
	prevJD := jdStart

	for jd := jdStart; jd < jdEnd; jd += 1.0 {
		sunPos, err := ephemeris.CalcPlanet(jd, ephemeris.Sun, flags)
		if err != nil {
			continue
		}
		moonPos, err := ephemeris.CalcPlanet(jd, ephemeris.Moon, flags)
		if err != nil {
			continue
		}

		// Signed difference Moon - Sun (0-360)
		diff := math.Mod(moonPos.Lon-sunPos.Lon+360, 360)

		if jd > jdStart {
			// New Moon: diff crosses 0° (from ~350 to ~10)
			if prevDiff > 300 && diff < 60 {
				exactJD := interpolateExact(prevJD, jd, prevDiff, diff, 0)
				lunations = append(lunations, buildLunation("nueva", exactJD, sunPos.Lon, chart.Cusps, natalPoints))
			}

			// Full Moon: diff crosses 180° (from ~170 to ~190)
			if prevDiff < 180 && diff >= 180 && prevDiff > 100 {
				exactJD := interpolateExact(prevJD, jd, prevDiff, diff, 180)
				lunations = append(lunations, buildLunation("llena", exactJD, moonPos.Lon, chart.Cusps, natalPoints))
			}
		}

		prevDiff = diff
		prevJD = jd
	}

	return &LunationResult{
		Lunations: lunations,
	}, nil
}

// refineLunationJD uses Newton iteration to find the exact JD when Sun-Moon
// elongation equals the target (0° for New Moon, 180° for Full Moon).
// Starts from a rough estimate between jd1 and jd2.
func interpolateExact(jd1, jd2, diff1, diff2, target float64) float64 {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// Linear interpolation for initial guess
	if target == 0 && diff1 > 300 { diff1 -= 360 }
	frac := 0.5
	if diff2-diff1 != 0 {
		frac = (target - diff1) / (diff2 - diff1)
		if frac < 0 { frac = 0 }
		if frac > 1 { frac = 1 }
	}
	jd := jd1 + frac*(jd2-jd1)

	// Newton refinement: converge to exact elongation = target
	for iter := 0; iter < 20; iter++ {
		sunPos, err := ephemeris.CalcPlanet(jd, ephemeris.Sun, flags)
		if err != nil { return jd }
		moonPos, err := ephemeris.CalcPlanet(jd, ephemeris.Moon, flags)
		if err != nil { return jd }

		elongation := math.Mod(moonPos.Lon-sunPos.Lon+360, 360)
		diff := elongation - target

		// Normalize diff to [-180, 180]
		if diff > 180 { diff -= 360 }
		if diff < -180 { diff += 360 }

		if math.Abs(diff) < 0.001 { // < 0.001° = < 4 seconds of arc
			return jd
		}

		// Relative speed: Moon speed - Sun speed (~12°/day)
		relSpeed := moonPos.Speed - sunPos.Speed
		if relSpeed == 0 { relSpeed = 12.0 }

		jd -= diff / relSpeed
	}
	return jd
}

// buildLunation creates a Lunation with natal aspect checking.
func buildLunation(typ string, jd, lon float64, cusps []float64, natalPoints map[string]float64) Lunation {
	_, m, d, _ := ephemeris.RevJul(jd)
	l := Lunation{
		Type:  typ,
		JD:    jd,
		Month: m,
		Day:   d,
		Lon:   lon,
		Sign:  astromath.SignName(lon),
		House: astromath.HouseForLon(lon, cusps),
	}

	// Find closest natal point within orb
	bestOrb := lunationOrb + 1
	for name, natLon := range natalPoints {
		asp := astromath.FindAspect(lon, natLon, lunationOrb)
		if asp != nil && asp.Orb < bestOrb {
			bestOrb = asp.Orb
			l.NatPoint = name
			l.NatOrb = asp.Orb
			l.NatAspect = asp.Name
		}
	}

	return l
}
