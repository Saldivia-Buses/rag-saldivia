package astromath

import (
	"fmt"
	"math"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

// AspectPattern represents a detected chart pattern.
type AspectPattern struct {
	Type    string   `json:"type"`    // "T-Square", "Grand Trine", "Grand Cross", "Yod", "Stellium"
	Planets []string `json:"planets"`
	Sign    string   `json:"sign,omitempty"` // for stelliums
	Element string   `json:"element,omitempty"` // for grand trines
}

// DetectAspectPatterns finds T-Squares, Grand Trines, Yods, Grand Crosses, and Stelliums.
func DetectAspectPatterns(planets map[string]*ephemeris.PlanetPos) []AspectPattern {
	var patterns []AspectPattern
	names := []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno", "Urano", "Neptuno", "Plutón"}

	var lons []struct{ name string; lon float64 }
	for _, n := range names {
		if p, ok := planets[n]; ok {
			lons = append(lons, struct{ name string; lon float64 }{n, p.Lon})
		}
	}

	// Stelliums: 3+ planets within 10° in same sign
	signGroups := make(map[int][]string)
	for _, l := range lons {
		signGroups[SignIndex(l.lon)] = append(signGroups[SignIndex(l.lon)], l.name)
	}
	for si, group := range signGroups {
		if len(group) >= 3 {
			patterns = append(patterns, AspectPattern{
				Type: "Stellium", Planets: group, Sign: Signs[si],
			})
		}
	}

	// T-Square: 2 planets in opposition + 1 squaring both
	for i := 0; i < len(lons); i++ {
		for j := i + 1; j < len(lons); j++ {
			if isAspect(lons[i].lon, lons[j].lon, 180, 8) {
				for k := 0; k < len(lons); k++ {
					if k == i || k == j { continue }
					if isAspect(lons[k].lon, lons[i].lon, 90, 8) && isAspect(lons[k].lon, lons[j].lon, 90, 8) {
						patterns = append(patterns, AspectPattern{
							Type: "T-Square", Planets: []string{lons[i].name, lons[j].name, lons[k].name},
						})
					}
				}
			}
		}
	}

	// Grand Trine: 3 planets each 120° apart
	for i := 0; i < len(lons); i++ {
		for j := i + 1; j < len(lons); j++ {
			if !isAspect(lons[i].lon, lons[j].lon, 120, 8) { continue }
			for k := j + 1; k < len(lons); k++ {
				if isAspect(lons[j].lon, lons[k].lon, 120, 8) && isAspect(lons[k].lon, lons[i].lon, 120, 8) {
					elem := signElement[SignIndex(lons[i].lon)]
					patterns = append(patterns, AspectPattern{
						Type: "Grand Trine", Planets: []string{lons[i].name, lons[j].name, lons[k].name},
						Element: elem,
					})
				}
			}
		}
	}

	// Grand Cross: 4 planets at 90° intervals
	for i := 0; i < len(lons); i++ {
		for j := i + 1; j < len(lons); j++ {
			if !isAspect(lons[i].lon, lons[j].lon, 90, 8) { continue }
			for k := j + 1; k < len(lons); k++ {
				if !isAspect(lons[j].lon, lons[k].lon, 90, 8) { continue }
				for l := k + 1; l < len(lons); l++ {
					if isAspect(lons[k].lon, lons[l].lon, 90, 8) && isAspect(lons[l].lon, lons[i].lon, 90, 8) {
						patterns = append(patterns, AspectPattern{
							Type: "Grand Cross",
							Planets: []string{lons[i].name, lons[j].name, lons[k].name, lons[l].name},
						})
					}
				}
			}
		}
	}

	// Yod: 2 planets sextile (60°) + 1 quincunx (150°) to both
	for i := 0; i < len(lons); i++ {
		for j := i + 1; j < len(lons); j++ {
			if !isAspect(lons[i].lon, lons[j].lon, 60, 5) { continue }
			for k := 0; k < len(lons); k++ {
				if k == i || k == j { continue }
				if isAspect(lons[k].lon, lons[i].lon, 150, 3) && isAspect(lons[k].lon, lons[j].lon, 150, 3) {
					patterns = append(patterns, AspectPattern{
						Type: "Yod", Planets: []string{lons[i].name, lons[j].name, lons[k].name + " (apex)"},
					})
				}
			}
		}
	}

	return patterns
}

func isAspect(lon1, lon2, angle, orb float64) bool {
	return math.Abs(AngDiff(lon1, lon2)-angle) <= orb
}

// ChartShape classifies the chart using Jones patterns.
type ChartShape struct {
	Pattern     string `json:"pattern"` // "Bowl", "Bucket", "Locomotive", "Splay", "Bundle", "Splash", "Seesaw"
	Description string `json:"description"`
}

func DetectChartShape(planets map[string]*ephemeris.PlanetPos) *ChartShape {
	var lons []float64
	for _, n := range []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno", "Urano", "Neptuno", "Plutón"} {
		if p, ok := planets[n]; ok {
			lons = append(lons, p.Lon)
		}
	}
	if len(lons) < 5 { return &ChartShape{Pattern: "unknown"} }

	// Sort longitudes
	for i := 0; i < len(lons); i++ {
		for j := i + 1; j < len(lons); j++ {
			if lons[j] < lons[i] { lons[i], lons[j] = lons[j], lons[i] }
		}
	}

	// Find largest gap between consecutive planets
	maxGap := 0.0
	for i := 0; i < len(lons); i++ {
		next := lons[(i+1)%len(lons)]
		gap := math.Mod(next-lons[i]+360, 360)
		if gap > maxGap { maxGap = gap }
	}

	// Occupied span
	span := 360 - maxGap

	switch {
	case span <= 120:
		return &ChartShape{"Bundle", "Planetas concentrados en 120° — enfoque intenso en pocas áreas"}
	case span <= 180:
		return &ChartShape{"Bowl", "Planetas en un hemisferio — autocontenido, busca completar la mitad vacía"}
	case maxGap >= 60 && maxGap <= 90 && span > 270:
		return &ChartShape{"Locomotive", "Un hueco de 60-90° — locomotora, empuje constante"}
	case span > 300:
		return &ChartShape{"Splash", "Planetas distribuidos en todo el zodíaco — versatilidad, dispersión"}
	default:
		return &ChartShape{"Splay", "Distribución irregular — individualismo, talentos múltiples"}
	}
}

// HemisphericDistribution analyzes planet distribution by hemisphere and quadrant.
type HemisphericDist struct {
	Eastern  int    `json:"eastern"`  // houses 10-3 (self-directed)
	Western  int    `json:"western"`  // houses 4-9 (other-directed)
	Northern int    `json:"northern"` // houses 1-6 (private)
	Southern int    `json:"southern"` // houses 7-12 (public)
	Dominant string `json:"dominant"` // which hemisphere dominates
}

func CalcHemisphericDist(planets map[string]*ephemeris.PlanetPos, cusps []float64) *HemisphericDist {
	d := &HemisphericDist{}
	for _, p := range planets {
		house := HouseForLon(p.Lon, cusps)
		if house >= 10 || house <= 3 { d.Eastern++ } else { d.Western++ }
		if house >= 1 && house <= 6 { d.Northern++ } else { d.Southern++ }
	}
	switch {
	case d.Eastern > d.Western+3: d.Dominant = "oriental — autodeterminado"
	case d.Western > d.Eastern+3: d.Dominant = "occidental — orientado a otros"
	case d.Southern > d.Northern+3: d.Dominant = "sur — vida pública prominente"
	case d.Northern > d.Southern+3: d.Dominant = "norte — vida interior rica"
	default: d.Dominant = "equilibrado"
	}
	return d
}

// PlanetaryAge returns the current planetary age period (Valens/Ptolemy).
// Moon=0-4, Mercury=4-14, Venus=14-22, Sun=22-41, Mars=41-56, Jupiter=56-68, Saturn=68+
type PlanetaryAgePeriod struct {
	Planet string  `json:"planet"`
	AgeFrom int    `json:"age_from"`
	AgeTo   int    `json:"age_to"`
	Theme  string  `json:"theme"`
}

var planetaryAges = []PlanetaryAgePeriod{
	{"Luna", 0, 4, "infancia, dependencia, nutrición"},
	{"Mercurio", 4, 14, "aprendizaje, lenguaje, curiosidad"},
	{"Venus", 14, 22, "relaciones, deseo, estética"},
	{"Sol", 22, 41, "identidad, carrera, voluntad"},
	{"Marte", 41, 56, "ambición, conflicto, energía"},
	{"Júpiter", 56, 68, "sabiduría, expansión, legado"},
	{"Saturno", 68, 120, "madurez, restricción, trascendencia"},
}

func CurrentPlanetaryAge(age float64) *PlanetaryAgePeriod {
	intAge := int(age)
	for _, pa := range planetaryAges {
		if intAge >= pa.AgeFrom && intAge < pa.AgeTo {
			return &pa
		}
	}
	return &planetaryAges[len(planetaryAges)-1]
}

// FullDignityTable builds a complete essential + accidental dignity table.
type DignityEntry struct {
	Planet     string `json:"planet"`
	Sign       string `json:"sign"`
	Domicile   bool   `json:"domicile"`
	Exaltation bool   `json:"exaltation"`
	Triplicity bool   `json:"triplicity"`
	Term       bool   `json:"term"`
	Face       bool   `json:"face"`
	Detriment  bool   `json:"detriment"`
	Fall       bool   `json:"fall"`
	Score      int    `json:"score"` // net dignity score
	Retrograde bool   `json:"retrograde"`
	Speed      string `json:"speed"` // "fast", "average", "slow", "stationary"
}

func BuildFullDignityTable(planets map[string]*ephemeris.PlanetPos, diurnal bool) []DignityEntry {
	var entries []DignityEntry
	for _, name := range ClassicalPlanets {
		p, ok := planets[string(name)]
		if !ok { continue }
		signIdx := SignIndex(p.Lon)
		score, _ := DignityScoreAt(string(name), p.Lon, diurnal)

		// Speed classification
		speed := "average"
		if p.Speed < 0 { speed = "retrograde" }
		absSpeed := math.Abs(p.Speed)
		if absSpeed < 0.01 { speed = "stationary" }

		entries = append(entries, DignityEntry{
			Planet:     string(name),
			Sign:       Signs[signIdx],
			Domicile:   isDomicile(string(name), signIdx),
			Exaltation: isExaltation(string(name), signIdx),
			Triplicity: TriplicityLord(signIdx, diurnal) == string(name),
			Term:       TermLord(p.Lon) == string(name),
			Face:       FaceLord(p.Lon) == string(name),
			Detriment:  DetrimentOf[signIdx] == string(name),
			Fall:       FallOf[signIdx] == string(name),
			Score:      score,
			Retrograde: p.Speed < 0,
			Speed:      speed,
		})
	}
	return entries
}

func isDomicile(planet string, signIdx int) bool {
	for _, s := range Domicile[planet] {
		if s == signIdx { return true }
	}
	return false
}

func isExaltation(planet string, signIdx int) bool {
	if ex, ok := Exaltation[planet]; ok {
		return ex.Sign == signIdx
	}
	return false
}

// FormatNatalAnalysis generates a comprehensive text for all natal sub-analyses.
func FormatNatalAnalysis(
	patterns []AspectPattern,
	shape *ChartShape,
	hemispheres *HemisphericDist,
	dignities []DignityEntry,
	age *PlanetaryAgePeriod,
) string {
	var b strings.Builder

	if len(patterns) > 0 {
		b.WriteString("## PATRONES DE ASPECTOS\n\n")
		for _, p := range patterns {
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", p.Type, strings.Join(p.Planets, ", ")))
			if p.Element != "" { b.WriteString(fmt.Sprintf("  Elemento: %s\n", p.Element)) }
		}
		b.WriteString("\n")
	}

	if shape != nil {
		b.WriteString(fmt.Sprintf("## FORMA DE CARTA: %s\n%s\n\n", shape.Pattern, shape.Description))
	}

	if hemispheres != nil {
		b.WriteString(fmt.Sprintf("## DISTRIBUCIÓN HEMISFÉRICA\nOriental=%d Occidental=%d Norte=%d Sur=%d\nDominante: %s\n\n",
			hemispheres.Eastern, hemispheres.Western, hemispheres.Northern, hemispheres.Southern, hemispheres.Dominant))
	}

	if len(dignities) > 0 {
		b.WriteString("## TABLA DE DIGNIDADES\n\n")
		for _, d := range dignities {
			markers := ""
			if d.Domicile { markers += " DOM" }
			if d.Exaltation { markers += " EXA" }
			if d.Detriment { markers += " DET" }
			if d.Fall { markers += " CAÍ" }
			if d.Retrograde { markers += " Rx" }
			b.WriteString(fmt.Sprintf("- %s en %s (score %d)%s\n", d.Planet, d.Sign, d.Score, markers))
		}
		b.WriteString("\n")
	}

	if age != nil {
		b.WriteString(fmt.Sprintf("## EDAD PLANETARIA: %s (%d-%d años)\n%s\n\n",
			age.Planet, age.AgeFrom, age.AgeTo, age.Theme))
	}

	return b.String()
}
