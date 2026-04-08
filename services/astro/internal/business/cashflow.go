package business

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// CalcCashFlow forecasts monthly cash flow based on H2 (income) and H8 (expenses) transits.
// Returns 12 months of scored forecasts.
func CalcCashFlow(companyChart *natal.Chart, year int) []CashFlowMonth {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// H2 cusp = income, H8 cusp = expenses/debt
	h2Cusp := 0.0
	h8Cusp := 0.0
	if len(companyChart.Cusps) > 8 {
		h2Cusp = companyChart.Cusps[2]
		h8Cusp = companyChart.Cusps[8]
	}

	// H2 ruler and H8 ruler
	h2RulerName := astromath.DomicileOf[astromath.SignIndex(h2Cusp)]
	h8RulerName := astromath.DomicileOf[astromath.SignIndex(h8Cusp)]
	h2RulerLon := 0.0
	h8RulerLon := 0.0
	if p, ok := companyChart.Planets[h2RulerName]; ok {
		h2RulerLon = p.Lon
	}
	if p, ok := companyChart.Planets[h8RulerName]; ok {
		h8RulerLon = p.Lon
	}

	// Planets that indicate money flow
	type moneyPlanet struct {
		name   string
		id     int
		income float64 // positive = helps income
		expense float64 // positive = increases expenses
	}
	planets := []moneyPlanet{
		{"Júpiter", ephemeris.Jupiter, 15, -5},    // Jupiter: expands income, reduces pressure
		{"Venus", ephemeris.Venus, 10, -3},         // Venus: harmonious finances
		{"Saturno", ephemeris.Saturn, -10, 12},     // Saturn: restricts income, increases obligations
		{"Marte", ephemeris.Mars, -5, 8},           // Mars: conflict with money, sudden expenses
		{"Plutón", ephemeris.Pluto, -3, 5},         // Pluto: transformative pressure
	}

	months := make([]CashFlowMonth, 12)
	for m := 0; m < 12; m++ {
		jd := ephemeris.JulDay(year, m+1, 15, 12.0)
		inflow := 50.0  // baseline
		outflow := 50.0 // baseline
		var details []string

		for _, mp := range planets {
			pos, err := ephemeris.CalcPlanet(jd, mp.id, flags)
			if err != nil {
				continue
			}

			// Transit to H2 cusp or H2 ruler (income)
			for _, target := range []float64{h2Cusp, h2RulerLon} {
				asp := astromath.FindAspect(pos.Lon, target, 5.0)
				if asp != nil {
					switch asp.Name {
					case "trine", "sextile", "conjunction":
						inflow += mp.income
						if mp.income > 0 {
							details = append(details, fmt.Sprintf("%s favorable a H2", mp.name))
						}
					case "square", "opposition":
						inflow -= abs(mp.income) * 0.7
						if mp.income > 0 {
							details = append(details, fmt.Sprintf("%s tenso a H2", mp.name))
						}
					}
				}
			}

			// Transit to H8 cusp or H8 ruler (expenses)
			for _, target := range []float64{h8Cusp, h8RulerLon} {
				asp := astromath.FindAspect(pos.Lon, target, 5.0)
				if asp != nil {
					switch asp.Name {
					case "conjunction", "square", "opposition":
						outflow += mp.expense
						if mp.expense > 5 {
							details = append(details, fmt.Sprintf("%s activa H8", mp.name))
						}
					case "trine", "sextile":
						outflow -= mp.expense * 0.5
					}
				}
			}
		}

		// Clamp
		if inflow < 0 { inflow = 0 }
		if inflow > 100 { inflow = 100 }
		if outflow < 0 { outflow = 0 }
		if outflow > 100 { outflow = 100 }

		net := inflow - outflow
		rating := "neutro"
		if net > 15 {
			rating = "abundancia"
		} else if net < -15 {
			rating = "presión"
		}

		detailStr := ""
		if len(details) > 0 {
			for i, d := range details {
				if i > 0 { detailStr += "; " }
				detailStr += d
				if i >= 2 { break } // max 3
			}
		}

		months[m] = CashFlowMonth{
			Month:   m + 1,
			Label:   MonthLabels[m],
			Inflow:  inflow,
			Outflow: outflow,
			Net:     net,
			Rating:  rating,
			Details: detailStr,
		}
	}

	return months
}
