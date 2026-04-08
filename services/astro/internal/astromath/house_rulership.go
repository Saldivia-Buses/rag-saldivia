package astromath

// HouseRulerResult holds the ruler analysis for all 12 houses.
type HouseRulerResult struct {
	Houses []HouseInfo `json:"houses"`
}

// HouseInfo describes one house's ruler and theme.
type HouseInfo struct {
	House      int    `json:"house"`       // 1-12
	CuspSign   string `json:"cusp_sign"`   // sign on the cusp
	Ruler      string `json:"ruler"`       // traditional ruler of the cusp sign
	RulerSign  string `json:"ruler_sign"`  // sign the ruler is in
	RulerHouse int    `json:"ruler_house"` // house the ruler occupies
	Theme      string `json:"theme"`       // house theme (Spanish)
}

// HouseThemes maps house number (1-12) to its theme description.
var HouseThemes = map[int]string{
	1:  "identidad, cuerpo, apariencia, inicio",
	2:  "recursos propios, finanzas, valores materiales",
	3:  "comunicación, hermanos, contratos, entorno cercano",
	4:  "hogar, familia, raíces, padre (tradición helenística)",
	5:  "creatividad, hijos, placer, romance",
	6:  "salud, trabajo diario, servicio, empleados",
	7:  "relaciones, socios, contratos, enemigos abiertos",
	8:  "crisis, transformación, muerte, recursos ajenos, deuda",
	9:  "viajes largos, estudios superiores, filosofía, extranjero",
	10: "carrera, reputación, autoridad, madre (tradición helenística)",
	11: "amigos, grupos, proyectos, esperanzas",
	12: "retiro, espiritualidad, lo oculto, enemigos secretos, encierro",
}

// CalcHouseRulers computes the ruler chain for all 12 houses.
func CalcHouseRulers(cusps []float64, planets map[string]float64) *HouseRulerResult {
	result := &HouseRulerResult{
		Houses: make([]HouseInfo, 12),
	}

	for i := 0; i < 12; i++ {
		house := i + 1
		cuspLon := 0.0
		if len(cusps) > i+1 {
			cuspLon = cusps[i+1] // cusps[0] unused, [1]-[12] are houses
		}
		cuspSignIdx := SignIndex(cuspLon)
		ruler := DomicileOf[cuspSignIdx]

		rulerLon := 0.0
		if lon, ok := planets[ruler]; ok {
			rulerLon = lon
		}
		rulerSignIdx := SignIndex(rulerLon)
		rulerHouse := HouseForLon(rulerLon, cusps)

		result.Houses[i] = HouseInfo{
			House:      house,
			CuspSign:   Signs[cuspSignIdx],
			Ruler:      ruler,
			RulerSign:  Signs[rulerSignIdx],
			RulerHouse: rulerHouse,
			Theme:      HouseThemes[house],
		}
	}

	return result
}
