package astromath

import "github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"

// HylegResult holds the Hyleg/Alcochoden longevity analysis.
type HylegResult struct {
	Hyleg       string  `json:"hyleg"`        // planet selected as Hyleg (giver of life)
	HylegLon    float64 `json:"hyleg_lon"`
	HylegSign   string  `json:"hyleg_sign"`
	HylegHouse  int     `json:"hyleg_house"`
	Alcochoden  string  `json:"alcochoden"`   // planet with most dignity at Hyleg's position
	AlcYears    int     `json:"alc_years"`    // base longevity years from alcochoden
	Reasoning   string  `json:"reasoning"`    // explanation of selection
}

// hylegicPlaces are the houses where the Hyleg can be found (1, 7, 9, 10, 11).
var hylegicPlaces = map[int]bool{1: true, 7: true, 9: true, 10: true, 11: true}

// alcochodenYears maps planets to their "greater/middle/lesser years" for longevity.
// Using middle years (most common in practice).
var alcochodenYears = map[string]int{
	"Sol":      69,
	"Luna":     108, // greater years
	"Mercurio": 48,
	"Venus":    82,  // greater years
	"Marte":    66,
	"Júpiter":  79,
	"Saturno":  57,
}

// CalcHyleg determines the Hyleg (Apheta) and Alcochoden (Kadhkhudah).
// Simplified Ptolemaic method:
// 1. Check if Sun is in a hylegic place (diurnal chart) or Moon (nocturnal)
// 2. Fallback to the other luminary, then ASC
// 3. Alcochoden = planet with highest essential dignity at the Hyleg's position
func CalcHyleg(
	planets map[string]*ephemeris.PlanetPos,
	cusps []float64, asc float64,
	diurnal bool,
) *HylegResult {
	sunPos := planets["Sol"]
	moonPos := planets["Luna"]

	sunHouse := HouseForLon(sunPos.Lon, cusps)
	moonHouse := HouseForLon(moonPos.Lon, cusps)

	var hyleg string
	var hylegLon float64
	var reasoning string

	if diurnal {
		// Day chart: try Sun first
		if hylegicPlaces[sunHouse] {
			hyleg = "Sol"
			hylegLon = sunPos.Lon
			reasoning = "Carta diurna: Sol en casa hiléguica"
		} else if hylegicPlaces[moonHouse] {
			hyleg = "Luna"
			hylegLon = moonPos.Lon
			reasoning = "Carta diurna: Sol fuera de lugar hiléguico, Luna como respaldo"
		}
	} else {
		// Night chart: try Moon first
		if hylegicPlaces[moonHouse] {
			hyleg = "Luna"
			hylegLon = moonPos.Lon
			reasoning = "Carta nocturna: Luna en casa hiléguica"
		} else if hylegicPlaces[sunHouse] {
			hyleg = "Sol"
			hylegLon = sunPos.Lon
			reasoning = "Carta nocturna: Luna fuera de lugar hiléguico, Sol como respaldo"
		}
	}

	// Fallback to ASC
	if hyleg == "" {
		hyleg = "ASC"
		hylegLon = asc
		reasoning = "Ninguna luminaria en casa hiléguica: ASC como Hyleg"
	}

	// Find Alcochoden: planet with highest essential dignity at Hyleg's position
	alcochoden := ""
	bestScore := 0
	for _, planet := range ClassicalPlanets {
		score, _ := DignityScoreAt(planet, hylegLon, diurnal)
		if score > bestScore {
			bestScore = score
			alcochoden = planet
		}
	}

	years := 0
	if alcochoden != "" {
		years = alcochodenYears[alcochoden]
	}

	return &HylegResult{
		Hyleg:      hyleg,
		HylegLon:   hylegLon,
		HylegSign:  Signs[SignIndex(hylegLon)],
		HylegHouse: HouseForLon(hylegLon, cusps),
		Alcochoden: alcochoden,
		AlcYears:   years,
		Reasoning:  reasoning,
	}
}
