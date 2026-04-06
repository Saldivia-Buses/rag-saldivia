package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ZRPeriod represents a Zodiacal Releasing period at any level.
type ZRPeriod struct {
	Level     int     `json:"level"`      // 1=major, 2=sub, 3=bound
	Sign      string  `json:"sign"`
	SignIndex int     `json:"sign_index"`
	Lord      string  `json:"lord"`
	StartAge  float64 `json:"start_age"`  // years from birth
	EndAge    float64 `json:"end_age"`
	Duration  float64 `json:"duration"`   // years
	Loosing   bool    `json:"loosing"`    // opposite to starting sign = peak/crisis
}

// ZRResult holds Zodiacal Releasing analysis for a target age.
type ZRResult struct {
	Lot       string     `json:"lot"`       // "Fortune" or "Spirit"
	LotLon    float64    `json:"lot_lon"`
	LotSign   string     `json:"lot_sign"`
	Level1    *ZRPeriod  `json:"level_1"`
	Level2    *ZRPeriod  `json:"level_2"`
	Level3    *ZRPeriod  `json:"level_3,omitempty"` // bound sub-period
}

// CalcZodiacalReleasing calculates ZR from a given Lot for a target age.
// lotName: "Fortune" uses Fortuna, "Spirit" uses Espíritu.
func CalcZodiacalReleasing(chart *natal.Chart, lotName string, targetAge float64) *ZRResult {
	var lotLon float64
	switch lotName {
	case "Fortune":
		if pos, ok := chart.Planets["Fortuna"]; ok {
			lotLon = pos.Lon
		}
	case "Spirit":
		if pos, ok := chart.Planets["Espíritu"]; ok {
			lotLon = pos.Lon
		}
	default:
		return nil
	}

	lotSignIdx := astromath.SignIndex(lotLon)

	// Level 1: Major periods — cycle through signs starting from Lot's sign
	level1 := findActivePeriod(lotSignIdx, targetAge, 1)

	// Level 2: Sub-periods within the major period
	level2 := findSubPeriod(level1, targetAge)

	// Level 3: Bound sub-periods within the sub-period
	var level3 *ZRPeriod
	if level2 != nil {
		level3 = findBoundPeriod(level2, targetAge)
	}

	return &ZRResult{
		Lot:     lotName,
		LotLon:  lotLon,
		LotSign: astromath.Signs[lotSignIdx],
		Level1:  level1,
		Level2:  level2,
		Level3:  level3,
	}
}

// findActivePeriod finds which major period (Level 1) is active at targetAge.
func findActivePeriod(startSignIdx int, targetAge float64, level int) *ZRPeriod {
	var cumAge float64
	// ZR cycles through all 12 signs repeatedly
	for cycle := 0; cycle < 10; cycle++ { // 10 cycles covers ~200+ years
		for i := 0; i < 12; i++ {
			signIdx := (startSignIdx + i) % 12
			duration := astromath.ZRSignYears[signIdx]
			endAge := cumAge + duration

			if targetAge < endAge {
				// Check loosing of the bond: opposite sign (6 signs away)
				loosing := i == 6

				return &ZRPeriod{
					Level:     level,
					Sign:      astromath.Signs[signIdx],
					SignIndex: signIdx,
					Lord:      astromath.SignLord[signIdx],
					StartAge:  cumAge,
					EndAge:    endAge,
					Duration:  duration,
					Loosing:   loosing,
				}
			}
			cumAge = endAge
		}
	}
	return nil
}

// findSubPeriod divides a major period proportionally among 12 signs.
func findSubPeriod(major *ZRPeriod, targetAge float64) *ZRPeriod {
	if major == nil {
		return nil
	}

	totalYears := astromath.ZRSignYears[0] + astromath.ZRSignYears[1] + astromath.ZRSignYears[2] +
		astromath.ZRSignYears[3] + astromath.ZRSignYears[4] + astromath.ZRSignYears[5] +
		astromath.ZRSignYears[6] + astromath.ZRSignYears[7] + astromath.ZRSignYears[8] +
		astromath.ZRSignYears[9] + astromath.ZRSignYears[10] + astromath.ZRSignYears[11]

	ageInMajor := targetAge - major.StartAge
	cumAge := major.StartAge

	for i := 0; i < 12; i++ {
		signIdx := (major.SignIndex + i) % 12
		subDuration := (astromath.ZRSignYears[signIdx] / totalYears) * major.Duration
		endAge := cumAge + subDuration

		if targetAge < endAge {
			return &ZRPeriod{
				Level:     2,
				Sign:      astromath.Signs[signIdx],
				SignIndex: signIdx,
				Lord:      astromath.SignLord[signIdx],
				StartAge:  cumAge,
				EndAge:    endAge,
				Duration:  subDuration,
				Loosing:   i == 6,
			}
		}
		cumAge = endAge
	}

	// Fallback: rounding — use last sub-period
	_ = ageInMajor
	return nil
}

// findBoundPeriod divides a sub-period by Egyptian bounds within its sign.
func findBoundPeriod(sub *ZRPeriod, targetAge float64) *ZRPeriod {
	if sub == nil {
		return nil
	}

	bounds := astromath.EgyptianBounds[sub.SignIndex]
	cumAge := sub.StartAge

	for _, b := range bounds {
		boundDuration := ((b.ToDeg - b.FromDeg) / 30.0) * sub.Duration
		endAge := cumAge + boundDuration

		if targetAge < endAge {
			return &ZRPeriod{
				Level:    3,
				Sign:     sub.Sign,
				Lord:     b.Lord,
				StartAge: cumAge,
				EndAge:   endAge,
				Duration: boundDuration,
			}
		}
		cumAge = endAge
	}
	return nil
}
