package intelligence

import "strings"

// Interpretation holds a pre-computed interpretation for a technique result.
type Interpretation struct {
	Technique   string `json:"technique"`
	Summary     string `json:"summary"`      // one-line verdict
	Detail      string `json:"detail"`       // 2-3 sentence explanation
	Significance string `json:"significance"` // "alta", "media", "baja"
}

// planetKeywords maps planets to thematic keywords for interpretation.
var planetKeywords = map[string]string{
	"Sol":      "identidad, voluntad, autoridad",
	"Luna":     "emociones, hogar, instinto, nutrición",
	"Mercurio": "comunicación, contratos, viajes cortos",
	"Venus":    "relaciones, dinero, placer, armonía",
	"Marte":    "acción, conflicto, energía, cirugía",
	"Júpiter":  "expansión, oportunidad, abundancia",
	"Saturno":  "restricción, estructura, madurez, tiempo",
	"Urano":    "cambio repentino, innovación, libertad",
	"Neptuno":  "ilusión, espiritualidad, confusión",
	"Plutón":   "transformación profunda, poder, crisis",
}

// houseKeywords maps houses to life areas.
var houseKeywords = map[int]string{
	1: "identidad", 2: "dinero", 3: "comunicación", 4: "hogar",
	5: "creatividad", 6: "salud/trabajo", 7: "relaciones", 8: "crisis",
	9: "viajes/filosofía", 10: "carrera", 11: "amigos/proyectos", 12: "espiritualidad",
}

// InterpretTransit produces a human-readable interpretation of a transit.
func InterpretTransit(transitPlanet, aspect, natalPoint string) *Interpretation {
	trKeywords := planetKeywords[transitPlanet]
	nature := "neutral"
	switch aspect {
	case "trine", "sextile":
		nature = "armónico"
	case "square", "opposition":
		nature = "tenso"
	case "conjunction":
		nature = "intenso"
	}

	summary := transitPlanet + " " + aspect + " " + natalPoint + " — " + nature
	detail := ""
	if trKeywords != "" {
		detail = "Temas activos: " + trKeywords + ". "
	}
	switch nature {
	case "armónico":
		detail += "Flujo favorable de energía. Oportunidad para avanzar."
	case "tenso":
		detail += "Tensión que requiere acción consciente. Potencial de crecimiento."
	case "intenso":
		detail += "Fusión de energías. Momento clave de activación."
	}

	return &Interpretation{
		Technique:    "tránsito",
		Summary:      summary,
		Detail:       detail,
		Significance: transitSignificance(transitPlanet),
	}
}

func transitSignificance(planet string) string {
	switch planet {
	case "Plutón", "Urano", "Neptuno":
		return "alta"
	case "Saturno", "Júpiter":
		return "media"
	default:
		return "baja"
	}
}

// InterpretHouse produces a brief interpretation of a planet in a house.
func InterpretHouse(planet string, house int) string {
	pk := planetKeywords[planet]
	hk := houseKeywords[house]
	if pk == "" || hk == "" {
		return ""
	}
	return strings.Title(pk) + " aplicado al área de " + hk
}
