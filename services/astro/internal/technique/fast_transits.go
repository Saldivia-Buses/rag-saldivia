package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// FastTransitActivation holds a fast planet transit hitting a natal point.
type FastTransitActivation struct {
	Transit    string  `json:"transit"`
	Aspect     string  `json:"aspect"`
	Natal      string  `json:"natal"`
	Orb        float64 `json:"orb"`
	TrPos      string  `json:"tr_pos"`
	Retrograde bool    `json:"retrograde"`
	Month      int     `json:"month"`
	Day        int     `json:"day"`
	Nature     string  `json:"nature"`
}

// fastPlanetIDs are the inner planets for fast transit analysis.
var fastPlanetIDs = []struct {
	Name string
	ID   int
}{
	{"Sol", ephemeris.Sun},
	{"Mercurio", ephemeris.Mercury},
	{"Venus", ephemeris.Venus},
	{"Marte", ephemeris.Mars},
}

const fastSampleDays = 1 // sample every day (vs 5 for slow transits)

// CalcFastTransits samples inner planets daily and finds aspects to natal points.
// Only records the closest pass per transit-natal pair.
func CalcFastTransits(chart *natal.Chart, year int) []FastTransitActivation {
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(year+1, 1, 1, 0.0)
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	orb := astromath.OrbDefaults.FastTransit

	// Natal points to check against
	type natalPoint struct {
		name string
		lon  float64
	}
	var natPoints []natalPoint

	// Only check against slow natal planets + angles (fast-to-fast is noise)
	slowNatal := []string{"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón"}
	for _, name := range slowNatal {
		if p, ok := chart.Planets[name]; ok {
			natPoints = append(natPoints, natalPoint{name, p.Lon})
		}
	}
	natPoints = append(natPoints, natalPoint{"ASC", chart.ASC})
	natPoints = append(natPoints, natalPoint{"MC", chart.MC})

	// Track best hit per transit-aspect-natal combo to avoid duplicates
	type hitKey struct {
		transit, aspect, natalPt string
	}
	bestHits := make(map[hitKey]*FastTransitActivation)

	for _, fp := range fastPlanetIDs {
		for jd := jdStart; jd < jdEnd; jd += fastSampleDays {
			pos, err := ephemeris.CalcPlanet(jd, fp.ID, flags)
			if err != nil {
				continue
			}

			for _, np := range natPoints {
				asp := astromath.FindAspect(pos.Lon, np.lon, orb)
				if asp == nil {
					continue
				}

				key := hitKey{fp.Name, asp.Name, np.name}
				existing, exists := bestHits[key]
				if !exists || asp.Orb < existing.Orb {
					_, m, d, _ := ephemeris.RevJul(jd)
					bestHits[key] = &FastTransitActivation{
						Transit:    fp.Name,
						Aspect:     asp.Name,
						Natal:      np.name,
						Orb:        math.Round(asp.Orb*100) / 100,
						TrPos:      astromath.PosToStr(pos.Lon),
						Retrograde: pos.Speed < 0,
						Month:      m,
						Day:        d,
						Nature:     fastNature(asp.Name),
					}
				}
			}
		}
	}

	results := make([]FastTransitActivation, 0, len(bestHits))
	for _, v := range bestHits {
		results = append(results, *v)
	}
	return results
}

func fastNature(aspect string) string {
	switch aspect {
	case "trine", "sextile":
		return "fácil"
	case "square", "opposition":
		return "tenso"
	default:
		return "neutral"
	}
}
