package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// PlanetaryReturn holds a planetary return chart (when a planet returns to natal position).
type PlanetaryReturn struct {
	Planet   string  `json:"planet"`
	ReturnJD float64 `json:"return_jd"`
	Year     int     `json:"year"`
	Month    int     `json:"month"`
	Day      int     `json:"day"`
	ReturnLon float64 `json:"return_lon"`
	Sign     string  `json:"sign"`
}

// returnPlanets are planets that have meaningful returns within a human life.
var returnPlanets = []struct {
	name   string
	id     int
	period float64 // approximate period in years
}{
	{"Júpiter", ephemeris.Jupiter, 11.86},
	{"Saturno", ephemeris.Saturn, 29.46},
	{"Marte", ephemeris.Mars, 1.88},
}

// CalcPlanetaryReturns finds when slow planets return to their natal longitude.
// Searches the target year for exact returns using Newton iteration.
func CalcPlanetaryReturns(chart *natal.Chart, year int) []PlanetaryReturn {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	var returns []PlanetaryReturn

	for _, rp := range returnPlanets {
		natPos, ok := chart.Planets[rp.name]
		if !ok {
			continue
		}
		targetLon := natPos.Lon

		// Search the year in monthly steps for proximity
		jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
		for month := 0; month < 12; month++ {
			jd := jdStart + float64(month)*30.44
			pos, err := ephemeris.CalcPlanet(jd, rp.id, flags)
			if err != nil {
				continue
			}
			orb := astromath.AngDiff(pos.Lon, targetLon)
			if orb > 10 {
				continue
			}

			// Newton iteration to find exact return
			exactJD, err := findPlanetCross(rp.id, targetLon, jd-15, flags)
			if err != nil {
				continue
			}
			y, m, d, _ := ephemeris.RevJul(exactJD)
			if y != year {
				continue
			}
			returns = append(returns, PlanetaryReturn{
				Planet: rp.name, ReturnJD: exactJD,
				Year: y, Month: m, Day: d,
				ReturnLon: targetLon, Sign: astromath.SignName(targetLon),
			})
		}
	}

	return returns
}

// findPlanetCross finds when a planet's longitude crosses targetLon after jdStart.
func findPlanetCross(planetID int, targetLon, jdStart float64, flags int) (float64, error) {
	const maxIter = 30
	const tolerance = 1e-5
	jd := jdStart

	for iter := 0; iter < maxIter; iter++ {
		pos, err := ephemeris.CalcPlanet(jd, planetID, flags)
		if err != nil {
			return 0, err
		}
		diff := astromath.SignedDiff(pos.Lon, targetLon)
		if diff < 0 {
			diff = -diff
		}
		if diff < tolerance {
			return jd, nil
		}
		speed := pos.Speed
		if speed == 0 {
			speed = 0.1
		}
		step := astromath.SignedDiff(pos.Lon, targetLon) / speed
		if step > 30 {
			step = 30
		} else if step < -30 {
			step = -30
		}
		jd += step
	}
	return jd, nil
}
