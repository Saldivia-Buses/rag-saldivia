package technique

import (
	"fmt"
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// PlanetaryCycle represents a planetary return or major cycle event.
type PlanetaryCycle struct {
	Planet      string  `json:"planet"`
	CycleType   string  `json:"cycle_type"`  // "return", "opposition", "square"
	NatalLon    float64 `json:"natal_lon"`
	TransitLon  float64 `json:"transit_lon"`
	Orb         float64 `json:"orb"`
	Month       int     `json:"month"`
	Description string  `json:"description"`
}

// cycleDefinition describes a planetary cycle to check.
type cycleDefinition struct {
	Name    string
	ID      int
	Period  float64 // approximate period in years
	Aspects []float64 // which aspects to check (0=return, 180=opposition, 90=square)
}

var cycleDefs = []cycleDefinition{
	{"Saturno", ephemeris.Saturn, 29.46, []float64{0, 90, 180}},    // Saturn return (~29y), squares (~7y), opposition (~14y)
	{"Júpiter", ephemeris.Jupiter, 11.86, []float64{0, 90, 180}},   // Jupiter return (~12y)
	{"Urano", ephemeris.Uranus, 84.01, []float64{0, 90, 180}},      // Uranus square (~21y), opposition (~42y)
	{"Neptuno", ephemeris.Neptune, 164.8, []float64{0, 90}},         // Neptune square (~41y)
	{"Plutón", ephemeris.Pluto, 248.1, []float64{0, 90}},            // Pluto square (~62y)
}

// cycleNames maps aspect angle to Spanish description.
var cycleNames = map[float64]string{
	0:   "retorno",
	90:  "cuadratura",
	180: "oposición",
}

const cycleOrb = 3.0

// CalcPlanetaryCycles checks for major planetary cycle events during a year.
// These are when slow transiting planets aspect their own natal position.
func CalcPlanetaryCycles(chart *natal.Chart, year int) []PlanetaryCycle {
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	var cycles []PlanetaryCycle

	for _, def := range cycleDefs {
		natPos, ok := chart.Planets[def.Name]
		if !ok {
			continue
		}
		natLon := natPos.Lon

		// Sample monthly (slow planets don't move much day-to-day)
		for month := 1; month <= 12; month++ {
			jd := jdStart + float64(month-1)*30.44 + 15 // mid-month
			trPos, err := ephemeris.CalcPlanet(jd, def.ID, flags)
			if err != nil {
				continue
			}

			for _, aspAngle := range def.Aspects {
				targetLon := astromath.Normalize360(natLon + aspAngle)
				orbVal := astromath.AngDiff(trPos.Lon, targetLon)

				if orbVal <= cycleOrb {
					cycleName := cycleNames[aspAngle]
					desc := fmt.Sprintf("%s %s a %s natal (%s)",
						def.Name, cycleName, def.Name,
						astromath.PosToStr(natLon))

					cycles = append(cycles, PlanetaryCycle{
						Planet:      def.Name,
						CycleType:   cycleName,
						NatalLon:    natLon,
						TransitLon:  trPos.Lon,
						Orb:         math.Round(orbVal*100) / 100,
						Month:       month,
						Description: desc,
					})
					break // only record once per aspect per year
				}
			}
		}
	}

	return cycles
}
