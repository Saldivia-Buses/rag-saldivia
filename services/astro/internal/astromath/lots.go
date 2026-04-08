package astromath

import "github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"

// LotDefinition describes a Hellenistic Lot formula.
// Day formula: ASC + A - B
// Night formula: ASC + B - A (unless NoReverse=true)
type LotDefinition struct {
	Key         string // internal key
	Name        string // Spanish display name
	Description string // what it signifies
	A           string // operand A (planet name, "Fortune", "Spirit", "SunExalt", "DSC", "Ruler7")
	B           string // operand B
	NoReverse   bool   // if true, same formula day and night (e.g., Basis)
}

// LotResult holds a calculated lot position.
type LotResult struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Lon         float64 `json:"lon"`
	Sign        string  `json:"sign"`
	Pos         string  `json:"pos"`
	House       int     `json:"house"`
	Lord        string  `json:"lord"` // traditional ruler of the sign
}

// LotDefinitions contains the 20 Hellenistic Lots.
// References: Paulus Alexandrinus, Vettius Valens.
var LotDefinitions = []LotDefinition{
	// Core lots (already computed in natal chart, included for completeness)
	{"fortune", "Fortuna", "suerte, cuerpo, circunstancias materiales", "Luna", "Sol", false},
	{"spirit", "Espíritu", "voluntad, alma, acción deliberada", "Sol", "Luna", false},

	// Planetary lots
	{"eros", "Eros", "amor, deseo, atracción", "Venus", "Spirit", false},
	{"necessity", "Necesidad", "restricción, obligación, karma", "Fortune", "Spirit", true}, // Basis: always same formula
	{"valor", "Valor", "coraje, audacia, acción", "Fortune", "Marte", false},
	{"victory", "Victoria", "éxito, logro, triunfo", "Fortune", "Júpiter", false},
	{"nemesis", "Némesis", "enemigos ocultos, justicia kármica", "Fortune", "Saturno", false},

	// Life lots
	{"marriage", "Matrimonio", "uniones, contratos legales", "Venus", "Saturno", false},
	{"children", "Hijos", "fertilidad, descendencia", "Júpiter", "Saturno", false},
	{"father", "Padre", "padre, figuras de autoridad", "Sol", "Saturno", false},
	{"mother", "Madre", "madre, figuras maternales", "Venus", "Luna", false},
	{"siblings", "Hermanos", "hermanos, relaciones fraternales", "Saturno", "Júpiter", false},
	{"death", "Muerte", "mortalidad, transformaciones radicales", "Luna", "Saturno", false}, // night: ASC + Sat - Moon

	// Business lots
	{"commerce", "Comercio", "negocios, transacciones", "Mercurio", "Fortune", false},
	{"prosperity", "Prosperidad", "abundancia, riqueza acumulada", "Fortune", "Spirit", false},
	{"acquisition", "Adquisición", "ganancias, ingresos", "Fortune", "Júpiter", false},
	{"travel", "Viajes", "desplazamientos, cambios de residencia", "Saturno", "Marte", false},
	{"illness", "Enfermedad", "vulnerabilidad física, dolencias", "Saturno", "Marte", false},

	// Profession/vocation
	{"profession", "Profesión", "vocación, carrera, estatus", "Luna", "Sol", false}, // same as fortune formula but used differently
	{"exaltation", "Exaltación", "honor, reconocimiento público", "Sol", "SunExalt", true}, // always: ASC + Sol - 19°Aries
}

// resolveOperand converts a lot operand key to an ecliptic longitude.
func resolveOperand(key string, planets map[string]*ephemeris.PlanetPos, asc, fortuneLon, spiritLon float64, cusps []float64) float64 {
	switch key {
	case "Fortune":
		return fortuneLon
	case "Spirit":
		return spiritLon
	case "SunExalt":
		return 19.0 // 19° Aries = 19.0 ecliptic
	case "DSC":
		return Normalize360(asc + 180)
	case "Ruler7":
		cusp7Lon := Normalize360(asc + 180)
		if len(cusps) >= 8 {
			cusp7Lon = cusps[7]
		}
		signIdx := SignIndex(cusp7Lon)
		ruler := DomicileOf[signIdx]
		if p, ok := planets[ruler]; ok {
			return p.Lon
		}
		return 0
	default:
		if p, ok := planets[key]; ok {
			return p.Lon
		}
		return 0
	}
}

// lotFormula computes ASC + A - B (normalized to 0-360).
func lotFormula(asc, a, b float64) float64 {
	return Normalize360(asc + a - b)
}

// CalcLot calculates a single lot by key.
func CalcLot(def *LotDefinition, planets map[string]*ephemeris.PlanetPos, asc float64, diurnal bool, cusps []float64) float64 {
	sunLon := 0.0
	moonLon := 0.0
	if p, ok := planets["Sol"]; ok {
		sunLon = p.Lon
	}
	if p, ok := planets["Luna"]; ok {
		moonLon = p.Lon
	}

	var fortuneLon, spiritLon float64
	if diurnal {
		fortuneLon = lotFormula(asc, moonLon, sunLon)
		spiritLon = lotFormula(asc, sunLon, moonLon)
	} else {
		fortuneLon = lotFormula(asc, sunLon, moonLon)
		spiritLon = lotFormula(asc, moonLon, sunLon)
	}

	aLon := resolveOperand(def.A, planets, asc, fortuneLon, spiritLon, cusps)
	bLon := resolveOperand(def.B, planets, asc, fortuneLon, spiritLon, cusps)

	if diurnal || def.NoReverse {
		return lotFormula(asc, aLon, bLon)
	}
	return lotFormula(asc, bLon, aLon)
}

// CalcAllLots calculates all hellenistic lots.
func CalcAllLots(planets map[string]*ephemeris.PlanetPos, asc float64, diurnal bool, cusps []float64) []LotResult {
	results := make([]LotResult, 0, len(LotDefinitions))
	for i := range LotDefinitions {
		def := &LotDefinitions[i]
		lon := CalcLot(def, planets, asc, diurnal, cusps)
		signIdx := SignIndex(lon)
		results = append(results, LotResult{
			Key:         def.Key,
			Name:        def.Name,
			Description: def.Description,
			Lon:         lon,
			Sign:        Signs[signIdx],
			Pos:         PosToStr(lon),
			House:       HouseForLon(lon, cusps),
			Lord:        DomicileOf[signIdx],
		})
	}
	return results
}

// CalcLotsActivations checks which lots are activated by Solar Arc in a given year.
// Returns lot results with activations noted.
type LotActivation struct {
	LotName   string  `json:"lot_name"`
	LotLon    float64 `json:"lot_lon"`
	SAPlanet  string  `json:"sa_planet"`
	SALon     float64 `json:"sa_lon"`
	Aspect    string  `json:"aspect"`
	Orb       float64 `json:"orb"`
}

func CalcLotsActivations(lots []LotResult, planets map[string]*ephemeris.PlanetPos, natalJD float64, targetYear int) []LotActivation {
	jdMid := float64(targetYear-2000)*365.25 + 2451545.0 + 182.5 // approx July 1
	years := (jdMid - natalJD) / 365.25
	arc := years * NaibodRate

	saNames := []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno"}
	orb := 2.0

	var activations []LotActivation
	for _, sa := range saNames {
		p, ok := planets[sa]
		if !ok {
			continue
		}
		saLon := Normalize360(p.Lon + arc)
		for _, lot := range lots {
			asp := FindAspect(saLon, lot.Lon, orb)
			if asp != nil {
				activations = append(activations, LotActivation{
					LotName:  lot.Name,
					LotLon:   lot.Lon,
					SAPlanet: sa,
					SALon:    saLon,
					Aspect:   asp.Name,
					Orb:      asp.Orb,
				})
			}
		}
	}
	return activations
}
