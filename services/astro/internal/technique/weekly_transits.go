package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// WeeklyTransit represents a transit event in the next 7 days.
type WeeklyTransit struct {
	Planet   string  `json:"planet"`
	Event    string  `json:"event"`      // "aspect", "ingress", "station"
	Detail   string  `json:"detail"`     // human-readable description
	JD       float64 `json:"jd"`
	Day      int     `json:"day"`
	Month    int     `json:"month"`
	Severity string  `json:"severity"`   // "alta", "media", "baja"
}

// allPlanetIDs includes all planets for weekly digest.
var allPlanetIDs = []struct {
	Name string
	ID   int
}{
	{"Sol", ephemeris.Sun}, {"Luna", ephemeris.Moon},
	{"Mercurio", ephemeris.Mercury}, {"Venus", ephemeris.Venus},
	{"Marte", ephemeris.Mars}, {"Júpiter", ephemeris.Jupiter},
	{"Saturno", ephemeris.Saturn},
}

// CalcWeeklyTransits produces a 7-day transit digest from a given start date.
// Checks: sign ingresses, aspects to natal points, stations.
func CalcWeeklyTransits(chart *natal.Chart, startJD float64) []WeeklyTransit {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	orb := 1.5 // tighter orb for weekly — precision matters

	// Natal points
	natalLons := make(map[string]float64)
	for name, pos := range chart.Planets {
		natalLons[name] = pos.Lon
	}
	natalLons["ASC"] = chart.ASC
	natalLons["MC"] = chart.MC

	var events []WeeklyTransit

	for _, p := range allPlanetIDs {
		var prevLon float64
		var prevSpeed float64
		var prevSign int

		for day := 0; day < 7; day++ {
			jd := startJD + float64(day)
			pos, err := ephemeris.CalcPlanet(jd, p.ID, flags)
			if err != nil {
				continue
			}

			_, m, d, _ := ephemeris.RevJul(jd)
			curSign := astromath.SignIndex(pos.Lon)

			if day > 0 {
				// Sign ingress
				if curSign != prevSign {
					events = append(events, WeeklyTransit{
						Planet:   p.Name,
						Event:    "ingress",
						Detail:   p.Name + " ingresa a " + astromath.Signs[curSign],
						JD:       jd,
						Day:      d,
						Month:    m,
						Severity: ingressSeverity(p.Name),
					})
				}

				// Station (direction change)
				if prevSpeed > 0 && pos.Speed < 0 {
					events = append(events, WeeklyTransit{
						Planet:   p.Name,
						Event:    "station",
						Detail:   p.Name + " estación retrógrada en " + astromath.PosToStr(pos.Lon),
						JD:       jd,
						Day:      d,
						Month:    m,
						Severity: "alta",
					})
				} else if prevSpeed < 0 && pos.Speed > 0 {
					events = append(events, WeeklyTransit{
						Planet:   p.Name,
						Event:    "station",
						Detail:   p.Name + " estación directa en " + astromath.PosToStr(pos.Lon),
						JD:       jd,
						Day:      d,
						Month:    m,
						Severity: "alta",
					})
				}
			}

			// Aspects to natal points (only for slow planets to avoid noise)
			if p.ID >= ephemeris.Mars { // Mars and slower
				for natName, natLon := range natalLons {
					asp := astromath.FindAspect(pos.Lon, natLon, orb)
					if asp != nil {
						events = append(events, WeeklyTransit{
							Planet:   p.Name,
							Event:    "aspect",
							Detail:   p.Name + " " + asp.Name + " " + natName + " (orbe " + fmtOrb(asp.Orb) + "°)",
							JD:       jd,
							Day:      d,
							Month:    m,
							Severity: aspectSeverity(asp.Name),
						})
					}
				}
			}

			prevLon = pos.Lon
			prevSpeed = pos.Speed
			prevSign = curSign
			_ = prevLon
		}
	}

	return events
}

func ingressSeverity(planet string) string {
	switch planet {
	case "Saturno", "Júpiter", "Marte":
		return "alta"
	case "Sol", "Venus", "Mercurio":
		return "media"
	default:
		return "baja"
	}
}

func aspectSeverity(aspect string) string {
	switch aspect {
	case "conjunction", "opposition", "square":
		return "alta"
	case "trine", "sextile":
		return "media"
	default:
		return "baja"
	}
}

func fmtOrb(orb float64) string {
	rounded := math.Round(orb*10) / 10
	whole := int(rounded)
	frac := int((rounded - float64(whole)) * 10)
	if frac == 0 {
		return string(rune('0'+whole)) + ".0"
	}
	return string(rune('0'+whole)) + "." + string(rune('0'+frac))
}
