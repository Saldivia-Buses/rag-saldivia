package technique

import "time"

// TimeLordSummary provides a unified view of all active time lords.
// Combines profections, firdaria, ZR, and decennials into one snapshot.
type TimeLordSummary struct {
	Year              int               `json:"year"`
	Age               float64           `json:"age"`
	Lords             []ActiveTimeLord  `json:"lords"`
	DominantPlanet    string            `json:"dominant_planet"`    // most frequent across techniques
	DominantCount     int               `json:"dominant_count"`
	Agreement         string            `json:"agreement"`          // "alta", "media", "baja"
}

// ActiveTimeLord is one active time lord from a specific technique.
type ActiveTimeLord struct {
	Technique string `json:"technique"` // "profecciones", "firdaria", "ZR_fortuna", "ZR_espiritu", "deceniales"
	Lord      string `json:"lord"`
	Level     string `json:"level"`     // "mayor", "sub", "L1", "L2"
}

// CalcTimeLords produces a unified time lord summary from pre-computed techniques.
func CalcTimeLords(
	profection *Profection,
	firdaria *Firdaria,
	zrFortune *ZRResult,
	zrSpirit *ZRResult,
	decennials *DecennialResult,
	birthDate time.Time,
	year int,
) *TimeLordSummary {
	midYear := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)
	age := midYear.Sub(birthDate).Hours() / (24 * 365.25)

	summary := &TimeLordSummary{Year: year, Age: age}
	counts := make(map[string]int)

	if profection != nil {
		summary.Lords = append(summary.Lords, ActiveTimeLord{
			Technique: "profecciones", Lord: profection.Lord, Level: "mayor",
		})
		counts[profection.Lord]++
	}

	if firdaria != nil {
		summary.Lords = append(summary.Lords, ActiveTimeLord{
			Technique: "firdaria", Lord: firdaria.MajorLord, Level: "mayor",
		})
		counts[firdaria.MajorLord]++
		if firdaria.SubLord != "" {
			summary.Lords = append(summary.Lords, ActiveTimeLord{
				Technique: "firdaria", Lord: firdaria.SubLord, Level: "sub",
			})
			counts[firdaria.SubLord]++
		}
	}

	if zrFortune != nil && zrFortune.Level1 != nil {
		summary.Lords = append(summary.Lords, ActiveTimeLord{
			Technique: "ZR_fortuna", Lord: zrFortune.Level1.Lord, Level: "L1",
		})
		counts[zrFortune.Level1.Lord]++
		if zrFortune.Level2 != nil {
			summary.Lords = append(summary.Lords, ActiveTimeLord{
				Technique: "ZR_fortuna", Lord: zrFortune.Level2.Lord, Level: "L2",
			})
			counts[zrFortune.Level2.Lord]++
		}
	}

	if zrSpirit != nil && zrSpirit.Level1 != nil {
		summary.Lords = append(summary.Lords, ActiveTimeLord{
			Technique: "ZR_espiritu", Lord: zrSpirit.Level1.Lord, Level: "L1",
		})
		counts[zrSpirit.Level1.Lord]++
	}

	if decennials != nil {
		summary.Lords = append(summary.Lords, ActiveTimeLord{
			Technique: "deceniales", Lord: decennials.MajorPlanet, Level: "mayor",
		})
		counts[decennials.MajorPlanet]++
		if decennials.SubPlanet != "" {
			summary.Lords = append(summary.Lords, ActiveTimeLord{
				Technique: "deceniales", Lord: decennials.SubPlanet, Level: "sub",
			})
			counts[decennials.SubPlanet]++
		}
	}

	// Find dominant planet
	bestCount := 0
	for planet, count := range counts {
		if count > bestCount {
			bestCount = count
			summary.DominantPlanet = planet
			summary.DominantCount = count
		}
	}

	// Agreement level
	totalTechs := 0
	if profection != nil { totalTechs++ }
	if firdaria != nil { totalTechs++ }
	if zrFortune != nil { totalTechs++ }
	if decennials != nil { totalTechs++ }

	if totalTechs > 0 {
		ratio := float64(bestCount) / float64(totalTechs)
		switch {
		case ratio >= 0.75:
			summary.Agreement = "alta"
		case ratio >= 0.5:
			summary.Agreement = "media"
		default:
			summary.Agreement = "baja"
		}
	}

	return summary
}
