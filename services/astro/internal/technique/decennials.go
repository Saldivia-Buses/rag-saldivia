package technique

import (
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// DecennialResult holds the active decennial period for a target year.
type DecennialResult struct {
	Diurnal       bool              `json:"diurnal"`
	SectLight     string            `json:"sect_light"`
	Age           float64           `json:"age"`
	MajorPlanet   string            `json:"major_planet"`
	MajorYears    int               `json:"major_years"`
	MajorStartAge float64           `json:"major_start_age"`
	MajorEndAge   float64           `json:"major_end_age"`
	SubPlanet     string            `json:"sub_planet"`
	SubStartAge   float64           `json:"sub_start_age"`
	SubEndAge     float64           `json:"sub_end_age"`
	NextMajor     string            `json:"next_major_planet"`
	Cycle         int               `json:"cycle"`
	Sequence      []DecennialEntry  `json:"sequence"`
}

// DecennialEntry is one period in the decennial sequence.
type DecennialEntry struct {
	Planet    string  `json:"planet"`
	Years     int     `json:"years"`
	StartAge  float64 `json:"start_age"`
	EndAge    float64 `json:"end_age"`
	IsCurrent bool    `json:"is_current"`
}

// Decennial period durations (Vettius Valens, Anthologies IV).
var decennialYears = map[string]int{
	"Sol": 19, "Luna": 25, "Mercurio": 20, "Venus": 8,
	"Marte": 15, "Júpiter": 12, "Saturno": 27,
}

const decennialCycleTotal = 126 // sum of all periods

// Aspect priorities for whole-sign aspects from sect light.
// Conjunction(0) > Trine(4,8) > Square(3,9) > Opposition(6) > Sextile(2,10)
// Distances 1,5,7,11 = no major aspect.
var aspectPriority = map[int]int{
	0: 0, // conjunction
	4: 1, 8: 1, // trine
	3: 2, 9: 2, // square
	6: 3, // opposition
	2: 4, 10: 4, // sextile
}

// buildDecennialOrder determines the planet order based on whole-sign aspects
// to the sect light. Valens method.
func buildDecennialOrder(chart *natal.Chart) []string {
	sectLight := "Sol"
	if !chart.Diurnal {
		sectLight = "Luna"
	}

	sectPos, ok := chart.Planets[sectLight]
	if !ok {
		// Fallback: standard order starting from sect light
		return fallbackOrder(sectLight)
	}
	sectSign := astromath.SignIndex(sectPos.Lon)

	type scored struct {
		name     string
		priority int
		signIdx  int
	}

	var planets []scored
	for _, name := range astromath.ClassicalPlanets {
		p, ok := chart.Planets[string(name)]
		if !ok {
			continue
		}
		pSign := astromath.SignIndex(p.Lon)
		distance := (pSign - sectSign + 12) % 12

		prio, hasAspect := aspectPriority[distance]
		if !hasAspect {
			prio = 5 // no major aspect
		}
		// Sect light itself gets conjunction priority
		if string(name) == sectLight {
			prio = 0
		}

		planets = append(planets, scored{string(name), prio, pSign})
	}

	// Sort: priority ascending, then sign index ascending (zodiacal order)
	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			if planets[j].priority < planets[i].priority ||
				(planets[j].priority == planets[i].priority && planets[j].signIdx < planets[i].signIdx) {
				planets[i], planets[j] = planets[j], planets[i]
			}
		}
	}

	order := make([]string, len(planets))
	for i, p := range planets {
		order[i] = p.name
	}
	return order
}

func fallbackOrder(sectLight string) []string {
	order := []string{sectLight}
	for _, p := range astromath.ClassicalPlanets {
		if string(p) != sectLight {
			order = append(order, string(p))
		}
	}
	return order
}

// CalcDecennials calculates the active decennial period for a target year.
func CalcDecennials(chart *natal.Chart, birthDate time.Time, targetYear int) *DecennialResult {
	targetDate := time.Date(targetYear, 7, 1, 0, 0, 0, 0, time.UTC)
	age := targetDate.Sub(birthDate).Hours() / (24 * 365.25)

	order := buildDecennialOrder(chart)

	// Build sequence with years
	type seqEntry struct {
		planet string
		years  int
	}
	var sequence []seqEntry
	for _, p := range order {
		y, ok := decennialYears[p]
		if !ok {
			continue
		}
		sequence = append(sequence, seqEntry{p, y})
	}

	if len(sequence) == 0 {
		return nil
	}

	// Cycle handling (126-year cycle)
	cycle := int(age/float64(decennialCycleTotal)) + 1
	ageInCycle := age - float64(cycle-1)*float64(decennialCycleTotal)

	// Find active major period
	var cumAge float64
	majorIdx := 0
	for i, entry := range sequence {
		end := cumAge + float64(entry.years)
		if ageInCycle < end {
			majorIdx = i
			break
		}
		cumAge = end
		if i == len(sequence)-1 {
			majorIdx = i
		}
	}

	major := sequence[majorIdx]
	majorStartAge := float64(cycle-1)*float64(decennialCycleTotal) + cumAge
	majorEndAge := majorStartAge + float64(major.years)

	// Sub-periods: proportional subdivision among all 7 planets in same order
	ageInMajor := age - majorStartAge
	subCumAge := 0.0
	subIdx := 0
	for i, entry := range sequence {
		subDuration := float64(major.years) * float64(entry.years) / float64(decennialCycleTotal)
		if ageInMajor < subCumAge+subDuration {
			subIdx = i
			break
		}
		subCumAge += subDuration
		if i == len(sequence)-1 {
			subIdx = i
		}
	}

	sub := sequence[subIdx]
	subDuration := float64(major.years) * float64(sub.years) / float64(decennialCycleTotal)
	subStartAge := majorStartAge + subCumAge
	subEndAge := subStartAge + subDuration

	// Next major
	nextIdx := (majorIdx + 1) % len(sequence)

	// Build full sequence for display
	var entries []DecennialEntry
	cAge := float64(cycle-1) * float64(decennialCycleTotal)
	for _, entry := range sequence {
		e := DecennialEntry{
			Planet:   entry.planet,
			Years:    entry.years,
			StartAge: cAge,
			EndAge:   cAge + float64(entry.years),
		}
		if entry.planet == major.planet && abs64(cAge-majorStartAge) < 0.01 {
			e.IsCurrent = true
		}
		entries = append(entries, e)
		cAge += float64(entry.years)
	}

	sectLight := "Sol"
	if !chart.Diurnal {
		sectLight = "Luna"
	}

	return &DecennialResult{
		Diurnal:       chart.Diurnal,
		SectLight:     sectLight,
		Age:           age,
		MajorPlanet:   major.planet,
		MajorYears:    major.years,
		MajorStartAge: majorStartAge,
		MajorEndAge:   majorEndAge,
		SubPlanet:     sub.planet,
		SubStartAge:   subStartAge,
		SubEndAge:     subEndAge,
		NextMajor:     sequence[nextIdx].planet,
		Cycle:         cycle,
		Sequence:      entries,
	}
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
