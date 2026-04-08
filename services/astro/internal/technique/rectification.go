package technique

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// RectificationEvent is a known life event used to test birth time candidates.
type RectificationEvent struct {
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"` // "marriage", "career", "health", "move", "child", "loss"
}

// RectificationCandidate is a scored birth time hypothesis.
type RectificationCandidate struct {
	BirthHour float64  `json:"birth_hour"`  // local hour
	Score     float64  `json:"score"`
	Factors   []string `json:"factors"`
	ASCSign   string   `json:"asc_sign"`
	ASCDeg    float64  `json:"asc_deg"`
	MCSign    string   `json:"mc_sign"`
}

// RectificationResult holds the output.
type RectificationResult struct {
	Candidates []RectificationCandidate `json:"candidates"`
	BestHour   float64                  `json:"best_hour"`
	BestScore  float64                  `json:"best_score"`
	SearchMin  float64                  `json:"search_min"` // hours searched
	SearchMax  float64                  `json:"search_max"`
	StepMin    float64                  `json:"step_minutes"`
}

// categoryHouses maps event categories to relevant natal houses.
var categoryHouses = map[string][]int{
	"marriage": {7},
	"career":   {10},
	"health":   {6, 8},
	"move":     {4, 9},
	"child":    {5},
	"loss":     {8, 12},
}

// Rectify searches for the most likely birth time by testing candidates
// against known life events using primary directions and solar arcs.
//
// IMPORTANT: This is the most expensive technique. Each candidate builds
// a full natal chart via BuildNatal (which acquires CalcMu).
// The handler MUST limit concurrency via semaphore (max 1 concurrent).
// Lock is acquired per-chart, NOT held for the entire scan.
func Rectify(
	year, month, day int,
	lat, lon, alt float64,
	utcOffset int,
	searchRange [2]float64, // [minHour, maxHour] in local time
	stepMinutes float64, // e.g., 4.0 for 4-minute steps
	events []RectificationEvent,
) (*RectificationResult, error) {
	if stepMinutes <= 0 {
		stepMinutes = 4.0
	}
	step := stepMinutes / 60.0 // convert to hours

	if searchRange[0] >= searchRange[1] {
		return nil, fmt.Errorf("invalid search range: [%.2f, %.2f]", searchRange[0], searchRange[1])
	}

	var candidates []RectificationCandidate

	for hour := searchRange[0]; hour < searchRange[1]; hour += step {
		// Build chart for this candidate time
		// CalcMu is acquired and released within BuildNatal — NOT held for the loop
		chart, err := natal.BuildNatal(year, month, day, hour, lat, lon, alt, utcOffset)
		if err != nil {
			continue
		}

		score, factors := scoreCandidate(chart, events, time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC))

		candidates = append(candidates, RectificationCandidate{
			BirthHour: math.Round(hour*100) / 100,
			Score:     score,
			Factors:   factors,
			ASCSign:   astromath.SignName(chart.ASC),
			ASCDeg:    chart.ASC,
			MCSign:    astromath.SignName(chart.MC),
		})
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Keep top 10
	if len(candidates) > 10 {
		candidates = candidates[:10]
	}

	bestHour := 0.0
	bestScore := 0.0
	if len(candidates) > 0 {
		bestHour = candidates[0].BirthHour
		bestScore = candidates[0].Score
	}

	return &RectificationResult{
		Candidates: candidates,
		BestHour:   bestHour,
		BestScore:  bestScore,
		SearchMin:  searchRange[0],
		SearchMax:  searchRange[1],
		StepMin:    stepMinutes,
	}, nil
}

// scoreCandidate evaluates a candidate birth time against life events.
// Uses Solar Arc directions to relevant houses.
func scoreCandidate(chart *natal.Chart, events []RectificationEvent, birthDate time.Time) (float64, []string) {
	score := 0.0
	var factors []string

	for _, event := range events {
		ageDays := event.Date.Sub(birthDate).Hours() / 24
		ageYears := ageDays / 365.25

		// Solar Arc: arc = age * Naibod rate
		arc := ageYears * astromath.NaibodRate

		// Get the house cusps relevant to this event category
		houses := categoryHouses[event.Category]
		if len(houses) == 0 {
			houses = []int{1} // fallback
		}

		for _, houseNum := range houses {
			if houseNum < 1 || houseNum > 12 || len(chart.Cusps) < 13 {
				continue
			}
			cuspLon := chart.Cusps[houseNum]

			// Check if any SA planet hits this cusp
			for _, name := range []string{"Sol", "Luna", "Marte", "Saturno", "Júpiter"} {
				pos, ok := chart.Planets[name]
				if !ok {
					continue
				}
				saLon := astromath.Normalize360(pos.Lon + arc)
				orbVal := astromath.AngDiff(saLon, cuspLon)

				if orbVal < 2.0 { // within 2° orb
					eventScore := (2.0 - orbVal) * 10 // closer = higher score
					score += eventScore
					factors = append(factors, fmt.Sprintf("SA %s → cúspide %d (orbe %.1f°) [%s]",
						name, houseNum, orbVal, event.Description))
				}
			}

			// Also check if MC/ASC is SA'd by a relevant planet
			if houseNum == 10 {
				// Career events: check SA to MC
				for _, name := range []string{"Sol", "Júpiter", "Saturno"} {
					pos, ok := chart.Planets[name]
					if !ok {
						continue
					}
					saLon := astromath.Normalize360(pos.Lon + arc)
					orbVal := astromath.AngDiff(saLon, chart.MC)
					if orbVal < 2.0 {
						score += (2.0 - orbVal) * 15 // MC hits are very significant
						factors = append(factors, fmt.Sprintf("SA %s → MC (orbe %.1f°) [%s]",
							name, orbVal, event.Description))
					}
				}
			}
		}
	}

	return math.Round(score*100) / 100, factors
}
