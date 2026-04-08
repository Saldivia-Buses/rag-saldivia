package business

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// CalcNegotiationTiming finds optimal days to negotiate with a counterparty.
// Based on transits to company H7 cusp (partnerships) and counterparty H2/H10.
func CalcNegotiationTiming(
	companyChart, counterpartyChart *natal.Chart,
	counterpartyName string,
	year, month int,
) []TimingWindow {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// Company's H7 cusp = partnership axis
	companyH7 := 0.0
	if len(companyChart.Cusps) > 7 {
		companyH7 = companyChart.Cusps[7]
	}

	// Counterparty's H2 cusp = their money
	counterH2 := 0.0
	if len(counterpartyChart.Cusps) > 2 {
		counterH2 = counterpartyChart.Cusps[2]
	}

	// Favorable planets for business: Jupiter, Venus
	// Challenging: Saturn, Mars
	type transitCheck struct {
		name   string
		id     int
		weight float64 // positive = good for negotiation
	}
	checks := []transitCheck{
		{"Júpiter", ephemeris.Jupiter, 15},
		{"Venus", ephemeris.Venus, 10},
		{"Saturno", ephemeris.Saturn, -12},
		{"Marte", ephemeris.Mars, -8},
	}

	jdStart := ephemeris.JulDay(year, month, 1, 12.0)
	var windows []TimingWindow

	// Score each day of the month
	daysInMonth := 30 // simplified
	for day := 1; day <= daysInMonth; day++ {
		jd := jdStart + float64(day-1)
		score := 50.0 // baseline
		var factors []string

		for _, chk := range checks {
			pos, err := ephemeris.CalcPlanet(jd, chk.id, flags)
			if err != nil {
				continue
			}

			// Check aspect to company H7
			asp := astromath.FindAspect(pos.Lon, companyH7, 3.0)
			if asp != nil {
				modifier := chk.weight
				if asp.Name == "square" || asp.Name == "opposition" {
					modifier = -abs(modifier)
				} else if asp.Name == "trine" || asp.Name == "sextile" {
					modifier = abs(modifier)
				}
				score += modifier
				factors = append(factors, fmt.Sprintf("%s %s H7 empresa", chk.name, asp.Name))
			}

			// Check aspect to counterparty H2
			asp2 := astromath.FindAspect(pos.Lon, counterH2, 3.0)
			if asp2 != nil {
				if asp2.Name == "trine" || asp2.Name == "sextile" || asp2.Name == "conjunction" {
					score += 5
					factors = append(factors, fmt.Sprintf("%s %s H2 contraparte", chk.name, asp2.Name))
				}
			}
		}

		// Mercury Rx penalty
		mercPos, err := ephemeris.CalcPlanet(jd, ephemeris.Mercury, flags)
		if err == nil && mercPos.Speed < 0 {
			score -= 10
			factors = append(factors, "Mercurio Rx")
		}

		if score >= 55 { // only include positive windows
			nature := "favorable"
			if score >= 75 {
				nature = "excelente"
			} else if score < 65 {
				nature = "aceptable"
			}
			windows = append(windows, TimingWindow{
				Counterparty: counterpartyName,
				Month:        month,
				DayStart:     day,
				DayEnd:       day,
				Score:        score,
				Nature:       nature,
				Factors:      factors,
			})
		}
	}

	// Consolidate consecutive days into ranges
	consolidated := consolidateWindows(windows)
	return consolidated
}

// consolidateWindows merges consecutive days into ranges.
func consolidateWindows(windows []TimingWindow) []TimingWindow {
	if len(windows) == 0 {
		return nil
	}
	var result []TimingWindow
	current := windows[0]
	for i := 1; i < len(windows); i++ {
		w := windows[i]
		if w.DayStart == current.DayEnd+1 && w.Counterparty == current.Counterparty {
			current.DayEnd = w.DayEnd
			if w.Score > current.Score {
				current.Score = w.Score
				current.Nature = w.Nature
			}
			current.Factors = append(current.Factors, w.Factors...)
		} else {
			result = append(result, current)
			current = w
		}
	}
	result = append(result, current)

	// Deduplicate factors
	for i := range result {
		result[i].Factors = dedupStrings(result[i].Factors)
	}
	return result
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func dedupStrings(ss []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
