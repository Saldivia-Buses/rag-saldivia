package business

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// VocationalResult holds career/vocation analysis based on natal chart.
type VocationalResult struct {
	MCSign      string   `json:"mc_sign"`
	MCRuler     string   `json:"mc_ruler"`
	MCRulerSign string   `json:"mc_ruler_sign"`
	MCRulerHouse int     `json:"mc_ruler_house"`
	H6Sign      string   `json:"h6_sign"`       // daily work
	H10Planets  []string `json:"h10_planets"`    // planets in house 10
	Vocations   []string `json:"vocations"`      // suggested vocations
	KeyPlanets  []string `json:"key_planets"`    // most career-relevant planets
}

// vocationalMap maps planet-sign combinations to career suggestions.
var vocationalMap = map[string][]string{
	"Sol":      {"liderazgo", "dirección", "gobierno", "entretenimiento"},
	"Luna":     {"cuidado", "alimentación", "hotelería", "enfermería"},
	"Mercurio": {"comunicación", "escritura", "comercio", "tecnología", "enseñanza"},
	"Venus":    {"arte", "diseño", "diplomacia", "moda", "hospitalidad"},
	"Marte":    {"deporte", "cirugía", "ingeniería", "militar", "emprendimiento"},
	"Júpiter":  {"derecho", "educación", "religión", "turismo", "expansión"},
	"Saturno":  {"administración", "construcción", "minería", "gobierno", "auditoría"},
}

// CalcVocational analyzes career indicators in the natal chart.
func CalcVocational(chart *natal.Chart) *VocationalResult {
	result := &VocationalResult{}

	// MC sign and ruler
	mcSignIdx := astromath.SignIndex(chart.MC)
	result.MCSign = astromath.Signs[mcSignIdx]
	result.MCRuler = astromath.DomicileOf[mcSignIdx]

	if p, ok := chart.Planets[result.MCRuler]; ok {
		result.MCRulerSign = astromath.SignName(p.Lon)
		result.MCRulerHouse = astromath.HouseForLon(p.Lon, chart.Cusps)
	}

	// H6 sign (daily work)
	if len(chart.Cusps) > 6 {
		result.H6Sign = astromath.SignName(chart.Cusps[6])
	}

	// Planets in H10
	for name, pos := range chart.Planets {
		house := astromath.HouseForLon(pos.Lon, chart.Cusps)
		if house == 10 {
			result.H10Planets = append(result.H10Planets, name)
		}
	}

	// Vocational suggestions based on MC ruler
	if v, ok := vocationalMap[result.MCRuler]; ok {
		result.Vocations = v
	}

	// Key career planets: MC ruler + planets in H10 + planets aspecting MC
	result.KeyPlanets = append(result.KeyPlanets, result.MCRuler)
	result.KeyPlanets = append(result.KeyPlanets, result.H10Planets...)

	_ = ephemeris.Sun // keep import for consistency

	return result
}
