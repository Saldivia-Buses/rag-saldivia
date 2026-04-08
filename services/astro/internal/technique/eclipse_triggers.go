package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// EclipseTrigger records when a trigger planet crosses an eclipse degree.
type EclipseTrigger struct {
	TriggerPlanet string  `json:"trigger_planet"`
	TriggerLon    float64 `json:"trigger_lon"`
	EclipseType   string  `json:"eclipse_type"` // "solar" or "lunar"
	EclipseLon    float64 `json:"eclipse_lon"`
	EclipseDate   string  `json:"eclipse_date"` // "MM/YYYY"
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
	Month         int     `json:"month"`
	Day           int     `json:"day"`
	NatPoint      string  `json:"natal_point,omitempty"` // if the eclipse degree is also near a natal point
	NatOrb        float64 `json:"natal_orb,omitempty"`
}

// triggerPlanets are the planets that can activate eclipse degrees.
var triggerPlanets = []struct {
	Name string
	ID   int
}{
	{"Marte", ephemeris.Mars},
	{"Júpiter", ephemeris.Jupiter},
	{"Saturno", ephemeris.Saturn},
}

const triggerOrb = 2.0

// CalcEclipseTriggers finds when Mars/Jupiter/Saturn cross recent eclipse degrees.
// Checks eclipses from the current year AND previous year.
func CalcEclipseTriggers(chart *natal.Chart, year int) ([]EclipseTrigger, error) {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// Collect eclipses from current + previous year
	var eclipses []Eclipse
	for _, y := range []int{year - 1, year} {
		ecl, err := FindEclipses(y)
		if err != nil {
			continue
		}
		eclipses = append(eclipses, ecl...)
	}

	if len(eclipses) == 0 {
		return nil, nil
	}

	// Build natal points for cross-reference
	natalPts := make(map[string]float64)
	for name, pos := range chart.Planets {
		natalPts[name] = pos.Lon
	}
	natalPts["ASC"] = chart.ASC
	natalPts["MC"] = chart.MC

	// Scan trigger planets monthly through the year
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	var triggers []EclipseTrigger

	for _, tp := range triggerPlanets {
		for month := 1; month <= 12; month++ {
			jd := jdStart + float64(month-1)*30.44 + 15 // mid-month
			pos, err := ephemeris.CalcPlanet(jd, tp.ID, flags)
			if err != nil {
				continue
			}

			for _, ecl := range eclipses {
				asp := astromath.FindAspect(pos.Lon, ecl.Lon, triggerOrb)
				if asp == nil {
					continue
				}

				_, em, _, _ := ephemeris.RevJul(ecl.JD)
				_, _, d, _ := ephemeris.RevJul(jd)

				trigger := EclipseTrigger{
					TriggerPlanet: tp.Name,
					TriggerLon:    pos.Lon,
					EclipseType:   ecl.Type,
					EclipseLon:    ecl.Lon,
					EclipseDate:   intToStr(em) + "/" + intToStr(ecl.Month),
					Aspect:        asp.Name,
					Orb:           asp.Orb,
					Month:         month,
					Day:           d,
				}

				// Check if eclipse degree is near a natal point (amplified impact)
				for natName, natLon := range natalPts {
					natAsp := astromath.FindAspect(ecl.Lon, natLon, 3.0)
					if natAsp != nil && natAsp.Name == "conjunction" {
						trigger.NatPoint = natName
						trigger.NatOrb = natAsp.Orb
						break
					}
				}

				triggers = append(triggers, trigger)
			}
		}
	}

	return triggers, nil
}

func intToStr(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
