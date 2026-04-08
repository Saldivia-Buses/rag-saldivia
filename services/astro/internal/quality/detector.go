package quality

import "strings"

// TechniqueStatus indicates how a technique was used in the LLM response.
type TechniqueStatus string

const (
	StatusUsed    TechniqueStatus = "USADO"   // data informed response AND mentioned
	StatusPartial TechniqueStatus = "PARCIAL" // in brief but not clearly mentioned
	StatusOmitted TechniqueStatus = "OMITIDO" // in brief, high priority, not mentioned
)

// TechniqueUsage tracks a single technique's presence in brief vs response.
type TechniqueUsage struct {
	ID     string          `json:"id"`
	Status TechniqueStatus `json:"status"`
	Weight float64         `json:"weight"` // domain priority weight (0-1)
}

// DetectionResult holds the full technique detection output.
type DetectionResult struct {
	Used    []string `json:"used"`
	Partial []string `json:"partial"`
	Omitted []string `json:"omitted"`
	Details []TechniqueUsage `json:"details"`
}

// DetectTechniques cross-references the intelligence brief against the LLM
// response to determine which techniques actually informed the output.
// Each technique section in the brief is checked for mention in the response.
func DetectTechniques(brief, response string) *DetectionResult {
	result := &DetectionResult{}
	if brief == "" || response == "" {
		return result
	}

	lowerBrief := strings.ToLower(brief)
	lowerResponse := strings.ToLower(response)

	for _, td := range techniqueDetectors {
		// Check if technique has data in the brief
		inBrief := false
		for _, marker := range td.briefMarkers {
			if strings.Contains(lowerBrief, marker) {
				inBrief = true
				break
			}
		}
		if !inBrief {
			continue // technique not in brief, skip
		}

		// Check if technique is mentioned in the response
		inResponse := false
		for _, marker := range td.responseMarkers {
			if strings.Contains(lowerResponse, marker) {
				inResponse = true
				break
			}
		}

		usage := TechniqueUsage{ID: td.id, Weight: td.weight}
		if inResponse {
			usage.Status = StatusUsed
			result.Used = append(result.Used, td.id)
		} else if td.weight >= 0.7 {
			usage.Status = StatusOmitted
			result.Omitted = append(result.Omitted, td.id)
		} else {
			usage.Status = StatusPartial
			result.Partial = append(result.Partial, td.id)
		}
		result.Details = append(result.Details, usage)
	}

	return result
}

// techniqueDetector defines how to detect a technique in brief and response.
type techniqueDetector struct {
	id              string
	weight          float64  // importance weight (0-1)
	briefMarkers    []string // lowercase strings that indicate technique is in brief
	responseMarkers []string // lowercase strings that indicate technique was used in response
}

// techniqueDetectors — 30+ technique detection patterns.
// briefMarkers: section headers or data patterns in the intelligence brief.
// responseMarkers: keywords the LLM typically uses when citing a technique.
var techniqueDetectors = []techniqueDetector{
	{
		id: "direcciones_primarias", weight: 0.9,
		briefMarkers:    []string{"## dp", "## direcciones primarias", "dirección primaria"},
		responseMarkers: []string{"dirección primaria", "direcciones primarias", "dp ", "primaria"},
	},
	{
		id: "arcos_solares", weight: 0.9,
		briefMarkers:    []string{"## sa", "## arcos solares", "sa "},
		responseMarkers: []string{"arco solar", "arcos solares", "solar arc"},
	},
	{
		id: "transitos", weight: 0.8,
		briefMarkers:    []string{"## tr", "## tránsitos", "tránsito"},
		responseMarkers: []string{"tránsito", "transito", "transita", "transitando"},
	},
	{
		id: "profecciones", weight: 0.8,
		briefMarkers:    []string{"## prof", "profección", "cronócrata"},
		responseMarkers: []string{"profección", "profeccion", "cronócrata", "casa activa"},
	},
	{
		id: "firdaria", weight: 0.7,
		briefMarkers:    []string{"## fird", "firdaria"},
		responseMarkers: []string{"firdaria", "período mayor", "sub-período"},
	},
	{
		id: "zodiacal_releasing", weight: 0.7,
		briefMarkers:    []string{"## zr", "zodiacal releasing", "zr fortuna"},
		responseMarkers: []string{"zodiacal releasing", "liberación zodiacal", "loosing of the bond", "zr "},
	},
	{
		id: "progresiones", weight: 0.6,
		briefMarkers:    []string{"## prog2", "## progresiones", "prog en"},
		responseMarkers: []string{"progresión", "progresion", "luna progresada", "progresado"},
	},
	{
		id: "eclipses", weight: 0.6,
		briefMarkers:    []string{"## ecl", "## eclipses", "eclipse"},
		responseMarkers: []string{"eclipse"},
	},
	{
		id: "revolucion_solar", weight: 0.5,
		briefMarkers:    []string{"## rs", "## revolución solar", "asc rs"},
		responseMarkers: []string{"revolución solar", "revolucion solar", "retorno solar", "carta solar"},
	},
	{
		id: "estaciones", weight: 0.5,
		briefMarkers:    []string{"## estaciones", "estación"},
		responseMarkers: []string{"estación", "estacion", "retrograda", "retrógrado"},
	},
	{
		id: "convergencia", weight: 1.0,
		briefMarkers:    []string{"## conv", "## matriz", "convergencia"},
		responseMarkers: []string{"convergen", "convergencia", "confluyen", "coinciden"},
	},
	{
		id: "lunaciones", weight: 0.4,
		briefMarkers:    []string{"lunación", "luna nueva", "luna llena"},
		responseMarkers: []string{"luna nueva", "luna llena", "lunación", "plenilunio"},
	},
	{
		id: "estrellas_fijas", weight: 0.5,
		briefMarkers:    []string{"## stars", "## estrellas", "estrella fija"},
		responseMarkers: []string{"estrella fija", "régulo", "spica", "algol", "aldebarán", "antares"},
	},
	{
		id: "lotes", weight: 0.6,
		briefMarkers:    []string{"## lotes", "lote de", "fortuna", "espíritu"},
		responseMarkers: []string{"lote de", "parte de", "fortuna", "espíritu"},
	},
	{
		id: "almuten", weight: 0.5,
		briefMarkers:    []string{"almutén", "almuten"},
		responseMarkers: []string{"almutén", "almuten", "planeta dominante"},
	},
	{
		id: "disposicion", weight: 0.4,
		briefMarkers:    []string{"dispositor", "disposición"},
		responseMarkers: []string{"dispositor", "cadena de disposición", "recepción mutua"},
	},
	{
		id: "secta", weight: 0.3,
		briefMarkers:    []string{"secta", "diurna", "nocturna"},
		responseMarkers: []string{"secta", "carta diurna", "carta nocturna", "luz de secta"},
	},
	{
		id: "hyleg", weight: 0.5,
		briefMarkers:    []string{"hyleg", "alcochoden"},
		responseMarkers: []string{"hyleg", "alcochoden", "longevidad"},
	},
	{
		id: "temperamento", weight: 0.4,
		briefMarkers:    []string{"temperamento", "colérico", "sanguíneo", "melancólico", "flemático"},
		responseMarkers: []string{"temperamento", "colérico", "sanguíneo", "melancólico", "flemático"},
	},
	{
		id: "melotesia", weight: 0.3,
		briefMarkers:    []string{"melotesia", "zona corporal"},
		responseMarkers: []string{"melotesia", "zona corporal", "cuerpo"},
	},
	{
		id: "sinastria", weight: 0.8,
		briefMarkers:    []string{"sinastría", "sinastria"},
		responseMarkers: []string{"sinastría", "sinastria", "compatibilidad", "aspectos cruzados"},
	},
	{
		id: "electiva", weight: 0.7,
		briefMarkers:    []string{"electiva", "ventana electiva"},
		responseMarkers: []string{"electiva", "mejor fecha", "fecha ideal", "ventana"},
	},
	{
		id: "horaria", weight: 0.7,
		briefMarkers:    []string{"horaria", "radicality"},
		responseMarkers: []string{"horaria", "carta horaria", "pregunta horaria"},
	},
	{
		id: "puntos_medios", weight: 0.3,
		briefMarkers:    []string{"## midpoints", "punto medio", "ebertin"},
		responseMarkers: []string{"punto medio", "midpoint"},
	},
	{
		id: "declinaciones", weight: 0.3,
		briefMarkers:    []string{"## decl", "paralelo", "declinación"},
		responseMarkers: []string{"paralelo", "contraparalelo", "declinación", "fuera de límite"},
	},
	{
		id: "transitos_rapidos", weight: 0.4,
		briefMarkers:    []string{"tránsito rápido", "fast transit", "marte transita"},
		responseMarkers: []string{"marte transita", "venus transita", "sol transita"},
	},
	{
		id: "deceniales", weight: 0.5,
		briefMarkers:    []string{"decenial", "valens"},
		responseMarkers: []string{"decenial", "valens"},
	},
	{
		id: "cadenas_activacion", weight: 0.6,
		briefMarkers:    []string{"## chains", "cadena de activación"},
		responseMarkers: []string{"cadena de activación", "triple activación", "cascada"},
	},
	{
		id: "ventanas_timing", weight: 0.7,
		briefMarkers:    []string{"## timing", "ventana de timing"},
		responseMarkers: []string{"ventana", "timing", "momento óptimo", "momento ideal"},
	},
}
