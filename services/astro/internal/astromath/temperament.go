package astromath

import "github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"

// TemperamentResult holds the classical four-humour analysis.
type TemperamentResult struct {
	Primary   string             `json:"primary"`   // dominant temperament
	Secondary string             `json:"secondary"`  // secondary temperament
	Scores    map[string]float64 `json:"scores"`    // choleric, sanguine, melancholic, phlegmatic
	Elements  map[string]int     `json:"elements"`  // fire, earth, air, water planet counts
	Qualities map[string]int     `json:"qualities"` // cardinal, fixed, mutable planet counts
}

// Element classification for each sign index.
var signToElement = [12]string{
	"fire", "earth", "air", "water",
	"fire", "earth", "air", "water",
	"fire", "earth", "air", "water",
}

// Quality classification for each sign index.
var signToQuality = [12]string{
	"cardinal", "fixed", "mutable",
	"cardinal", "fixed", "mutable",
	"cardinal", "fixed", "mutable",
	"cardinal", "fixed", "mutable",
}

// Temperament mapping: element → classical humor.
var elementTemperament = map[string]string{
	"fire":  "colérico",
	"earth": "melancólico",
	"air":   "sanguíneo",
	"water": "flemático",
}

// CalcTemperament computes the classical four-humour temperament.
// Uses the 7 classical planets + ASC + MC as weighted points.
func CalcTemperament(planets map[string]*ephemeris.PlanetPos, asc, mc float64) *TemperamentResult {
	elements := map[string]int{"fire": 0, "earth": 0, "air": 0, "water": 0}
	qualities := map[string]int{"cardinal": 0, "fixed": 0, "mutable": 0}

	// Weight: luminaries (Sol, Luna) = 2, others = 1
	weights := map[string]int{
		"Sol": 2, "Luna": 2, "Mercurio": 1, "Venus": 1,
		"Marte": 1, "Júpiter": 1, "Saturno": 1,
	}

	for name, w := range weights {
		p, ok := planets[name]
		if !ok {
			continue
		}
		signIdx := SignIndex(p.Lon)
		elem := signToElement[signIdx]
		qual := signToQuality[signIdx]
		elements[elem] += w
		qualities[qual] += w
	}

	// Add ASC sign (weight 2 — very important for temperament)
	ascSign := SignIndex(asc)
	elements[signToElement[ascSign]] += 2
	qualities[signToQuality[ascSign]] += 2

	// Add MC sign (weight 1)
	mcSign := SignIndex(mc)
	elements[signToElement[mcSign]]++
	qualities[signToQuality[mcSign]]++

	// Convert element counts to temperament scores (normalize to 0-100)
	total := 0
	for _, v := range elements {
		total += v
	}
	scores := make(map[string]float64)
	for elem, count := range elements {
		temp := elementTemperament[elem]
		if total > 0 {
			scores[temp] = float64(count) / float64(total) * 100
		}
	}

	// Find primary and secondary
	type kv struct {
		k string
		v float64
	}
	var sorted []kv
	for k, v := range scores {
		sorted = append(sorted, kv{k, v})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	primary := ""
	secondary := ""
	if len(sorted) > 0 {
		primary = sorted[0].k
	}
	if len(sorted) > 1 {
		secondary = sorted[1].k
	}

	return &TemperamentResult{
		Primary:   primary,
		Secondary: secondary,
		Scores:    scores,
		Elements:  elements,
		Qualities: qualities,
	}
}
