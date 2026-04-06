package technique

import (
	"time"
)

// Firdaria holds the active firdaria period for a target date.
type Firdaria struct {
	Diurnal       bool    `json:"diurnal"`
	Age           float64 `json:"age"`
	MajorLord     string  `json:"major_lord"`
	MajorYears    int     `json:"major_years"`
	MajorStartAge float64 `json:"major_start_age"`
	MajorEndAge   float64 `json:"major_end_age"`
	SubLord       string  `json:"sub_lord"`       // "" for nodes
	SubStartAge   float64 `json:"sub_start_age"`
	SubEndAge     float64 `json:"sub_end_age"`
	NextMajorLord string  `json:"next_major_lord"`
	Cycle         int     `json:"cycle"`          // which 75-year cycle (1-based)
}

// firdariaEntry represents one major period in the sequence.
type firdariaEntry struct {
	Lord  string
	Years int
}

// Diurnal sequence (75-year cycle).
var diurnalSequence = []firdariaEntry{
	{"Sol", 10}, {"Venus", 8}, {"Mercurio", 13}, {"Luna", 9},
	{"Saturno", 11}, {"Júpiter", 12}, {"Marte", 7},
	{"Nodo Norte", 3}, {"Nodo Sur", 2},
}

// Nocturnal sequence (starts from Luna).
var nocturnalSequence = []firdariaEntry{
	{"Luna", 9}, {"Saturno", 11}, {"Júpiter", 12}, {"Marte", 7},
	{"Sol", 10}, {"Venus", 8}, {"Mercurio", 13},
	{"Nodo Norte", 3}, {"Nodo Sur", 2},
}

// Sub-period planets (7 planets cycle, no nodes).
var subPlanets = []string{"Sol", "Venus", "Mercurio", "Luna", "Saturno", "Júpiter", "Marte"}

// CalcFirdaria calculates the active firdaria period for a target date.
func CalcFirdaria(birthDate time.Time, diurnal bool, targetYear int) *Firdaria {
	targetDate := time.Date(targetYear, 7, 1, 0, 0, 0, 0, time.UTC)
	age := targetDate.Sub(birthDate).Hours() / (24 * 365.25)

	seq := diurnalSequence
	if !diurnal {
		seq = nocturnalSequence
	}

	// Total cycle = 75 years
	const cycleYears = 75.0
	cycle := int(age/cycleYears) + 1
	ageInCycle := age - float64(cycle-1)*cycleYears

	// Find the active major period
	var cumAge float64
	var majorIdx int
	for i, entry := range seq {
		end := cumAge + float64(entry.Years)
		if ageInCycle < end {
			majorIdx = i
			break
		}
		cumAge = end
	}

	major := seq[majorIdx]
	majorStartAge := float64(cycle-1)*cycleYears + cumAge
	majorEndAge := majorStartAge + float64(major.Years)

	// Sub-period: divide major period among 7 sub-planets
	var subLord string
	var subStartAge, subEndAge float64

	isNode := major.Lord == "Nodo Norte" || major.Lord == "Nodo Sur"
	if !isNode && len(subPlanets) > 0 {
		subDuration := float64(major.Years) / float64(len(subPlanets))
		ageInMajor := age - majorStartAge

		subIdx := int(ageInMajor / subDuration)
		if subIdx >= len(subPlanets) {
			subIdx = len(subPlanets) - 1
		}

		// Sub-period starts from the major lord's position in the sub sequence
		majorSubIdx := 0
		for i, p := range subPlanets {
			if p == major.Lord {
				majorSubIdx = i
				break
			}
		}
		actualSubIdx := (majorSubIdx + subIdx) % len(subPlanets)
		subLord = subPlanets[actualSubIdx]
		subStartAge = majorStartAge + float64(subIdx)*subDuration
		subEndAge = subStartAge + subDuration
	}

	// Next major lord
	nextIdx := (majorIdx + 1) % len(seq)
	nextMajorLord := seq[nextIdx].Lord

	return &Firdaria{
		Diurnal:       diurnal,
		Age:           age,
		MajorLord:     major.Lord,
		MajorYears:    major.Years,
		MajorStartAge: majorStartAge,
		MajorEndAge:   majorEndAge,
		SubLord:       subLord,
		SubStartAge:   subStartAge,
		SubEndAge:     subEndAge,
		NextMajorLord: nextMajorLord,
		Cycle:         cycle,
	}
}
