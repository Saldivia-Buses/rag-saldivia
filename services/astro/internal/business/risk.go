package business

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// riskHouseMap maps risk categories to the natal houses that indicate them.
var riskHouseMap = map[string][]int{
	"financiero":   {2, 8},
	"operativo":    {6, 10},
	"legal":        {7, 9},
	"personal":     {6, 11},
	"reputacional": {10, 1},
}

// maleficIDs are the planets that generate risk when transiting business houses.
var maleficIDs = []struct {
	name   string
	id     int
	weight int // risk severity weight
}{
	{"Saturno", ephemeris.Saturn, 3},
	{"Marte", ephemeris.Mars, 2},
	{"Plutón", ephemeris.Pluto, 2},
	{"Urano", ephemeris.Uranus, 1},
}

// CalcRiskHeatmap produces a risk level (0-5) for each category × month.
func CalcRiskHeatmap(companyChart *natal.Chart, year int) []RiskCell {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	// Pre-compute cusp longitudes for each category
	categoryTargets := make(map[string][]float64)
	for category, houses := range riskHouseMap {
		var targets []float64
		for _, h := range houses {
			if len(companyChart.Cusps) > h {
				targets = append(targets, companyChart.Cusps[h])
				// Also add the ruler of the house cusp
				signIdx := astromath.SignIndex(companyChart.Cusps[h])
				ruler := astromath.DomicileOf[signIdx]
				if p, ok := companyChart.Planets[ruler]; ok {
					targets = append(targets, p.Lon)
				}
			}
		}
		categoryTargets[category] = targets
	}

	var cells []RiskCell

	for _, category := range RiskCategories {
		targets := categoryTargets[category]
		for m := 1; m <= 12; m++ {
			jd := ephemeris.JulDay(year, m, 15, 12.0)
			riskScore := 0
			var alert string

			for _, mal := range maleficIDs {
				pos, err := ephemeris.CalcPlanet(jd, mal.id, flags)
				if err != nil {
					continue
				}
				for _, target := range targets {
					asp := astromath.FindAspect(pos.Lon, target, 5.0)
					if asp == nil {
						continue
					}
					switch asp.Name {
					case "conjunction":
						riskScore += mal.weight * 2
						alert = fmt.Sprintf("%s conjunción punto clave", mal.name)
					case "square":
						riskScore += mal.weight * 2
						if alert == "" {
							alert = fmt.Sprintf("%s cuadratura punto clave", mal.name)
						}
					case "opposition":
						riskScore += mal.weight
						if alert == "" {
							alert = fmt.Sprintf("%s oposición punto clave", mal.name)
						}
					}
				}
			}

			// Clamp to 0-5
			level := riskScore
			if level > 5 {
				level = 5
			}

			cells = append(cells, RiskCell{
				Category: category,
				Month:    m,
				Level:    level,
				Alert:    alert,
			})
		}
	}

	return cells
}
