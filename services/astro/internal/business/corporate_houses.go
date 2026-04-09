package business

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// CorporateHouse represents one house in a corporate chart interpretation.
type CorporateHouse struct {
	House       int      `json:"house"`
	Sign        string   `json:"sign"`
	Ruler       string   `json:"ruler"`
	RulerSign   string   `json:"ruler_sign"`
	RulerHouse  int      `json:"ruler_house"`
	Planets     []string `json:"planets,omitempty"`
	Theme       string   `json:"theme"`
	Interpretation string `json:"interpretation"`
}

// CorporateHousesResult holds the full corporate house analysis.
type CorporateHousesResult struct {
	Houses []CorporateHouse `json:"houses"`
}

// Corporate house themes (Georgia Stathis tradition).
var corporateThemes = [12]string{
	"CEO / Imagen corporativa",                    // House 1
	"Cash flow / Activos líquidos",                // House 2
	"Comunicaciones / Logística corta",            // House 3
	"Sede / Infraestructura / Propiedad",          // House 4
	"Creatividad / Especulación / I+D",            // House 5
	"Empleados / Operaciones diarias",             // House 6
	"Clientes / Partners / Contratos",             // House 7
	"Deudas / Inversores / Impuestos",             // House 8
	"Expansión / Legal / Internacional",           // House 9
	"Reputación / CEO público / Gobierno",         // House 10
	"Revenue / Ingresos / Networking",             // House 11
	"Enemigos ocultos / Problemas hidden / Pérdidas", // House 12
}

// CalcCorporateHouses analyzes a company's natal chart using the Georgia Stathis
// corporate house system. Each house maps to a business function.
func CalcCorporateHouses(chart *natal.Chart) *CorporateHousesResult {
	result := &CorporateHousesResult{
		Houses: make([]CorporateHouse, 12),
	}

	for i := 0; i < 12; i++ {
		houseNum := i + 1
		cuspLon := chart.Cusps[houseNum]
		signIdx := int(cuspLon/30) % 12
		signName := astromath.Signs[signIdx]
		ruler := astromath.SignLord[signIdx]

		// Find ruler's sign and house
		rulerSign := ""
		rulerHouse := 0
		if pos, ok := chart.Planets[ruler]; ok {
			rulerSignIdx := int(pos.Lon/30) % 12
			rulerSign = astromath.Signs[rulerSignIdx]
			rulerHouse = astromath.HouseForLon(pos.Lon, chart.Cusps)
		}

		// Find planets in this house
		var planetsInHouse []string
		for name, pos := range chart.Planets {
			if astromath.HouseForLon(pos.Lon, chart.Cusps) == houseNum {
				planetsInHouse = append(planetsInHouse, name)
			}
		}

		// Generate interpretation
		interp := fmt.Sprintf("%s en %s (regente %s en casa %d/%s)",
			corporateThemes[i], signName, ruler, rulerHouse, rulerSign)
		if len(planetsInHouse) > 0 {
			interp += fmt.Sprintf(" — planetas presentes: %s", joinStrings(planetsInHouse))
		}

		result.Houses[i] = CorporateHouse{
			House:          houseNum,
			Sign:           signName,
			Ruler:          ruler,
			RulerSign:      rulerSign,
			RulerHouse:     rulerHouse,
			Planets:        planetsInHouse,
			Theme:          corporateThemes[i],
			Interpretation: interp,
		}
	}

	return result
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += ", " + s
	}
	return result
}
