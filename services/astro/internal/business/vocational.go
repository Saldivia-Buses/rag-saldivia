package business

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// VocationalResult holds career/vocation analysis based on natal chart.
// 5 techniques: MC ruler, Almuten MC, Lot of Profession, H10 from Fortune, Sect light.
type VocationalResult struct {
	MCSign          string   `json:"mc_sign"`
	MCRuler         string   `json:"mc_ruler"`
	MCRulerSign     string   `json:"mc_ruler_sign"`
	MCRulerHouse    int      `json:"mc_ruler_house"`
	H6Sign          string   `json:"h6_sign"`           // daily work
	H10Planets      []string `json:"h10_planets"`        // planets in house 10
	Vocations       []string `json:"vocations"`          // suggested vocations
	KeyPlanets      []string `json:"key_planets"`        // most career-relevant planets
	AlmutenMC       string   `json:"almuten_mc"`         // most dignified planet at MC degree
	AlmutenMCScore  int      `json:"almuten_mc_score"`   // dignity score
	LotProfession   float64  `json:"lot_profession"`     // ecliptic longitude of Lot of Profession
	LotProfSign     string   `json:"lot_prof_sign"`      // sign of Lot of Profession
	H10FromFortune  string   `json:"h10_from_fortune"`   // sign 10 houses from Fortune (Valens)
	SectLightCareer string   `json:"sect_light_career"`  // sect light vocational indicator
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

	// ── Technique 2: Almuten of MC ──
	// Most dignified planet at the MC degree (al-Qabisi method)
	almuten := astromath.CalcAlmuten(chart.Planets, chart.ASC, chart.MC, chart.Diurnal)
	if almuten != nil {
		result.AlmutenMC = almuten.Winner
		result.AlmutenMCScore = almuten.Score
		if almuten.Winner != result.MCRuler {
			result.KeyPlanets = append(result.KeyPlanets, almuten.Winner)
		}
	}

	// ── Technique 3: Lot of Profession (Valens: ASC + Mercury - Sun) ──
	if merc, ok := chart.Planets["Mercurio"]; ok {
		if sol, ok := chart.Planets["Sol"]; ok {
			lotLon := astromath.Normalize360(chart.ASC + merc.Lon - sol.Lon)
			result.LotProfession = lotLon
			result.LotProfSign = astromath.SignName(lotLon)
		}
	}

	// ── Technique 4: House 10 from Fortune (Valens) ──
	// Count 10 signs from Lot of Fortune's sign
	moonLon := 0.0
	sunLon := 0.0
	if m, ok := chart.Planets["Luna"]; ok {
		moonLon = m.Lon
	}
	if s, ok := chart.Planets["Sol"]; ok {
		sunLon = s.Lon
	}
	fortuneLon := astromath.PartOfFortune(chart.ASC, moonLon, sunLon, chart.Diurnal)
	fortuneSignIdx := int(fortuneLon/30) % 12
	h10FromFortuneIdx := (fortuneSignIdx + 9) % 12 // 10th sign = +9 (0-indexed)
	result.H10FromFortune = astromath.Signs[h10FromFortuneIdx]

	// ── Technique 5: Sect light vocational indicator ──
	if chart.Diurnal {
		result.SectLightCareer = "Sol (carta diurna) — vocación visible, pública, de liderazgo"
	} else {
		result.SectLightCareer = "Luna (carta nocturna) — vocación adaptable, intuitiva, de servicio"
	}

	_ = ephemeris.Sun // keep import for consistency

	return result
}
